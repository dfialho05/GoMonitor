package cpu

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dfialho05/GoMonitor/application/pck/common"
	"github.com/shirou/gopsutil/v3/cpu"
)

// GeneralStats contains general information about the system CPU
// This structure aggregates static data (model, cores) and dynamic data (current usage)
type GeneralStats struct {
	Percentage  float64 // Global CPU usage percentage (0-100%)
	Cores       int     // Number of physical CPU cores
	ClockSpeed  float64 // Clock speed in MHz
	ModelName   string  // CPU model name (e.g. "Intel Core i7-8550U")
	VendorID    string  // Vendor identifier (e.g. "GenuineIntel", "AuthenticAMD")
	Microcode   string  // CPU microcode version
	CacheSize   int32   // CPU cache size in KB
	Flags       string  // CPU flags/capabilities (e.g. "sse", "avx", "aes")
	Temperature int     // CPU temperature in degrees Celsius (0 if not available)
}

// GetGeneralStats collects general information about the system CPU
// This function aggregates static data (model, cores, cache) and dynamic data (current usage)
// Similar to the output of 'lscpu' command
//
// Returns:
//   - GeneralStats filled with CPU information
//   - error if unable to get the information
func GetGeneralStats() (GeneralStats, error) {
	// 1. Get global CPU usage percentage
	// Wait 1 second to get an accurate reading
	// false = return only one global value (average of all cores)
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return GeneralStats{}, fmt.Errorf("error getting CPU usage percentage: %w", err)
	}

	// Extract the first (and only) value from the slice
	percentage := 0.0
	if len(cpuPercent) > 0 {
		percentage = cpuPercent[0]
	}

	// 2. Get static CPU information
	cpuInfo, err := cpu.Info()
	if err != nil {
		return GeneralStats{}, fmt.Errorf("error getting CPU information: %w", err)
	}

	// 3. Initialize the return structure with usage percentage
	stats := GeneralStats{
		Percentage: percentage,
	}

	// 4. Fill static fields if information is available
	// Normally the cpuInfo slice contains one entry per logical core,
	// but they all have the same static information, so we use the first one
	if len(cpuInfo) > 0 {
		info := cpuInfo[0]
		stats.ModelName = info.ModelName
		stats.Cores = int(info.Cores)
		stats.ClockSpeed = info.Mhz // gopsutil uses 'Mhz' for frequency in MHz
		stats.VendorID = info.VendorID
		stats.Microcode = info.Microcode
		stats.CacheSize = info.CacheSize
		// Join flags into a space-separated string
		stats.Flags = strings.Join(info.Flags, " ")
	}

	// 5. Get CPU temperature
	stats.Temperature = getCPUTemperature()

	return stats, nil
}

// GetProcessStats collects CPU information for all active processes
// This function is a wrapper that reuses common process collection logic
// Similar to task manager output
//
// Returns:
//   - slice of ProcessInfo sorted by CPU usage (descending)
//   - error if unable to get the data
func GetProcessStats() ([]common.ProcessInfo, error) {
	// 1. Collect information from all processes using the common function
	processes, err := common.CollectAllProcessInfo()
	if err != nil {
		return nil, fmt.Errorf("error collecting processes: %w", err)
	}

	// 2. Sort processes by CPU usage (highest to lowest)
	// This makes it easier to view the most active processes
	common.SortProcessesByField(processes, "cpu", true)

	return processes, nil
}

// PrintGeneralStats prints general CPU statistics in a formatted way
// This function presents a complete summary of CPU capabilities and current usage
//
// Parameters:
//   - stats: GeneralStats structure with data to present
func PrintGeneralStats(stats GeneralStats) {
	fmt.Printf("\n╔══════════════════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  %-80s  ║\n", "General CPU Information")
	fmt.Printf("╠══════════════════════════════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║  Model:           %-62s  ║\n", common.TruncateString(stats.ModelName, 62))
	fmt.Printf("║  Vendor:          %-62s  ║\n", stats.VendorID)
	fmt.Printf("║  Cores:           %-62d  ║\n", stats.Cores)
	fmt.Printf("║  Frequency:       %-58.2f MHz  ║\n", stats.ClockSpeed)
	fmt.Printf("║  Current Usage:   %-58.2f %%    ║\n", stats.Percentage)
	fmt.Printf("║  Cache:           %-58d KB  ║\n", stats.CacheSize)
	fmt.Printf("║  Microcode:       %-62s  ║\n", stats.Microcode)

	// Show temperature if available
	if stats.Temperature > 0 {
		fmt.Printf("║  Temperature:     %-58d °C  ║\n", stats.Temperature)
	} else {
		fmt.Printf("║  Temperature:     %-62s  ║\n", "N/A (not available)")
	}

	fmt.Printf("╚══════════════════════════════════════════════════════════════════════════════════╝\n")

	// Note: Flags are not printed by default as they are very long
	// Uncomment the line below if you want to see all CPU flags
	// fmt.Printf("\nFlags: %s\n", stats.Flags)
}

// PrintTopProcessesByCPU prints the N processes with highest CPU usage
// This function provides a formatted view of the most active processes
//
// Parameters:
//   - n: number of processes to show (top N)
//
// Returns:
//   - error if unable to get the data
func PrintTopProcessesByCPU(n int) error {
	// Get processes sorted by CPU
	processes, err := GetProcessStats()
	if err != nil {
		return err
	}

	// Use the common function to print the table
	title := fmt.Sprintf("Top %d Processes by CPU Usage", n)
	common.PrintProcessTable(processes, n, title)

	return nil
}

// GetCPUUsageByPID gets the CPU usage of a specific process
// This function is useful for monitoring an individual process
//
// Parameters:
//   - pid: process ID
//
// Returns:
//   - percentage of CPU usage by the process
//   - error if the process doesn't exist or is not accessible
func GetCPUUsageByPID(pid int32) (float64, error) {
	// Get total system memory
	totalMem, err := common.GetSystemMemoryTotal()
	if err != nil {
		return 0, err
	}

	// Get the process
	p, err := common.GetProcessByPID(pid)
	if err != nil {
		return 0, err
	}

	// Get process information
	info, err := common.GetProcessInfo(p, totalMem)
	if err != nil {
		return 0, err
	}

	return info.CPUPercentage, nil
}

// getCPUTemperature gets the system CPU temperature
// Searches for thermal zones that contain CPU temperature (x86_pkg_temp, coretemp, etc.)
//
// Returns:
//   - temperature in degrees Celsius (0 if not available)
func getCPUTemperature() int {
	// List of thermal zone types that contain CPU temperature
	// x86_pkg_temp is the CPU package temperature (most common on Intel systems)
	// acpitz can also contain CPU temperature on some systems
	targetTypes := []string{"x86_pkg_temp", "coretemp", "cpu_thermal", "acpitz"}

	// Search all available thermal zones
	for i := 0; i < 20; i++ {
		zonePath := fmt.Sprintf("/sys/class/thermal/thermal_zone%d/", i)

		// Read the thermal zone type
		typeBuf, err := os.ReadFile(zonePath + "type")
		if err != nil {
			continue // This zone doesn't exist or is not accessible
		}

		zoneType := strings.TrimSpace(string(typeBuf))

		// Check if it's a CPU thermal zone
		isCPUZone := false
		for _, targetType := range targetTypes {
			if zoneType == targetType || strings.Contains(zoneType, targetType) {
				isCPUZone = true
				break
			}
		}

		if !isCPUZone {
			continue
		}

		// Read the temperature from this zone
		tempBuf, err := os.ReadFile(zonePath + "temp")
		if err != nil {
			continue
		}

		// Convert from string to integer
		tempMilliC, err := strconv.Atoi(strings.TrimSpace(string(tempBuf)))
		if err != nil {
			continue
		}

		// Convert from millidegrees Celsius to degrees Celsius
		temp := tempMilliC / 1000

		// Validate if temperature is reasonable (between 0 and 150°C)
		if temp > 0 && temp < 150 {
			return temp
		}
	}

	// If not found, return 0 (not available)
	return 0
}
