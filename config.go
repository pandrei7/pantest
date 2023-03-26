package main

import (
	"fmt"
	"os"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v2"
)

type Config struct {
	MaxWorkers int      `yaml:"maxWorkers" default:"10"`
	InputDir   string   `yaml:"inputDir"`
	RefDir     string   `yaml:"refDir"`
	TestGenCmd []string `yaml:"testGenCmd"`
	TestGenDir string   `yaml:"testGenDir"`
	Execs      []Exec   `yaml:"execs"`
}

type Exec struct {
	Name    string   `yaml:"name"`
	Cmd     []string `yaml:"cmd"`
	Timeout float64  `yaml:"time" default:"1"`
	Ignore  bool     `yaml:"ignore" default:"false"`
}

func ParseConfig(filename string) (Config, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	defaults.Set(&config)

	if err := yaml.Unmarshal([]byte(bytes), &config); err != nil {
		return Config{}, fmt.Errorf("failed to parse config file: %w", err)
	}

	for i := 0; i < len(config.Execs); i += 1 {
		defaults.Set(&config.Execs[i])
		if config.Execs[i].Ignore {
			config.Execs = append(config.Execs[:i], config.Execs[i+1:]...)
			i -= 1
		}
	}

	for _, exec := range config.Execs {
		if len(exec.Cmd) <= 0 {
			msg := "the cmd of an exec cannot be empty"
			if len(exec.Name) > 0 {
				msg = fmt.Sprintf("the cmd of %s cannot be empty", exec.Name)
			}
			return Config{}, fmt.Errorf(msg)
		}
	}

	// Add "default" names for execs which don't have one specified.
	for i, exec := range config.Execs {
		if exec.Name == "" {
			config.Execs[i].Name = exec.Cmd[0]
		}
	}

	return config, nil
}
