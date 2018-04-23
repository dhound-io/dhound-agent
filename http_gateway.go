package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"

	"crypto/tls"
	"net/http"
	"net/url"
)

type ServerRequestMessage struct {
	AccessToken              string              `json:"token"`
	ServerKey                string              `json:"hd"`
	LocalTimeUtcNumber       int64               `json:"ult"`
	LocalTimeUtcOffsetNumber int                 `json:"ulto,omitempty"`
	Version                  string              `json:"v,omitempty"`
	Events                   []*SecurityEvent    `json:"events,omitempty"`
	IpServices               map[string][]string `json:"ipsrvs,omitempty"`
}

type ServerResponseMessage struct {
	Success      bool   `json:"success,omitempty"`
	ErrorMessage string `json:"error,omitempty"`
	ErrorCode    int    `json:"errorcode,omitempty"`
}

type HttpGateway struct {
	Input                chan []*SecurityEventsContainer `json:- yaml:-`
	SystemState          *SystemState
	Options              *Options
	MainConfig           *MainConfig
	_serverUrl           string
	_timeOffsetInSeconds int
	_client              *http.Client
	_firstMessageSent    bool
}

func (gate *HttpGateway) Init() {

	gate._firstMessageSent = false

	_, timeOffsetInSeconds := time.Now().In(time.Local).Zone()

	gate._timeOffsetInSeconds = timeOffsetInSeconds

	serverUrl := "https://gate.dhound.io/collect"

	config := gate.MainConfig.Output

	if config.Environment == "DEV" {
		serverUrl = "http://localhost:5000/collect"
	} else if config.Environment == "TEST" {
		serverUrl = "https://gate-test.dhound.io/collect"
	} else {
		serverUrl = "https://gate.dhound.io/collect"
		//exit(exitStat.faulted, "Environment %s not supported\n", config.Environment)
	}

	gate._serverUrl = serverUrl

	proxy := config.Proxy
	//emit(logLevel.verbose, proxy)

	if len(proxy) > 0 {
		emit(logLevel.verbose, "Server url: %s via proxy: %s\n", gate._serverUrl, proxy)
	} else {
		emit(logLevel.verbose, "Server url: %s\n", serverUrl)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{},
		//		DialContext: (&net.Dialer{
		//			Timeout:   30 * time.Second, // default: 30
		//			KeepAlive: 0,                // default: 30
		//		}).DialContext,
		DisableKeepAlives:  true,
		MaxIdleConns:       -1,               // default: 100
		IdleConnTimeout:    10 * time.Second, // default: 90
		DisableCompression: true,
	}

	if len(proxy) > 0 {

		proxyUrl, err := url.Parse(proxy)
		if err != nil {
			emit(logLevel.critical, "Error parsing proxy URL: %s, error: %s\n", proxy, err.Error())
			panic("Exit!")
		}

		transport.Proxy = http.ProxyURL(proxyUrl)
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(time.Millisecond * 15000),
	}

	gate._client = client
}

func (gate *HttpGateway) Run() {

	// wait events from channel input
	for eventsContainers := range gate.Input {
		// debug("GATE: try to send %d events.", len(eventsContainers))

		gate.SendToServer(eventsContainers)

		// sync source state
		gate.SystemState.Input <- eventsContainers
	}
}

func (gate HttpGateway) SendToServer(eventsContainers []*SecurityEventsContainer) {

	CreateDirIfNotExist(".state", 0765)

	config := gate.MainConfig.Output

	serverMessage := ServerRequestMessage{
		AccessToken:        config.AccessToken,
		ServerKey:          config.ServerKey,
		LocalTimeUtcNumber: DateToCustomLong(time.Now()),
	}

	if gate._firstMessageSent == false {
		serverMessage.Version = Version
		serverMessage.LocalTimeUtcOffsetNumber = gate._timeOffsetInSeconds
	}

	// extract all events
	secEvents := make([]*SecurityEvent, 0)

	for _, eventsContainer := range eventsContainers {
		if eventsContainer != nil && len(eventsContainer.SecurityEvents) > 0 {
			secEvents = append(secEvents, eventsContainer.SecurityEvents...)
		}

		for ip, services := range eventsContainer.IpToServiceMap {
			if serverMessage.IpServices == nil {
				serverMessage.IpServices = make(map[string][]string)
			}
			serverMessage.IpServices[ip] = services
		}
	}

	// check previous failed request and try to resend it
	files, _ := filepath.Glob(".state/.net_*")
	if len(files) > 0 {
		sort.Strings(files)
		for _, netfile := range files {
			netContent, err := ioutil.ReadFile(netfile)
			if err != nil {
				emitLine(logLevel.important, "failed read net file %s, error: %s\n", netfile, err.Error())
				continue
			}

			resp, errs := gate._client.Post(gate._serverUrl, "application/json", bytes.NewBuffer(netContent))
			if errs == nil {
				resp.Body.Close()
				err := os.Remove(netfile)
				if err != nil {
					emit(logLevel.important, "Failed removing net file %s, error: %s\n", netfile, err.Error())
				}
			}
		}
	}

	if len(secEvents) > 0 {
		serverMessage.Events = secEvents
	}

	messageJson, _ := json.Marshal(serverMessage)

	resp, errs := gate._client.Post(gate._serverUrl, "application/json", bytes.NewBuffer([]byte(messageJson)))

	//defer transport.CloseIdleConnections()

	failed := false

	if errs != nil {
		errsJson, _ := json.Marshal(errs)
		emit(logLevel.important, "Failed sending message to server. JsonSize: %d. Errors: %s.\n", len(messageJson), errsJson)
		failed = true
	} else {
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			emit(logLevel.important, "Failed sending message to server. JsonSize: %d. Status Code: %d.\n", len(messageJson), resp.StatusCode)
			failed = true
		} else { // status 200, but body can contain error
			response := ServerResponseMessage{}
			body, _ := ioutil.ReadAll(resp.Body)
			err := json.Unmarshal(body, &response)
			if err != nil {
				emit(logLevel.important, "Failed converting server response to json. Response: %s. \n", string(body))
				failed = true
			} else {
				// successfully received response, check that success is true
				if response.Success != true {
					// if server send error code 1  (wrong json format, let loose all current events to prevent it in future
					if response.ErrorCode == 1 {
						emit(logLevel.important, "Failed sending requests to server. Server error: %s (%d). %d events will be lost. \n", response.ErrorMessage, response.ErrorCode, len(serverMessage.Events))
						failed = false
					} else {
						emit(logLevel.important, "Failed sending requests to server. Server error: %s (%d). \n", response.ErrorMessage, response.ErrorCode)
						failed = true
					}
				}
			}
		}
	}

	if failed == true {
		// don't lose any security events, save it into temp file
		if len(serverMessage.Events) > 0 {

			publisherTmpFile := fmt.Sprintf(".state/.net_%d", time.Now().Unix())
			tmpFileContent, _ := json.Marshal(serverMessage)
			err := ioutil.WriteFile(publisherTmpFile, tmpFileContent, 0664)
			if err != nil {
				emit(logLevel.important, "Failed to create tempfile (%s) for writing: %s. Events will be lost.\n", publisherTmpFile, err.Error())
			}
		}
	} else {
		gate._firstMessageSent = true

		if len(serverMessage.Events) > 0 {
			emit(logLevel.verbose, "Sent request on server. Body size: %d. Body: %s.", len(messageJson), messageJson)
		}
	}
}
