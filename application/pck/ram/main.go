package main

import (
	"fmt"

	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

type RamGeneral struct {
	Total uint64
	Free  uint64
}

type RamProcess struct {
	PID        int32 // gopsutil uses int32 for PIDs
	Percentage float32
}

func getRamGeneral() (RamGeneral, error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return RamGeneral{}, err
	}

	return RamGeneral{
		Total: vm.Total,
		Free:  vm.Free,
	}, nil
}

func getRamProcess() ([]RamProcess, error) {
	// 1. Obtain total system memory ONCE before the loop
	vm, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	totalSystemMem := float64(vm.Total)

	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var ramProcessInfo []RamProcess

	for _, p := range processes {
		// 2. In gopsutil v3, p.Pid is a field (int32), not a function
		pid := p.Pid

		// MemoryPercent() is more reliable and faster than calculating manually
		// but if you want to use MemoryInfo:
		memInfo, err := p.MemoryInfo()
		if err != nil {
			continue
		}

		// RSS (Resident Set Size) is the actual physical RAM used by the process
		rss := float64(memInfo.RSS)

		// 3. Calculate the percentage using the total memory obtained in step 1
		percentage := (rss / totalSystemMem) * 100

		if rss > 0 {
			ramProcessInfo = append(ramProcessInfo, RamProcess{
				PID:        pid,
				Percentage: float32(percentage),
			})
		}
	}

	return ramProcessInfo, nil
}

func main() {
	fmt.Println("=== Teste de RAM ===")

	general, _ := getRamGeneral()
	fmt.Printf("Total: %d MB | Livre: %d MB\n", general.Total/1024/1024, general.Free/1024/1024)

	procRam, _ := getRamProcess()
	fmt.Printf("Monitorizando %d processos...\n", len(procRam))

	// Sort from highest to lowest RAM usage and show only the first 5 for testing
	if len(procRam) > 1 {
		// Descending sort by Percentage (simple selection sort to avoid importing packages)
		for i := 0; i < len(procRam)-1; i++ {
			maxIdx := i
			for j := i + 1; j < len(procRam); j++ {
				if procRam[j].Percentage > procRam[maxIdx].Percentage {
					maxIdx = j
				}
			}
			if maxIdx != i {
				procRam[i], procRam[maxIdx] = procRam[maxIdx], procRam[i]
			}
		}
	}
	for i := 0; i < len(procRam); i++ {
		fmt.Printf("PID: %d | RAM: %.2f%%\n", procRam[i].PID, procRam[i].Percentage)
	}
}
