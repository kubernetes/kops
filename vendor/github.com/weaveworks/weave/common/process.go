package common

import (
	"os"
	"strconv"
)

func AllPids(procRoot string) ([]int, error) {
	fh, err := os.Open(procRoot)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	dirNames, err := fh.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	pids := make([]int, len(dirNames))
	for _, dirName := range dirNames {
		pid, err := strconv.Atoi(dirName)
		if err != nil { // Only interested in numeric entries - skip /proc/stat, etc.
			continue
		}
		pids = append(pids, pid)
	}
	return pids, err
}
