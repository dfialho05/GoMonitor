package common

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

// ProcessInfo contains detailed information about a process
// This structure is used in all modules to represent process data
type ProcessInfo struct {
	PID           int32   // Process ID in the operating system
	Name          string  // Process/executable name
	CPUPercentage float64 // CPU usage percentage (0-100+, can exceed 100 on multi-core systems)
	RAMPercentage float32 // RAM usage percentage relative to total system memory
	RAMBytes      uint64  // RAM memory used in bytes (RSS - Resident Set Size)
}

// GetSystemMemoryTotal gets the total system memory once
// This function is optimized to be called only once and the result reused
// Returns: total memory in bytes and error (if any)
func GetSystemMemoryTotal() (uint64, error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return 0, fmt.Errorf("error getting system memory information: %w", err)
	}
	return vm.Total, nil
}

// GetProcessInfo collects complete information about a specific process
// This function centralizes the logic of obtaining process data, avoiding duplication
//
// Parameters:
//   - p: pointer to the process (gopsutil/process.Process)
//   - totalSystemMem: total system memory in bytes (to calculate percentages)
//
// Returns: filled ProcessInfo and error (if any)
func GetProcessInfo(p *process.Process, totalSystemMem uint64) (*ProcessInfo, error) {
	// 1. Get the process PID
	pid := p.Pid

	// 2. Get the process name
	// Note: Some system/kernel processes don't allow reading the name without root privileges
	name, err := p.Name()
	if err != nil {
		return nil, fmt.Errorf("error getting process name PID %d: %w", pid, err)
	}

	// 3. Get CPU usage percentage
	// CPUPercent() returns CPU utilization since the last call
	// If it's the first call, it may return 0.0 or a not very accurate value
	cpuPercent, err := p.CPUPercent()
	if err != nil {
		// If there's an error getting CPU, don't fail - just assume 0%
		cpuPercent = 0.0
	}

	// 4. Get memory usage information
	memInfo, err := p.MemoryInfo()
	if err != nil {
		return nil, fmt.Errorf("error getting memory information for process PID %d: %w", pid, err)
	}

	// 5. Calculate RAM usage percentage
	// RSS (Resident Set Size) is the amount of physical RAM actually used by the process
	// Does not include swap memory or shared memory that is not loaded
	rss := float64(memInfo.RSS)
	ramPercentage := float32((rss / float64(totalSystemMem)) * 100)

	// 6. Return structured process information
	return &ProcessInfo{
		PID:           pid,
		Name:          name,
		CPUPercentage: cpuPercent,
		RAMPercentage: ramPercentage,
		RAMBytes:      memInfo.RSS,
	}, nil
}

// GetAllProcesses gets the list of all active processes in the system
// This function is an optimized wrapper for process.Processes() with error handling
//
// Returns: slice of processes and error (if any)
func GetAllProcesses() ([]*process.Process, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("error getting system process list: %w", err)
	}
	return processes, nil
}

// GetProcessByPID creates a process object from a specific PID
// Checks if the process exists and is accessible
//
// Parameters:
//   - pid: Process ID to search for
//
// Returns: pointer to the process and error (if it doesn't exist or is not accessible)
func GetProcessByPID(pid int32) (*process.Process, error) {
	p, err := process.NewProcess(pid)
	if err != nil {
		return nil, fmt.Errorf("process with PID %d not found or inaccessible: %w", pid, err)
	}
	return p, nil
}

// CollectAllProcessInfo collects complete information from all active processes
// This is the main function that should be used by modules to get process data
// Centralizes all iteration and error handling logic
//
// Returns: slice of ProcessInfo with all valid processes and error (if any)
func CollectAllProcessInfo() ([]ProcessInfo, error) {
	// 1. Get total system memory (we do this only once)
	totalSystemMem, err := GetSystemMemoryTotal()
	if err != nil {
		return nil, err
	}

	// 2. Get all active processes
	allProcesses, err := GetAllProcesses()
	if err != nil {
		return nil, err
	}

	// 3. Pre-allocate the slice with estimated capacity to avoid reallocations
	processInfoList := make([]ProcessInfo, 0, len(allProcesses))

	// 4. Iterate through each process and collect its statistics
	for _, p := range allProcesses {
		// Try to get process information
		info, err := GetProcessInfo(p, totalSystemMem)
		if err != nil {
			// If we can't get information, skip this process
			// This is common for system processes or processes that have terminated in the meantime
			continue
		}

		// Add process information to the list
		processInfoList = append(processInfoList, *info)
	}

	return processInfoList, nil
}

// SortProcessesByField sorts a slice of ProcessInfo by a specific field
// Uses a simple sorting algorithm (selection sort) to avoid external dependencies
//
// Parameters:
//   - processes: slice of ProcessInfo to sort (is modified in-place)
//   - field: field to sort by ("cpu", "ram", "pid", "name")
//   - descending: true for descending order (largest -> smallest), false for ascending
func SortProcessesByField(processes []ProcessInfo, field string, descending bool) {
	n := len(processes)
	if n <= 1 {
		return // Nothing to sort
	}

	// Selection sort - simple and sufficient for most cases
	for i := 0; i < n-1; i++ {
		selectedIdx := i
		for j := i + 1; j < n; j++ {
			shouldSwap := false

			// Determine if we should swap based on field and order
			switch field {
			case "cpu":
				if descending {
					shouldSwap = processes[j].CPUPercentage > processes[selectedIdx].CPUPercentage
				} else {
					shouldSwap = processes[j].CPUPercentage < processes[selectedIdx].CPUPercentage
				}
			case "ram":
				if descending {
					shouldSwap = processes[j].RAMPercentage > processes[selectedIdx].RAMPercentage
				} else {
					shouldSwap = processes[j].RAMPercentage < processes[selectedIdx].RAMPercentage
				}
			case "pid":
				if descending {
					shouldSwap = processes[j].PID > processes[selectedIdx].PID
				} else {
					shouldSwap = processes[j].PID < processes[selectedIdx].PID
				}
			case "name":
				if descending {
					shouldSwap = processes[j].Name > processes[selectedIdx].Name
				} else {
					shouldSwap = processes[j].Name < processes[selectedIdx].Name
				}
			}

			if shouldSwap {
				selectedIdx = j
			}
		}

		// Swap elements if necessary
		if selectedIdx != i {
			processes[i], processes[selectedIdx] = processes[selectedIdx], processes[i]
		}
	}
}

// TruncateString truncates a string to a maximum length
// Adds "..." at the end if the string is truncated
//
// Parameters:
//   - s: string to truncate
//   - maxLen: maximum allowed length
//
// Returns: truncated string (if necessary)
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen] // If maxLen is too small, just cut
	}
	return s[:maxLen-3] + "..."
}

// FormatBytes converts bytes to a readable string (MB, GB, etc.)
// Useful for presenting memory sizes in a user-friendly way
//
// Parameters:
//   - bytes: number of bytes to format
//
// Returns: formatted string (e.g. "256.5 MB", "1.2 GB")
func FormatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// MonitorProcessContinuously continuously monitors a specific process
// Prints statistics at each specified interval until the process terminates or Ctrl+C
//
// Parameters:
//   - targetPID: PID of the process to monitor
//   - intervalSeconds: interval between updates in seconds
//
// Returns: error if the process cannot be monitored
func MonitorProcessContinuously(targetPID int32, intervalSeconds int) error {
	fmt.Printf("\n╔══════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  Monitoring process PID %d every %d seconds              ║\n", targetPID, intervalSeconds)
	fmt.Printf("║  Press Ctrl+C to stop                                       ║\n")
	fmt.Printf("╚══════════════════════════════════════════════════════════════╝\n\n")

	// Get total system memory once
	totalSystemMem, err := GetSystemMemoryTotal()
	if err != nil {
		return err
	}

	// Infinite monitoring loop
	for {
		// Get the process
		p, err := GetProcessByPID(targetPID)
		if err != nil {
			return fmt.Errorf("process terminated or is not accessible: %w", err)
		}

		// Get process statistics
		info, err := GetProcessInfo(p, totalSystemMem)
		if err != nil {
			return fmt.Errorf("error getting process statistics: %w", err)
		}

		// Print formatted statistics
		timestamp := time.Now().Format("15:04:05")
		fmt.Printf("┌─ [%s] ────────────────────────────────────────────────┐\n", timestamp)
		fmt.Printf("│ PID:  %-50d │\n", info.PID)
		fmt.Printf("│ Name: %-50s │\n", TruncateString(info.Name, 50))
		fmt.Printf("│ CPU:  %-6.2f%% %-42s │\n", info.CPUPercentage, "")
		fmt.Printf("│ RAM:  %-6.2f%% (%-36s) │\n", info.RAMPercentage, FormatBytes(info.RAMBytes))
		fmt.Printf("└───────────────────────────────────────────────────────────┘\n\n")

		// Wait for the specified interval before the next update
		time.Sleep(time.Duration(intervalSeconds) * time.Second)
	}
}

// PrintProcessTable prints a formatted table of processes
// Used to present process lists consistently across all modules
//
// Parameters:
//   - processes: slice of ProcessInfo to print
//   - maxProcesses: maximum number of processes to show (0 = all)
//   - title: table title
func PrintProcessTable(processes []ProcessInfo, maxProcesses int, title string) {
	// Limit to the requested number of processes
	if maxProcesses > 0 && maxProcesses < len(processes) {
		processes = processes[:maxProcesses]
	}

	// Print header
	fmt.Printf("\n╔══════════════════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  %-80s  ║\n", title)
	fmt.Printf("╠══════════════════════════════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║ %-8s │ %-30s │ %-10s │ %-10s │ %-12s ║\n", "PID", "Name", "CPU %", "RAM %", "RAM")
	fmt.Printf("╠══════════════════════════════════════════════════════════════════════════════════╣\n")

	// Print each process
	for _, p := range processes {
		fmt.Printf("║ %-8d │ %-30s │ %9.2f%% │ %9.2f%% │ %12s ║\n",
			p.PID,
			TruncateString(p.Name, 30),
			p.CPUPercentage,
			p.RAMPercentage,
			FormatBytes(p.RAMBytes))
	}

	fmt.Printf("╚══════════════════════════════════════════════════════════════════════════════════╝\n")
}
