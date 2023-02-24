package main

import (
	_ "embed"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	tm "github.com/buger/goterm"
)

type Status int

// configTemplateBytes contains the default configuration file, ready to be
// used as a template for new files.
//
//go:embed config.yml
var configTemplateBytes []byte

const (
	NONE Status = iota
	STARTING
	FAILED
	WA
	TLE
	OK
)

type Event struct {
	execIndex int
	testIndex int
	status    Status
	msgs      []string
}

func (e Event) Exec(index int) Event {
	e.execIndex = index
	return e
}

func (e Event) Test(index int) Event {
	e.testIndex = index
	return e
}

func (e Event) Status(status Status) Event {
	e.status = status
	return e
}

func (e Event) Msg(msg string) Event {
	e.msgs = append(e.msgs, msg)
	return e
}

func runCli(configFilename string) {
	config, err := ParseConfig(configFilename)
	if err != nil {
		log.Fatal(err)
	}

	runChecks(config)
}

// initConfigFile creates a new configuration file based on the template.
func initConfigFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("file %v already exists", path)
	}

	if err := os.WriteFile(path, configTemplateBytes, 0666); err != nil {
		return fmt.Errorf("failed to write the config file: %w", err)
	}

	return nil
}

func runChecks(config Config) {
	tests, err := getTests(config)
	if err != nil {
		log.Fatalf("failed to read tests: %v\n", err)
	}
	tm.Printf("Using %d tests.\n", len(tests))

	maxWorkers := decideMaxWorkers(config)
	tm.Printf("Using %d max workers.\n", maxWorkers)
	tm.Println()
	tm.Flush()

	events := make(chan Event)
	monitorDone := make(chan struct{}, 1)
	go func() {
		summarizer := NewSummarizer(config.Execs)
		ui := NewUI(config.Execs, tests)
		ui.InitDisplay()
		for event := range events {
			summarizer.Ingest(event)
			ui.Display(event)
		}
		tm.Println()
		tm.Flush()
		// goterm does not seem to handle big strings? Use fmt instead.
		fmt.Print(summarizer.Summary())

		monitorDone <- struct{}{}
	}()

	var wg sync.WaitGroup
	limiter := make(chan struct{}, maxWorkers)
	for i := range config.Execs {
		wg.Add(1)
		go func(index int, exe Exec) {
			limiter <- struct{}{}
			defer wg.Done()
			testExec(exe, tests, Event{execIndex: index}, events)
			<-limiter
		}(i, config.Execs[i])
	}
	wg.Wait()
	close(events)

	<-monitorDone
}

func getTests(config Config) ([]TestCase, error) {
	files, err := os.ReadDir(config.InputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open input dir: %w", err)
	}

	tests := make([]TestCase, 0, len(files))
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".in" {
			continue
		}

		testName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		inPath := filepath.Join(config.InputDir, file.Name())
		refPath := filepath.Join(config.RefDir, fmt.Sprintf("%s.ref", testName))

		if _, err := os.Stat(refPath); errors.Is(err, os.ErrNotExist) {
			tm.Printf("Warning: could not find %s. Skipping...\n", refPath)
			tm.Flush()
			continue
		}

		tests = append(tests, TestCase{
			name:          testName,
			inputFilename: inPath,
			refFilename:   refPath,
		})
	}

	sort.Sort(ByTestNumber(tests))
	return tests, nil
}

func decideMaxWorkers(config Config) int {
	// Don't allow the checker to use all cores - trust me.
	workers := runtime.NumCPU() - 2
	if config.MaxWorkers < workers {
		workers = config.MaxWorkers
	}
	if workers < 1 {
		workers = 1
	}
	return workers
}

func testExec(exe Exec, tests []TestCase, baseEvent Event, events chan<- Event) {
	for i, test := range tests {
		runTest(exe, test, baseEvent.Test(i), events)
	}
}

func runTest(exe Exec, test TestCase, baseEvent Event, events chan<- Event) {
	events <- baseEvent.Status(STARTING)

	inputBytes, err := ioutil.ReadFile(test.inputFilename)
	if err != nil {
		events <- baseEvent.Status(FAILED).Msg(fmt.Sprintf("failed to read input: %v", err))
		return
	}
	input := string(inputBytes)

	refBytes, err := ioutil.ReadFile(test.refFilename)
	if err != nil {
		events <- baseEvent.Status(FAILED).Msg(fmt.Sprintf("failed to read ref: %v", err))
		return
	}

	outputBytes, err := runProgram(input, exe.Cmd, exe.Timeout)
	if err != nil {
		if isTimeout(err) {
			events <- baseEvent.Status(TLE).Msg(fmt.Sprintf("timed out: %v", err))
		} else {
			events <- baseEvent.Status(FAILED).Msg(fmt.Sprintf("failed to run: %v", err))
		}
		return
	}

	if compatibleOutputs(outputBytes, refBytes) {
		events <- baseEvent.Status(OK)
	} else {
		events <- baseEvent.Status(WA)
	}
}

func runProgram(input string, cmdArgs []string, timeout float64) ([]byte, error) {
	args := append([]string{fmt.Sprintf("%f", timeout)}, cmdArgs...)
	cmd := exec.Command("timeout", args...)
	cmd.Stdin = strings.NewReader(input)
	return cmd.Output()
}

func isTimeout(err error) bool {
	exitError, ok := err.(*exec.ExitError)
	return ok && exitError.ExitCode() == 124
}

func compatibleOutputs(out []byte, ref []byte) bool {
	linesOut := strings.Split(strings.TrimSpace(string(out)), "\n")
	linesRef := strings.Split(strings.TrimSpace(string(ref)), "\n")
	if len(linesOut) != len(linesRef) {
		return false
	}

	for i := range linesOut {
		if strings.TrimSpace(linesOut[i]) != strings.TrimSpace(linesRef[i]) {
			return false
		}
	}
	return true
}
