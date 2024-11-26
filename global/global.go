package global

import (
	"os"
	"strconv"

	"github.com/projectdiscovery/utils/sysutil"
)

var OS_MAX_THREADS int

func init() {
	OS_MAX_THREADS = func() int {
		if value, err := strconv.Atoi(os.Getenv("OS_MAX_THREADS")); err == nil && value != 0 {
			return value
		}
		return 10000
	}()

	_ = sysutil.SetMaxThreads(OS_MAX_THREADS)
}
