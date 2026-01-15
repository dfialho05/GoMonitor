package disk

import "strings"

// ignoredFsTypes contains a map of filesystem types to ignore
// Uses map[string]struct{} because struct{} doesn't occupy memory space
// Allows O(1) lookups unlike slices which are O(n)
//
// These filesystems are virtual or temporary and do not represent
// real physical storage devices
var ignoredFsTypes = map[string]struct{}{
	"tmpfs":      {}, // Temporary filesystem in RAM
	"devtmpfs":   {}, // Temporary device filesystem
	"sysfs":      {}, // Virtual filesystem for kernel information
	"proc":       {}, // Process filesystem (kernel information)
	"devpts":     {}, // Device pseudo-terminals
	"securityfs": {}, // Security filesystem
	"cgroup":     {}, // Control groups (v1)
	"cgroup2":    {}, // Control groups (v2)
	"pstore":     {}, // Persistent storage for crash logs
	"efivarfs":   {}, // EFI variables
	"bpf":        {}, // Berkeley Packet Filter
	"autofs":     {}, // Auto-mount filesystem
	"mqueue":     {}, // POSIX message queues
	"hugetlbfs":  {}, // Large memory pages
	"debugfs":    {}, // Kernel debug filesystem
	"tracefs":    {}, // Kernel tracing filesystem
	"configfs":   {}, // Configuration filesystem
	"fusectl":    {}, // FUSE (Filesystem in Userspace) control
	"squashfs":   {}, // Compressed read-only filesystem (used by snaps)
}

// ignoredPrefixes contains mountpoint prefixes to ignore
// These are virtual or temporary paths that should not be considered
// as real physical disks
//
// Note: For prefixes we still need to do an O(n) loop, but the list is small
var ignoredPrefixes = []string{
	"/sys",       // Virtual kernel filesystems
	"/proc",      // Process and kernel information
	"/dev",       // Devices (except real mounts)
	"/run/snapd", // Ubuntu/Debian snaps
	"/run/lock",  // Temporary lock files
	"/run/user",  // User runtime directories
	"/snap",      // Mounted snap applications
	"/boot/efi",  // EFI partition (usually small and system)
	"/var/snap",  // Snap data
}

// IsRealDisk checks if a mountpoint represents a real physical disk
// This function filters out virtual, temporary and system filesystems
//
// Parameters:
//   - mountpoint: path where the filesystem is mounted (e.g. "/", "/home")
//   - fstype: filesystem type (e.g. "ext4", "ntfs", "tmpfs")
//
// Returns:
//   - true if it's a real disk that should be monitored
//   - false if it's a virtual/temporary filesystem that should be ignored
//
// Examples:
//   - IsRealDisk("/", "ext4") -> true (root disk)
//   - IsRealDisk("/home", "ext4") -> true (home partition)
//   - IsRealDisk("/dev/shm", "tmpfs") -> false (temporary RAM)
//   - IsRealDisk("/proc", "proc") -> false (virtual filesystem)
func IsRealDisk(mountpoint string, fstype string) bool {
	// 1. Instant check in map (O(1))
	// If the filesystem type is in the ignored list, it's not a real disk
	if _, isIgnored := ignoredFsTypes[fstype]; isIgnored {
		return false
	}

	// 2. Prefix check (O(n), but n is small)
	// If the mountpoint starts with an ignored prefix, it's not a real disk
	for _, prefix := range ignoredPrefixes {
		if strings.HasPrefix(mountpoint, prefix) {
			return false
		}
	}

	// 3. If it passed both checks, it's considered a real disk
	return true
}

// AddIgnoredFsType adds a filesystem type to the ignored list
// Useful for customizing which filesystem types should be filtered
//
// Parameters:
//   - fstype: filesystem type to ignore (e.g. "btrfs", "zfs")
func AddIgnoredFsType(fstype string) {
	ignoredFsTypes[fstype] = struct{}{}
}

// RemoveIgnoredFsType removes a filesystem type from the ignored list
// Useful if you want to monitor a filesystem type that's in the default list
//
// Parameters:
//   - fstype: filesystem type to stop ignoring
func RemoveIgnoredFsType(fstype string) {
	delete(ignoredFsTypes, fstype)
}

// AddIgnoredPrefix adds a path prefix to the ignored list
// Useful for customizing which paths should be filtered
//
// Parameters:
//   - prefix: path prefix to ignore (e.g. "/mnt/temp")
func AddIgnoredPrefix(prefix string) {
	ignoredPrefixes = append(ignoredPrefixes, prefix)
}

// GetIgnoredFsTypes returns a list of all ignored filesystem types
// Useful for debugging or showing the user which types are being filtered
//
// Returns:
//   - slice with all filesystem types in the ignored list
func GetIgnoredFsTypes() []string {
	types := make([]string, 0, len(ignoredFsTypes))
	for fstype := range ignoredFsTypes {
		types = append(types, fstype)
	}
	return types
}

// GetIgnoredPrefixes returns a list of all ignored path prefixes
// Useful for debugging or showing the user which paths are being filtered
//
// Returns:
//   - slice with all path prefixes in the ignored list
func GetIgnoredPrefixes() []string {
	// Return a copy to prevent external modifications
	prefixes := make([]string, len(ignoredPrefixes))
	copy(prefixes, ignoredPrefixes)
	return prefixes
}
