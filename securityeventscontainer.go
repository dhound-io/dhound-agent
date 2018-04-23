package main

type SecurityEventsContainer struct {
	SourceId string `json:"sourceid,omitempty"`
	Offset   int64  `json:"offset,omitempty"`
	Source   string `json:"source,omitempty"`

	IpToServiceMap map[string][]string `json:"ipmap,omitempty"`
	SecurityEvents []*SecurityEvent    `json:"events,omitempty"`
}

func (container *SecurityEventsContainer) CleanSecurityEventsFromDublicates() {
	container.SecurityEvents = CleanSecurityEventsFromDublicates(container.SecurityEvents)
}
