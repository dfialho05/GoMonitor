package disk

import (
	"fmt"

	"github.com/dfialho05/GoMonitor/application/pck/common"
	"github.com/shirou/gopsutil/v3/disk"
)

// StorageDevice represents information about a storage device
// This structure contains data about total, used and free space on a disk
type StorageDevice struct {
	Mountpoint string  // Disk mount point (e.g. "/", "/home", "C:\")
	Fstype     string  // File system type (e.g. "ext4", "ntfs", "btrfs")
	Total      uint64  // Total disk space in bytes
	Used       uint64  // Used disk space in bytes
	Free       uint64  // Free disk space in bytes
	Percent    float64 // Usage percentage (0-100%)
}

const (
	// MinStorageSize defines the minimum size to consider a valid disk
	// Disks smaller than 2GB are usually boot or recovery partitions
	MinStorageSize = 2 * (1024 * 1024 * 1024) // 2GB in bytes
)

// GetAllStorageDevices collects information about all storage devices
// This function automatically filters virtual and temporary file systems
//
// Returns:
//   - slice of StorageDevice with all real physical disks in the system
//   - error if unable to get the information
func GetAllStorageDevices() ([]StorageDevice, error) {
	// 1. Get all system partitions
	// false = don't include virtual partitions (but we still need to filter manually)
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, fmt.Errorf("error getting disk partitions: %w", err)
	}

	// 2. Pre-allocate slice with estimated capacity to avoid reallocations
	storageList := make([]StorageDevice, 0, len(partitions))

	// 3. Iterate through each partition and collect its statistics
	for _, partition := range partitions {
		// 3.1. Check if it's a real disk (not virtual/temporary)
		if !IsRealDisk(partition.Mountpoint, partition.Fstype) {
			continue
		}

		// 3.2. Get usage statistics for this partition
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			// If we can't get usage, skip this partition
			// This can happen if the disk is removed or not accessible
			continue
		}

		// 3.3. Filter very small disks (boot partitions, EFI, etc.)
		if usage.Total < MinStorageSize {
			continue
		}

		// 3.4. Add storage device to the list
		storageList = append(storageList, StorageDevice{
			Mountpoint: partition.Mountpoint,
			Fstype:     partition.Fstype,
			Total:      usage.Total,
			Used:       usage.Used,
			Free:       usage.Free,
			Percent:    usage.UsedPercent,
		})
	}

	return storageList, nil
}

// GetStorageByMountpoint gets information about a specific disk by its mount point
// This function is useful for monitoring a specific disk
//
// Parameters:
//   - mountpoint: disk mount point (e.g. "/", "/home", "C:\")
//
// Returns:
//   - pointer to StorageDevice with disk information
//   - error if the disk is not found or not accessible
func GetStorageByMountpoint(mountpoint string) (*StorageDevice, error) {
	// Get usage statistics for the specified mount point
	usage, err := disk.Usage(mountpoint)
	if err != nil {
		return nil, fmt.Errorf("error getting disk information %s: %w", mountpoint, err)
	}

	// Get information about the file system type
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, fmt.Errorf("error getting partitions: %w", err)
	}

	// Search for the partition corresponding to the mount point
	fstype := "unknown"
	for _, partition := range partitions {
		if partition.Mountpoint == mountpoint {
			fstype = partition.Fstype
			break
		}
	}

	// Return disk information
	return &StorageDevice{
		Mountpoint: mountpoint,
		Fstype:     fstype,
		Total:      usage.Total,
		Used:       usage.Used,
		Free:       usage.Free,
		Percent:    usage.UsedPercent,
	}, nil
}

// PrintStorageDevices prints information about all storage devices
// This function presents a formatted table with data from all disks
//
// Returns:
//   - error if unable to get disk data
func PrintStorageDevices() error {
	// Get all storage devices
	devices, err := GetAllStorageDevices()
	if err != nil {
		return err
	}

	// Check if devices were found
	if len(devices) == 0 {
		fmt.Println("\nNo real storage devices found.")
		return nil
	}

	// Print header
	fmt.Printf("\n╔══════════════════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  %-80s  ║\n", "Storage Devices")
	fmt.Printf("╠══════════════════════════════════════════════════════════════════════════════════╣\n")

	// Print each device
	for i, device := range devices {
		if i > 0 {
			fmt.Printf("╟──────────────────────────────────────────────────────────────────────────────────╢\n")
		}

		fmt.Printf("║  Mount Point:       %-58s  ║\n", common.TruncateString(device.Mountpoint, 58))
		fmt.Printf("║  File System:       %-58s  ║\n", device.Fstype)
		fmt.Printf("║  Total:             %-58s  ║\n", common.FormatBytes(device.Total))
		fmt.Printf("║  Used:              %-58s  ║\n", common.FormatBytes(device.Used))
		fmt.Printf("║  Free:              %-58s  ║\n", common.FormatBytes(device.Free))
		fmt.Printf("║  Usage:             %-58.2f %%    ║\n", device.Percent)
	}

	fmt.Printf("╚══════════════════════════════════════════════════════════════════════════════════╝\n")

	return nil
}

// PrintStorageDevice prints information about a single storage device
// This function is useful for showing details of a specific disk
//
// Parameters:
//   - device: StorageDevice with data to present
func PrintStorageDevice(device StorageDevice) {
	fmt.Printf("\n╔══════════════════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  %-80s  ║\n", "Disk Information")
	fmt.Printf("╠══════════════════════════════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║  Mount Point:       %-58s  ║\n", common.TruncateString(device.Mountpoint, 58))
	fmt.Printf("║  File System:       %-58s  ║\n", device.Fstype)
	fmt.Printf("║  Total:             %-58s  ║\n", common.FormatBytes(device.Total))
	fmt.Printf("║  Used:              %-58s  ║\n", common.FormatBytes(device.Used))
	fmt.Printf("║  Free:              %-58s  ║\n", common.FormatBytes(device.Free))
	fmt.Printf("║  Usage:             %-58.2f %%    ║\n", device.Percent)
	fmt.Printf("╚══════════════════════════════════════════════════════════════════════════════════╝\n")
}

// GetTotalStorageStats calculates total statistics from all disks
// This function aggregates information from all storage devices
//
// Returns:
//   - total: sum of all available storage space in bytes
//   - used: sum of all used space in bytes
//   - free: sum of all free space in bytes
//   - error if unable to get the data
func GetTotalStorageStats() (uint64, uint64, uint64, error) {
	devices, err := GetAllStorageDevices()
	if err != nil {
		return 0, 0, 0, err
	}

	var totalSpace, usedSpace, freeSpace uint64

	for _, device := range devices {
		totalSpace += device.Total
		usedSpace += device.Used
		freeSpace += device.Free
	}

	return totalSpace, usedSpace, freeSpace, nil
}

// PrintTotalStorageStats prints aggregated statistics from all disks
// This function shows a summary of total storage space in the system
//
// Returns:
//   - error if unable to get the data
func PrintTotalStorageStats() error {
	total, used, free, err := GetTotalStorageStats()
	if err != nil {
		return err
	}

	percent := 0.0
	if total > 0 {
		percent = (float64(used) / float64(total)) * 100
	}

	fmt.Printf("\n╔══════════════════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  %-80s  ║\n", "Total System Storage")
	fmt.Printf("╠══════════════════════════════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║  Total:             %-58s  ║\n", common.FormatBytes(total))
	fmt.Printf("║  Used:              %-58s  ║\n", common.FormatBytes(used))
	fmt.Printf("║  Free:              %-58s  ║\n", common.FormatBytes(free))
	fmt.Printf("║  Usage:             %-58.2f %%    ║\n", percent)
	fmt.Printf("╚══════════════════════════════════════════════════════════════════════════════════╝\n")

	return nil
}

// GetIOCounters gets I/O statistics (read/write) from disks
// This function provides information about read and write activity
//
// Returns:
//   - map with disk name as key and IOCountersStat as value
//   - error if unable to get the data
func GetIOCounters() (map[string]disk.IOCountersStat, error) {
	ioCounters, err := disk.IOCounters()
	if err != nil {
		return nil, fmt.Errorf("error getting I/O counters: %w", err)
	}
	return ioCounters, nil
}
