package process

import (
	stringsutil "github.com/projectdiscovery/utils/strings"
	ps "github.com/shirou/gopsutil/v3/process"
)

// CloseProcesses part
func CloseProcesses(filter func(process *ps.Process) bool, skipPids map[int32]struct{}) {
	processes, err := ps.Processes()
	if err != nil {
		return
	}

	for _, process := range processes {
		// skip non-chrome processes
		if !IsChromeProcess(process) {
			continue
		}
		// skip chrome processes that were already running
		if _, ok := skipPids[process.Pid]; ok {
			continue
		}
		_ = process.Kill()
	}
}

// FindProcesses finds chrome process running on host
func FindProcesses(filter func(process *ps.Process) bool) map[int32]struct{} {
	processes, _ := ps.Processes()
	list := make(map[int32]struct{})
	for _, process := range processes {
		if filter(process) {
			list[process.Pid] = struct{}{}
			if ppid, err := process.Ppid(); err == nil {
				list[ppid] = struct{}{}
			}
		}
	}
	return list
}

// IsChromeProcess checks if a process is chrome/chromium
func IsChromeProcess(process *ps.Process) bool {
	name, _ := process.Name()
	executable, _ := process.Exe()
	return stringsutil.ContainsAny(name, "chrome", "chromium") || stringsutil.ContainsAny(executable, "chrome", "chromium")
}
