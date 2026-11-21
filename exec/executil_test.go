package executil

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var newLineMarker string

func init() {
	if runtime.GOOS == "windows" {
		newLineMarker = "\r\n"
	} else {
		newLineMarker = "\n"
	}
}

func TestRun(t *testing.T) {
	// try to run the echo command
	s, err := Run("echo test")
	require.Nil(t, err, "failed execution", err)
	require.Equal(t, "test"+newLineMarker, s, "output doesn't contain expected result", s)
}

func TestRunAdv(t *testing.T) {
	testcases := []struct {
		GOOS     string // OS
		Command  string
		Expected string // expected output
		Contains string // expected output contains
	}{
		// Tests With Flags
		{"darwin", "uname -s", "Darwin", ""},
		{"linux", "uname -s", "Linux", ""},
		{"windows", "cmd /c ver", "", "Windows"},
		// Tests With CMD PIPE
		{"windows", `systeminfo | findstr /B /C:"OS Name"`, "", "Windows"},
		{"darwin", `sw_vers | grep -i "ProductName"`, "", "macOS"},
		{"linux", `uname -a  | cut -d " " -f 1`, "Linux", ""},
		// Other Shell Specific Features
		{"windows", `cmd /c " echo This && echo Works"`, "This \r\nWorks", ""},
		{"linux", "true && echo This Works", "This Works", ""},
		{"darwin", "true && echo This Works", "This Works", ""},
	}

	runFunc := func(cmd string, expected string, contains string) {
		s, err := Run(cmd)
		require.Nilf(t, err, "%v failed to execute", cmd)
		if expected != "" {
			require.Equal(t, expected+newLineMarker, s)
		} else if contains != "" {
			require.Contains(t, s, contains)
		} else {
			t.Logf("Malformed test case : %v", cmd)
		}
		t.Logf("Test Successful: %v", cmd)
	}

	for _, v := range testcases {
		switch v.GOOS {
		case "windows":
			if runtime.GOOS != "windows" {
				continue
			}
			runFunc(v.Command, v.Expected, v.Contains)
		case "darwin":
			if runtime.GOOS != "darwin" {
				continue
			}
			runFunc(v.Command, v.Expected, v.Contains)
		case "linux":
			if runtime.GOOS != "linux" {
				continue
			}
			runFunc(v.Command, v.Expected, v.Contains)
		default:
			t.Logf("No Unit Test Available for this platform")

		}
	}
}

func TestRunSafe(t *testing.T) {
	_, err := RunSafe(`whoami | grep Hello`)
	require.Error(t, err)
}

func TestRunSh(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	// try to run the echo command
	s, err := RunSh("echo", "test")
	require.Nil(t, err, "failed execution", err)
	require.Equal(t, "test"+newLineMarker, s, "output doesn't contain expected result", s)
}

func TestRunPS(t *testing.T) {
	if runtime.GOOS != "windows" {
		return
	}
	// run powershell command (runs in both ps1 and ps2)
	s, err := RunPS("get-host")
	require.Nil(t, err, "failed execution", err)
	require.Contains(t, s, "Microsoft.PowerShell", "failed to run powershell command get-host")
}

// runWithTimeout fails the test instead of hanging when cmdFunc does not return.
func runWithTimeout(t *testing.T, cmdFunc func() (string, error)) (string, error) {
	t.Helper()

	type result struct {
		out string
		err error
	}

	done := make(chan result, 1)

	go func() {
		out, err := cmdFunc()
		done <- result{out: out, err: err}
	}()

	select {
	case res := <-done:
		return res.out, res.err
	case <-time.After(10 * time.Second):
		t.Fatal("command did not return within timeout")

		return "", nil
	}
}

type pipeRunner struct {
	name string
	run  func(cmd string) (string, error)
	// shell reports whether the runner interprets shell syntax.
	shell bool
}

type pipeCase struct {
	name    string
	cmd     string
	shell   bool
	wantOut string
	wantErr bool
}

func pipeRunners() []pipeRunner {
	return []pipeRunner{
		{name: "Run", run: Run, shell: true},
		{name: "RunSh", run: func(cmd string) (string, error) { return RunSh(cmd) }, shell: true},
		{name: "RunSafe", run: func(cmd string) (string, error) { return RunSafe(strings.Fields(cmd)...) }, shell: false},
	}
}

func pipeCases() []pipeCase {
	zeros := strings.Repeat("\x00", 200000)

	return []pipeCase{
		{name: "reads stdin", cmd: "cat", shell: false, wantOut: "", wantErr: false},
		{name: "large stdout", cmd: "head -c 200000 /dev/zero", shell: false, wantOut: zeros, wantErr: false},
		{
			name: "large stderr then success", cmd: "head -c 200000 /dev/zero 1>&2; echo ok",
			shell: true, wantOut: "ok\n", wantErr: false,
		},
		{
			name: "large stdout and stderr", cmd: "head -c 200000 /dev/zero 1>&2; head -c 200000 /dev/zero",
			shell: true, wantOut: zeros, wantErr: false,
		},
		{
			name: "large stderr then failure", cmd: "head -c 200000 /dev/zero 1>&2; exit 3",
			shell: true, wantOut: zeros, wantErr: true,
		},
	}
}

func TestRunPipes(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("POSIX shell commands")
	}

	for _, runner := range pipeRunners() {
		for _, testCase := range pipeCases() {
			if testCase.shell && !runner.shell {
				continue
			}

			t.Run(runner.name+"/"+testCase.name, func(t *testing.T) {
				t.Parallel()

				out, err := runWithTimeout(t, func() (string, error) { return runner.run(testCase.cmd) })
				if testCase.wantErr {
					require.Error(t, err)
					// The error carries the command's stderr text.
					require.Contains(t, err.Error(), strings.Repeat("\x00", 1000))
					require.Contains(t, err.Error(), "exit error")
				} else {
					require.NoError(t, err)
				}

				require.Equal(t, testCase.wantOut, out)
			})
		}
	}
}

func TestRunShStdoutOnlyOnSuccess(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("POSIX shell commands")
	}

	out, err := runWithTimeout(t, func() (string, error) { return RunSh("echo out; echo err 1>&2") })
	require.NoError(t, err)
	require.Equal(t, "out\n", out)
}

func TestRunShFailureCarriesStderr(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("POSIX shell commands")
	}

	out, err := runWithTimeout(t, func() (string, error) { return RunSh("echo out; echo boom 1>&2; exit 2") })
	require.Error(t, err)
	require.Equal(t, "boom\n", out)
	require.Contains(t, err.Error(), "boom")
	require.Contains(t, err.Error(), "exit error")
	require.Contains(t, err.Error(), "exit status 2")

	exitErr, ok := errors.AsType[*exec.ExitError](err)
	require.True(t, ok, "error should wrap *exec.ExitError")
	require.Equal(t, 2, exitErr.ExitCode())
}
