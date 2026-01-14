package main

import (
	"fmt"

	"github.com/shirou/gopsutil/v3/disk"
)

// StorageGeneral represents general information about a storage device
type StorageGeneral struct {
	Mountpoint string
	Total      uint64
	Used       uint64
	Free       uint64
}

const (
	MinStorageSize = 2 * (1024 * 1024 * 1024) // 2Gb minimum storage size in bytes
)

// GetAllStorageDevices retrieves information about all storage devices on the system
func GetAllStorageDevices() ([]StorageGeneral, error) {
	// Get all disk partitions
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}

	var storageList []StorageGeneral
	for _, partition := range partitions {
		// Skip partitions that are not real disks
		if !IsRealDisk(partition.Mountpoint, partition.Fstype) {
			continue
		}

		// Get usage statistics for the partition
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue
		}

		// Skip disks smaller than the minimum size
		if usage.Total < MinStorageSize {
			continue
		}

		// Add the storage device to the list
		storageList = append(storageList, StorageGeneral{
			Mountpoint: partition.Mountpoint,
			Total:      usage.Total,
			Used:       usage.Used,
			Free:       usage.Free,
		})
	}
	return storageList, nil
}

func main() {
	// Get all storage devices
	storageList, err := GetAllStorageDevices()
	if err != nil {
		fmt.Println("Error getting storage information:", err)
		return
	}

	// Display information for each storage device
	for _, storage := range storageList {
		// Convert bytes to GB: bytes / 1024 / 1024 / 1024 = GB
		totalGB := storage.Total / 1024 / 1024 / 1024
		usedGB := storage.Used / 1024 / 1024 / 1024
		freeGB := storage.Free / 1024 / 1024 / 1024

		fmt.Printf("Disk: %s\n", storage.Mountpoint)
		fmt.Printf("Total: %d GB\n", totalGB)
		fmt.Printf("Used: %d GB\n", usedGB)
		fmt.Printf("Free: %d GB\n\n", freeGB)
	}
}
