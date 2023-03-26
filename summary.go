package main

import (
	"fmt"
	"strings"
)

type VerdictCounter map[Status]int

func (c VerdictCounter) Add(status Status) {
	if old, has := c[status]; has {
		c[status] = old + 1
	} else {
		c[status] = 1
	}
}

type Summarizer struct {
	names    []string
	counters []VerdictCounter
	msgs     []string
}

func NewSummarizer(execs []Exec) *Summarizer {
	names := make([]string, len(execs))
	for i := range execs {
		names[i] = execs[i].Name
	}

	counters := make([]VerdictCounter, len(execs))
	for i := range counters {
		counters[i] = make(VerdictCounter)
	}

	msgs := []string{}

	return &Summarizer{names: names, counters: counters, msgs: msgs}
}

func (s *Summarizer) Ingest(event Event) {
	s.counters[event.execIndex].Add(event.status)

	for _, msg := range event.msgs {
		msg := fmt.Sprintf("%s: Test %d: %s", s.names[event.execIndex], event.testIndex, msg)
		s.msgs = append(s.msgs, msg)
	}
}

func (s *Summarizer) Summary(isRun bool) string {
	var summary strings.Builder

	maxNameLen := 0
	for _, name := range s.names {
		if len(name) > maxNameLen {
			maxNameLen = len(name)
		}
	}

	for i, name := range s.names {
		summary.WriteString(fmt.Sprintf("%d. %-*s", i+1, maxNameLen, name))
		if isRun {
			summary.WriteString(fmt.Sprintf("   OK %2d", s.counters[i][OK]))
			summary.WriteString(fmt.Sprintf(" | WA %2d", s.counters[i][WA]))
		} else {
			summary.WriteString(fmt.Sprintf("   SAME %2d", s.counters[i][SAME_OUT]))
			summary.WriteString(fmt.Sprintf(" | DIFF %2d", s.counters[i][MISMATCH]))
		}
		summary.WriteString(fmt.Sprintf(" | TLE %2d", s.counters[i][TLE]))
		if failed := s.counters[i][FAILED]; failed > 0 {
			summary.WriteString(fmt.Sprintf(" (%d failed)", failed))
		}
		summary.WriteString("\n")
	}

	if len(s.msgs) > 0 {
		summary.WriteString("\n")
		for _, msg := range s.msgs {
			summary.WriteString(fmt.Sprintf("%s\n", msg))
		}
	}

	return summary.String()
}
