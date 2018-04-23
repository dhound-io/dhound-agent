package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type MainConfig struct {
	Output OutputConfig `json:output yaml:output`
	Input  InputConfig  `json:input yaml:input`
}

type OutputConfig struct {
	AccessToken string `json:accesstoken yaml:accesstoken`
	ServerKey   string `json:serverkey yaml:serverkey`
	Environment string `json:environment,omitempty yaml:environment,omitempty`
	Proxy       string `json:proxy,omitempty yaml:proxy,omitempty`
}

type InputConfig struct {
	AllRules         bool         `json:allrules yaml:allrules`
	Rules            []string     `json:rules yaml:rules`
	NetworkInterface string       `json:networkinterface yaml:networkinterface`
	TrackDnsTraffic  bool         `json:trackdnstraffic yaml:trackdnstraffic`
	RuleConfigs      []RuleConfig `json:- yaml:-`
}

type RuleConfig struct {
	Source            string                `json:source yaml:source`
	Paths             []string              `json:paths yaml:paths`
	Encoding          string                `json:encoding yaml:encoding`
	DeadTime          string                `json:deadtime yaml:deadtime`
	ExcludeFilesRegex string                `json:excludefilesregex yaml:excludefilesregex`
	Events            []SecurityEventConfig `json:events yaml:events`
	EventTimeFormat   string                `json:eventtimeformat yaml:eventtimeformat`
	deadtime          time.Duration         `json:- yaml:-`

	CompiledExcludeFilesRegex *regexp.Regexp `json:- yaml:-`

	RuleFileName string `json:- yaml:-`
}

type SecurityEventConfig struct {
	Sid                  uint                      `json:sid yaml:sid`
	Gid                  uint                      `json:gid yaml:gid`
	WinEventIds          []uint                    `json:wineventids yaml:wineventids`
	Message              string                    `json:message yaml:message`
	Regex                string                    `json:regex yaml:regex`
	Fields               map[string]string         `json:fields yaml:fields`
	Exclude              map[string]string         `json:exclude yaml:exclude`
	ExcludeCompiledRegex map[string]*regexp.Regexp `json:- yaml:-`
	CompiledRegex        *regexp.Regexp            `json:- yaml:-`
	Critical             bool                      `json:critical yaml:critical`
}

func DiscoverYamlConfigs(directory string) (files []string, err error) {
	fi, err := os.Stat(directory)
	if err != nil {
		return nil, err
	}
	files = make([]string, 0)
	if fi.IsDir() {
		entries, err := ioutil.ReadDir(directory)
		if err != nil {
			return nil, err
		}
		for _, filename := range entries {
			if path.Ext(filename.Name()) == ".yml" {
				files = append(files, path.Join(directory, filename.Name()))
			}
		}
	}
	return files, nil
}

func LoadConfig(options *Options) (config MainConfig, err error) {

	directory := options.ConfigDir

	mainConfig := path.Join(directory, "config.yml")
	err = LoadYamlFile(mainConfig, &config)
	if err != nil {
		emitLine(logLevel.critical, "Failed loading main config file '%s': %s", mainConfig, err)
		return
	}

	rulesDir := path.Join(directory, "rules.d")

	ruleFiles, err := DiscoverYamlConfigs(rulesDir)
	if err != nil {
		emit(logLevel.critical, "Failed loading rules config files (*.yml) in directory: %s. Error: %s\n", rulesDir, err)
		return
	}

	if len(ruleFiles) < 1 {
		emit(logLevel.critical, "No *.yml rule config files found in directory: %s.\n", rulesDir)
		return
	}

	rules := make([]RuleConfig, 0)
	for _, ruleFile := range ruleFiles {
		var rule RuleConfig
		ruleConfigErr := LoadYamlFile(ruleFile, &rule)

		if ruleConfigErr != nil {
			emit(logLevel.important, "Failed loading file '%s': %s.\n", ruleFile, ruleConfigErr)
		} else {

			_, fileName := path.Split(ruleFile)
			fileName = strings.ToLower(fileName)
			rule.RuleFileName = strings.TrimSuffix(fileName, ".yml")

			// normalize rule
			if rule.DeadTime == "" {
				rule.DeadTime = options.DefaultFileDeadtime
			}

			if rule.ExcludeFilesRegex == "" {
				rule.ExcludeFilesRegex = options.DefaultExcludeFileFilter
			}

			rule.deadtime, err = time.ParseDuration(rule.DeadTime)
			if err != nil {
				emit(logLevel.important, "Failed parsing deadtime in file '%s': %s.\n", ruleFile, err)
				continue
			}

			if len(rule.ExcludeFilesRegex) > 0 {
				excludeFilesRegex, err := regexp.Compile(rule.ExcludeFilesRegex)
				if err != nil {
					emit(logLevel.important, "Failed parsing excludefilesregex '%s' in config file: '%s'. Error: %s\n", rule.ExcludeFilesRegex, ruleFile, err)
					continue
				}
				rule.CompiledExcludeFilesRegex = excludeFilesRegex
			}

			events := make([]SecurityEventConfig, 0)
			for _, event := range rule.Events {
				regex := event.Regex
				compiledRegex, err := regexp.Compile(regex)
				if err != nil {
					emit(logLevel.important, "Failed parsing regex '%s' in config file '%s'. Error: %s\n", regex, ruleFile, err)
					continue
				}

				event.CompiledRegex = compiledRegex

				if config.Input.AllRules == false {
					if Contains(config.Input.Rules, rule.RuleFileName) == false {
						continue
					}
				}

				// compile exlude filter
				if event.Exclude != nil && len(event.Exclude) > 0 {
					event.ExcludeCompiledRegex = make(map[string]*regexp.Regexp)
					for key, value := range event.Exclude {
						excludeCompiledRegex, err := regexp.Compile(value)
						if err != nil {
							emit(logLevel.important, "Failed parsing exclude regex: '%s' in config file '%s'. Error: %s\n", value, ruleFile, err)
							continue
						}
						event.ExcludeCompiledRegex[key] = excludeCompiledRegex
					}
				}

				events = append(events, event)
			}

			rule.Events = events

			if len(rule.Events) > 0 {
				rules = append(rules, rule)
			}
		}
	}

	config.Input.RuleConfigs = rules
	// emitJson(logLevel.verbose, config.Input)

	ruleFiles = make([]string, 0)
	for _, rule := range config.Input.RuleConfigs {
		ruleFiles = append(ruleFiles, "'"+rule.RuleFileName+"'")
	}
	emit(logLevel.important, "Loaded rule files: %s.\n", strings.Join(ruleFiles, ", "))

	return
}

func LoadYamlFile(path string, out interface{}) error {
	ymlFile, err := os.Open(path)
	if err != nil {
		return err
	}

	fi, _ := ymlFile.Stat()

	buffer := make([]byte, fi.Size())
	_, err = ymlFile.Read(buffer)

	buffer = []byte(os.ExpandEnv(string(buffer)))

	err = yaml.Unmarshal(buffer, out)
	if err != nil {
		return errors.New(fmt.Sprintf("Incorrect yaml format: %s", err))
	}

	return nil
}
