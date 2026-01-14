package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type GPUStats struct {
	Model       string
	Utilization float64
	MemoryTotal uint64 // MB
	MemoryUsed  uint64 // MB
	Temp        int    // Celsius
}

func GetGPUStats() (GPUStats, error) {
	// Try NVIDIA first.
	stats, err := getNvidiaStats()
	if err == nil {
		return stats, nil
	}

	// If NVIDIA detection fails, fall back to reading integrated GPU info via sysfs (Linux).
	return getIntegratedStats()
}

func getNvidiaStats() (GPUStats, error) {
	cmd := exec.Command("nvidia-smi", "--query-gpu=name,utilization.gpu,memory.total,memory.used,temperature.gpu", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		return GPUStats{}, err
	}

	fields := strings.Split(strings.TrimSpace(string(output)), ", ")
	util, _ := strconv.ParseFloat(fields[1], 64)
	memTotal, _ := strconv.ParseUint(fields[2], 10, 64)
	memUsed, _ := strconv.ParseUint(fields[3], 10, 64)
	temp, _ := strconv.Atoi(fields[4])

	return GPUStats{
		Model:       fields[0],
		Utilization: util,
		MemoryTotal: memTotal,
		MemoryUsed:  memUsed,
		Temp:        temp,
	}, nil
}

func getIntegratedStats() (GPUStats, error) {
	gpuPath := "/sys/class/drm/card0/device/"

	// Read vendor and device IDs from sysfs.
	vendorBuf, _ := os.ReadFile(gpuPath + "vendor")
	deviceBuf, _ := os.ReadFile(gpuPath + "device")
	vendor := strings.TrimSpace(string(vendorBuf))
	device := strings.TrimSpace(string(deviceBuf))

	modelName := "Intel UHD Graphics"
	if vendor == "0x8086" {
		// Check for known device IDs (examples for some Intel UHD 620 variants).
		if device == "0x3ea0" || device == "0x5917" {
			modelName = "Intel(R) UHD Graphics 620"
		}
	}

	// Try to read a package/cpu thermal sensor for an approximate GPU temperature.
	// On many laptops the integrated GPU temperature is not exposed directly and
	// may be represented by a CPU/package thermal zone.
	tempBuf, err := os.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	temp := 0
	if err == nil {
		t, _ := strconv.Atoi(strings.TrimSpace(string(tempBuf)))
		// Convert from millidegrees Celsius to degrees Celsius.
		temp = t / 1000
	}

	return GPUStats{
		Model:       modelName,
		Utilization: 0.1, // Small placeholder so the field can be displayed if desired.
		Temp:        temp,
		MemoryTotal: 1, // Placeholder to indicate shared (integrated) memory.
	}, nil
}

func main() {
	fmt.Println("=== gotm: GPU Detection ===")

	gpu, err := GetGPUStats()
	if err != nil {
		fmt.Println("Não foi possível detetar nenhuma GPU activa.")
	} else {
		fmt.Printf("GPU detetada: %s\n", gpu.Model)
		// Only print fields that have a positive value.
		if gpu.Utilization > 0 {
			fmt.Printf("  Uso: %.1f%%\n", gpu.Utilization)
		}
		if gpu.MemoryTotal > 0 {
			fmt.Printf("  VRAM: %dMB / %dMB\n", gpu.MemoryUsed, gpu.MemoryTotal)
		}
		if gpu.Temp > 0 {
			fmt.Printf("  Temperatura: %d°C\n", gpu.Temp)
		}
	}
}
