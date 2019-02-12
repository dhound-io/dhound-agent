package main

import (
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type Options struct {
	ConfigDir                string
	LogFile                  string
	IdleTimeoutInSeconds     int
	IdleTimeout              time.Duration
	Version                  bool
	Verbose                  bool
	Pprof                    string
	NetTimeout               int64
	DefaultFileDeadtime      string
	DefaultExcludeFileFilter string
}

func (options *Options) ParseArguments() {

	options.NetTimeout = 15
	options.DefaultFileDeadtime = "360h"
	options.DefaultExcludeFileFilter = "((.gz)|(.zip)|(.tar)|(.zip))"

	flag.StringVar(&options.ConfigDir, "config-dir", options.ConfigDir, "path to dhound-agent configuration directory")
	flag.StringVar(&options.LogFile, "log-file", options.LogFile, "path to the dhound log file")

	flag.IntVar(&options.IdleTimeoutInSeconds, "timeout", 60, "frequency in seconds to send data on the server")

	flag.BoolVar(&options.Verbose, "verbose", options.Verbose, "log more detailed and debug information")
	flag.BoolVar(&options.Version, "version", options.Version, "dhound-agent version")

	flag.StringVar(&options.Pprof, "pprof", options.Pprof, "profiling option (for internal using)")

	flag.Parse()

	if runtime.GOOS == "windows" {
		// for windows all files are located on the same folder, current directory should be set up in the code
		execPath, err := Executable()
		if err != nil {
			exit(exitStat.faulted, "failed receiving current executable path. error: %s\n", err)
			return
		}

		workingDir := filepath.Dir(execPath)
		// set current working directory
		err = os.Chdir(workingDir)
		if err != nil {
			exit(exitStat.faulted, "failed settings working dir. error: %s\n", err)
			return
		}

		if len(options.ConfigDir) < 1 {
			options.ConfigDir = "config"
		}
	}

	if options.IdleTimeoutInSeconds > 0 {
		options.IdleTimeout = time.Second * time.Duration(options.IdleTimeoutInSeconds)
	}
}
