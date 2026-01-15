package ui

import (
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"
	"unsafe"

	"github.com/dfialho05/GoMonitor/application/pck/common"
)

// ANSI escape code constants
const (
	// Colors
	resetColor   = "\033[0m"
	redColor     = "\033[31m"
	greenColor   = "\033[32m"
	yellowColor  = "\033[33m"
	blueColor    = "\033[34m"
	magentaColor = "\033[35m"
	cyanColor    = "\033[36m"
	whiteColor   = "\033[37m"
	boldColor    = "\033[1m"

	// Background colors
	bgBlack   = "\033[40m"
	bgRed     = "\033[41m"
	bgGreen   = "\033[42m"
	bgYellow  = "\033[43m"
	bgBlue    = "\033[44m"
	bgMagenta = "\033[45m"
	bgCyan    = "\033[46m"
	bgWhite   = "\033[47m"

	// Cursor controls
	clearScreen   = "\033[2J"
	moveCursor    = "\033[%d;%dH"
	hideCursor    = "\033[?25l"
	showCursor    = "\033[?25h"
	clearLine     = "\033[2K"
	saveCursor    = "\033[s"
	restoreCursor = "\033[u"
)

// SortMode defines the process sorting mode
type SortMode int

const (
	SortByCPU SortMode = iota // Sort by CPU usage
	SortByRAM                 // Sort by RAM usage
	SortByPID                 // Sort by PID
)

// InteractiveTUI represents the interactive TUI interface
type InteractiveTUI struct {
	processes     []common.ProcessInfo // Process list
	selectedIndex int                  // Selected process index
	scrollOffset  int                  // Scroll offset
	sortMode      SortMode             // Current sort mode
	running       bool                 // Flag to control main loop
	width         int                  // Terminal width
	height        int                  // Terminal height
}

// NewInteractiveTUI creates a new TUI interface instance
// Returns a pointer to configured InteractiveTUI
func NewInteractiveTUI() *InteractiveTUI {
	return &InteractiveTUI{
		selectedIndex: 0,
		scrollOffset:  0,
		sortMode:      SortByCPU,
		running:       true,
		width:         120,
		height:        30,
	}
}

// Run starts the interactive TUI interface
// This is the main method that controls the entire interface flow
func (tui *InteractiveTUI) Run() error {
	// Configure terminal for raw mode (capture keys without buffer)
	oldState, err := setRawMode()
	if err != nil {
		return fmt.Errorf("error configuring terminal: %w", err)
	}
	defer restoreTerminal(oldState)

	// Hide cursor
	fmt.Print(hideCursor)
	defer fmt.Print(showCursor)

	// Configure Ctrl+C handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Channel for key capture
	keyChan := make(chan byte, 10)
	go tui.captureKeys(keyChan)

	// First data update
	tui.updateProcesses()
	tui.render()

	// Main interface loop
	for tui.running {
		// Wait for events
		select {
		case <-sigChan:
			// Ctrl+C pressed - exit
			tui.running = false

		case key := <-keyChan:
			// Process pressed key
			tui.handleKey(key)

		default:
			time.Sleep(50 * time.Millisecond)
		}
	}

	// Clear screen and show cursor before exiting
	fmt.Print(clearScreen)
	fmt.Printf(moveCursor, 1, 1)
	fmt.Print(showCursor)

	return nil
}

// updateProcesses updates the process list and sorts according to current mode
func (tui *InteractiveTUI) updateProcesses() {
	// Collect all processes
	processes, err := common.CollectAllProcessInfo()
	if err != nil {
		return
	}

	// Sort according to selected mode
	tui.sortProcesses(processes)

	// Update the list
	tui.processes = processes

	// Adjust selected index if necessary
	if tui.selectedIndex >= len(tui.processes) {
		tui.selectedIndex = len(tui.processes) - 1
	}
	if tui.selectedIndex < 0 {
		tui.selectedIndex = 0
	}
}

// sortProcesses sorts the process list according to current mode
func (tui *InteractiveTUI) sortProcesses(processes []common.ProcessInfo) {
	switch tui.sortMode {
	case SortByCPU:
		sort.Slice(processes, func(i, j int) bool {
			return processes[i].CPUPercentage > processes[j].CPUPercentage
		})
	case SortByRAM:
		sort.Slice(processes, func(i, j int) bool {
			return processes[i].RAMPercentage > processes[j].RAMPercentage
		})
	case SortByPID:
		sort.Slice(processes, func(i, j int) bool {
			return processes[i].PID < processes[j].PID
		})
	}
}

// render renders the entire interface on screen
func (tui *InteractiveTUI) render() {
	// Clear screen
	fmt.Print(clearScreen)
	fmt.Printf(moveCursor, 1, 1)

	// Render header
	tui.renderHeader()

	// Render info bar
	tui.renderInfoBar()

	// Render table header
	tui.renderTableHeader()

	// Render process list
	tui.renderProcessList()

	// Render footer with controls
	tui.renderFooter()
}

// renderHeader renders the header with logo
func (tui *InteractiveTUI) renderHeader() {
	fmt.Println(cyanColor + boldColor + "╔════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╗" + resetColor)
	fmt.Println(cyanColor + boldColor + "║" + greenColor + "    ██████╗  ██████╗ ███╗   ███╗" + cyanColor + "                    GOMONITOR - Interactive Process Manager                    " + "║" + resetColor)
	fmt.Println(cyanColor + boldColor + "║" + greenColor + "   ██╔════╝ ██╔═══██╗████╗ ████║" + cyanColor + "                     Real-time System Resource Monitor                         " + "║" + resetColor)
	fmt.Println(cyanColor + boldColor + "║" + greenColor + "   ██║  ███╗██║   ██║██╔████╔██║" + cyanColor + "                                                                               " + "║" + resetColor)
	fmt.Println(cyanColor + boldColor + "║" + greenColor + "   ██║   ██║██║   ██║██║╚██╔╝██║" + cyanColor + "                                                                               " + "║" + resetColor)
	fmt.Println(cyanColor + boldColor + "║" + greenColor + "   ╚██████╔╝╚██████╔╝██║ ╚═╝ ██║" + cyanColor + "                                                                               " + "║" + resetColor)
	fmt.Println(cyanColor + boldColor + "║" + greenColor + "    ╚═════╝  ╚═════╝ ╚═╝     ╚═╝" + cyanColor + "                                                                               " + "║" + resetColor)
	fmt.Println(cyanColor + boldColor + "╚════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╝" + resetColor)
	fmt.Println()
}

// renderInfoBar renders the bar with system information
func (tui *InteractiveTUI) renderInfoBar() {
	// Calculate total statistics
	var totalCPU float64
	var totalRAM float32
	processCount := len(tui.processes)

	for _, p := range tui.processes {
		totalCPU += p.CPUPercentage
		totalRAM += p.RAMPercentage
	}

	// Get total memory information
	totalMemory, err := common.GetSystemMemoryTotal()
	totalMemoryGB := 0.0
	if err == nil {
		totalMemoryGB = float64(totalMemory) / 1024 / 1024 / 1024
	}

	// Current sort mode
	sortModeStr := ""
	switch tui.sortMode {
	case SortByCPU:
		sortModeStr = yellowColor + "CPU ▼" + resetColor
	case SortByRAM:
		sortModeStr = yellowColor + "RAM ▼" + resetColor
	case SortByPID:
		sortModeStr = yellowColor + "PID ▲" + resetColor
	}

	fmt.Printf("  %s%sProcesses:%s %d  ", boldColor, cyanColor, resetColor, processCount)
	fmt.Printf("%s%sTotal CPU:%s %.2f%%  ", boldColor, greenColor, resetColor, totalCPU)
	fmt.Printf("%s%sTotal RAM:%s %.2f%% (%.2f GB)  ", boldColor, magentaColor, resetColor, totalRAM, totalMemoryGB)
	fmt.Printf("%s%sSort by:%s %s", boldColor, whiteColor, resetColor, sortModeStr)
	fmt.Println()
	fmt.Println()
}

// renderTableHeader renders the process table header
func (tui *InteractiveTUI) renderTableHeader() {
	fmt.Print(boldColor)
	fmt.Printf("  %-8s %-35s %10s %10s %15s\n", "PID", "NAME", "CPU %", "RAM %", "MEMORY")
	fmt.Print(resetColor)
	fmt.Println("  " + "─────────────────────────────────────────────────────────────────────────────────────────────────────────────────")
}

// renderProcessList renders the process list with scroll
func (tui *InteractiveTUI) renderProcessList() {
	// Determine how many lines we can show (height - headers - footer)
	maxLines := 20

	// Adjust scroll offset if necessary
	if tui.selectedIndex < tui.scrollOffset {
		tui.scrollOffset = tui.selectedIndex
	}
	if tui.selectedIndex >= tui.scrollOffset+maxLines {
		tui.scrollOffset = tui.selectedIndex - maxLines + 1
	}

	// Render visible processes
	for i := 0; i < maxLines && i+tui.scrollOffset < len(tui.processes); i++ {
		index := i + tui.scrollOffset
		p := tui.processes[index]

		// Check if this process is selected
		isSelected := index == tui.selectedIndex

		// Apply selection style
		if isSelected {
			fmt.Print(bgBlue + whiteColor + boldColor)
		}

		// Format memory
		memoryStr := common.FormatBytes(p.RAMBytes)

		// Truncate name if necessary
		name := p.Name
		if len(name) > 35 {
			name = name[:32] + "..."
		}

		// Print process line
		fmt.Printf("  %-8d %-35s %9.2f%% %9.2f%% %15s", p.PID, name, p.CPUPercentage, p.RAMPercentage, memoryStr)

		if isSelected {
			fmt.Print(resetColor)
		}
		fmt.Println()
	}

	// Fill empty lines if necessary
	visibleCount := maxLines
	if len(tui.processes)-tui.scrollOffset < maxLines {
		visibleCount = len(tui.processes) - tui.scrollOffset
	}
	for i := visibleCount; i < maxLines; i++ {
		fmt.Println()
	}
}

// renderFooter renders the footer with control instructions
func (tui *InteractiveTUI) renderFooter() {
	fmt.Println()
	fmt.Println("  " + "─────────────────────────────────────────────────────────────────────────────────────────────────────────────────")
	fmt.Printf("  %s[↑/↓]%s Navigate  ", cyanColor+boldColor, resetColor)
	fmt.Printf("%s[F5/R]%s Refresh  ", yellowColor+boldColor, resetColor)
	fmt.Printf("%s[C]%s CPU  ", greenColor+boldColor, resetColor)
	fmt.Printf("%s[M]%s RAM  ", magentaColor+boldColor, resetColor)
	fmt.Printf("%s[P]%s PID  ", yellowColor+boldColor, resetColor)
	fmt.Printf("%s[D/DEL]%s Kill Process  ", redColor+boldColor, resetColor)
	fmt.Printf("%s[Q/ESC]%s Quit", whiteColor+boldColor, resetColor)
	fmt.Println()
}

// handleKey processes a pressed key
func (tui *InteractiveTUI) handleKey(key byte) {
	switch key {
	case 'q', 'Q', 27: // q, Q or ESC
		tui.running = false

	case 65: // Up arrow
		if tui.selectedIndex > 0 {
			tui.selectedIndex--
		}
		tui.render()

	case 66: // Down arrow
		if tui.selectedIndex < len(tui.processes)-1 {
			tui.selectedIndex++
		}
		tui.render()

	case 'r', 'R': // Refresh
		tui.updateProcesses()
		tui.render()

	case 'c', 'C': // Sort by CPU
		tui.sortMode = SortByCPU
		tui.updateProcesses()
		tui.render()

	case 'm', 'M': // Sort by RAM (Memory)
		tui.sortMode = SortByRAM
		tui.updateProcesses()
		tui.render()

	case 'p', 'P': // Sort by PID
		tui.sortMode = SortByPID
		tui.updateProcesses()
		tui.render()

	case 127, 'd', 'D': // Delete or D - kill process
		tui.killSelectedProcess()
		tui.render()
	}
}

// killSelectedProcess kills the selected process using the system's kill command
func (tui *InteractiveTUI) killSelectedProcess() {
	if tui.selectedIndex < 0 || tui.selectedIndex >= len(tui.processes) {
		return
	}

	selectedProcess := tui.processes[tui.selectedIndex]
	pid := selectedProcess.PID

	// Use system's kill command to kill the process
	// First try SIGTERM (15) for graceful termination
	err := syscall.Kill(int(pid), syscall.SIGTERM)

	// If SIGTERM fails, try SIGKILL (9) for force
	if err != nil {
		syscall.Kill(int(pid), syscall.SIGKILL)
	}

	// Wait a bit and update the process list
	time.Sleep(100 * time.Millisecond)
	tui.updateProcesses()
}

// captureKeys captures keys from the terminal in raw mode
func (tui *InteractiveTUI) captureKeys(keyChan chan byte) {
	buf := make([]byte, 6)
	for tui.running {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			continue
		}

		if n > 0 {
			// Detect escape sequences (arrows and F-keys)
			if buf[0] == 27 && n >= 3 {
				// F5 key: ESC [ 1 5 ~
				if n >= 5 && buf[1] == '[' && buf[2] == '1' && buf[3] == '5' && buf[4] == '~' {
					keyChan <- 'r' // Treat F5 as refresh (same as 'R')
					// Escape sequence for arrows: ESC [ A/B/C/D
				} else if buf[1] == '[' {
					keyChan <- buf[2] // A=65 (↑), B=66 (↓), C=67 (→), D=68 (←)
				} else {
					keyChan <- buf[0] // Simple ESC
				}
			} else {
				keyChan <- buf[0]
			}
		}
	}
}

// setRawMode configures the terminal in raw mode to capture keys
func setRawMode() (*syscall.Termios, error) {
	// Get current terminal settings
	var oldState syscall.Termios
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		syscall.TCGETS,
		uintptr(unsafe.Pointer(&oldState))); err != 0 {
		return nil, err
	}

	// Create new configuration based on old one
	newState := oldState

	// Disable canonical mode and echo
	newState.Lflag &^= syscall.ICANON | syscall.ECHO | syscall.ISIG

	// Apply new configuration
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		syscall.TCSETS,
		uintptr(unsafe.Pointer(&newState))); err != 0 {
		return nil, err
	}

	return &oldState, nil
}

// restoreTerminal restores the terminal to its original state
func restoreTerminal(oldState *syscall.Termios) {
	if oldState != nil {
		syscall.Syscall(syscall.SYS_IOCTL,
			uintptr(syscall.Stdin),
			syscall.TCSETS,
			uintptr(unsafe.Pointer(oldState)))
	}
}
