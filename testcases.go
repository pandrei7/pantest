package main

import (
	"regexp"
	"strconv"
)

type TestCase struct {
	name          string
	inputFilename string
	refFilename   string
}

// ByTestNumber is a class which helps order tests cases in increasing order
// of the numbers embeded in their names.
type ByTestNumber []TestCase

func (t ByTestNumber) Len() int { return len(t) }

func (t ByTestNumber) Swap(i, j int) { t[i], t[j] = t[j], t[i] }

func (t ByTestNumber) Less(i, j int) bool {
	numI := t[i].extractNum()
	numJ := t[j].extractNum()
	if numI != numJ {
		return numI < numJ
	}
	return t[i].name < t[j].name
}

func (t TestCase) extractNum() int {
	numExtractor := regexp.MustCompile("[\\d]+")
	if seq := numExtractor.Find([]byte(t.name)); seq == nil {
		return -1
	} else if num, err := strconv.Atoi(string(seq)); err != nil {
		return -1
	} else {
		return num
	}
}
