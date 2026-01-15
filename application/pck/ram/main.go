package ram

import (
	"fmt"

	"github.com/dfialho05/GoMonitor/application/pck/common"
	"github.com/shirou/gopsutil/v3/mem"
)

// RamGeneral contains general information about system RAM
// This structure provides a global view of memory usage
type RamGeneral struct {
	Total     uint64  // Total RAM installed in the system (in bytes)
	Used      uint64  // RAM currently in use (in bytes)
	Free      uint64  // Free/available RAM (in bytes)
	Available uint64  // Available memory for new processes (in bytes, includes reusable cache)
	Percent   float64 // Memory usage percentage (0-100%)
}

// GetRamGeneral collects general information about system RAM
// This function provides global memory usage statistics
//
// Returns:
//   - RamGeneral filled with memory statistics
//   - error if unable to get the information
func GetRamGeneral() (RamGeneral, error) {
	// Get virtual memory (RAM) statistics
	vm, err := mem.VirtualMemory()
	if err != nil {
		return RamGeneral{}, fmt.Errorf("error getting memory information: %w", err)
	}

	// Fill the structure with the obtained data
	return RamGeneral{
		Total:     vm.Total,
		Used:      vm.Used,
		Free:      vm.Free,
		Available: vm.Available,
		Percent:   vm.UsedPercent,
	}, nil
}

// GetProcessStatsByRAM collects RAM information for all active processes
// Processes are automatically sorted by RAM usage (highest to lowest)
//
// Returns:
//   - slice of ProcessInfo sorted by RAM usage (descending)
//   - error if unable to get the data
func GetProcessStatsByRAM() ([]common.ProcessInfo, error) {
	// 1. Collect information from all processes using the common function
	processes, err := common.CollectAllProcessInfo()
	if err != nil {
		return nil, fmt.Errorf("error collecting processes: %w", err)
	}

	// 2. Sort processes by RAM usage (highest to lowest)
	// This makes it easier to identify processes that consume the most memory
	common.SortProcessesByField(processes, "ram", true)

	return processes, nil
}

// GetRAMUsageByPID gets the RAM usage of a specific process
// This function is useful for monitoring an individual process's memory
//
// Parameters:
//   - pid: process ID
//
// Returns:
//   - percentage of RAM usage by the process
//   - bytes of RAM used by the process
//   - error if the process doesn't exist or is not accessible
func GetRAMUsageByPID(pid int32) (float32, uint64, error) {
	// Get total system memory
	totalMem, err := common.GetSystemMemoryTotal()
	if err != nil {
		return 0, 0, err
	}

	// Get the process
	p, err := common.GetProcessByPID(pid)
	if err != nil {
		return 0, 0, err
	}

	// Get process information
	info, err := common.GetProcessInfo(p, totalMem)
	if err != nil {
		return 0, 0, err
	}

	return info.RAMPercentage, info.RAMBytes, nil
}

// PrintGeneralStats prints general RAM statistics in a formatted way
// This function presents a complete summary of system memory usage
//
// Parameters:
//   - stats: RamGeneral structure with data to present
func PrintGeneralStats(stats RamGeneral) {
	fmt.Printf("\n╔══════════════════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  %-80s  ║\n", "General RAM Memory Information")
	fmt.Printf("╠══════════════════════════════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║  Total:           %-62s  ║\n", common.FormatBytes(stats.Total))
	fmt.Printf("║  Used:            %-62s  ║\n", common.FormatBytes(stats.Used))
	fmt.Printf("║  Free:            %-62s  ║\n", common.FormatBytes(stats.Free))
	fmt.Printf("║  Available:       %-62s  ║\n", common.FormatBytes(stats.Available))
	fmt.Printf("║  Usage:           %-58.2f %%    ║\n", stats.Percent)
	fmt.Printf("╚══════════════════════════════════════════════════════════════════════════════════╝\n")
}

// PrintTopProcessesByRAM prints the N processes with highest RAM usage
// This function provides a formatted view of processes that consume the most memory
//
// Parameters:
//   - n: number of processes to show (top N)
//
// Returns:
//   - error if unable to get the data
func PrintTopProcessesByRAM(n int) error {
	// Get processes sorted by RAM
	processes, err := GetProcessStatsByRAM()
	if err != nil {
		return err
	}

	// Use the common function to print the table
	title := fmt.Sprintf("Top %d Processes by RAM Usage", n)
	common.PrintProcessTable(processes, n, title)

	return nil
}

// GetSwapMemory gets information about system swap memory
// Swap is virtual memory on disk used when RAM is full
//
// Returns:
//   - total: total swap size in bytes
//   - used: swap currently in use in bytes
//   - percent: swap usage percentage
//   - error if unable to get the information
func GetSwapMemory() (uint64, uint64, float64, error) {
	swapMem, err := mem.SwapMemory()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("error getting swap information: %w", err)
	}

	return swapMem.Total, swapMem.Used, swapMem.UsedPercent, nil
}

// PrintSwapStats prints swap memory statistics in a formatted way
// Shows information about system virtual memory (swap) usage
func PrintSwapStats() error {
	total, used, percent, err := GetSwapMemory()
	if err != nil {
		return err
	}

	free := total - used

	fmt.Printf("\n╔══════════════════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  %-80s  ║\n", "Swap Memory Information")
	fmt.Printf("╠══════════════════════════════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║  Total:           %-62s  ║\n", common.FormatBytes(total))
	fmt.Printf("║  Used:            %-62s  ║\n", common.FormatBytes(used))
	fmt.Printf("║  Free:            %-62s  ║\n", common.FormatBytes(free))
	fmt.Printf("║  Usage:           %-58.2f %%    ║\n", percent)
	fmt.Printf("╚══════════════════════════════════════════════════════════════════════════════════╝\n")

	return nil
}
