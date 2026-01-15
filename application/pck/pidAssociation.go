package pck

import (
	"fmt"

	"github.com/dfialho05/GoMonitor/application/pck/common"
)

// GetProcessAssociation collects and associates CPU and RAM statistics for each process
// This function is a simplified wrapper that uses common functions to avoid duplication
//
// Returns:
//   - slice of ProcessInfo with all active processes in the system
//   - error if unable to get the data
func GetProcessAssociation() ([]common.ProcessInfo, error) {
	// Delegates all logic to the common function that centralizes data collection
	return common.CollectAllProcessInfo()
}

// GetProcessAssociationSorted collects and returns processes sorted by CPU usage
// Processes are sorted in descending order (highest usage first)
//
// Returns:
//   - slice of ProcessInfo sorted by CPU usage (descending)
//   - error if unable to get the data
func GetProcessAssociationSorted() ([]common.ProcessInfo, error) {
	// 1. Get all processes with their statistics
	processes, err := common.CollectAllProcessInfo()
	if err != nil {
		return nil, fmt.Errorf("error getting processes: %w", err)
	}

	// 2. Sort processes by CPU usage (highest to lowest)
	// Uses the common sorting function that implements selection sort
	common.SortProcessesByField(processes, "cpu", true)

	return processes, nil
}

// GetProcessAssociationByPID searches and returns statistics for a specific process
// This function is optimized to search only one process by its PID
//
// Parameters:
//   - targetPID: process ID to search for
//
// Returns:
//   - pointer to ProcessInfo with process statistics
//   - error if the process is not found or not accessible
func GetProcessAssociationByPID(targetPID int32) (*common.ProcessInfo, error) {
	// 1. Get total system memory (needed to calculate percentages)
	totalSystemMem, err := common.GetSystemMemoryTotal()
	if err != nil {
		return nil, err
	}

	// 2. Create a process object with the provided PID
	p, err := common.GetProcessByPID(targetPID)
	if err != nil {
		return nil, err
	}

	// 3. Get all process information using the common function
	info, err := common.GetProcessInfo(p, totalSystemMem)
	if err != nil {
		return nil, fmt.Errorf("error getting information for process PID %d: %w", targetPID, err)
	}

	return info, nil
}

// MonitorProcessContinuous continuously monitors a specific process
// Prints statistics at each specified interval until the process terminates or Ctrl+C
//
// Parameters:
//   - targetPID: process ID to monitor
//   - intervalSeconds: interval between updates in seconds
//
// Returns:
//   - error if the process cannot be monitored
func MonitorProcessContinuous(targetPID int32, intervalSeconds int) error {
	// Delegates to the common function that implements all monitoring logic
	return common.MonitorProcessContinuously(targetPID, intervalSeconds)
}

// PrintTopProcesses prints the N processes with highest CPU usage
// This function provides a formatted view of the most active processes
//
// Parameters:
//   - n: number of processes to show (top N)
//
// Returns:
//   - error if unable to get process data
func PrintTopProcesses(n int) error {
	// 1. Get processes sorted by CPU usage
	processes, err := GetProcessAssociationSorted()
	if err != nil {
		return fmt.Errorf("error getting sorted processes: %w", err)
	}

	// 2. Use the common function to print the formatted table
	title := fmt.Sprintf("Top %d Processes (sorted by CPU usage)", n)
	common.PrintProcessTable(processes, n, title)

	return nil
}
