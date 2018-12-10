package main

import (
	"time"

	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

type SysProcessManager struct {
	_pidToProcessInfoMap map[int32]*ProcessInfo
	_localPortOnPidMap   map[uint32]int32
	_firstRun            bool
}

type ProcessInfo struct {
	Name, CommandLine string
	Pid               int32
}

func (manager *SysProcessManager) Init() {
	manager._pidToProcessInfoMap = make(map[int32]*ProcessInfo)
	manager._localPortOnPidMap = make(map[uint32]int32)
}

func (manager *SysProcessManager) Run() {

	// run sync local port on pid
	go func() {
		firstRun := manager._syncLocalPortsOnPids()
		if firstRun {
			emitLine(logLevel.verbose, "Info about local ports started collecting. detected local open ports: %d", len(manager._localPortOnPidMap))
			for _ = range time.NewTicker(time.Second * 5).C {
				manager._syncLocalPortsOnPids()
			}
		} else {
			emitLine(logLevel.important, "Info about local ports won't be collected")
		}
	}()

	// run sync pid on process info
	go func() {
		firstRun := manager._syncProcessInfoOnPids(true)
		if firstRun {
			emitLine(logLevel.important, "Info about local current processes started collecting. detected processes: %d", len(manager._pidToProcessInfoMap))
			for _ = range time.NewTicker(time.Second * 30).C {
				manager._syncProcessInfoOnPids(false)
			}
		} else {
			emitLine(logLevel.important, "Info about local current processes won't be collected")
		}
	}()
}

func (manager *SysProcessManager) _syncLocalPortsOnPids() bool {
	connections, err := net.Connections("inet")
	if err != nil {
		if err != nil {
			emitLine(logLevel.important, "could not get connections usage: %s", err.Error())
			return false
		}
	}

	lportsMap := manager._localPortOnPidMap

	lports := make([]uint32, 0)

	if connections != nil {
		for _, connection := range connections {
			// track output connections only
			// if connection.Family != 0 && connection.Status != "LISTEN" {
			lportsMap[connection.Laddr.Port] = connection.Pid
			lports = append(lports, connection.Laddr.Port)
			// }
		}
	}

	// debugJson(lportsMap)

	for lport, _ := range lportsMap {
		if ContainsUint32(lports, lport) == false {
			// debug("Removed port %d", lport)
			delete(lportsMap, lport)
		}
	}

	return true
}

func (manager *SysProcessManager) _syncProcessInfoOnPids(quick bool) bool {

	processes, err := process.Processes()
	if err != nil {
		emitLine(logLevel.important, "could not get processes: %s", err.Error())
		return false
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

		if quick == false {
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
	return true
}

func (manager *SysProcessManager) FindProcessInfoByLocalPort(localPort uint32) *ProcessInfo {

	var processId int32 = -1

	// find pid
	if pid, ok := manager._localPortOnPidMap[localPort]; ok {
		processId = pid
	} else {
		manager._syncProcessInfoOnPids(true)
		if pid, ok := manager._localPortOnPidMap[localPort]; ok {
			processId = pid
		}
	}

	if processId > 0 {
		result := &ProcessInfo{}

		if processInfo, ok := manager._pidToProcessInfoMap[processId]; ok {
			result = processInfo
		} else {
			manager._syncProcessInfoOnPids(true)
			if processInfo, ok := manager._pidToProcessInfoMap[processId]; ok {
				result = processInfo
			}
		}

		result.Pid = processId

		return result
	} else {
		// debug("Local port not found for local port %d", localPort)
	}

	return nil
}
