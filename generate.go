package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	taskGenertor "task-generator/lib"
	"task-generator/lib/common"
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
	RunParallel        bool    `yaml:"run_parallel"`
	Verbose            int     `yaml:"verbose"`
}

var logger *common.VerboseLogger

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

	//Set verbose level
	if config.Verbose <= 4 && config.Verbose >= 0 {
		// Create a verbose logger with a verbosity level of Info.
		logger = common.NewVerboseLogger("", config.Verbose)
	} else {
		fmt.Println("Error: Invalid verbose level")
		os.Exit(1)
	}

	// Print a warning that the automotive method is not consider number of tasks
	if config.PeriodDistribution == "automotive" {
		logger.LogWarning("The automotive method does not consider the number of tasks")
	}

	//	then we need to create the task sets
	// 	we can run the task generation in parallel if the config file specifies it
	if config.RunParallel {
		taskGenertor.CreateTaskSetsParallel(config.Path, config.NumSets, config.Tasks,
			config.Utilization, config.PeriodDistribution, config.ExecVariation, config.Jitter, config.IsPreemptive,
			config.ConstantJitter, config.MaxJobs, logger)
	} else {
		taskGenertor.CreateTaskSets(config.Path, config.NumSets, config.Tasks,
			config.Utilization, config.PeriodDistribution, config.ExecVariation, config.Jitter, config.IsPreemptive,
			config.ConstantJitter, config.MaxJobs, logger)
	}

}
