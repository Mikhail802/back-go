// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crashmonitor_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"golang.org/x/telemetry"
	"golang.org/x/telemetry/internal/counter"
	"golang.org/x/telemetry/internal/crashmonitor"
	"golang.org/x/telemetry/internal/testenv"
)

func TestMain(m *testing.M) {
	entry := os.Getenv("CRASHMONITOR_TEST_ENTRYPOINT")
	switch entry {
	case "via-stderr.panic", "via-stderr.trap":
		// This mode bypasses Start and debug.SetCrashOutput;
		// the crash is printed to stderr.
		debug.SetTraceback("system")
		crashmonitor.WriteSentinel(os.Stderr)

		if entry == "via-stderr.panic" {
			childPanic() // this line is "TestMain:+10"
		} else {
			childTrap() // this line is "TestMain:+12"
		}
		panic("unreachable")

	case "start.panic", "start.trap", "start.exit":
		// These modes uses Start and debug.SetCrashOutput.
		// We stub the actual telemetry by instead writing to a file.
		crashmonitor.SetIncrementCounter(func(name string) {
			os.WriteFile(os.Getenv("CRASHMONITOR_TELEMETRY_FILE"), []byte(name), 0666)
		})
		crashmonitor.SetChildExitHook(func() {
			os.WriteFile(os.Getenv("CRASHMONITOR_TELEMETRY_EXIT_FILE"), nil, 0666)
		})
		telemetry.Start(telemetry.Config{
			ReportCrashes: true,
			TelemetryDir:  os.Getenv("CRASHMONITOR_TELEMETRY_DIR"),
		})
		switch entry {
		case "start.panic":
			go func() {
				childPanic() // this line is "TestMain.func3:+1"
			}()
			select {} // deadlocks when reached
		case "start.trap":
			go func() {
				childTrap() // this line is "TestMain.func4:+1"
			}()
			select {} // deadlocks when reached
		case "start.exit":
			os.Exit(42)
		}

	default:
		os.Exit(m.Run()) // run tests as normal
	}
}

func childPanic() {
	fmt.Println("hello")
	grandchildPanic() // this line is "childPanic:+2"
}

func grandchildPanic() {
	panic("oops") // this line is "grandchildPanic:=79" (the call from child is inlined)
}

var sinkPtr *int

func childTrap() {
	fmt.Println("hello")
	grandchildTrap(sinkPtr) // this line is "childTrap:+2"
}

func grandchildTrap(i *int) {
	*i = 42 // this line is "grandchildTrap:=90" (the call from childTrap is inlined)
}

// TestViaStderr is an internal test that asserts that the telemetry
// stack generated by the panic in grandchild is correct. It uses
// stderr, and does not rely on [start.Start] or [debug.SetCrashOutput].
func TestViaStderr(t *testing.T) {
	// Standard panic.
	t.Run("panic", func(t *testing.T) {
		_, _, stderr := runSelf(t, "via-stderr.panic")
		got, err := crashmonitor.TelemetryCounterName(stderr)
		if err != nil {
			t.Fatal(err)
		}
		got = sanitize(counter.DecodeStack(got))
		want := "crash/crash\n" +
		"runtime.gopanic:--\n" +
		"golang.org/x/telemetry/internal/crashmonitor_test.grandchildPanic:=79\n" +
		"golang.org/x/telemetry/internal/crashmonitor_test.childPanic:+2\n" +
		"golang.org/x/telemetry/internal/crashmonitor_test.TestMain:+10\n" +
		"main.main:--\n" +
		"runtime.main:--\n" +
		"runtime.goexit:--"

		if !crashmonitor.Supported() { // !go1.23
			// Traceback excludes PCs for inlined frames. Before
			// go1.23 (https://go.dev/cl/571798 specifically),
			// passing the set of PCs in the traceback to
			// runtime.CallersFrames, would report only the
			// innermost inlined frame and none of the inline
			// "callers".
			//
			// Thus, here we must drop the caller of the inlined
			// frame.
			want = strings.ReplaceAll(want, "golang.org/x/telemetry/internal/crashmonitor_test.childPanic:+2\n", "")
		}

		if got != want {
			t.Errorf("got counter name <<%s>>, want <<%s>>", got, want)
		}
	})

	// Panic via trap.
	t.Run("trap", func(t *testing.T) {
		_, _, stderr := runSelf(t, "via-stderr.trap")
		got, err := crashmonitor.TelemetryCounterName(stderr)
		if err != nil {
			t.Fatal(err)
		}
		got = sanitize(counter.DecodeStack(got))
		want := "crash/crash\n" +
		"runtime.gopanic:--\n" +
		"runtime.panicmem:--\n" +
		"runtime.sigpanic:--\n" +
		"golang.org/x/telemetry/internal/crashmonitor_test.grandchildTrap:=90\n" +
		"golang.org/x/telemetry/internal/crashmonitor_test.childTrap:+2\n" +
		"golang.org/x/telemetry/internal/crashmonitor_test.TestMain:+12\n" +
		"main.main:--\n" +
		"runtime.main:--\n" +
		"runtime.goexit:--"

		if !crashmonitor.Supported() { // !go1.23
			// See above.
			want = strings.ReplaceAll(want, "runtime.sigpanic:--\n", "")
			want = strings.ReplaceAll(want, "golang.org/x/telemetry/internal/crashmonitor_test.childTrap:+2\n", "")
		}

		if got != want {
			t.Errorf("got counter name <<%s>>, want <<%s>>", got, want)
		}
	})
}

func waitForExitFile(t *testing.T, exitFile string) {
	deadline := time.Now().Add(10 * time.Second)
	for {
		_, err := os.ReadFile(exitFile)
		if err == nil {
			break // success
		}

		if !os.IsNotExist(err) {
			t.Fatalf("failed to read exit file: %v", err)
		}
		// The crashmonitor has not written the file yet.
		// Allow it more time.
		if time.Now().After(deadline) {
			t.Fatalf("crashmonitor failed to write file in a timely manner")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// TestStart is an integration test of the crashmonitor feature of [telemetry.Start].
// Requires go1.23+.
func TestStart(t *testing.T) {
	testenv.SkipIfUnsupportedPlatform(t)

	if !crashmonitor.Supported() {
		t.Skip("crashmonitor not supported")
	}

	// Assert that the crash monitor does nothing when the child
	// process merely exits.
	t.Run("exit", func(t *testing.T) {
		telemetryFile, exitFile, _ := runSelf(t, "start.exit")
		waitForExitFile(t, exitFile)
		data, err := os.ReadFile(telemetryFile)
		if err == nil {
			t.Fatalf("telemetry counter <<%s>> was unexpectedly incremented", data)
		}
	})

	// Assert that the crash monitor increments a telemetry counter of the
	// correct name when the child process panics.
	t.Run("panic", func(t *testing.T) {
		// Gather a stack trace from executing the panic statement above.
		telemetryFile, exitFile, _ := runSelf(t, "start.panic")
		waitForExitFile(t, exitFile)
		data, err := os.ReadFile(telemetryFile)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		got := sanitize(counter.DecodeStack(string(data)))
		want := "crash/crash\n" +
			"runtime.gopanic:--\n" +
			"golang.org/x/telemetry/internal/crashmonitor_test.grandchildPanic:=79\n" +
			"golang.org/x/telemetry/internal/crashmonitor_test.childPanic:+2\n" +
			"golang.org/x/telemetry/internal/crashmonitor_test.TestMain.func3:+1\n" +
			"runtime.goexit:--"
		if got != want {
			t.Errorf("got counter name <<%s>>, want <<%s>>", got, want)
		}
	})

	// Assert that the crash monitor increments a telemetry counter of the
	// correct name when the child process panics via trap.
	t.Run("trap", func(t *testing.T) {
		// Gather a stack trace from executing the panic statement above.
		telemetryFile, exitFile, _ := runSelf(t, "start.trap")
		waitForExitFile(t, exitFile)
		data, err := os.ReadFile(telemetryFile)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		got := sanitize(counter.DecodeStack(string(data)))
		want := "crash/crash\n" +
			"runtime.gopanic:--\n" +
			"runtime.panicmem:--\n" +
			"runtime.sigpanic:--\n" +
			"golang.org/x/telemetry/internal/crashmonitor_test.grandchildTrap:=90\n" +
			"golang.org/x/telemetry/internal/crashmonitor_test.childTrap:+2\n" +
			"golang.org/x/telemetry/internal/crashmonitor_test.TestMain.func4:+1\n" +
			"runtime.goexit:--"
		if got != want {
			t.Errorf("got counter name <<%s>>, want <<%s>>", got, want)
		}
	})
}

// runSelf fork+exec's this test executable using an alternate entry point.
// It returns the child's stderr, the name of the file
// to which any incremented counter name will be written, and
// the name of the file that will be written to when the crashmonitor
// exits.
func runSelf(t *testing.T, entrypoint string) (string, string, []byte) {
	testenv.MustHaveExec(t)

	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	tmpdir := t.TempDir()

	// Provide the names via the environment of the files the child is stubbed
	// to write to.

	// The exit file is created by the crashmonitor when it is finished.
	telemetryExitFile := filepath.Join(tmpdir, "exit.telemetry")

	// The telemetry file will contain the name of the incremented counter.
	telemetryFile := filepath.Join(tmpdir, "fake.telemetry")

	cmd := exec.Command(exe)
	cmd.Env = append(os.Environ(),
		"CRASHMONITOR_TEST_ENTRYPOINT="+entrypoint,
		"CRASHMONITOR_TELEMETRY_FILE="+telemetryFile,
		"CRASHMONITOR_TELEMETRY_EXIT_FILE="+telemetryExitFile,
		"CRASHMONITOR_TELEMETRY_DIR="+t.TempDir(),
	)
	cmd.Stderr = new(bytes.Buffer)
	cmd.Run() // failure is expected
	stderr := cmd.Stderr.(*bytes.Buffer).Bytes()
	if true { // debugging
		t.Logf("stderr: %s", stderr)
	}
	return telemetryFile, telemetryExitFile, stderr
}

// sanitize redacts the line numbers that we don't control from a counter name.
func sanitize(name string) string {
	lines := strings.Split(name, "\n")
	for i, line := range lines {
		if symbol, _, ok := strings.Cut(line, ":"); ok &&
			!strings.HasPrefix(line, "golang.org/x/telemetry/internal/crashmonitor") {
			lines[i] = symbol + ":--"
		}
	}
	return strings.Join(lines, "\n")
}
