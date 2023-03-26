package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"sync"

	tm "github.com/buger/goterm"
)

func runSame(configFilename string, rounds int, execName1, execName2 string) {
	config, err := ParseConfig(configFilename)
	if err != nil {
		log.Fatal(err)
	}

	exec1, err := findExec(config.Execs, execName1)
	if err != nil {
		log.Fatal(err)
	}
	exec2, err := findExec(config.Execs, execName2)
	if err != nil {
		log.Fatal(err)
	}
	config.Execs = []Exec{exec1, exec2}

	runSameTests(config, rounds, exec1, exec2)
}

func findExec(execs []Exec, execName string) (Exec, error) {
	var resultExec Exec
	matches := 0

	for _, exec := range execs {
		if exec.Name == execName {
			resultExec = exec
			matches += 1
		}
	}

	if matches > 1 {
		return Exec{}, fmt.Errorf("multiple executables (%v) found for '%v'", matches, execName)
	}
	if matches < 1 {
		return Exec{}, fmt.Errorf("executable '%v' not found", execName)
	}
	return resultExec, nil
}

func runSameTests(config Config, rounds int, exec1, exec2 Exec) {
	if err := os.MkdirAll(config.TestGenDir, 0777); err != nil {
		log.Fatal(fmt.Errorf("failed to create directory for inputs: %w", err))
	}

	maxWorkers := decideMaxWorkers(config)
	displayStartupInfo(config, rounds, maxWorkers)

	events := make(chan Event)
	monitorDone := make(chan struct{}, 1)
	go func() {
		// Generate fictive tests for the UI. This should be fixed in the future.
		fakeTestcases := make([]TestCase, rounds)
		for i := 0; i < rounds; i += 1 {
			fakeTestcases[i] = TestCase{name: fmt.Sprintf("%v", i)}
		}

		summarizer := NewSummarizer(config.Execs)
		ui := NewUI(config.Execs, fakeTestcases)
		ui.InitDisplay()
		for event := range events {
			summarizer.Ingest(event)
			ui.Display(event)
		}
		tm.Println()
		tm.Flush()
		// goterm does not seem to handle big strings? Use fmt instead.
		fmt.Print(summarizer.Summary(false))

		monitorDone <- struct{}{}
	}()

	var wg sync.WaitGroup
	limiter := make(chan struct{}, maxWorkers)
	for i := 0; i < rounds; i += 1 {
		wg.Add(1)
		go func(index int) {
			limiter <- struct{}{}
			defer wg.Done()
			runSameTest(index, config, exec1, exec2, Event{testIndex: index}, events)
			<-limiter
		}(i)
	}
	wg.Wait()
	close(events)

	<-monitorDone
}

func runSameTest(index int, config Config, exec1, exec2 Exec, baseEvent Event, events chan<- Event) {
	events <- baseEvent.Exec(0).Status(GENERATING)
	events <- baseEvent.Exec(1).Status(GENERATING)

	inputFilename := path.Join(config.TestGenDir, strconv.Itoa(index))
	if err := generateInput(config.TestGenCmd, inputFilename); err != nil {
		msg := fmt.Sprintf("failed to generate input: %v", err)
		events <- baseEvent.Exec(0).Status(FAILED).Msg(msg)
		events <- baseEvent.Exec(1).Status(FAILED)
		return
	}

	inputFile, err := os.Open(inputFilename)
	if err != nil {
		msg := fmt.Sprintf("failed to open input file: %v", err)
		events <- baseEvent.Exec(0).Status(FAILED).Msg(msg)
		events <- baseEvent.Exec(1).Status(FAILED)
		return
	}
	defer inputFile.Close()

	events <- baseEvent.Exec(0).Status(STARTING)
	out1, err1 := runProgram(inputFile, exec1.Cmd, exec1.Timeout)
	if err1 != nil {
		events <- baseEvent.Exec(0).Status(determineStatus(err1))
	}

	// The first run "consumed" our file.
	inputFile.Seek(0, io.SeekStart)

	events <- baseEvent.Exec(1).Status(STARTING)
	out2, err2 := runProgram(inputFile, exec2.Cmd, exec2.Timeout)
	if err2 != nil {
		events <- baseEvent.Exec(1).Status(determineStatus(err2))
	}

	if err1 != nil || err2 != nil {
		// This is a hack to ensure both execs have the same "SAME" count.
		if err1 == nil {
			events <- baseEvent.Exec(0).Status(OK)
		}
		if err2 == nil {
			events <- baseEvent.Exec(1).Status(OK)
		}
		return
	}

	if !compatibleOutputs(out1, out2) {
		events <- baseEvent.Exec(0).Status(MISMATCH)
		events <- baseEvent.Exec(1).Status(MISMATCH)
		return
	}

	events <- baseEvent.Exec(0).Status(SAME_OUT)
	events <- baseEvent.Exec(1).Status(SAME_OUT)

	// Remove the input file, since it's not interesting anymore.
	if err := os.Remove(inputFilename); err != nil {
		msg := fmt.Sprintf("failed to remove %s: %v", inputFilename, err)
		events <- baseEvent.Exec(0).Msg(msg)
	}
}

func generateInput(cmdArgs []string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdout = file
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}

func determineStatus(err error) Status {
	if err == nil {
		return NONE
	}
	if isTimeout(err) {
		return TLE
	}
	return FAILED
}

func displayStartupInfo(config Config, rounds int, maxWorkers int) {
	tm.Printf("Running %v rounds with %v workers.\n", rounds, maxWorkers)
	for i, exec := range config.Execs {
		tm.Printf("%d. %s\t(%.2fs timeout)\n", i+1, exec.Name, exec.Timeout)
	}
	tm.Println()
	tm.Flush()
}
