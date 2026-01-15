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

// GoMonitor ASCII Logo
// Each logo line is stored in the slice to facilitate side-by-side printing
var logoLines = []string{
	"",
	colorCyan + colorBold + "            ╔════════════════════════╗" + colorReset,
	colorCyan + colorBold + "            ║                        ║" + colorReset,
	colorCyan + colorBold + "            ║     " + colorGreen + "██████╗  ██████╗" + colorReset + colorCyan + colorBold + "    ║" + colorReset,
	colorCyan + colorBold + "            ║     " + colorGreen + "██╔════╝██╔═══██╗" + colorReset + colorCyan + colorBold + "   ║" + colorReset,
	colorCyan + colorBold + "            ║     " + colorGreen + "██║  ███╗██║   ██║" + colorReset + colorCyan + colorBold + "   ║" + colorReset,
	colorCyan + colorBold + "            ║     " + colorGreen + "██║   ██║██║   ██║" + colorReset + colorCyan + colorBold + "   ║" + colorReset,
	colorCyan + colorBold + "            ║     " + colorGreen + "╚██████╔╝╚██████╔╝" + colorReset + colorCyan + colorBold + "   ║" + colorReset,
	colorCyan + colorBold + "            ║     " + colorGreen + " ╚═════╝  ╚═════╝" + colorReset + colorCyan + colorBold + "    ║" + colorReset,
	colorCyan + colorBold + "            ║                        ║" + colorReset,
	colorCyan + colorBold + "            ║       " + colorYellow + "███╗   ███╗" + colorReset + colorCyan + colorBold + "       ║" + colorReset,
	colorCyan + colorBold + "            ║       " + colorYellow + "████╗ ████║" + colorReset + colorCyan + colorBold + "       ║" + colorReset,
	colorCyan + colorBold + "            ║       " + colorYellow + "██╔████╔██║" + colorReset + colorCyan + colorBold + "       ║" + colorReset,
	colorCyan + colorBold + "            ║       " + colorYellow + "██║╚██╔╝██║" + colorReset + colorCyan + colorBold + "       ║" + colorReset,
	colorCyan + colorBold + "            ║       " + colorYellow + "██║ ╚═╝ ██║" + colorReset + colorCyan + colorBold + "       ║" + colorReset,
	colorCyan + colorBold + "            ║       " + colorYellow + "╚═╝     ╚═╝" + colorReset + colorCyan + colorBold + "       ║" + colorReset,
	colorCyan + colorBold + "            ║                        ║" + colorReset,
	colorCyan + colorBold + "            ║   " + colorWhite + "System Monitor v1.0" + colorReset + colorCyan + colorBold + "  ║" + colorReset,
	colorCyan + colorBold + "            ║                        ║" + colorReset,
	colorCyan + colorBold + "            ╚════════════════════════╝" + colorReset,
	"",
}

// SystemInfo contains all system information to be displayed
type SystemInfo struct {
	Username     string
	Hostname     string
	OS           string
	Kernel       string
	Uptime       string
	Shell        string
	CPUModel     string
	CPUCores     int
	CPUUsage     float64
	CPUTemp      int
	RAMTotal     string
	RAMUsed      string
	RAMPercent   float64
	DiskTotal    string
	DiskUsed     string
	DiskPercent  float64
	GPUModel     string
	GPUTemp      int
	ProcessCount int
}

// PrintDefaultStyle prints the default style interface
// Shows GoMonitor logo on the left and system information on the right
func PrintDefaultStyle() error {
	// Collect all system information
	sysInfo, err := collectSystemInfo()
	if err != nil {
		return fmt.Errorf("error collecting system information: %w", err)
	}

	// Prepare system information lines
	infoLines := formatSystemInfo(sysInfo)

	// Print top separator line
	fmt.Println()

	// Calculate maximum number of lines (logo or info)
	maxLines := len(logoLines)
	if len(infoLines) > maxLines {
		maxLines = len(infoLines)
	}

	// Print logo and information side by side
	for i := 0; i < maxLines; i++ {
		// Print logo line (or empty space if finished)
		if i < len(logoLines) {
			fmt.Print(logoLines[i])
		} else {
			// Empty space the size of the logo (without colors)
			fmt.Print(strings.Repeat(" ", 40))
		}

		// Add spacing between logo and information
		fmt.Print("    ")

		// Print information line (or empty if finished)
		if i < len(infoLines) {
			fmt.Print(infoLines[i])
		}

		fmt.Println()
	}

	// Bottom separator line
	fmt.Println()

	return nil
}

// collectSystemInfo collects all system information
// This function aggregates data from all modules (CPU, RAM, GPU, Disk)
func collectSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{}

	// 1. Get user and hostname information
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

	// 2. Get operating system information
	info.OS = getOSInfo()
	info.Kernel = getKernelVersion()

	// 3. Get system uptime (approximate via /proc/uptime on Linux)
	info.Uptime = getSystemUptime()

	// 4. Get user shell
	info.Shell = os.Getenv("SHELL")
	if info.Shell == "" {
		info.Shell = "unknown"
	}

	// 5. Get CPU information
	cpuStats, err := cpu.GetGeneralStats()
	if err == nil {
		info.CPUModel = cpuStats.ModelName
		info.CPUCores = cpuStats.Cores
		info.CPUUsage = cpuStats.Percentage
		info.CPUTemp = cpuStats.Temperature
	}

	// 6. Get RAM information
	ramStats, err := ram.GetRamGeneral()
	if err == nil {
		info.RAMTotal = formatBytes(ramStats.Total)
		info.RAMUsed = formatBytes(ramStats.Used)
		info.RAMPercent = ramStats.Percent
	}

	// 7. Get Disk information
	diskTotal, diskUsed, _, err := disk.GetTotalStorageStats()
	if err == nil {
		info.DiskTotal = formatBytes(diskTotal)
		info.DiskUsed = formatBytes(diskUsed)
		if diskTotal > 0 {
			info.DiskPercent = (float64(diskUsed) / float64(diskTotal)) * 100
		}
	}

	// 8. Get GPU information
	gpuStats, err := gpu.GetGPUStats()
	if err == nil {
		info.GPUModel = gpuStats.Model
		info.GPUTemp = gpuStats.Temp
	} else {
		info.GPUModel = "Not detected"
		info.GPUTemp = 0
	}

	// 9. Count processes (approximation)
	info.ProcessCount = 0 // Can be implemented if needed

	return info, nil
}

// formatSystemInfo formats system information into text lines
// Each line contains a colored label and its value
func formatSystemInfo(info *SystemInfo) []string {
	lines := []string{}

	// Title line: username@hostname
	titleLine := colorBold + colorGreen + info.Username + colorReset + colorBold + "@" + colorGreen + info.Hostname + colorReset
	lines = append(lines, titleLine)

	// Separator line (dashes the size of the title without colors)
	separatorLength := len(info.Username) + 1 + len(info.Hostname)
	lines = append(lines, colorBold+strings.Repeat("─", separatorLength)+colorReset)

	// Operating System
	lines = append(lines, formatInfoLine("OS", info.OS, colorBlue))

	// Kernel
	lines = append(lines, formatInfoLine("Kernel", info.Kernel, colorBlue))

	// Uptime
	lines = append(lines, formatInfoLine("Uptime", info.Uptime, colorBlue))

	// Shell
	lines = append(lines, formatInfoLine("Shell", info.Shell, colorBlue))

	// CPU
	cpuInfo := fmt.Sprintf("%s (%d cores)", truncateString(info.CPUModel, 40), info.CPUCores)
	lines = append(lines, formatInfoLine("CPU", cpuInfo, colorCyan))

	// CPU Usage
	cpuUsage := fmt.Sprintf("%.2f%%", info.CPUUsage)
	lines = append(lines, formatInfoLine("CPU Usage", cpuUsage, colorCyan))

	// CPU Temperature
	if info.CPUTemp > 0 {
		cpuTemp := fmt.Sprintf("%d°C", info.CPUTemp)
		lines = append(lines, formatInfoLine("CPU Temp", cpuTemp, colorCyan))
	}

	// RAM
	ramInfo := fmt.Sprintf("%s / %s (%.1f%%)", info.RAMUsed, info.RAMTotal, info.RAMPercent)
	lines = append(lines, formatInfoLine("RAM", ramInfo, colorYellow))

	// Disk
	diskInfo := fmt.Sprintf("%s / %s (%.1f%%)", info.DiskUsed, info.DiskTotal, info.DiskPercent)
	lines = append(lines, formatInfoLine("Disk", diskInfo, colorMagenta))

	// GPU
	gpuInfo := truncateString(info.GPUModel, 50)
	if info.GPUTemp > 0 {
		gpuInfo = fmt.Sprintf("%s (%d°C)", truncateString(info.GPUModel, 40), info.GPUTemp)
	}
	lines = append(lines, formatInfoLine("GPU", gpuInfo, colorGreen))

	// Empty line
	lines = append(lines, "")

	// Color bar (default style)
	colorBar := ""
	colors := []string{colorRed, colorYellow, colorGreen, colorCyan, colorBlue, colorMagenta, colorWhite}
	for _, c := range colors {
		colorBar += c + "███" + colorReset
	}
	lines = append(lines, colorBar)

	return lines
}

// formatInfoLine formats an information line with label and value
// Returns a formatted string with colors
func formatInfoLine(label, value, labelColor string) string {
	// Label with color and bold, followed by colon and value
	return labelColor + colorBold + label + colorReset + ": " + value
}

// formatBytes converts bytes to a readable string (KB, MB, GB, TB)
// Helper function to format memory and disk sizes
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

// truncateString truncates a string to a maximum length
// Adds "..." at the end if the string is truncated
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// getSystemUptime gets the system uptime
// This function reads from /proc/uptime on Linux or returns a generic message
func getSystemUptime() string {
	// Try to read /proc/uptime (Linux)
	if runtime.GOOS == "linux" {
		data, err := os.ReadFile("/proc/uptime")
		if err == nil {
			// The first number in /proc/uptime is the uptime in seconds
			var uptimeSeconds float64
			fmt.Sscanf(string(data), "%f", &uptimeSeconds)

			// Convert to days, hours, minutes
			duration := time.Duration(uptimeSeconds) * time.Second
			days := int(duration.Hours() / 24)
			hours := int(duration.Hours()) % 24
			minutes := int(duration.Minutes()) % 60

			if days > 0 {
				return fmt.Sprintf("%d days, %d hours, %d mins", days, hours, minutes)
			} else if hours > 0 {
				return fmt.Sprintf("%d hours, %d mins", hours, minutes)
			} else {
				return fmt.Sprintf("%d mins", minutes)
			}
		}
	}

	// Fallback for other operating systems
	return "unknown"
}

// PrintColorTest prints a test of all available colors
// Useful to check if the terminal supports ANSI colors
func PrintColorTest() {
	fmt.Println("\n" + colorBold + "ANSI Color Test:" + colorReset)
	fmt.Println(colorRed + "■ Red" + colorReset)
	fmt.Println(colorGreen + "■ Green" + colorReset)
	fmt.Println(colorYellow + "■ Yellow" + colorReset)
	fmt.Println(colorBlue + "■ Blue" + colorReset)
	fmt.Println(colorMagenta + "■ Magenta" + colorReset)
	fmt.Println(colorCyan + "■ Cyan" + colorReset)
	fmt.Println(colorWhite + "■ White" + colorReset)
	fmt.Println(colorBold + "■ Bold" + colorReset)
	fmt.Println()
}

// getOSInfo gets detailed operating system information
// Reads /etc/os-release on Linux to get the distribution name
func getOSInfo() string {
	// Try to read /etc/os-release (Linux)
	if runtime.GOOS == "linux" {
		data, err := os.ReadFile("/etc/os-release")
		if err == nil {
			// Look for the PRETTY_NAME line
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "PRETTY_NAME=") {
					// Extract the value between quotes
					osName := strings.TrimPrefix(line, "PRETTY_NAME=")
					osName = strings.Trim(osName, "\"")
					return osName
				}
			}
			// If PRETTY_NAME not found, look for NAME
			for _, line := range lines {
				if strings.HasPrefix(line, "NAME=") {
					osName := strings.TrimPrefix(line, "NAME=")
					osName = strings.Trim(osName, "\"")
					return osName
				}
			}
		}
	}

	// Fallback to generic OS
	switch runtime.GOOS {
	case "linux":
		return "Linux"
	case "darwin":
		return "macOS"
	case "windows":
		return "Windows"
	default:
		return runtime.GOOS
	}
}

// getKernelVersion gets the system kernel version
// On Linux, reads from /proc/version or executes uname -r
func getKernelVersion() string {
	if runtime.GOOS == "linux" {
		// Try to read /proc/version_signature (Ubuntu/Debian)
		data, err := os.ReadFile("/proc/version_signature")
		if err == nil {
			version := strings.TrimSpace(string(data))
			// Get only the version, not all the text
			parts := strings.Fields(version)
			if len(parts) >= 3 {
				return parts[2] // Third field is usually the version
			}
		}

		// Try to read /proc/version
		data, err = os.ReadFile("/proc/version")
		if err == nil {
			version := strings.TrimSpace(string(data))
			// Extract kernel version (usually after "Linux version")
			if strings.Contains(version, "Linux version") {
				parts := strings.Split(version, " ")
				if len(parts) >= 3 {
					return parts[2] // Version is in the third position
				}
			}
		}
	}

	// Fallback to Go version (since we can't easily get the kernel)
	return runtime.Version()
}
