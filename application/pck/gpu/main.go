package gpu

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// GPUStats contains GPU usage statistics
// This structure supports both dedicated GPUs (NVIDIA) and integrated GPUs (Intel)
type GPUStats struct {
	Model        string  // GPU model name (e.g. "NVIDIA GeForce RTX 3060", "Intel UHD Graphics 620")
	Utilization  float64 // GPU utilization percentage (0-100%)
	MemoryTotal  uint64  // Total GPU memory in MB (VRAM)
	MemoryUsed   uint64  // Used GPU memory in MB
	Temp         int     // GPU temperature in degrees Celsius
	IsIntegrated bool    // Indicates if it's an integrated GPU (true) or dedicated (false)
}

// GetGPUStats detects and collects statistics from the active GPU in the system
// This function first tries to detect an NVIDIA GPU using nvidia-smi
// If that fails, it tries to detect an integrated GPU through sysfs (Linux)
//
// Returns:
//   - GPUStats filled with GPU information
//   - error if no GPU is detected or if there's an error reading
func GetGPUStats() (GPUStats, error) {
	// 1. Try to detect NVIDIA GPU first
	// NVIDIA GPUs are easier to monitor through nvidia-smi
	stats, err := getNvidiaStats()
	if err == nil {
		stats.IsIntegrated = false
		return stats, nil
	}

	// 2. If NVIDIA detection fails, try integrated GPU
	// Integrated GPUs (Intel, AMD APU) use shared RAM memory
	stats, err = getIntegratedStats()
	if err == nil {
		stats.IsIntegrated = true
		return stats, nil
	}

	return GPUStats{}, fmt.Errorf("could not detect any GPU in the system")
}

// getNvidiaStats collects statistics from an NVIDIA GPU using the nvidia-smi command
// This command provides detailed information about usage, memory and temperature
//
// Returns:
//   - GPUStats filled with NVIDIA GPU data
//   - error if nvidia-smi is not available or fails
func getNvidiaStats() (GPUStats, error) {
	// Execute nvidia-smi with specific query to get structured data
	// --query-gpu: specifies which fields we want
	// --format=csv,noheader,nounits: output format without headers and units
	cmd := exec.Command("nvidia-smi",
		"--query-gpu=name,utilization.gpu,memory.total,memory.used,temperature.gpu",
		"--format=csv,noheader,nounits")

	output, err := cmd.Output()
	if err != nil {
		return GPUStats{}, fmt.Errorf("nvidia-smi not available or failed: %w", err)
	}

	// Parse CSV output
	// Expected format: "Name, Utilization, Total Memory, Used Memory, Temperature"
	fields := strings.Split(strings.TrimSpace(string(output)), ", ")
	if len(fields) < 5 {
		return GPUStats{}, fmt.Errorf("unexpected format in nvidia-smi output")
	}

	// Convert numeric values
	util, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		util = 0.0
	}

	memTotal, err := strconv.ParseUint(fields[2], 10, 64)
	if err != nil {
		memTotal = 0
	}

	memUsed, err := strconv.ParseUint(fields[3], 10, 64)
	if err != nil {
		memUsed = 0
	}

	temp, err := strconv.Atoi(fields[4])
	if err != nil {
		temp = 0
	}

	return GPUStats{
		Model:       strings.TrimSpace(fields[0]),
		Utilization: util,
		MemoryTotal: memTotal,
		MemoryUsed:  memUsed,
		Temp:        temp,
	}, nil
}

// getIntegratedStats collects statistics from an integrated GPU through sysfs (Linux)
// Integrated GPUs share memory with the system and have limited monitoring capabilities
//
// Returns:
//   - GPUStats filled with basic integrated GPU data
//   - error if unable to read from sysfs
func getIntegratedStats() (GPUStats, error) {
	// Search for GPU in card0, card1, card2, etc.
	// The GPU can be on any card depending on system configuration
	var vendor, device string
	var foundGPU bool

	for i := 0; i < 10; i++ {
		gpuPath := fmt.Sprintf("/sys/class/drm/card%d/device/", i)

		// Try to read vendor ID
		vendorBuf, err := os.ReadFile(gpuPath + "vendor")
		if err != nil {
			continue // Try next card
		}

		// Try to read device ID
		deviceBuf, err := os.ReadFile(gpuPath + "device")
		if err != nil {
			continue // Try next card
		}

		vendor = strings.TrimSpace(string(vendorBuf))
		device = strings.TrimSpace(string(deviceBuf))

		// Check if it's an Intel or AMD GPU (integrated)
		if vendor == "0x8086" || vendor == "0x1002" {
			foundGPU = true
			break
		}
	}

	if !foundGPU {
		return GPUStats{}, fmt.Errorf("could not find integrated GPU in the system")
	}

	// Determine model name based on IDs
	modelName := identifyGPUModel(vendor, device)

	// Try to read temperature from a thermal zone
	// Search for thermal zones that may have GPU temperature
	temp := readGPUTemperature()

	return GPUStats{
		Model:       modelName,
		Utilization: 0.0, // Integrated GPU: utilization not easily available
		MemoryTotal: 0,   // Integrated GPU: uses shared RAM (not fixed value)
		MemoryUsed:  0,
		Temp:        temp,
	}, nil
}

// identifyGPUModel identifies the GPU model based on vendor/device IDs
// This function maps hexadecimal codes to readable model names
//
// Parameters:
//   - vendor: vendor ID (e.g. "0x8086" for Intel)
//   - device: specific device ID
//
// Returns:
//   - string with GPU model name
func identifyGPUModel(vendor, device string) string {
	// Intel: 0x8086
	if vendor == "0x8086" {
		switch device {
		case "0x3ea0", "0x5917":
			return "Intel UHD Graphics 620"
		case "0x5916":
			return "Intel HD Graphics 620"
		case "0x1916":
			return "Intel HD Graphics 520"
		case "0x191b":
			return "Intel HD Graphics 530"
		case "0x3e9b":
			return "Intel UHD Graphics 630"
		case "0x9a49":
			return "Intel Iris Xe Graphics"
		default:
			return "Intel Integrated Graphics"
		}
	}

	// AMD: 0x1002
	if vendor == "0x1002" {
		return "AMD Radeon Graphics"
	}

	// If not recognized, return generic
	return fmt.Sprintf("Integrated Graphics (Vendor: %s, Device: %s)", vendor, device)
}

// readThermalZone reads the temperature from a system thermal zone
// This function tries to read from thermal_zone0, which usually represents CPU/GPU on laptops
//
// Returns:
//   - temperature in degrees Celsius (0 if not available)
func readThermalZone() int {
	// Try to read from the first thermal zone
	// The value is in millidegrees Celsius (e.g. 45000 = 45°C)
	tempBuf, err := os.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	if err != nil {
		return 0
	}

	// Convert from string to integer
	tempMilliC, err := strconv.Atoi(strings.TrimSpace(string(tempBuf)))
	if err != nil {
		return 0
	}

	// Convert from millidegrees to degrees Celsius
	return tempMilliC / 1000
}

// readGPUTemperature tries to read GPU temperature from various thermal zones
// Specifically searches for zones that may contain GPU temperature
//
// Returns:
//   - temperature in degrees Celsius (0 if not available)
func readGPUTemperature() int {
	// List of thermal zone types that may contain GPU temperature
	targetTypes := []string{"INT3400", "acpitz", "pch_skylake", "B0D4"}

	// Search all thermal zones
	for i := 0; i < 20; i++ {
		zonePath := fmt.Sprintf("/sys/class/thermal/thermal_zone%d/", i)

		// Read the thermal zone type
		typeBuf, err := os.ReadFile(zonePath + "type")
		if err != nil {
			continue
		}

		zoneType := strings.TrimSpace(string(typeBuf))

		// Check if it's one of the types we're looking for
		isTarget := false
		for _, targetType := range targetTypes {
			if strings.Contains(zoneType, targetType) {
				isTarget = true
				break
			}
		}

		if !isTarget {
			continue
		}

		// Read the temperature from this zone
		tempBuf, err := os.ReadFile(zonePath + "temp")
		if err != nil {
			continue
		}

		tempMilliC, err := strconv.Atoi(strings.TrimSpace(string(tempBuf)))
		if err != nil {
			continue
		}

		// Convert from millidegrees to degrees Celsius
		temp := tempMilliC / 1000

		// Return the first valid temperature found
		if temp > 0 && temp < 150 { // Sanity check
			return temp
		}
	}

	// If no specific GPU temperature found, use thermal zone 0
	return readThermalZone()
}

// PrintGPUStats prints GPU statistics in a formatted way
// This function presents all available GPU information clearly
//
// Parameters:
//   - stats: GPUStats structure with data to present
func PrintGPUStats(stats GPUStats) {
	fmt.Printf("\n╔══════════════════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  %-80s  ║\n", "GPU Information")
	fmt.Printf("╠══════════════════════════════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║  Model:           %-62s  ║\n", truncateString(stats.Model, 62))

	// GPU type (integrated or dedicated)
	gpuType := "Dedicated"
	if stats.IsIntegrated {
		gpuType = "Integrated"
	}
	fmt.Printf("║  Type:            %-62s  ║\n", gpuType)

	// Utilization (only if available)
	if stats.Utilization > 0 {
		fmt.Printf("║  Utilization:     %-58.1f %%    ║\n", stats.Utilization)
	} else {
		fmt.Printf("║  Utilization:     %-62s  ║\n", "N/A (not available)")
	}

	// Memory (only if available)
	if stats.MemoryTotal > 0 {
		fmt.Printf("║  VRAM Total:      %-58d MB  ║\n", stats.MemoryTotal)
		fmt.Printf("║  VRAM Used:       %-58d MB  ║\n", stats.MemoryUsed)
		memPercent := float64(stats.MemoryUsed) / float64(stats.MemoryTotal) * 100
		fmt.Printf("║  VRAM Usage:      %-58.1f %%    ║\n", memPercent)
	} else {
		fmt.Printf("║  VRAM:            %-62s  ║\n", "Shared (system RAM)")
	}

	// Temperature (only if available)
	if stats.Temp > 0 {
		fmt.Printf("║  Temperature:     %-58d °C  ║\n", stats.Temp)
	} else {
		fmt.Printf("║  Temperature:     %-62s  ║\n", "N/A (not available)")
	}

	fmt.Printf("╚══════════════════════════════════════════════════════════════════════════════════╝\n")
}

// HasNvidiaGPU checks if the system has an NVIDIA GPU available
// This function is useful to determine if it's worth trying to use nvidia-smi
//
// Returns:
//   - true if nvidia-smi is available and functional
func HasNvidiaGPU() bool {
	cmd := exec.Command("nvidia-smi", "-L")
	err := cmd.Run()
	return err == nil
}

// truncateString truncates a string to a maximum length
// Adds "..." at the end if the string is truncated
//
// Parameters:
//   - s: string to truncate
//   - maxLen: maximum allowed length
//
// Returns:
//   - truncated string (if necessary)
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
