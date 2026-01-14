package main

import "strings"

// Convert slices to a map for O(1) lookups
// Use struct{} because it occupies no space — we only need the key
var ignoredFsTypes = map[string]struct{}{
	"tmpfs":      {},
	"devtmpfs":   {},
	"sysfs":      {},
	"proc":       {},
	"devpts":     {},
	"securityfs": {},
	"cgroup":     {},
	"cgroup2":    {},
	"pstore":     {},
	"efivarfs":   {},
	"bpf":        {},
	"autofs":     {},
	"mqueue":     {},
	"hugetlbfs":  {},
	"debugfs":    {},
	"tracefs":    {},
	"configfs":   {},
	"fusectl":    {},
	"squashfs":   {},
}

// For prefixes we still need to loop with strings.HasPrefix,
// but we can keep the list as small as possible.
var ignoredPrefixes = []string{
	"/sys", "/proc", "/dev", "/run/snapd", "/run/lock", "/run/user", "/snap",
}

// IsRealDisk contains the optimized validation logic
func IsRealDisk(mountpoint string, fstype string) bool {
	// 1. Instant lookup in the map (O(1))
	if _, isIgnored := ignoredFsTypes[fstype]; isIgnored {
		return false
	}

	// 2. Prefix check (still O(n), but the list is small)
	for _, prefix := range ignoredPrefixes {
		if strings.HasPrefix(mountpoint, prefix) {
			return false
		}
	}

	return true
}
