package main

import (
	"encoding/json"
	"fmt"
	"github.com/judwhite/go-svc/svc"
	"log"
	"os"
)

var program = &Program{}

func main() {

	defer func() {
		p := recover()
		if p == nil {
			return
		}
		fault("recovered panic: %v", p)
	}()

	options := &Options{}
	options.ParseArguments()

	if options.Version {
		fmt.Println(Version)
		return
	}

	program.Options = options

	// Call svc.Run to start your program/service.
	if err := svc.Run(program); err != nil {
		log.Fatal(err)
	}
}

func debug(msgfmt string, args ...interface{}) {
	emitLine(logLevel.verbose, msgfmt, args...)
}

func emitLine(level int, msgfmt string, args ...interface{}) {
	emit(level, msgfmt+"\n", args...)
}

func emit(level int, msgfmt string, args ...interface{}) {
	if program.Options.Verbose == false && level == logLevel.verbose {
		return
	}

	log.Printf(msgfmt, args...)
}

func debugJson(v interface{}) {
	emitJson(logLevel.verbose, v)
}

func emitJson(level int, v interface{}) {
	js, err := json.Marshal(v)
	if err != nil {
		emit(logLevel.important, "Failed convert object to json. Error: %s\n", err)
	} else {
		emit(level, "%s\n", js)
	}
}

func fault(msgfmt string, args ...interface{}) {
	exit(exitStat.faulted, msgfmt, args...)
}

func exit(stat int, msgfmt string, args ...interface{}) {
	log.Printf(msgfmt, args...)
	log.Printf("Exit!")
	os.Exit(stat)
}

var exitStat = struct {
	ok, usageError, faulted int
}{
	ok:      0,
	faulted: 2,
}

var logLevel = struct {
	verbose, important, critical int
}{
	verbose:   0,
	important: 1,
	critical:  2,
}
