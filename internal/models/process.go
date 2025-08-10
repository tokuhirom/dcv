package models

import (
	"strconv"
	"strings"
)

// Process represents a single process in the container
type Process struct {
	UID   string
	PID   string
	PPID  string
	C     string
	STIME string
	TTY   string
	TIME  string
	CMD   string

	// CPU and Memory stats from docker stats
	CPUPerc  float64
	MemUsage string
	MemPerc  float64
}

// ParsePercentage parses percentage strings like "1.23%" to float64
func ParsePercentage(s string) float64 {
	s = strings.TrimSuffix(s, "%")
	val, _ := strconv.ParseFloat(s, 64)
	return val
}
