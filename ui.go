package main

import (
	"fmt"

	tm "github.com/buger/goterm"
)

type UI struct {
	names      []string
	nameOffset int
	tests      int
}

func NewUI(execs []Exec, tests []TestCase) *UI {
	names := make([]string, len(execs))
	nameOffset := 0
	for i, exec := range execs {
		names[i] = exec.Name

		nameLen := len(fmt.Sprintf("%d. %s", i+1, names[i]))
		if nameLen > nameOffset {
			nameOffset = nameLen
		}
	}

	nameOffset += 3 // Leave some space between the names and the results.

	return &UI{names: names, nameOffset: nameOffset, tests: len(tests)}
}

func (u *UI) InitDisplay() {
	for i, name := range u.names {
		tm.Printf("%d. %s\n", i+1, name)
	}

	for i := range u.names {
		for j := 0; j < u.tests; j += 1 {
			u.renderTestCell(i, j, NONE)
		}
	}
	tm.Flush()
}

func (u *UI) Display(event Event) {
	if event.status == NONE {
		return
	}

	u.renderTestCell(event.execIndex, event.testIndex, event.status)
	tm.Flush()
}

func (u *UI) renderTestCell(execIndex, testIndex int, status Status) {
	symbol := map[Status]string{
		NONE:     "_",
		STARTING: "?",
		FAILED:   "⚠",
		WA:       "✖",
		TLE:      "⏱",
		OK:       "✓",
	}[status]

	color := map[Status]int{
		NONE:     tm.WHITE,
		STARTING: tm.YELLOW,
		FAILED:   tm.MAGENTA,
		WA:       tm.RED,
		TLE:      tm.BLUE,
		OK:       tm.GREEN,
	}[status]

	rowOffset := len(u.names) - execIndex
	colOffset := testIndex*2 + testIndex/5 + u.nameOffset
	tm.MoveCursorUp(rowOffset)
	tm.MoveCursorForward(colOffset)
	tm.Print(tm.Color(symbol, color))
	tm.MoveCursorBackward(colOffset + 1)
	tm.MoveCursorDown(rowOffset)
}
