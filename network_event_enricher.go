package main

type NetworkEventType int

const (
	TcpConnectionInitiatedByHost NetworkEventType = iota
	TcpConnectionSetUp
	UdpSendByHost
	DnsResponseReceived
)

type NetworkEvent struct {
	Type       NetworkEventType
	Connection *NetworkConnectionData
	Dns        *DnsAnswer
	_processId int32
}

type DnsAnswer struct {
	DomainName  string
	IpAddresses *[]string
}

type NetworkConnectionData struct {
	LocalIpAddress     string
	LocalPort          uint32
	RemoteIpAddress    string
	RemotePort         uint32
	Sequence           uint32
	EventTimeUtcNumber int64
}

type NetworkEventEnricher struct {
	Input      chan *NetworkEvent
	Output     chan *SecurityEventsContainer
	SysManager *SysProcessManager
	_cache     []*NetworkEvent
}

func (enricher *NetworkEventEnricher) Init() {

}

func (enricher *NetworkEventEnricher) Run() {
	for networkEvent := range enricher.Input {
		if networkEvent != nil {
			// if networkEvent.Type == TcpConnectionInitiatedByHost {
			debugJson(networkEvent)
			// }

			// enricher.SysManager.FindProcessInfoByLocalPort(networkEvent.)
		}
	}
}
