package git

import (
	"fmt"
	"regexp"
	"strconv"
)

// CheckGitVersion verifies that the installed git version meets the minimum requirements
func CheckGitVersion(minMajor, minMinor int) error {
	output, err := Run("--version")
	if err != nil {
		return fmt.Errorf("failed to get git version: %w", err)
	}

	major, minor, err := parseGitVersion(output)
	if err != nil {
		return err
	}

	if major < minMajor || (major == minMajor && minor < minMinor) {
		return fmt.Errorf("git version %d.%d is required, but found %d.%d", minMajor, minMinor, major, minor)
	}

	return nil
}

// parseGitVersion extracts major and minor version from git --version output
// e.g., "git version 2.39.0" -> (2, 39, nil)
func parseGitVersion(output string) (major, minor int, err error) {
	re := regexp.MustCompile(`git version (\d+)\.(\d+)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) < 3 {
		return 0, 0, fmt.Errorf("unable to parse git version from: %s", output)
	}

	major, err = strconv.Atoi(matches[1])
	if err != nil {
		return 0, 0, fmt.Errorf("unable to parse major version: %w", err)
	}

	minor, err = strconv.Atoi(matches[2])
	if err != nil {
		return 0, 0, fmt.Errorf("unable to parse minor version: %w", err)
	}

	return major, minor, nil
}
