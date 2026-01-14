package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/dfialho05/GoMonitor/application/pck"
)

func main() {
	fmt.Println("=== GoMonitor - Monitor de Processos ===\n")

	// Verificar se foi fornecido um argumento (PID específico ou número de processos)
	if len(os.Args) > 1 {
		// Tentar converter o argumento para um número
		arg := os.Args[1]
		num, err := strconv.Atoi(arg)

		if err != nil {
			fmt.Printf("Erro: '%s' não é um número válido\n", arg)
			fmt.Println("\nUso:")
			fmt.Println("  ./main           - Lista os top 10 processos")
			fmt.Println("  ./main 20        - Lista os top 20 processos")
			fmt.Println("  ./main -p 1234   - Monitoriza o processo com PID 1234")
			return
		}

		// Se o argumento anterior foi "-p", monitorizar esse PID
		if len(os.Args) > 2 && os.Args[1] == "-p" {
			pid, err := strconv.Atoi(os.Args[2])
			if err != nil {
				fmt.Printf("Erro: PID inválido '%s'\n", os.Args[2])
				return
			}

			// Monitorizar processo específico a cada 2 segundos
			err = pck.MonitorProcessContinuous(int32(pid), 2)
			if err != nil {
				fmt.Printf("Erro ao monitorizar processo: %v\n", err)
			}
			return
		}

		// Caso contrário, listar top N processos
		err = pck.PrintTopProcesses(num)
		if err != nil {
			fmt.Printf("Erro ao obter processos: %v\n", err)
		}
		return
	}

	// Comportamento padrão: mostrar top 10 processos
	err := pck.PrintTopProcesses(10)
	if err != nil {
		fmt.Printf("Erro ao obter processos: %v\n", err)
		return
	}

	fmt.Println("\n=== Exemplo de uso de GetProcessAssociation ===")

	// Obter todos os processos
	processes, err := pck.GetProcessAssociation()
	if err != nil {
		fmt.Printf("Erro: %v\n", err)
		return
	}

	fmt.Printf("\nTotal de processos monitorizados: %d\n", len(processes))

	// Mostrar estatísticas gerais
	var totalCPU float64
	var totalRAM float32

	for _, p := range processes {
		totalCPU += p.CPUPercentage
		totalRAM += p.RAMPercentage
	}

	fmt.Printf("Uso total de CPU (soma de todos os processos): %.2f%%\n", totalCPU)
	fmt.Printf("Uso total de RAM (soma de todos os processos): %.2f%%\n", totalRAM)

	fmt.Println("\n=== Procurar processo específico ===")

	// Exemplo: procurar pelo processo init (PID 1)
	if len(processes) > 0 {
		// Pegar no primeiro processo como exemplo
		examplePID := processes[0].PID
		info, err := pck.GetProcessAssociationByPID(examplePID)
		if err != nil {
			fmt.Printf("Erro ao procurar processo: %v\n", err)
		} else {
			fmt.Printf("Processo encontrado:\n")
			fmt.Printf("  PID: %d\n", info.PID)
			fmt.Printf("  Nome: %s\n", info.Name)
			fmt.Printf("  CPU: %.2f%%\n", info.CPUPercentage)
			fmt.Printf("  RAM: %.2f%% (%.2f MB)\n",
				info.RAMPercentage,
				float64(info.RAMBytes)/1024/1024)
		}
	}

	fmt.Println("\n=== Dica ===")
	fmt.Println("Para monitorizar continuamente um processo específico:")
	fmt.Println("  ./main -p <PID>")
	fmt.Println("\nPara listar mais processos:")
	fmt.Println("  ./main 20")
}
