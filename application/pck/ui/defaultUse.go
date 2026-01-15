package ui

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/dfialho05/GoMonitor/application/pck/cpu"
	"github.com/dfialho05/GoMonitor/application/pck/disk"
	"github.com/dfialho05/GoMonitor/application/pck/gpu"
	"github.com/dfialho05/GoMonitor/application/pck/ram"

	"golang.org/x/term"
)

// ANSI color constants
const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorWhite   = "\033[37m"
	colorBold    = "\033[1m"
)

// GOM Horizontal logo
// IMPORTANT: All visual lines must have the same length for alignment to work.
// The box has a visual width of 42 characters.
var logoLines = []string{
	"",
	colorCyan + colorBold + "  ╔════════════════════════════════════════╗" + colorReset,
	colorCyan + colorBold + "  ║                                        ║" + colorReset,
	colorCyan + colorBold + "  ║  " + colorGreen + " ██████╗  ██████╗ ███╗   ███╗ " + colorReset + colorCyan + colorBold + "       ║" + colorReset,
	colorCyan + colorBold + "  ║  " + colorGreen + "██╔════╝ ██╔═══██╗████╗ ████║ " + colorReset + colorCyan + colorBold + "       ║" + colorReset,
	colorCyan + colorBold + "  ║  " + colorGreen + "██║  ███╗██║   ██║██╔████╔██║ " + colorReset + colorCyan + colorBold + "       ║" + colorReset,
	colorCyan + colorBold + "  ║  " + colorGreen + "██║   ██║██║   ██║██║╚██╔╝██║ " + colorReset + colorCyan + colorBold + "       ║" + colorReset,
	colorCyan + colorBold + "  ║  " + colorGreen + "╚██████╔╝╚██████╔╝██║ ╚═╝ ██║ " + colorReset + colorCyan + colorBold + "       ║" + colorReset,
	colorCyan + colorBold + "  ║  " + colorGreen + " ╚═════╝  ╚═════╝ ╚═╝     ╚═╝ " + colorReset + colorCyan + colorBold + "       ║" + colorReset,
	colorCyan + colorBold + "  ║                                        ║" + colorReset,
	colorCyan + colorBold + "  ║                                        ║" + colorReset,
	colorCyan + colorBold + "  ╚════════════════════════════════════════╝" + colorReset,
	"",
}

// System data structure
type SystemInfo struct {
	Username    string
	Hostname    string
	OS          string
	Kernel      string
	Uptime      string
	Shell       string
	CPUModel    string
	CPUCores    int
	CPUUsage    float64
	CPUTemp     int
	RAMTotal    string
	RAMUsed     string
	RAMPercent  float64
	DiskTotal   string
	DiskUsed    string
	DiskPercent float64
	GPUModel    string
	GPUTemp     int
}

// PrintDefaultStyle prints the interface
func PrintDefaultStyle() error {
	sysInfo, err := collectSystemInfo()
	if err != nil {
		return fmt.Errorf("error collecting system information: %w", err)
	}

	infoLines := formatSystemInfo(sysInfo)

	// Detect terminal width
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 100 // safe default value
	}

	fmt.Println()

	// VISUAL SAFETY LOGIC:
	// The box is 44 chars wide. The text is about 40-50 chars.
	// We need at least 100-110 chars of width to avoid wrapping the line.
	// If less than 110, use Vertical mode.
	if width < 110 {
		// Vertical mode (small screen)
		for _, line := range logoLines {
			fmt.Println(line)
		}
		for _, line := range infoLines {
			fmt.Println("   " + line) // Small indentation to look nice
		}
	} else {
		// Side-by-side mode (large screen)
		maxLines := len(logoLines)
		if len(infoLines) > maxLines {
			maxLines = len(infoLines)
		}

		for i := 0; i < maxLines; i++ {
			// Print logo line
			if i < len(logoLines) {
				fmt.Print(logoLines[i])
			} else {
				// 44 spaces to compensate for the logo box width when it ends
				fmt.Print(strings.Repeat(" ", 44))
			}

			// Spacing between logo and text
			fmt.Print("    ")

			// Print info line
			if i < len(infoLines) {
				fmt.Print(infoLines[i])
			}

			fmt.Println()
		}
	}

	fmt.Println()
	return nil
}

// collectSystemInfo gathers the data (same as before)
func collectSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{}

	currentUser, err := user.Current()
	if err == nil {
		info.Username = currentUser.Username
	} else {
		info.Username = "unknown"
	}

	hostname, err := os.Hostname()
	if err == nil {
		info.Hostname = hostname
	} else {
		info.Hostname = "unknown"
	}

	info.OS = getOSInfo()
	info.Kernel = getKernelVersion()
	info.Uptime = getSystemUptime()
	info.Shell = os.Getenv("SHELL")

	cpuStats, err := cpu.GetGeneralStats()
	if err == nil {
		info.CPUModel = cpuStats.ModelName
		info.CPUCores = cpuStats.Cores
		info.CPUUsage = cpuStats.Percentage
		info.CPUTemp = cpuStats.Temperature
	}

	ramStats, err := ram.GetRamGeneral()
	if err == nil {
		info.RAMTotal = formatBytes(ramStats.Total)
		info.RAMUsed = formatBytes(ramStats.Used)
		info.RAMPercent = ramStats.Percent
	}

	diskTotal, diskUsed, _, err := disk.GetTotalStorageStats()
	if err == nil {
		info.DiskTotal = formatBytes(diskTotal)
		info.DiskUsed = formatBytes(diskUsed)
		if diskTotal > 0 {
			info.DiskPercent = (float64(diskUsed) / float64(diskTotal)) * 100
		}
	}

	gpuStats, err := gpu.GetGPUStats()
	if err == nil {
		info.GPUModel = gpuStats.Model
		info.GPUTemp = gpuStats.Temp
	} else {
		info.GPUModel = "Not detected"
		info.GPUTemp = 0
	}

	return info, nil
}

// formatSystemInfo formats the text
func formatSystemInfo(info *SystemInfo) []string {
	lines := []string{}

	// Start with empty line to align with the top of the box
	lines = append(lines, "")

	lines = append(lines, formatInfoLine("OS", info.OS, colorBlue))
	lines = append(lines, formatInfoLine("Kernel", info.Kernel, colorBlue))
	lines = append(lines, formatInfoLine("Uptime", info.Uptime, colorBlue))
	lines = append(lines, formatInfoLine("Shell", info.Shell, colorBlue))

	// More aggressive truncation (25 chars) to avoid line wrap
	cpuInfo := fmt.Sprintf("%s (%d cores)", truncateString(info.CPUModel, 25), info.CPUCores)
	lines = append(lines, formatInfoLine("CPU", cpuInfo, colorCyan))
	lines = append(lines, formatInfoLine("CPU Usage", fmt.Sprintf("%.2f%%", info.CPUUsage), colorCyan))

	if info.CPUTemp > 0 {
		cpuTemp := fmt.Sprintf("%d°C", info.CPUTemp)
		lines = append(lines, formatInfoLine("CPU Temp", cpuTemp, colorCyan))
	}

	ramInfo := fmt.Sprintf("%s / %s (%.0f%%)", info.RAMUsed, info.RAMTotal, info.RAMPercent)
	lines = append(lines, formatInfoLine("RAM", ramInfo, colorYellow))

	diskInfo := fmt.Sprintf("%s / %s (%.0f%%)", info.DiskUsed, info.DiskTotal, info.DiskPercent)
	lines = append(lines, formatInfoLine("Disk", diskInfo, colorMagenta))

	gpuInfo := truncateString(info.GPUModel, 25)
	if info.GPUTemp > 0 {
		gpuInfo = fmt.Sprintf("%s (%d°C)", gpuInfo, info.GPUTemp)
	}
	lines = append(lines, formatInfoLine("GPU", gpuInfo, colorGreen))

	return lines
}

func formatInfoLine(label, value, labelColor string) string {
	return labelColor + colorBold + label + colorReset + ": " + value
}

func formatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)
	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func getSystemUptime() string {
	if runtime.GOOS == "linux" {
		data, err := os.ReadFile("/proc/uptime")
		if err == nil {
			var uptimeSeconds float64
			fmt.Sscanf(string(data), "%f", &uptimeSeconds)
			duration := time.Duration(uptimeSeconds) * time.Second

			// Simplified formatting
			hours := int(duration.Hours())
			minutes := int(duration.Minutes()) % 60

			if hours > 24 {
				days := hours / 24
				hours = hours % 24
				return fmt.Sprintf("%d days, %d hours, %d mins", days, hours, minutes)
			}
			return fmt.Sprintf("%d hours, %d mins", hours, minutes)
		}
	}
	return "unknown"
}

func getOSInfo() string {
	if runtime.GOOS == "linux" {
		data, err := os.ReadFile("/etc/os-release")
		if err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "PRETTY_NAME=") {
					return strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
				}
			}
		}
	}
	return runtime.GOOS
}

func getKernelVersion() string {
	if runtime.GOOS == "linux" {
		data, err := os.ReadFile("/proc/version_signature")
		if err == nil {
			parts := strings.Fields(string(data))
			if len(parts) >= 3 {
				return parts[2]
			}
		}
		// Fallback
		data, err = os.ReadFile("/proc/version")
		if err == nil {
			version := string(data)
			if strings.Contains(version, "Linux version") {
				parts := strings.Split(version, " ")
				if len(parts) >= 3 {
					return parts[2]
				}
			}
		}
	}
	return runtime.Version()
}
