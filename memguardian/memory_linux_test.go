//go:build linux

package memguardian

import (
	"os"
	"path/filepath"
	"testing"
)

func writeMemoryFixture(t *testing.T, path, value string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(value), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestCgroupV2MemoryOverridesHostMemory(t *testing.T) {
	root := t.TempDir()
	procPath := filepath.Join(root, "proc", "cgroup")
	group := filepath.Join(root, "sys", "worker")
	writeMemoryFixture(t, procPath, "0::/worker\n")
	writeMemoryFixture(t, filepath.Join(group, "memory.max"), "1000\n")
	writeMemoryFixture(t, filepath.Join(group, "memory.current"), "900\n")
	writeMemoryFixture(t, filepath.Join(group, "memory.stat"), "anon 800\ninactive_file 100\n")

	info, err := getSysInfoForPaths(procPath, filepath.Join(root, "sys"))
	if err != nil {
		t.Fatal(err)
	}
	if got := info.TotalRam(); got != 1000 {
		t.Fatalf("total RAM = %d, want 1000", got)
	}
	if got := info.UsedRam(); got != 800 {
		t.Fatalf("used RAM = %d, want 800", got)
	}
	if got := info.UsedPercent(); got != 80 {
		t.Fatalf("used percent = %f, want 80", got)
	}
}

func TestCgroupV1MemoryController(t *testing.T) {
	root := t.TempDir()
	procPath := filepath.Join(root, "proc", "cgroup")
	group := filepath.Join(root, "sys", "memory", "worker")
	writeMemoryFixture(t, procPath, "7:cpu,cpuacct:/worker\n6:memory:/worker\n")
	writeMemoryFixture(t, filepath.Join(group, "memory.limit_in_bytes"), "1000\n")
	writeMemoryFixture(t, filepath.Join(group, "memory.usage_in_bytes"), "950\n")
	writeMemoryFixture(t, filepath.Join(group, "memory.stat"), "total_inactive_file 100\n")

	used, limit, limited, err := cgroupMemoryUsage(procPath, filepath.Join(root, "sys"))
	if err != nil {
		t.Fatal(err)
	}
	if !limited || used != 850 || limit != 1000 {
		t.Fatalf("cgroup memory = %d/%d limited=%t, want 850/1000 true", used, limit, limited)
	}
}

func TestCgroupV1RemappedMountRoot(t *testing.T) {
	root := t.TempDir()
	procPath := filepath.Join(root, "proc", "cgroup")
	mountPoint := filepath.Join(root, "sys", "memory")
	writeMemoryFixture(t, procPath, "6:memory:/docker/worker\n")
	writeMemoryFixture(t, filepath.Join(root, "proc", "mountinfo"),
		"27 22 0:28 /docker "+mountPoint+" rw - cgroup cgroup rw,memory\n")
	writeMemoryFixture(t, filepath.Join(mountPoint, "memory.limit_in_bytes"), "1000\n")
	writeMemoryFixture(t, filepath.Join(mountPoint, "memory.usage_in_bytes"), "320\n")
	writeMemoryFixture(t, filepath.Join(mountPoint, "memory.stat"), "total_inactive_file 100\n")
	worker := filepath.Join(mountPoint, "worker")
	writeMemoryFixture(t, filepath.Join(worker, "memory.limit_in_bytes"), "400\n")
	writeMemoryFixture(t, filepath.Join(worker, "memory.usage_in_bytes"), "300\n")
	writeMemoryFixture(t, filepath.Join(worker, "memory.stat"), "total_inactive_file 100\n")

	used, limit, limited, err := cgroupMemoryUsage(procPath, filepath.Join(root, "sys"))
	if err != nil {
		t.Fatal(err)
	}
	if !limited || used != 200 || limit != 400 {
		t.Fatalf("cgroup memory = %d/%d limited=%t, want 200/400 true", used, limit, limited)
	}
}

func TestCgroupV2RemappedMountRoot(t *testing.T) {
	root := t.TempDir()
	procPath := filepath.Join(root, "proc", "cgroup")
	mountPoint := filepath.Join(root, "sys")
	writeMemoryFixture(t, procPath, "0::/parent/worker\n")
	writeMemoryFixture(t, filepath.Join(root, "proc", "mountinfo"),
		"27 22 0:28 /parent "+mountPoint+" rw - cgroup2 cgroup rw\n")
	writeMemoryFixture(t, filepath.Join(mountPoint, "memory.max"), "1000\n")
	writeMemoryFixture(t, filepath.Join(mountPoint, "memory.current"), "320\n")
	writeMemoryFixture(t, filepath.Join(mountPoint, "memory.stat"), "inactive_file 100\n")
	worker := filepath.Join(mountPoint, "worker")
	writeMemoryFixture(t, filepath.Join(worker, "memory.max"), "400\n")
	writeMemoryFixture(t, filepath.Join(worker, "memory.current"), "300\n")
	writeMemoryFixture(t, filepath.Join(worker, "memory.stat"), "inactive_file 100\n")

	used, limit, limited, err := cgroupMemoryUsage(procPath, mountPoint)
	if err != nil {
		t.Fatal(err)
	}
	if !limited || used != 200 || limit != 400 {
		t.Fatalf("cgroup memory = %d/%d limited=%t, want 200/400 true", used, limit, limited)
	}
}

func TestHybridCgroupFallsBackToLimitedMemoryController(t *testing.T) {
	root := t.TempDir()
	procPath := filepath.Join(root, "proc", "cgroup")
	group := filepath.Join(root, "sys", "memory", "worker")
	writeMemoryFixture(t, procPath, "0::/worker\n6:memory:/worker\n")
	writeMemoryFixture(t, filepath.Join(group, "memory.limit_in_bytes"), "1000\n")
	writeMemoryFixture(t, filepath.Join(group, "memory.usage_in_bytes"), "900\n")
	writeMemoryFixture(t, filepath.Join(group, "memory.stat"), "total_inactive_file 100\n")

	used, limit, limited, err := cgroupMemoryUsage(procPath, filepath.Join(root, "sys"))
	if err != nil {
		t.Fatal(err)
	}
	if !limited || used != 800 || limit != 1000 {
		t.Fatalf("cgroup memory = %d/%d limited=%t, want 800/1000 true", used, limit, limited)
	}
}

func TestCgroupV2UsesConstrainedAncestor(t *testing.T) {
	root := t.TempDir()
	procPath := filepath.Join(root, "proc", "cgroup")
	parent := filepath.Join(root, "sys", "parent")
	child := filepath.Join(parent, "child")
	writeMemoryFixture(t, procPath, "0::/parent/child\n")
	writeMemoryFixture(t, filepath.Join(child, "memory.max"), "max\n")
	writeMemoryFixture(t, filepath.Join(parent, "memory.max"), "1000\n")
	writeMemoryFixture(t, filepath.Join(parent, "memory.current"), "950\n")
	writeMemoryFixture(t, filepath.Join(parent, "memory.stat"), "inactive_file 50\n")

	used, limit, limited, err := cgroupMemoryUsage(procPath, filepath.Join(root, "sys"))
	if err != nil {
		t.Fatal(err)
	}
	if !limited || used != 900 || limit != 1000 {
		t.Fatalf("cgroup memory = %d/%d limited=%t, want 900/1000 true", used, limit, limited)
	}
}

func TestUnlimitedCgroupUsesHostMemory(t *testing.T) {
	root := t.TempDir()
	procPath := filepath.Join(root, "proc", "cgroup")
	group := filepath.Join(root, "sys", "worker")
	writeMemoryFixture(t, procPath, "0::/worker\n")
	writeMemoryFixture(t, filepath.Join(group, "memory.max"), "max\n")

	info, err := getSysInfoForPaths(procPath, filepath.Join(root, "sys"))
	if err != nil {
		t.Fatal(err)
	}
	if info.TotalRam() == 0 {
		t.Fatal("unlimited cgroup did not fall back to host memory")
	}
}

func TestCgroupLimitAboveHostUsesHostMemory(t *testing.T) {
	root := t.TempDir()
	procPath := filepath.Join(root, "proc", "cgroup")
	group := filepath.Join(root, "sys", "worker")
	writeMemoryFixture(t, procPath, "0::/worker\n")
	writeMemoryFixture(t, filepath.Join(group, "memory.max"), "36028797018963968\n")
	writeMemoryFixture(t, filepath.Join(group, "memory.current"), "1000\n")
	writeMemoryFixture(t, filepath.Join(group, "memory.stat"), "inactive_file 0\n")

	info, err := getSysInfoForPaths(procPath, filepath.Join(root, "sys"))
	if err != nil {
		t.Fatal(err)
	}
	if info.TotalRam() >= 1<<55 {
		t.Fatalf("total RAM = %d, want host memory below ineffective cgroup limit", info.TotalRam())
	}
}

func TestInvalidLimitedCgroupDoesNotHideMemoryPressure(t *testing.T) {
	root := t.TempDir()
	procPath := filepath.Join(root, "proc", "cgroup")
	group := filepath.Join(root, "sys", "worker")
	writeMemoryFixture(t, procPath, "0::/worker\n")
	writeMemoryFixture(t, filepath.Join(group, "memory.max"), "1000\n")
	writeMemoryFixture(t, filepath.Join(group, "memory.current"), "invalid\n")
	writeMemoryFixture(t, filepath.Join(group, "memory.stat"), "inactive_file 0\n")

	if _, err := getSysInfoForPaths(procPath, filepath.Join(root, "sys")); err == nil {
		t.Fatal("expected an error for malformed limited cgroup usage")
	}
}
