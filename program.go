package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/judwhite/go-svc/svc"
	"gopkg.in/natefinch/lumberjack.v2"
)

import _ "net/http/pprof"

type Program struct {
	Options *Options
	Wg      sync.WaitGroup
	Quit    chan struct{}
}

func (program *Program) Init(env svc.Environment) error {

	options := program.Options

	// configure logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	if len(options.LogFile) > 0 {
		fmt.Println("See output in  " + options.LogFile)

		CreateDirIfNotExist(filepath.Dir(options.LogFile), 0765)

		log.SetOutput(&lumberjack.Logger{
			Filename:   options.LogFile,
			MaxSize:    50, // megabytes
			MaxBackups: 3,
			MaxAge:     28, //days
		})
	}

	return nil
}

func (program *Program) InternalRun() {

	options := program.Options

	emitLine(logLevel.important, "The agent %s is started. Options: config='%s' log-file='%s' timeout='%v' verbose='%t'", Version, options.ConfigDir, options.LogFile, options.IdleTimeout, options.Verbose)

	if len(options.Pprof) > 0 {
		go func() {
			emit(logLevel.verbose, "run profiler on http://%s/debug/pprof/\n", options.Pprof)
			err := http.ListenAndServe(options.Pprof, nil)
			if err != nil {
				emit(logLevel.important, "failed running profiler: %s \n", err.Error())
			}
		}()
	}

	config, err := LoadConfig(options)
	if err != nil {
		exit(exitStat.faulted, "Failed loading config files")
		return
	}

	if len(config.Input.RuleConfigs) == 0 {
		exit(exitStat.faulted, "Input section in the config file does not contains any rules.")
		return
	}

	sysProcessManager := &SysProcessManager{}
	sysProcessManager.Init()
	sysProcessManager.Run()

	networkEventEnricher := &NetworkEventEnricher{
		Input: make(chan *NetworkEvent),
		SysManager: sysProcessManager
	}

	networkMonitor := &NetworkMonitor{
		Output: networkEventEnricher.Input,
	}
	networkMonitor.Run()
	networkEventEnricher.Run()

	// init all channels
	systemState := &SystemState{
		Input: make(chan []*SecurityEventsContainer),
	}
	systemState.Restore()

	gate := &HttpGateway{
		SystemState: systemState,
		Options:     options,
		MainConfig:  &config,
		Input:       make(chan []*SecurityEventsContainer),
	}
	gate.Init()

	queue := &Queue{
		Options:     options,
		Input:       make(chan *SecurityEventsContainer),
		NextChannel: gate.Input,
	}
	queue.Init()

	ipEnricher := &IpEnricher{
		Input:       make(chan *SecurityEventsContainer),
		NextChannel: queue.Input,
		Options:     options,
		Config:      &config,
	}
	ipEnricher.Init()

	fileRules := make([]*RuleConfig, 0)
	winEventRules := make([]*RuleConfig, 0)

	for _, config := range config.Input.RuleConfigs {
		ruleConfig := config
		if ruleConfig.Source == "wineventlog" {
			winEventRules = append(winEventRules, &ruleConfig)
		} else {
			fileRules = append(fileRules, &ruleConfig)
		}
	}

	// run processing messages from channels
	go systemState.Sync()
	go ipEnricher.Run()
	go queue.Run()
	go gate.Run()

	// run crawler over files
	if len(fileRules) > 0 {
		filesCrawler := &FilesCrawler{
			Rules:       fileRules,
			SystemState: systemState,
			NextChannel: ipEnricher.Input,
		}
		filesCrawler.Init()
		go filesCrawler.Run()
	}

	if runtime.GOOS == "windows" {
		// run crawler over win event logs
		if len(winEventRules) > 0 {
			winEventCrawler := &WinEventLogCrawler{
				Rules:       winEventRules,
				Options:     options,
				NextChannel: ipEnricher.Input,
				SystemState: systemState,
			}
			winEventCrawler.Init()
			go winEventCrawler.Run()
		}
	}
}

func (program *Program) Start() error {
	// The Start method must not block, or Windows may assume your service failed
	// to start. Launch a Goroutine here to do something interesting/blocking.

	program.Quit = make(chan struct{})

	program.Wg.Add(1)
	go func() {

		program.InternalRun()

		<-program.Quit
		// debug("Quit signal received...")
		program.Wg.Done()
	}()

	return nil
}

func (program *Program) Stop() error {
	// The Stop method is invoked by stopping the Windows service, or by pressing Ctrl+C on the console.
	// This method may block, but it's a good idea to finish quickly or your process may be killed by
	// Windows during a shutdown/reboot. As a general rule you shouldn't rely on graceful shutdown.

	// emitLine(logLevel.verbose, "Stopping...")
	close(program.Quit)
	program.Wg.Wait()
	emitLine(logLevel.verbose, "Stopped.")

	return nil
}
