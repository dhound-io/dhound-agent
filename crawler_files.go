package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/text/encoding/htmlindex"
)

type FilesCrawler struct {
	Rules                 []*RuleConfig
	SystemState           *SystemState
	NextChannel           chan *SecurityEventsContainer
	_firstRun             bool
	_inited               bool
	_crawlPeriod          time.Duration
	_defaultPeriodToParse time.Duration
}

func (crawler *FilesCrawler) Init() {

	if len(crawler.Rules) < 1 {
		return
	}

	crawler._crawlPeriod = 60 * time.Second
	crawler._defaultPeriodToParse = time.Hour * 24 * 30
	crawler._inited = true
}

func (crawler *FilesCrawler) Run() {

	if crawler._inited {
		emit(logLevel.verbose, "files crawler started")

		crawler._firstRun = true
		for {
			crawler._RunOnce()

			crawler._firstRun = false
			time.Sleep(crawler._crawlPeriod)
		}

	}
}

func (crawler *FilesCrawler) _RunOnce() {
	pathOnRulesMap := crawler._GetFilesListMap()
	// debugJson(pathOnRulesMap)

	for path, rules := range pathOnRulesMap {

		fileId := GetFileOsUniqueKey(path)
		if len(fileId) < 1 {
			continue
		}

		fileInfo, err := os.Stat(path)
		if err != nil {
			emit(logLevel.important, "failed reading file stat. file: %s, error: %s\n", path, err)
			continue
		}

		if fileInfo.IsDir() {
			continue
		}

		maxFileDeadTime := time.Second * 1
		// find max dead time among rules
		for _, rule := range rules {
			if rule.deadtime > maxFileDeadTime {
				maxFileDeadTime = rule.deadtime
			}
		}

		fileModified := fileInfo.ModTime()

		if time.Now().Sub(fileModified) > maxFileDeadTime {
			// the file is so old for parsing
			// emitLine(logLevel.verbose, "file '%s' so old. Last modified time: %s, max file deadtime: %s", path, fileModified.String(), maxFileDeadTime.String())
			continue
		}

		sourceState := crawler.SystemState.Find(fileId)
		// debugJson(sourceState)
		var position int64 = 0
		var linePosition int64 = 1
		if sourceState.Offset > 0 {
			position = sourceState.Offset
			linePosition = sourceState.Line
		}

		if crawler._firstRun {
			ruleNames := make([]string, 0)
			for _, rule := range rules {
				ruleNames = append(ruleNames, rule.RuleFileName)
			}

			if linePosition > 0 {
				emitLine(logLevel.important, "resume observing file '%s' from position %d (line:%d); rules: '%s'.", path, position, linePosition, strings.Join(ruleNames, ", "))
			} else {
				emitLine(logLevel.important, "resume observing file '%s' from position %d; rules: '%s'.", path, position, strings.Join(ruleNames, ", "))
			}

		}

		fileSize := fileInfo.Size()

		if position > fileSize {
			position = 0
			linePosition = 1
		} else if position == fileSize {
			continue
		}

		// read file from position to end
		file, err := ReadOpen(path)
		if err != nil {
			emitLine(logLevel.important, "failed reading file '%s'. Error: %s", path, err)
			continue
		}

		defer file.Close()

		// get encoding from the first rule
		encodingName := rules[0].Encoding
		if len(encodingName) > 0 {
			_, err = htmlindex.Get(encodingName)
			if err != nil {
				emitLine(logLevel.important, "incorrect encoding '%s' specified for file '%s'. default encoding will be used.", encodingName, path)
				encodingName = ""
			}
		}

		decodingReader, err := Utf8Reader(file, encodingName)
		if err != nil {
			emitLine(logLevel.important, "failed reading file '%s'. encoding: '%s'", path, encodingName)
			continue
		}

		if position > 0 {
			file.Seek(position, io.SeekStart)
		}

		reader := bufio.NewReaderSize(decodingReader, 10*1024)
		eventsContainer := &SecurityEventsContainer{}
		eventsContainer.SourceId = sourceState.SourceId
		eventsContainer.Source = path

		src := path

		for {
			segment, err := reader.ReadBytes('\n')

			line := string(segment)
			if len(line) > 0 {
				// remove \r\n from the end of the string
				if line[len(line)-1] == '\n' {
					drop := 1
					if len(line) > 1 && line[len(line)-2] == '\r' {
						drop = 2
					}
					line = line[:len(line)-drop]
				} else if err == io.EOF {
					// this is the last line before file END, need to decide - process or not to process
					// if file is still modifiying, break the last line
					if time.Now().Sub(fileModified) < time.Second*60 {
						// debug(time.Now().Sub(fileModified).String())
						break
					}
				}
			}
			if linePosition > 0 {
				linePosition++
			}

			position, _ := file.Seek(0, io.SeekCurrent)

			sourceState.Line = linePosition
			sourceState.Offset = position
			eventsContainer.Offset = sourceState.Offset
			eventsContainer.Line = sourceState.Line

			if len(line) > 0 {
				// process line by correspondent rules
				for _, rule := range rules {
					crawler.ParseLine(&src, linePosition, &line, rule, eventsContainer)
				}
			}

			if err != nil {
				if err == io.EOF {
					break
				} else if err == bufio.ErrBufferFull {
					continue
				} else {
					emitLine(logLevel.important, "unexpected error during reading the file '%s'. error: %s", path, err)
					break
				}
			}
		}

		eventsContainer.CleanSecurityEventsFromDublicates()
		crawler.NextChannel <- eventsContainer

		// debug("finished processing file %s. security events: %d. rules: %d", path, len(eventsContainer.SecurityEvents), len(rules))

	} // # end for path, rules := range pathOnRulesMap
}

func (crawler *FilesCrawler) ParseLine(source *string, linePosition int64, text *string, rule *RuleConfig, eventsContainer *SecurityEventsContainer) {

	for _, eventFilter := range rule.Events {

		regex := eventFilter.CompiledRegex
		securityId := eventFilter.Sid

		matches := regex.FindStringSubmatch(*text)

		if matches != nil && len(matches) > 0 {

			resultMap := make(map[string]string)

			// fill result map with predefined fields
			if len(eventFilter.Fields) > 0 {
				for key, value := range eventFilter.Fields {
					resultMap[key] = value
				}
			}

			RegexFindAllSubmatches(text, regex, &resultMap)

			// emitJson(logLevel.verbose, resultMap)

			// parse datetime
			eventTimeStr := resultMap["eventTime"]

			var eventTime time.Time
			var err error
			if len(rule.EventTimeFormat) < 1 {
				eventTime, err = time.ParseInLocation(eventTimeStr, eventTimeStr, time.Local)

				if err != nil {
					emit(logLevel.important, "FileReader: Failed parsing '%s'. Error: %s\n", eventTimeStr, err)
					continue
				}
			} else {
				eventTime, err = ExYearParseDate(rule.EventTimeFormat, eventTimeStr, time.Local)
				if err != nil {
					emit(logLevel.important, "FileReader: Failed parsing '%s' to format '%s'. Error: %s\n", eventTimeStr, rule.EventTimeFormat, err)
					continue
				}
			}

			eventTimeNumber := DateToCustomLong(eventTime)

			ipAddress := resultMap["ip"]

			skipEvent := ApplyExcludeFilterToSecurityEvents(eventFilter.ExcludeCompiledRegex, &resultMap)

			additionalFieldsMap := make(map[string]string)
			for key, value := range resultMap {
				if key != "ip" && key != "eventTime" {
					additionalFieldsMap[key] = value
				}
			}

			if len(additionalFieldsMap) < 1 {
				additionalFieldsMap = nil
			}

			if skipEvent == false {

				securityMessage := eventFilter.Message

				securityMessage = ApplyFilterToSecurityMessage(securityMessage, &resultMap)
				securityMessage = strings.Replace(securityMessage, "#ip", ipAddress, len(securityMessage))

				eventSource := *source
				if linePosition > 0 {
					eventSource = fmt.Sprintf("%s:%d", eventSource, linePosition-1)
				}

				securityEvent := &SecurityEvent{
					SecurityId:         securityId,
					SecurityGroupId:    eventFilter.Gid,
					EventTimeUtcNumber: eventTimeNumber,
					Message:            securityMessage,
					Critical:           eventFilter.Critical,
					IpAddress:          ipAddress,
					AdditionalFields:   additionalFieldsMap,
					Source:             &eventSource,
				}

				eventsContainer.SecurityEvents = append(eventsContainer.SecurityEvents, securityEvent)
				// debugJson(eventsContainer)
			}
		}
	}
}

func (crawler *FilesCrawler) _GetFilesListMap() map[string][]*RuleConfig {
	pathOnRulesMap := make(map[string][]*RuleConfig)

	// find files by all specified paths
	for _, rule := range crawler.Rules {
		for _, path := range rule.Paths {
			files, err := filepath.Glob(path)
			if err != nil {
				emit(logLevel.important, "Malformed specified path %s in the file %s. \t")
			}

			for _, file := range files {

				file = NormalizeFileName(file)

				// check if this rule is already assigned to this path
				mapRules := pathOnRulesMap[file]
				for _, mapRule := range mapRules {
					if mapRule.RuleFileName == rule.RuleFileName {
						continue
					}
				}

				// check if file is excluded by the current rule
				if rule.CompiledExcludeFilesRegex != nil {
					if rule.CompiledExcludeFilesRegex.MatchString(file) {
						continue
					}
				}

				pathOnRulesMap[file] = append(pathOnRulesMap[file], rule)
			}
		}
	}

	return pathOnRulesMap
}
