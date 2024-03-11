package detach

import (
	"os"
	"os/exec"
	"strings"

	"github.com/cespare/xxhash"
	sliceutil "github.com/projectdiscovery/utils/slice"
	"github.com/shirou/gopsutil/v3/process"
)

var (
	DetachArg = "-detach"
)

func isDetach() bool {
	return sliceutil.Contains(os.Args, DetachArg)
}

func hash(binary string, args ...string) uint64 {
	all := []string{binary}
	all = append(all, args...)
	fullCli := strings.TrimSpace(strings.ToLower(strings.Join(all, " ")))
	return xxhash.Sum64String(fullCli)
}

func isUnique() bool {
	myHash := hash(os.Args[0], os.Args[1:]...)
	processes, _ := process.Processes()
	var numberOfInstances int
	for _, process := range processes {
		cmdLine, err := process.CmdlineSlice()
		if err != nil {
			continue
		}
		binaryPath, err := process.Exe()
		if err != nil {
			continue
		}
		processHash := hash(binaryPath, cmdLine...)
		if myHash == processHash {
			numberOfInstances++
		}
	}
	return numberOfInstances == 1
}

// Run a function in detached mode - the child process will be disconnected from the parent one
func Run(f func() error) error {
	if isDetach() {
		return f()
	}

	cmd := exec.Command(os.Args[0], append(os.Args[1:], DetachArg)...)

	return cmd.Start()
}

// Run a function in detached mode one time and prevents overlapping executions
func RunSingleFlight(f func() error) error {
	if isDetach() && isUnique() {
		return f()
	}

	cmd := exec.Command(os.Args[0], append(os.Args[1:], DetachArg)...)

	return cmd.Start()
}
