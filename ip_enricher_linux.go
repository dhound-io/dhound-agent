// +build !windows,cgo

package main

import (
	"fmt"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func (enricher *IpEnricher) ShowDevices() {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		emitLine(logLevel.important, "Failed find network devices", err.Error())
	}

	var deviceLines []string
	for _, device := range devices {
		var ips []string
		for _, address := range device.Addresses {
			ips = append(ips, address.IP.String())
		}

		deviceLines = append(deviceLines, fmt.Sprintf("'%s' (%s)", device.Name, strings.Join(ips, ",")))
	}
	emitLine(logLevel.important, "Found network interfaces: %s", strings.Join(deviceLines, "; "))
}

func (enricher *IpEnricher) _InternalSync(runOnce bool) {
	if runOnce {
		return
	}

	// for linux default interface
	deviceName := "eth0"

	if len(enricher.Config.Input.NetworkInterface) > 0 {
		deviceName = enricher.Config.Input.NetworkInterface
	}

	// Open device
	handle, err := pcap.OpenLive(deviceName, 1600, false, pcap.BlockForever)
	if err != nil {
		emitLine(logLevel.important, "Failed listening device %s. Error: %s.", deviceName, err.Error())
		enricher.ShowDevices()
		return
	}
	defer handle.Close()

	emitLine(logLevel.important, "ip encricher enabled. listening to device %s.", deviceName)

	// Set filter
	var filter = "udp and port 53"
	err = handle.SetBPFFilter(filter)
	if err != nil {
		emitLine(logLevel.important, "Failed set BPF Filter. Error: %s.", err.Error())
		return
	}

	var eth layers.Ethernet
	var ip4 layers.IPv4
	var ip6 layers.IPv6
	var tcp layers.TCP
	var udp layers.UDP
	var dns layers.DNS
	var payload gopacket.Payload

	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ip4, &ip6, &tcp, &udp, &dns, &payload)

	decodedLayers := make([]gopacket.LayerType, 0, 10)
	for {
		data, _, err := handle.ReadPacketData()
		if err != nil {
			emitLine(logLevel.verbose, "Error reading packet data: %s", err)
			continue
		}

		err = parser.DecodeLayers(data, &decodedLayers)
		if err != nil {
			emitLine(logLevel.verbose, "Error decoding packet data: %s", err)
		}

		//		srcIp := ""
		//		dstIp := ""

		for _, typ := range decodedLayers {
			switch typ {
			//			case layers.LayerTypeIPv4:
			//				srcIp = ip4.SrcIP.String()
			//				dstIp = ip4.DstIP.String()
			//			case layers.LayerTypeIPv6:
			//				srcIp = ip6.SrcIP.String()
			//				dstIp = ip6.DstIP.String()
			case layers.LayerTypeDNS:
				dnsANCount := int(dns.ANCount)

				if dnsANCount > 0 {

					for _, dnsQuestion := range dns.Questions {

						dnsName := string(dnsQuestion.Name)
						var ips []string

						for _, dnsAnswer := range dns.Answers {
							ip := dnsAnswer.IP.String()
							if ip != "<nil>" {
								ips = append(ips, ip)
							}
						}

						if len(ips) > 0 {
							ipMap := enricher._ipToServicesMap

							if ipMap == nil {
								ipMap = make(map[string][]string)
							}

							unique := false
							for _, ip := range ips {
								if !Contains(ipMap[ip], dnsName) {
									ipMap[ip] = append(ipMap[ip], dnsName)
									unique = true
								}
							}

							if unique {
								debug("DNS: %s: %s", dnsName, strings.Join(ips, ","))
							}

							enricher._ipToServicesMap = ipMap

							// debugJson(enricher._ipToServicesMap)
						}
					}
				}

			}
		}
	}
}

func (enricher *IpEnricher) _ProcessEventsContainer(
	eventsContainer *SecurityEventsContainer) {
	eventsContainer.IpToServiceMap = enricher._ipToServicesMap
	enricher._ipToServicesMap = nil
}
