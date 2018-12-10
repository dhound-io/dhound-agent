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
	Input  chan *NetworkEvent
	_cache []*NetworkEvent
}

func (enricher *NetworkEventEnricher) Init() {

}

func (enricher *NetworkEventEnricher) Run() {
	for networkEvent := range enricher.Input {
		if networkEvent != nil {
			// if networkEvent.Type == TcpConnectionInitiatedByHost {
			debugJson(networkEvent)
			// }
		}
	}
}

/*processInfo := monitor.SysProcessManager.FindProcessInfoByLocalPort(uint32(tcp.DstPort))
pid := ""
process := ""
cmdLine := ""
if processInfo != nil {
	pid = fmt.Sprintf("%d", processInfo.Pid)
	process = processInfo.Name
	cmdLine = processInfo.CommandLine
}
*/
