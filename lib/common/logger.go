package common

import (
	"log"
	"os"
)

const (
	VerboseLevelDebug   = 4
	VerboseLevelInfo    = 3
	VerboseLevelWarning = 2
	VerboseLevelError   = 1
	VerboseLevelNone    = 0
)

// VerboseLogger is a custom logger that logs messages with varying levels of verbosity.
type VerboseLogger struct {
	*log.Logger
	verboseLevel int
}

// NewVerboseLogger creates a new VerboseLogger with the given prefix and verbosity level.
func NewVerboseLogger(prefix string, verboseLevel int) *VerboseLogger {
	return &VerboseLogger{
		//Logger:       log.New(os.Stdout, prefix, log.Ldate|log.Ltime),
		Logger:       log.New(os.Stdout, prefix, 0),
		verboseLevel: verboseLevel,
	}
}

// LogDebug logs a message with debug verbosity level.
func (vl *VerboseLogger) LogDebug(message string) {
	if VerboseLevelDebug <= vl.verboseLevel {
		vl.Printf("[DEBUG] %s\n", message)
	}
}

// LogInfo logs a message with info verbosity level.
func (vl *VerboseLogger) LogInfo(message string) {
	if VerboseLevelInfo <= vl.verboseLevel {
		vl.Printf("[INFO] %s\n", message)
	}
}

// LogWarning logs a message with warning verbosity level.
func (vl *VerboseLogger) LogWarning(message string) {
	if VerboseLevelWarning <= vl.verboseLevel {
		vl.Printf("[WARNING] %s\n", message)
	}
}

// LogError logs a message with error verbosity level.
func (vl *VerboseLogger) LogError(message string) {
	if VerboseLevelError <= vl.verboseLevel {
		vl.Printf("[ERROR] %s\n", message)
	}
}

// LogFatal logs a message with error verbosity level and then exits the program with status 1.
func (vl *VerboseLogger) LogFatal(message string) {
	vl.LogError(message)
	os.Exit(1)
}

// GetVerboseLevel returns the verbosity level of the logger.
func (vl *VerboseLogger) GetVerboseLevel() int {
	return vl.verboseLevel
}
