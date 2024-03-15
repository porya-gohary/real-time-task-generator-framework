package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"task-generator/lib"
	"task-generator/lib/common"
)

// Config represents the structure of the YAML configuration file
type Config struct {
	Path               string  `yaml:"path"`
	OutputFormat       string  `yaml:"output_format"`
	NumCores           int     `yaml:"number_of_cores"`
	UtilDistribution   string  `yaml:"utilization_distribution"`
	PeriodDistribution string  `yaml:"period_distribution"`
	PeriodRange        []int   `yaml:"period_range"`
	Periods            []int   `yaml:"periods"`
	NumSets            int     `yaml:"num_sets"`
	Tasks              int     `yaml:"tasks"`
	Utilization        float64 `yaml:"utilization"`
	ExecVariation      float64 `yaml:"exec_variation"`
	Jitter             float64 `yaml:"jitter"`
	ConstantJitter     bool    `yaml:"constant_jitter"`
	MaxJobs            int     `yaml:"max_jobs"`
	GenerateDAGs       bool    `yaml:"generate_dags"`
	MakeDotFile        bool    `yaml:"generate_dot"`
	DAGType            string  `yaml:"dag_type"`
	ForkProb           float64 `yaml:"fork_probability"`
	EdgeProb           float64 `yaml:"edge_probability"`
	MaxBranch          int     `yaml:"max_branches"`
	MaxVertices        int     `yaml:"max_vertices"`
	NumRoots           int     `yaml:"num_roots"`
	MaxDepth           int     `yaml:"max_depth"`
	GenerateJobs       bool    `yaml:"generate_job_sets"`
	PriorityAssignment string  `yaml:"priority_assignment"`
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

	if config.OutputFormat != "csv" && config.OutputFormat != "yaml" {
		logger.LogFatal("Invalid output format")
	}

	// Print a warning that the automotive method is not consider number of tasks
	if config.PeriodDistribution == "automotive" && config.UtilDistribution == "automotive" {
		logger.LogWarning("The automotive method does not consider the number of tasks")
	}
	// Print a fatal error if the period distribution is automotive and the utilization distribution is not automotive
	if config.UtilDistribution == "automotive" && config.PeriodDistribution != "automotive" {
		logger.LogFatal("The utilization distribution is automotive but the period distribution is not automotive")
	}

	//	then we need to create the task sets
	// 	we can run the task generation in parallel if the config file specifies it
	if config.RunParallel {
		lib.CreateTaskSetsParallel(config.Path, config.NumCores, config.NumSets, config.Tasks,
			config.Utilization, config.UtilDistribution, config.PeriodDistribution, config.PeriodRange, config.Periods,
			config.ExecVariation, config.Jitter, config.ConstantJitter, config.MaxJobs,
			config.OutputFormat, logger)
	} else {
		lib.CreateTaskSets(config.Path, config.NumCores, config.NumSets, config.Tasks,
			config.Utilization, config.UtilDistribution, config.PeriodDistribution, config.PeriodRange, config.Periods,
			config.ExecVariation, config.Jitter, config.ConstantJitter, config.MaxJobs,
			config.OutputFormat, logger)
	}

	// then we need to generate the DAGs
	if config.GenerateDAGs {
		if config.RunParallel {
			if config.DAGType == "fork-join" {
				lib.GenerateDAGSetsParallel(config.Path, config.ForkProb, config.EdgeProb, config.MaxBranch,
					config.MaxVertices, config.MaxDepth, config.MakeDotFile, config.OutputFormat)
			} else if config.DAGType == "random" {
				lib.GenerateRandomDAGsParallel(config.Path, config.NumRoots, config.MaxBranch, config.MaxDepth,
					config.MakeDotFile, config.OutputFormat)
			} else {
				logger.LogFatal("Invalid DAG type")
			}
		} else {
			if config.DAGType == "fork-join" {
				lib.GenerateDAGSets(config.Path, config.ForkProb, config.EdgeProb, config.MaxBranch, config.MaxVertices,
					config.MaxDepth, config.MakeDotFile, config.OutputFormat)
			} else if config.DAGType == "random" {
				lib.GenerateRandomDAGs(config.Path, config.NumRoots, config.MaxBranch, config.MaxDepth,
					config.MakeDotFile, config.OutputFormat)
			} else {
				logger.LogFatal("Invalid DAG type")
			}
		}
	}

	//	then we need to generate the job sets
	if config.GenerateJobs {
		// first change the priority assignment to an integer
		var priorityAssignment int
		switch config.PriorityAssignment {
		case "RM":
			priorityAssignment = lib.RM
		case "DM":
			priorityAssignment = lib.DM
		case "EDF":
			priorityAssignment = lib.EDF
		}
		if config.RunParallel {
			lib.GenerateJobSetsParallel(config.Path, priorityAssignment)
		} else {
			lib.GenerateJobSets(config.Path, priorityAssignment)
		}
	}

}
