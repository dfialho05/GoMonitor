package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/dfialho05/GoMonitor/application/pck"
	"github.com/dfialho05/GoMonitor/application/pck/common"
	"github.com/dfialho05/GoMonitor/application/pck/cpu"
	"github.com/dfialho05/GoMonitor/application/pck/disk"
	"github.com/dfialho05/GoMonitor/application/pck/gpu"
	"github.com/dfialho05/GoMonitor/application/pck/ram"
	"github.com/dfialho05/GoMonitor/application/pck/ui"
)

// Terminal color constants (ANSI codes)
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorBold   = "\033[1m"
)

func main() {
	// Process command line arguments
	if len(os.Args) > 1 {
		// Show header for commands that are not defaultUse and not interactive
		arg1 := os.Args[1]
		if arg1 != "-n" && arg1 != "--default" && arg1 != "-f" && arg1 != "--full" {
			printMainHeader()
		}
		handleCommandLineArgs()
		return
	}

	// Default behavior: show default interface
	showDefaultInterface()
}

// printMainHeader prints the main application header
// Displays the logo and basic information about GoMonitor
func printMainHeader() {
	fmt.Println(colorBold + colorCyan)
	fmt.Println("╔══════════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                                                                                  ║")
	fmt.Println("║                        ██████╗  ██████╗ ███╗   ███╗                              ║")
	fmt.Println("║                       ██╔════╝ ██╔═══██╗████╗ ████║                              ║")
	fmt.Println("║                       ██║  ███╗██║   ██║██╔████╔██║                              ║")
	fmt.Println("║                       ██║   ██║██║   ██║██║╚██╔╝██║                              ║")
	fmt.Println("║                       ╚██████╔╝╚██████╔╝██║ ╚═╝ ██║                              ║")
	fmt.Println("║                        ╚═════╝  ╚═════╝ ╚═╝     ╚═╝                              ║")
	fmt.Println("║                                                                                  ║")
	fmt.Println("║                           System Monitor in Go                                   ║")
	fmt.Println("║                           Resource Monitoring                                    ║")
	fmt.Println("║                                                                                  ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println(colorReset)
}

// handleCommandLineArgs processes command line arguments
// Supports various operation modes based on provided arguments
func handleCommandLineArgs() {
	arg1 := os.Args[1]

	// Help mode
	if arg1 == "-h" || arg1 == "--help" {
		printHelp()
		return
	}

	// Top processes listing mode
	if arg1 == "-t" || arg1 == "--top" {
		n := 10 // Default: top 10
		if len(os.Args) > 2 {
			if num, err := strconv.Atoi(os.Args[2]); err == nil {
				n = num
			}
		}

		showTopProcesses(n)
		return
	}

	// CPU information mode
	if arg1 == "-c" || arg1 == "--cpu" {
		showCPUInfo()
		return
	}

	// RAM information mode
	if arg1 == "-r" || arg1 == "--ram" {
		showRAMInfo()
		return
	}

	// GPU information mode
	if arg1 == "-g" || arg1 == "--gpu" {
		showGPUInfo()
		return
	}

	// Disk information mode
	if arg1 == "-d" || arg1 == "--disk" {
		showDiskInfo()
		return
	}

	// Complete system overview mode
	if arg1 == "-a" || arg1 == "--all" {
		showSystemOverview()
		return
	}

	// Interactive TUI mode (full/interactive mode)
	if arg1 == "-f" || arg1 == "--full" {
		showInteractiveTUI()
		return
	}

	// If we got here, unrecognized argument
	fmt.Printf(colorRed+"Error: Unrecognized argument '%s'\n"+colorReset, arg1)
	printUsage()
}

// printUsage prints basic usage information
func printUsage() {
	fmt.Println("\nUsage: gomonitor [options]")
	fmt.Println("\nFor more information, use: gomonitor --help")
}

// printHelp prints complete help with all available commands
func printHelp() {
	fmt.Println(colorBold + colorGreen + "\n=== GoMonitor - Help ===" + colorReset)
	fmt.Println("\nComplete system monitor written in Go")
	fmt.Println("\n" + colorBold + "USAGE:" + colorReset)
	fmt.Println("  gomonitor [options] [arguments]")

	fmt.Println("\n" + colorBold + "OPTIONS:" + colorReset)
	fmt.Println("  " + colorCyan + "-h, --help" + colorReset + "              Shows this help message")
	fmt.Println("  " + colorCyan + "-f, --full" + colorReset + "              Interactive TUI mode (navigate processes, kill, etc)")
	fmt.Println("  " + colorCyan + "-a, --all" + colorReset + "               Shows complete system overview")
	fmt.Println("  " + colorCyan + "-c, --cpu" + colorReset + "               Shows detailed CPU information")
	fmt.Println("  " + colorCyan + "-r, --ram" + colorReset + "               Shows detailed RAM information")
	fmt.Println("  " + colorCyan + "-g, --gpu" + colorReset + "               Shows GPU information")
	fmt.Println("  " + colorCyan + "-d, --disk" + colorReset + "              Shows disk information")
	fmt.Println("  " + colorCyan + "-t, --top" + colorReset + " [N]           Shows top N processes (default: 10)")

	fmt.Println("\n" + colorBold + "EXAMPLES:" + colorReset)
	fmt.Println("  gomonitor                    # Shows default interface")
	fmt.Println("  gomonitor -f                 # Interactive TUI mode")
	fmt.Println("  gomonitor --all              # Shows complete overview")
	fmt.Println("  gomonitor --cpu              # Shows only CPU information")
	fmt.Println("  gomonitor -t 20              # Shows top 20 processes")

	fmt.Println("\n" + colorBold + "Author:" + colorReset)
	fmt.Println("  GoMonitor is a system monitoring tool like neofetch based on Go")
	fmt.Println("  Author: David Fialho")
	fmt.Println()
}

// showSystemOverview shows a complete overview of all system resources
// This is the main function that aggregates information from all modules
func showSystemOverview() {
	fmt.Println(colorBold + colorYellow + "\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + colorReset)
	fmt.Println(colorBold + "                        SYSTEM OVERVIEW" + colorReset)
	fmt.Println(colorBold + colorYellow + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + colorReset)

	// 1. CPU Information
	fmt.Println(colorBold + colorBlue + "\n[1] PROCESSOR (CPU)" + colorReset)
	showCPUInfo()

	// 2. RAM Information
	fmt.Println(colorBold + colorBlue + "\n[2] RAM MEMORY" + colorReset)
	showRAMInfo()

	// 3. GPU Information
	fmt.Println(colorBold + colorBlue + "\n[3] GRAPHICS CARD (GPU)" + colorReset)
	showGPUInfo()

	// 4. Disk Information
	fmt.Println(colorBold + colorBlue + "\n[4] STORAGE" + colorReset)
	showDiskInfo()

	// 5. Top Processes
	fmt.Println(colorBold + colorBlue + "\n[5] MOST ACTIVE PROCESSES" + colorReset)
	showTopProcesses(10)

	// Footer with tips
	fmt.Println(colorBold + colorYellow + "\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + colorReset)
	fmt.Println(colorCyan + "\n💡 Tip: Use 'gomonitor --help' to see all available options" + colorReset)
	fmt.Println()
}

// showCPUInfo shows detailed information about the CPU
func showCPUInfo() {
	// Get general CPU statistics
	stats, err := cpu.GetGeneralStats()
	if err != nil {
		fmt.Printf(colorRed+"Error getting CPU information: %v\n"+colorReset, err)
		return
	}

	// Print general statistics
	cpu.PrintGeneralStats(stats)

	// Show top 5 processes by CPU usage
	fmt.Println(colorPurple + "\n→ Top 5 Processes by CPU Usage:" + colorReset)
	if err := cpu.PrintTopProcessesByCPU(5); err != nil {
		fmt.Printf(colorRed+"Error getting processes: %v\n"+colorReset, err)
	}
}

// showRAMInfo shows detailed information about RAM
func showRAMInfo() {
	// Get general RAM statistics
	stats, err := ram.GetRamGeneral()
	if err != nil {
		fmt.Printf(colorRed+"Error getting RAM information: %v\n"+colorReset, err)
		return
	}

	// Print general statistics
	ram.PrintGeneralStats(stats)

	// Show Swap information
	fmt.Println(colorPurple + "\n→ Swap Memory:" + colorReset)
	if err := ram.PrintSwapStats(); err != nil {
		fmt.Printf(colorRed+"Error getting swap information: %v\n"+colorReset, err)
	}

	// Show top 5 processes by RAM usage
	fmt.Println(colorPurple + "\n→ Top 5 Processes by RAM Usage:" + colorReset)
	if err := ram.PrintTopProcessesByRAM(5); err != nil {
		fmt.Printf(colorRed+"Error getting processes: %v\n"+colorReset, err)
	}
}

// showGPUInfo shows information about the GPU
func showGPUInfo() {
	// Get GPU statistics
	stats, err := gpu.GetGPUStats()
	if err != nil {
		fmt.Printf(colorYellow+"⚠ Could not detect GPU: %v\n"+colorReset, err)
		return
	}

	// Print GPU statistics
	gpu.PrintGPUStats(stats)
}

// showDiskInfo shows information about disks
func showDiskInfo() {
	// Show total statistics
	if err := disk.PrintTotalStorageStats(); err != nil {
		fmt.Printf(colorRed+"Error getting total statistics: %v\n"+colorReset, err)
		return
	}

	// Show all devices
	fmt.Println(colorPurple + "\n→ Individual Devices:" + colorReset)
	if err := disk.PrintStorageDevices(); err != nil {
		fmt.Printf(colorRed+"Error getting devices: %v\n"+colorReset, err)
	}
}

// showTopProcesses shows the N most active processes in the system
// Sorted by CPU usage
func showTopProcesses(n int) {
	if err := pck.PrintTopProcesses(n); err != nil {
		fmt.Printf(colorRed+"Error getting processes: %v\n"+colorReset, err)
	}
}

// Auxiliary function to get process association statistics
// (maintained for compatibility with existing code)
func getProcessAssociationStats() {
	fmt.Println(colorBold + colorGreen + "\n=== Process Association Statistics ===" + colorReset)

	// Get all processes
	processes, err := pck.GetProcessAssociation()
	if err != nil {
		fmt.Printf(colorRed+"Error getting processes: %v\n"+colorReset, err)
		return
	}

	fmt.Printf("\n"+colorCyan+"Total monitored processes: "+colorReset+"%d\n", len(processes))

	// Calculate aggregate statistics
	var totalCPU float64
	var totalRAM float32

	for _, p := range processes {
		totalCPU += p.CPUPercentage
		totalRAM += p.RAMPercentage
	}

	fmt.Printf(colorYellow+"Total CPU usage (sum of all processes): "+colorReset+"%.2f%%\n", totalCPU)
	fmt.Printf(colorYellow+"Total RAM usage (sum of all processes): "+colorReset+"%.2f%%\n", totalRAM)

	// Show example of specific process
	if len(processes) > 0 {
		fmt.Println(colorPurple + "\n→ Example - First Process:" + colorReset)
		p := processes[0]
		fmt.Printf("  PID:  %d\n", p.PID)
		fmt.Printf("  Name: %s\n", p.Name)
		fmt.Printf("  CPU:  %.2f%%\n", p.CPUPercentage)
		fmt.Printf("  RAM:  %.2f%% (%s)\n", p.RAMPercentage, common.FormatBytes(p.RAMBytes))
	}
}

// showDefaultInterface shows the default style interface
// GoMonitor logo on the left and system information on the right
func showDefaultInterface() {
	if err := ui.PrintDefaultStyle(); err != nil {
		fmt.Printf(colorRed+"Error showing default interface: %v\n"+colorReset, err)
	}
}

// showInteractiveTUI starts the interactive TUI interface
// Allows navigating through processes, killing processes, sorting, etc.
func showInteractiveTUI() {
	// Check if we're in an interactive terminal
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		fmt.Printf(colorRed+"Error: Could not access stdin: %v\n"+colorReset, err)
		fmt.Println(colorYellow + "\nInteractive mode requires a TTY terminal." + colorReset)
		fmt.Println("Use: gomonitor --all  to see information without interactivity")
		return
	}

	if (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		fmt.Printf(colorRed + "Error: Interactive mode requires a TTY terminal.\n" + colorReset)
		fmt.Println(colorYellow + "It seems that input is being redirected or executed in a pipe." + colorReset)
		fmt.Println("\nUse: gomonitor --all  to see information without interactivity")
		return
	}

	tui := ui.NewInteractiveTUI()
	if err := tui.Run(); err != nil {
		fmt.Printf(colorRed+"\nError running interactive interface: %v\n"+colorReset, err)
		fmt.Println(colorYellow + "\nTip: Make sure you're running in a real interactive terminal." + colorReset)
	}
}
