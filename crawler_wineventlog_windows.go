// +build !windows
package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type WinEventLogCrawler struct {
	Rules                  []*RuleConfig
	SystemState            *SystemState
	NextChannel            chan *SecurityEventsContainer
	Options                *Options
	_winEventFieldsRegex   *regexp.Regexp
	_defaultPeriodToParse  time.Duration
	_maxNumberEventsToRead int
	_inited                bool
	_crawlPeriod           time.Duration
	_firstRun              bool
}

func (crawler *WinEventLogCrawler) Init() {
	if len(crawler.Rules) < 1 {
		return
	}

	winEventFieldsRegex, err := regexp.Compile(`<EventID(.*?)>(?P<eventid>.*?)<\/EventID>|<TimeCreated SystemTime='(?P<_eventTime>.*?)'|<EventRecordID(.*?)>(?P<recordid>.*?)<\/EventRecordID>`)
	if err != nil {
		emit(logLevel.critical, "Failed parse base winevent regex. %s\n", err)
		return
	}

	crawler._winEventFieldsRegex = winEventFieldsRegex

	crawler._maxNumberEventsToRead = 1000
	crawler._defaultPeriodToParse = time.Hour * 24 * 30
	crawler._crawlPeriod = time.Second * 60

	crawler._inited = true
}

func (crawler *WinEventLogCrawler) Run() {

	if crawler._inited {
		emit(logLevel.verbose, "windows eventlog crawler started\n")

		crawler._firstRun = true
		for {
			crawler._RunOnce()
			crawler._firstRun = false
			time.Sleep(crawler._crawlPeriod)
		}
	}
}

func (crawler *WinEventLogCrawler) _RunOnce() {
	// find rules with unique pathes

	for _, rule := range crawler.Rules {
		for _, path := range rule.Paths {

			var winEventIdSearchList []string
			for _, event := range rule.Events {
				for _, winEventId := range event.WinEventIds {
					winEventIdStr := fmt.Sprintf("EventID=%d", winEventId)
					if !Contains(winEventIdSearchList, winEventIdStr) {
						winEventIdSearchList = append(winEventIdSearchList, winEventIdStr)
					}
				}
			}

			if len(winEventIdSearchList) < 1 {
				continue
			}

			periodToParse := crawler._defaultPeriodToParse

			sourceId := rule.RuleFileName + `_` + path
			sourceState := crawler.SystemState.Find(sourceId)
			if sourceState.Offset > 0 {
				// add one millisecond not to catch previous collected events
				lastParseTime := CustomLongToTime(sourceState.Offset)
				// emitJson(logLevel.verbose, lastParseTime)
				periodToParse = time.Now().UTC().Sub(lastParseTime)
			} else if len(rule.DeadTime) > 0 {
				periodToParse = rule.deadtime
			}

			if crawler._firstRun {
				emitLine(logLevel.important, "resume observing event log '%s' for events '%s'.", path, strings.Join(winEventIdSearchList, ", "))
			}

			timeToParseInMilliSeconds := int64(periodToParse / time.Millisecond)

			queryTime := time.Now()

			var query = `<QueryList>
							<Query Id="0" Path="` + path + `">
							<Select Path="` + path + `">
								*[System[(` + strings.Join(winEventIdSearchList, " or ") + `) and TimeCreated[timediff(@SystemTime) &lt;= ` + fmt.Sprintf("%d", timeToParseInMilliSeconds) + `]]]
							</Select>
							</Query>
						</QueryList>`

			// make request to windows logs
			result, err := _ReadEventLogs(query, crawler._maxNumberEventsToRead)
			if err != nil {
				if crawler._firstRun {
					emitLine(logLevel.important, "winevent_crawler query failed. Error: %s Query: '%s'. ", err.Error(), query)
				}
				continue
			}

			// emit(logLevel.verbose, query)
			// emitJson(logLevel.verbose, len(result))

			eventsContainer := &SecurityEventsContainer{
				SourceId: sourceId,
				Offset:   DateToCustomLong(queryTime),
				Source:   path,
			}

			sourceState.Offset = DateToCustomLong(queryTime)

			// emitJson(logLevel.verbose, rule)

			src := path

			for _, event := range rule.Events {

				if len(event.WinEventIds) < 1 {
					continue
				}

				for _, item := range result {

					resultMap := make(map[string]string)

					RegexFindAllSubmatches(&item, crawler._winEventFieldsRegex, &resultMap)

					// check if found item corresponds to current event
					eventId := ForceAtoui(resultMap["eventid"])
					if !ContainsUint(event.WinEventIds, eventId) {
						continue
					}

					// fill result map with predefined fields
					if len(event.Fields) > 0 {
						for key, value := range event.Fields {
							resultMap[key] = value
						}
					}

					RegexFindAllSubmatches(&item, event.CompiledRegex, &resultMap)

					skipEvent := ApplyExcludeFilterToSecurityEvents(event.ExcludeCompiledRegex, &resultMap)
					if skipEvent {
						continue
					}

					securityMessage := resultMap["message"]

					if len(event.Message) > 0 {
						securityMessage = event.Message
					}

					securityMessage = ApplyFilterToSecurityMessage(securityMessage, &resultMap)

					AdditionalFieldsMap := make(map[string]string)
					for key, value := range resultMap {
						if key != "ip" && key != "eventTime" && key != "_eventTime" {
							AdditionalFieldsMap[key] = value
						}
					}

					// parse event time
					eventTime, err := time.Parse(time.RFC3339Nano, resultMap["_eventTime"])
					if err != nil {
						if crawler._firstRun {
							emit(logLevel.important, "Failed parsed win event time '%s'. EventId: %s, Path: %s\n", resultMap["_eventTime"], resultMap["_eventid"], path)
						}
						continue
					}

					if len(resultMap["ip"]) < 1 {
						if crawler._firstRun {
							emit(logLevel.important, "Ip is not defined for event EventId: %s, Path: %s, File: %s\n", resultMap["_eventid"], path, rule.RuleFileName)
						}
						continue
					}

					securityEvent := &SecurityEvent{
						SecurityId:         event.Sid,
						SecurityGroupId:    event.Gid,
						EventTimeUtcNumber: DateToCustomLong(eventTime),
						Message:            securityMessage,
						Critical:           event.Critical,
						IpAddress:          resultMap["ip"],
						AdditionalFields:   AdditionalFieldsMap,
						Source:             &src,
					}

					eventsContainer.SecurityEvents = append(eventsContainer.SecurityEvents, securityEvent)
				}
			}

			eventsContainer.CleanSecurityEventsFromDublicates()
			crawler.NextChannel <- eventsContainer // send events to next channel
		}
	}

}
