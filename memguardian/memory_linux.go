//go:build linux

package memguardian

import "syscall"

func getSysInfo() (*SysInfo, error) {
	var sysInfo syscall.Sysinfo_t
	err := syscall.Sysinfo(&sysInfo)
	if err != nil {
		return nil, err
	}

	si := &SysInfo{
		Uptime:    sysInfo.Uptime,
		totalRam:  sysInfo.Totalram,
		freeRam:   sysInfo.Freeram,
		SharedRam: sysInfo.Freeram,
		BufferRam: sysInfo.Bufferram,
		TotalSwap: sysInfo.Totalswap,
		FreeSwap:  sysInfo.Freeswap,
		Unit:      uint64(sysInfo.Unit),
	}

	return si, nil
}
