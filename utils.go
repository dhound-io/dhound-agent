package main

import (
	"regexp"
	"strings"
)

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func ContainsUint(s []uint, e uint) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func RegexFindAllSubmatches(text *string, regex *regexp.Regexp, resultMap *map[string]string) {
	matches := regex.FindAllStringSubmatch(*text, -1)

	submatches := regex.SubexpNames()

	for _, match := range matches {
		// fill result map from results of regex match
		for i, name := range submatches {
			if len(name) > 0 {
				matchValue := match[i]
				if len(matchValue) > 0 {
					(*resultMap)[name] = matchValue
				}
			}
		}
	}
}

func RegexSplit(text string, delimeterRegex *regexp.Regexp) []string {
	indexes := delimeterRegex.FindAllStringIndex(text, -1)
	laststart := 0
	result := make([]string, len(indexes)+1)
	for i, element := range indexes {
		result[i] = text[laststart:element[0]]
		laststart = element[1]
	}
	result[len(indexes)] = text[laststart:len(text)]
	return result
}

func ApplyExcludeFilterToSecurityEvents(excludeRegexFilter map[string]*regexp.Regexp, resultMap *map[string]string) bool {
	skipEvent := false
	if len(excludeRegexFilter) > 0 {
		for name, excludeRegex := range excludeRegexFilter {
			if excludeRegex != nil {
				// check if field with this name presented in resultMap
				fieldValue := (*resultMap)[name]
				//emit("%s", fieldValue)
				if len(fieldValue) > 0 {
					// if regex pass, we should exlude this event
					excludeMatched := excludeRegex.Match([]byte(fieldValue))
					if excludeMatched == true {
						skipEvent = true
					}
				}
			}
		}
	}
	return skipEvent
}

func ApplyFilterToSecurityMessage(securityMessage string, resultMap *map[string]string) string {
	if len(securityMessage) > 0 && resultMap != nil {
		// it can be dynamic message
		for key, value := range *resultMap {
			securityMessage = strings.Replace(securityMessage, "#"+key, value, len(securityMessage))
			// emitJson(logLevel.important, securityMessage)
		}
	}
	return securityMessage
}

func CleanSecurityEventsFromDublicates(securityEvents []*SecurityEvent) []*SecurityEvent {
	if len(securityEvents) < 2 {
		return securityEvents
	}

	targetSecurityEvents := make([]*SecurityEvent, 0)
	for _, baseEvent := range securityEvents {
		foundDublicate := false
		for _, targetEvent := range targetSecurityEvents {
			if baseEvent.SecurityId == targetEvent.SecurityId && baseEvent.IpAddress == targetEvent.IpAddress && baseEvent.EventTimeUtcNumber == targetEvent.EventTimeUtcNumber {
				foundDublicate = true
				break
			}
		}

		if !foundDublicate {
			targetSecurityEvents = append(targetSecurityEvents, baseEvent)
		}
	}

	return targetSecurityEvents

}
