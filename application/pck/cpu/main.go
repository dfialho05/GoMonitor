package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/process"
)

// GeneralStats contains the global hardware information
type GeneralStats struct {
	Percentage float64
	Cores      int
	ClockSpeed float64
	ModelName  string
	VendorID   string
	Microcode  string
	CacheSize  int32
	Flags      string
}

// ProcessStats contains the information for each individual process
type ProcessStats struct {
	PID        int32
	Name       string
	Percentage float64
}

// GetGeneralStats collects global CPU data (Neofetch-like)
func GetGeneralStats() (GeneralStats, error) {
	// 1. Get the global usage percentage
	// Wait 1 second for an accurate reading
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return GeneralStats{}, err
	}

	percentage := 0.0
	if len(cpuPercent) > 0 {
		percentage = cpuPercent[0]
	}

	// 2. Get static CPU information
	cpuInfo, err := cpu.Info()
	if err != nil {
		return GeneralStats{}, err
	}

	stats := GeneralStats{
		Percentage: percentage,
	}

	if len(cpuInfo) > 0 {
		info := cpuInfo[0]
		stats.ModelName = info.ModelName
		stats.Cores = int(info.Cores)
		stats.ClockSpeed = info.Mhz // gopsutil uses 'Mhz' for ClockSpeed
		stats.VendorID = info.VendorID
		stats.Microcode = info.Microcode
		stats.CacheSize = info.CacheSize
		// Join flags into a single space-separated string
		stats.Flags = strings.Join(info.Flags, " ")
	}

	return stats, nil
}

// GetProcessStats collects the list of processes and their CPU usage (task manager-like)
func GetProcessStats() ([]ProcessStats, error) {
	// 1. Get the list of all active PIDs
	allProcesses, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var stats []ProcessStats

	for _, p := range allProcesses {
		// Get the process name
		name, err := p.Name()
		if err != nil {
			// Many system/kernel processes don't allow reading the name without root
			continue
		}

		// Get the CPU usage of the process
		// We use 0 to get an instantaneous value without blocking the loop
		percent, err := p.CPUPercent()
		if err != nil {
			continue
		}

		// Add to the results list
		stats = append(stats, ProcessStats{
			PID:        p.Pid,
			Name:       name,
			Percentage: percent,
		})
	}

	return stats, nil
}

func main() {

	fmt.Println("A iniciar verificação do CPU...")

	// Get and print general CPU statistics
	general, err := GetGeneralStats()
	if err != nil {
		fmt.Printf("Erro ao obter estatísticas gerais do CPU: %v\n", err)
	} else {
		fmt.Println("=== CPU Geral ===")
		fmt.Printf("Uso total: %.2f%%\n", general.Percentage)
		fmt.Printf("Cores: %d\n", general.Cores)
		fmt.Printf("ClockSpeed (MHz): %.2f\n", general.ClockSpeed)
		fmt.Printf("ModelName: %s\n", general.ModelName)
		fmt.Printf("VendorID: %s\n", general.VendorID)
		fmt.Printf("Microcode: %s\n", general.Microcode)
		fmt.Printf("CacheSize: %d\n", general.CacheSize)
		fmt.Printf("Flags: %s\n", general.Flags)
	}

	// Get and print per-process statistics
	processes, err := GetProcessStats()
	if err != nil {
		fmt.Printf("Erro ao obter estatísticas de processos: %v\n", err)
	} else {
		// Sort by Percentage descending (from largest to smallest)
		for i := 0; i < len(processes)-1; i++ {
			maxIdx := i
			for j := i + 1; j < len(processes); j++ {
				if processes[j].Percentage > processes[maxIdx].Percentage {
					maxIdx = j
				}
			}
			if maxIdx != i {
				processes[i], processes[maxIdx] = processes[maxIdx], processes[i]
			}
		}

		fmt.Printf("\nEncontrados %d processos. A listar usos de CPU por processo (do maior para o menor):\n", len(processes))
		for _, p := range processes {
			fmt.Printf("PID: %d\tNome: %s\tCPU: %.2f%%\n", p.PID, p.Name, p.Percentage)
		}
	}

	fmt.Println("Verificação terminada.")
}
