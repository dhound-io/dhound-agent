package main

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

type SysProcessManager struct {
	_pidToProcessInfoMap map[int32]*ProcessInfo
	_firstRun            bool
}

type ProcessInfo struct {
	Name, CommandLine string
}

func (manager *SysProcessManager) Init() {
	manager._pidToProcessInfoMap = make(map[int32]*ProcessInfo)
}

func (manager *SysProcessManager) Run() {
	// time ticker to flush events
	ticker := time.NewTicker(time.Second * 10)
	go func() {
		manager._syncProcessInfoOnPids()
		for _ = range ticker.C {
			manager._syncProcessInfoOnPids()
		}
	}()
}

func (manager *SysProcessManager) _syncProcessInfoOnPids() {

	processes, err := process.Processes()
	if err != nil {
		emitLine(logLevel.verbose, "could not get processes: %s", err.Error())
		return
	}

	pids := make([]int32, 0)
	for _, process := range processes {
		pids = append(pids, process.Pid)
	}

	pidsToProcess := make([]int32, 0)

	// sync map
	for _, pid := range pids {
		if _, ok := manager._pidToProcessInfoMap[pid]; !ok {
			manager._pidToProcessInfoMap[pid] = &ProcessInfo{}
			pidsToProcess = append(pidsToProcess, pid)
		}
	}

	obsoletePids := make([]int32, 0)
	for pid, _ := range manager._pidToProcessInfoMap {
		if ContainsInt32(pids, pid) == false {
			// debug("remove pid: %d (%s)", pid, value.Name)
			obsoletePids = append(obsoletePids, pid)
		}
	}

	for _, pid := range obsoletePids {
		delete(manager._pidToProcessInfoMap, pid)
	}

	if len(pidsToProcess) > 0 {

		// parse name
		for _, process := range processes {
			if ContainsInt32(pidsToProcess, process.Pid) {
				name, _ := process.Name()
				manager._pidToProcessInfoMap[process.Pid].Name = name
				// debug("new pid: %d (%s)", process.Pid, name)
			}
		}
		// parse commmandline
		for _, process := range processes {
			if ContainsInt32(pidsToProcess, process.Pid) {
				cmdLine, _ := process.Cmdline()
				manager._pidToProcessInfoMap[process.Pid].CommandLine = cmdLine
				// debug("new pid: %d (%s)", process.Pid, cmdLine)
			}
		}
	}
}

func SyncLocalPortsOnPids() {
	connections, err := net.Connections("inet")
	if err != nil {
		fmt.Errorf("could not get NetConnections: %v", err)
	}
	if len(connections) == 0 {
		fmt.Errorf("could not get NetConnections: %v", connections)
	}
	for _, connection := range connections {
		// track output connections only
		if connection.Family != 0 && connection.Status != "LISTEN" {
			debugJson(connection)
		}
	}
}
