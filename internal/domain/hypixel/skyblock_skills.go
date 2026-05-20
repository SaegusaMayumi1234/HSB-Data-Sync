package hypixel

import (
	"fmt"
	"strconv"
	"strings"
)

type SkyblockSkillsResponse struct {
	Success     bool   `json:"success"`
	LastUpdated int64  `json:"lastUpdated"`
	Version     string `json:"version"`
}

func IsNewerVersion(currVersion, newVersion string) (bool, error) {
	curr, err := parseVersion(currVersion)
	if err != nil {
		return false, fmt.Errorf("invalid current version %q: %w", currVersion, err)
	}

	next, err := parseVersion(newVersion)
	if err != nil {
		return false, fmt.Errorf("invalid new version %q: %w", newVersion, err)
	}

	// Pad the shorter slice so both have the same length
	for len(curr) < len(next) {
		curr = append(curr, 0)
	}
	for len(next) < len(curr) {
		next = append(next, 0)
	}

	for i := range curr {
		if next[i] > curr[i] {
			return true, nil
		}
		if next[i] < curr[i] {
			return false, nil
		}
	}

	return false, nil // versions are equal
}

// parseVersion splits a version string into a slice of integers.
func parseVersion(v string) ([]int, error) {
	parts := strings.Split(v, ".")
	nums := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("non-numeric segment %q", p)
		}
		if n < 0 {
			return nil, fmt.Errorf("negative segment %d", n)
		}
		nums = append(nums, n)
	}
	return nums, nil
}
