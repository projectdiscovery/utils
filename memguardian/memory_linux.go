//go:build linux

package memguardian

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func getSysInfo() (*SysInfo, error) {
	return getSysInfoForPaths("/proc/self/cgroup", "/sys/fs/cgroup")
}

func getSysInfoForPaths(procCgroupPath, cgroupRoot string) (*SysInfo, error) {
	var sysInfo syscall.Sysinfo_t
	err := syscall.Sysinfo(&sysInfo)
	if err != nil {
		return nil, err
	}

	si := &SysInfo{
		Uptime:    int64(sysInfo.Uptime),
		totalRam:  uint64(sysInfo.Totalram),
		freeRam:   uint64(sysInfo.Freeram),
		SharedRam: uint64(sysInfo.Freeram),
		BufferRam: uint64(sysInfo.Bufferram),
		TotalSwap: uint64(sysInfo.Totalswap),
		FreeSwap:  uint64(sysInfo.Freeswap),
		Unit:      uint64(sysInfo.Unit),
	}
	used, limit, limited, err := cgroupMemoryUsage(procCgroupPath, cgroupRoot)
	if err != nil {
		return nil, err
	}
	if limited && limit < si.TotalRam() {
		si.totalRam = limit
		si.freeRam = limit - used
		si.Unit = 1
	}

	return si, nil
}

func cgroupMemoryUsage(procCgroupPath, cgroupRoot string) (uint64, uint64, bool, error) {
	data, err := os.ReadFile(procCgroupPath)
	if os.IsNotExist(err) {
		return 0, 0, false, nil
	}
	if err != nil {
		return 0, 0, false, fmt.Errorf("read process cgroup: %w", err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.SplitN(line, ":", 3)
		if len(parts) != 3 {
			continue
		}
		group := strings.TrimPrefix(filepath.Clean("/"+parts[2]), "/")
		if parts[0] == "0" && parts[1] == "" {
			used, limit, limited, err := memoryUsageFromHierarchy(
				filepath.Join(cgroupRoot, group), cgroupRoot,
				"memory.current", "memory.max", "memory.stat",
				"inactive_file",
			)
			if err != nil || limited {
				return used, limit, limited, err
			}
			continue
		}
		for _, controller := range strings.Split(parts[1], ",") {
			if controller == "memory" {
				root := filepath.Join(cgroupRoot, "memory")
				return memoryUsageFromHierarchy(
					filepath.Join(root, group), root,
					"memory.usage_in_bytes", "memory.limit_in_bytes", "memory.stat",
					"total_inactive_file",
				)
			}
		}
	}
	return 0, 0, false, nil
}

func memoryUsageFromHierarchy(groupDir, rootDir, currentName, limitName, statName, inactiveFileKey string) (uint64, uint64, bool, error) {
	rootDir = filepath.Clean(rootDir)
	groupDir = filepath.Clean(groupDir)
	var selectedUsed, selectedLimit uint64
	var limited bool
	for {
		used, limit, hasLimit, err := memoryUsageFromFiles(
			filepath.Join(groupDir, currentName),
			filepath.Join(groupDir, limitName),
			filepath.Join(groupDir, statName),
			inactiveFileKey,
		)
		if err != nil {
			return 0, 0, false, err
		}
		if hasLimit && (!limited || float64(used)/float64(limit) > float64(selectedUsed)/float64(selectedLimit)) {
			selectedUsed, selectedLimit, limited = used, limit, true
		}
		if groupDir == rootDir {
			break
		}
		parent := filepath.Dir(groupDir)
		if parent == groupDir || (parent != rootDir && !strings.HasPrefix(parent, rootDir+string(filepath.Separator))) {
			break
		}
		groupDir = parent
	}
	return selectedUsed, selectedLimit, limited, nil
}

func memoryUsageFromFiles(currentPath, limitPath, statPath, inactiveFileKey string) (uint64, uint64, bool, error) {
	limitData, err := os.ReadFile(limitPath)
	if os.IsNotExist(err) {
		return 0, 0, false, nil
	}
	if err != nil {
		return 0, 0, false, fmt.Errorf("read cgroup memory limit: %w", err)
	}
	limitText := strings.TrimSpace(string(limitData))
	if limitText == "max" {
		return 0, 0, false, nil
	}
	limit, err := strconv.ParseUint(limitText, 10, 64)
	if err != nil {
		return 0, 0, false, fmt.Errorf("parse cgroup memory limit: %w", err)
	}
	if limit >= 1<<60 {
		return 0, 0, false, nil
	}
	if limit == 0 {
		return 0, 0, false, fmt.Errorf("cgroup memory limit is zero")
	}

	currentData, err := os.ReadFile(currentPath)
	if err != nil {
		return 0, 0, false, fmt.Errorf("read cgroup memory usage: %w", err)
	}
	current, err := strconv.ParseUint(strings.TrimSpace(string(currentData)), 10, 64)
	if err != nil {
		return 0, 0, false, fmt.Errorf("parse cgroup memory usage: %w", err)
	}
	statData, err := os.ReadFile(statPath)
	if err != nil {
		return 0, 0, false, fmt.Errorf("read cgroup memory stats: %w", err)
	}
	var inactive uint64
	for _, line := range strings.Split(string(statData), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || fields[0] != inactiveFileKey {
			continue
		}
		inactive, err = strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return 0, 0, false, fmt.Errorf("parse cgroup inactive memory: %w", err)
		}
		break
	}
	if inactive > current {
		inactive = current
	}
	used := current - inactive
	if used > limit {
		used = limit
	}
	return used, limit, true, nil
}
