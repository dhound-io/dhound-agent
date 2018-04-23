package main

type IpEnricher struct {
	Input            chan *SecurityEventsContainer
	NextChannel      chan *SecurityEventsContainer
	Options          *Options
	Config           *MainConfig
	_ipToServicesMap map[string][]string
	_firstRun        bool
}

func (enricher *IpEnricher) Init() {
	enricher._firstRun = true
	if enricher.Config.Input.TrackDnsTraffic {
		enricher._InternalSync(true)
	}
	enricher._firstRun = false
}

func (enricher *IpEnricher) Run() {

	go enricher.Sync()

	for eventsContainer := range enricher.Input {
		if enricher.Config.Input.TrackDnsTraffic {
			enricher._ProcessEventsContainer(eventsContainer)
		}
		// debugJson(eventsContainer)
		enricher.NextChannel <- eventsContainer
	}
}

func (enricher *IpEnricher) Sync() {
	if enricher.Config.Input.TrackDnsTraffic {
		enricher._InternalSync(false)
	}
}

func (enricher *IpEnricher) GetDnsByIp(ip string) []string {
	if enricher._ipToServicesMap != nil {
		services := enricher._ipToServicesMap[ip]

		return services
	}
	return nil
}
