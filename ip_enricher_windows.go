// +build windows

package main

import (
	"os/exec"
	"regexp"
	"strings"
	"time"
)

func (enricher *IpEnricher) _InternalSync(runOnce bool) {
	// update regulary from ipconfig /displaydns output
	const duration = 50 * time.Second

	aRecordsRegex, err := regexp.Compile(": (?P<name>.*?)\r*\n(.*?:.*?\n){4}.*?:( |\t)*(?P<ip>((\\d+\\.){3}[0-9]{1,3})|(([A-f0-9:]+:+)+[A-f0-9]+))")
	if err != nil {
		emitLine(logLevel.important, "wrong regex. Error: %s", err)
		return
	}

	cnameRecordsRegex, err := regexp.Compile(": (?P<name>.*?)\r*\n(.*?:.*?\n){4}( |\t)+([CNAME]).*?:( |\t)*(?P<cname>\\S+)")
	if err != nil {
		emitLine(logLevel.important, "wrong regex. Error: %s", err)
		return
	}

	delimiterRegex, err := regexp.Compile("(\r*)\n(\r*)\n")
	if err != nil {
		emitLine(logLevel.important, "wrong regex. Error: %s", err)
		return
	}

	internalIpRegex, err := regexp.Compile("^((0\\.)|(127\\.0\\.0\\.1)|(192\\.168\\.)|(10\\.)|(172\\.(1[6-9]|2[0-9]|3[0-1])\\.)|(fc00:)|(fe80:))")
	if err != nil {
		emitLine(logLevel.important, "wrong regex. Error: %s", err)
		return
	}

	for {
		path, err := exec.LookPath("ipconfig")
		if err != nil {
			emitLine(logLevel.important, "ipconfig utility not found. Error: %s", err)
			return
		}

		out, err := exec.Command(path, "/displaydns").Output()
		if err != nil {
			emitLine(logLevel.important, "failed receiving output from ipconfig /displaydns command. Error: %s", err)
		} else {
			output := string(out)

			sections := RegexSplit(output, delimiterRegex)

			aRecords := make(map[string][]string)
			cnameRecords := make(map[string][]string)

			for _, section := range sections {
				parsedFields := make(map[string]string)
				line := section
				RegexFindAllSubmatches(&line, aRecordsRegex, &parsedFields)

				if len(parsedFields["ip"]) > 0 {
					// A(AAA) record found
					ip := parsedFields["ip"]

					// exclude internal network ips
					if internalIpRegex.MatchString(ip) {
						continue
					}

					name := parsedFields["name"]
					if len(name) > 0 {
						name := strings.ToLower(name)
						if !Contains(aRecords[ip], name) {
							aRecords[ip] = append(aRecords[ip], name)
						}
					}
				} else {
					RegexFindAllSubmatches(&line, cnameRecordsRegex, &parsedFields)
					if len(parsedFields["cname"]) > 0 {
						// CNAME record found
						cname := strings.ToLower(parsedFields["cname"])
						name := strings.ToLower(parsedFields["name"])

						if !Contains(cnameRecords[cname], name) {
							cnameRecords[cname] = append(aRecords[cname], name)
						}
					}
				}
			}

			// merge a records and cname records
			for ip, names := range aRecords {
				for _, name := range names {
					secondNames := cnameRecords[name]
					for _, secondName := range secondNames {
						if !Contains(names, secondName) {
							aRecords[ip] = append(aRecords[ip], secondName)
						}
					}
				}
			}

			enricher._ipToServicesMap = aRecords

			if enricher._firstRun {
				emitLine(logLevel.verbose, "ip encricher enabled. found %d records in dns cache.", len(enricher._ipToServicesMap))
			}

			if runOnce {
				return
			}
		}

		time.Sleep(duration)
	}
}

func (enricher *IpEnricher) _ProcessEventsContainer(eventsContainer *SecurityEventsContainer) {
	for _, event := range eventsContainer.SecurityEvents {
		services := enricher.GetDnsByIp(event.IpAddress)
		if len(services) > 0 {
			if eventsContainer.IpToServiceMap == nil {
				eventsContainer.IpToServiceMap = make(map[string][]string)
			}
			eventsContainer.IpToServiceMap[event.IpAddress] = services
		}
	}
}
