//go:build !solution

package fileleak

import (
	"fmt"
	"os"
	"path/filepath"
)

type testingT interface {
	Errorf(msg string, args ...interface{})
	Cleanup(func())
}

// VerifyNone checks if the test leaked any file descriptors.
// It records the currently open file descriptors at the start
// and registers a cleanup function that verifies no new file
// descriptors remain open at the end of the test.
func VerifyNone(t testingT) {
	// Get the initial list of open file descriptors
	startingFDs, err := getOpenFileDescriptors()
	if err != nil {
		t.Errorf("fileleak: failed to get initial file descriptors: %v", err)
		return
	}

	// Register cleanup to check for leaks at the end of the test
	t.Cleanup(func() {
		// Get the final list of open file descriptors
		endingFDs, err := getOpenFileDescriptors()
		if err != nil {
			t.Errorf("fileleak: failed to get final file descriptors: %v", err)
			return
		}

		// Find new file descriptors that weren't open at the start
		var leakedFDs []string
		for fd, path := range endingFDs {
			if startPath, wasOpenAtStart := startingFDs[fd]; !wasOpenAtStart {
				leakedFDs = append(leakedFDs, fmt.Sprintf("%s (fd %s)", path, fd))
			} else if startPath != path {
				// File descriptor was reused for a different file
				leakedFDs = append(leakedFDs, fmt.Sprintf("%s (fd %s, was %s)", path, fd, startPath))
			}
		}

		// Report leaks if any were found
		if len(leakedFDs) > 0 {
			t.Errorf("fileleak: detected %d leaked file descriptors:\n%v", len(leakedFDs), leakedFDs)
		}
	})
}

// getOpenFileDescriptors returns a map of open file descriptors to their target paths
func getOpenFileDescriptors() (map[string]string, error) {
	fdDir := "/proc/self/fd"
	entries, err := os.ReadDir(fdDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", fdDir, err)
	}

	fds := make(map[string]string)
	for _, entry := range entries {
		fdPath := filepath.Join(fdDir, entry.Name())
		target, err := os.Readlink(fdPath)
		if err != nil {
			// Skip descriptors we can't read
			continue
		}
		fds[entry.Name()] = target
	}

	return fds, nil
}
