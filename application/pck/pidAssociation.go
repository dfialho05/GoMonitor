package pck

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

// ProcessInfo contém as estatísticas combinadas de CPU e RAM para cada processo
type ProcessInfo struct {
	PID           int32   // ID do processo
	Name          string  // Nome do processo
	CPUPercentage float64 // Percentagem de uso do CPU
	RAMPercentage float32 // Percentagem de uso da RAM
	RAMBytes      uint64  // Memória RAM utilizada em bytes
}

// GetProcessAssociation recolhe e associa as estatísticas de CPU e RAM para cada processo
// Retorna uma lista de ProcessInfo com todos os processos ativos no sistema
func GetProcessAssociation() ([]ProcessInfo, error) {
	// 1. Obter a lista de todos os processos ativos no sistema
	allProcesses, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter lista de processos: %w", err)
	}

	// 2. Obter a memória total do sistema (necessário para calcular percentagens de RAM)
	vm, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter informação de memória: %w", err)
	}
	totalSystemMem := float64(vm.Total)

	// 3. Slice para armazenar as informações de todos os processos
	var processInfoList []ProcessInfo

	// 4. Iterar por cada processo e recolher as suas estatísticas
	for _, p := range allProcesses {
		// 4.1. Obter o PID do processo
		pid := p.Pid

		// 4.2. Obter o nome do processo
		// Muitos processos de sistema/kernel não permitem ler o nome sem root
		name, err := p.Name()
		if err != nil {
			// Se não conseguirmos obter o nome, saltamos este processo
			continue
		}

		// 4.3. Obter a percentagem de uso do CPU
		// Usamos um tempo de espera curto para não bloquear demasiado
		cpuPercent, err := p.CPUPercent()
		if err != nil {
			// Se houver erro, assumimos 0% de CPU
			cpuPercent = 0.0
		}

		// 4.4. Obter informação sobre o uso de memória
		memInfo, err := p.MemoryInfo()
		if err != nil {
			// Se não conseguirmos obter a memória, saltamos este processo
			continue
		}

		// 4.5. Calcular a percentagem de RAM utilizada
		// RSS (Resident Set Size) é a memória física RAM realmente usada pelo processo
		rss := float64(memInfo.RSS)
		ramPercentage := float32((rss / totalSystemMem) * 100)

		// 4.6. Adicionar as informações do processo à lista
		processInfoList = append(processInfoList, ProcessInfo{
			PID:           pid,
			Name:          name,
			CPUPercentage: cpuPercent,
			RAMPercentage: ramPercentage,
			RAMBytes:      memInfo.RSS,
		})
	}

	return processInfoList, nil
}

// GetProcessAssociationSorted recolhe e retorna os processos ordenados por uso de CPU (descendente)
func GetProcessAssociationSorted() ([]ProcessInfo, error) {
	// 1. Obter todos os processos com as suas estatísticas
	processes, err := GetProcessAssociation()
	if err != nil {
		return nil, err
	}

	// 2. Ordenar os processos por uso de CPU (do maior para o menor)
	// Usando selection sort simples para evitar dependências externas
	for i := 0; i < len(processes)-1; i++ {
		maxIdx := i
		for j := i + 1; j < len(processes); j++ {
			if processes[j].CPUPercentage > processes[maxIdx].CPUPercentage {
				maxIdx = j
			}
		}
		// Trocar os elementos se necessário
		if maxIdx != i {
			processes[i], processes[maxIdx] = processes[maxIdx], processes[i]
		}
	}

	return processes, nil
}

// GetProcessAssociationByPID procura e retorna as estatísticas de um processo específico pelo seu PID
func GetProcessAssociationByPID(targetPID int32) (*ProcessInfo, error) {
	// 1. Criar um objeto de processo com o PID fornecido
	p, err := process.NewProcess(targetPID)
	if err != nil {
		return nil, fmt.Errorf("processo com PID %d não encontrado: %w", targetPID, err)
	}

	// 2. Obter a memória total do sistema
	vm, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter informação de memória: %w", err)
	}
	totalSystemMem := float64(vm.Total)

	// 3. Obter o nome do processo
	name, err := p.Name()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter nome do processo: %w", err)
	}

	// 4. Obter a percentagem de CPU
	cpuPercent, err := p.CPUPercent()
	if err != nil {
		cpuPercent = 0.0
	}

	// 5. Obter informação de memória
	memInfo, err := p.MemoryInfo()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter informação de memória do processo: %w", err)
	}

	// 6. Calcular a percentagem de RAM
	rss := float64(memInfo.RSS)
	ramPercentage := float32((rss / totalSystemMem) * 100)

	// 7. Retornar as informações do processo
	return &ProcessInfo{
		PID:           targetPID,
		Name:          name,
		CPUPercentage: cpuPercent,
		RAMPercentage: ramPercentage,
		RAMBytes:      memInfo.RSS,
	}, nil
}

// MonitorProcessContinuous monitoriza continuamente um processo específico e imprime as estatísticas
// Útil para debugging e testes
func MonitorProcessContinuous(targetPID int32, intervalSeconds int) error {
	fmt.Printf("A monitorizar processo PID %d a cada %d segundos...\n", targetPID, intervalSeconds)
	fmt.Println("Pressione Ctrl+C para parar")

	for {
		// Obter as estatísticas do processo
		info, err := GetProcessAssociationByPID(targetPID)
		if err != nil {
			return fmt.Errorf("erro ao monitorizar processo: %w", err)
		}

		// Imprimir as estatísticas
		fmt.Printf("\n[%s] PID: %d | Nome: %s\n",
			time.Now().Format("15:04:05"),
			info.PID,
			info.Name)
		fmt.Printf("  CPU: %.2f%% | RAM: %.2f%% (%.2f MB)\n",
			info.CPUPercentage,
			info.RAMPercentage,
			float64(info.RAMBytes)/1024/1024)

		// Esperar pelo intervalo especificado
		time.Sleep(time.Duration(intervalSeconds) * time.Second)
	}
}

// PrintTopProcesses imprime os N processos com maior uso de recursos
func PrintTopProcesses(n int) error {
	// Obter os processos ordenados
	processes, err := GetProcessAssociationSorted()
	if err != nil {
		return err
	}

	// Limitar ao número de processos solicitado
	if n > len(processes) {
		n = len(processes)
	}

	fmt.Printf("\n=== Top %d Processos (por uso de CPU) ===\n", n)
	fmt.Printf("%-8s %-30s %-10s %-10s %-15s\n", "PID", "Nome", "CPU %", "RAM %", "RAM (MB)")
	fmt.Println("--------------------------------------------------------------------------------")

	for i := 0; i < n; i++ {
		p := processes[i]
		ramMB := float64(p.RAMBytes) / 1024 / 1024
		fmt.Printf("%-8d %-30s %-10.2f %-10.2f %-15.2f\n",
			p.PID,
			truncateString(p.Name, 30),
			p.CPUPercentage,
			p.RAMPercentage,
			ramMB)
	}

	return nil
}

// truncateString trunca uma string para um comprimento máximo
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
