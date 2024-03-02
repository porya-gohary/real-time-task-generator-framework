package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	taskGenertor "task-generator/lib"
)

// Config represents the structure of the YAML configuration file
type Config struct {
	Path               string  `yaml:"path"`
	PeriodDistribution string  `yaml:"period_distribution"`
	NumSets            int     `yaml:"num_sets"`
	Tasks              int     `yaml:"tasks"`
	Utilization        float64 `yaml:"utilization"`
	ExecVariation      float64 `yaml:"exec_variation"`
	Jitter             float64 `yaml:"jitter"`
	ConstantJitter     bool    `yaml:"constant_jitter"`
	IsPreemptive       bool    `yaml:"is_preemptive"`
	MaxJobs            int     `yaml:"max_jobs"`
}

func main() {
	//	first we need to read the config file
	var configFile string
	flag.StringVar(&configFile, "config", "config.yaml", "path to the YAML config file")
	flag.Parse()

	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		os.Exit(1)
	}

	var config Config
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		fmt.Printf("Error parsing config file: %v\n", err)
		os.Exit(1)
	}

	//	then we need to create the task sets
	taskGenertor.CreateTaskSetsParallel(config.Path, config.NumSets, config.Tasks, config.Utilization, config.PeriodDistribution, config.ExecVariation, config.Jitter, config.IsPreemptive, config.ConstantJitter, config.MaxJobs)

}
