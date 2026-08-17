// Copyright 2026 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package memlimit

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"strconv"
	"strings"
)

var ErrUnbounded = errors.New("memory limit unbounded")

var root fs.FS = os.DirFS("/")

type Source int

const (
	_ Source = iota
	sourceCgroupV2
	sourceCgroupV1
	sourceProcMeminfo
)

func (src Source) String() string {
	switch src {
	case sourceCgroupV2:
		return "cgroup v2 (memory.max)"
	case sourceCgroupV1:
		return "cgroup v1 (memory.limit_in_bytes)"
	case sourceProcMeminfo:
		return "/proc/meminfo (node-wide, unbounded)"
	default:
		return "unknown"
	}
}

func readStatField(path, field string) (uint64, error) {
	f, err := root.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 && fields[0] == field {
			return strconv.ParseUint(fields[1], 10, 64)
		}
	}
	return 0, scanner.Err()
}

func residentBytes() (uint64, error) {
	data, err := fs.ReadFile(root, "proc/self/statm")
	if err != nil {
		return 0, err
	}
	fields := strings.Fields(string(data))
	if len(fields) < 2 {
		return 0, errors.New("unexpected statm format")
	}
	pages, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return 0, err
	}
	return pages * uint64(os.Getpagesize()), nil
}

func workingSet(usage, inactiveFile uint64) uint64 {
	if usage > inactiveFile {
		return usage - inactiveFile
	}
	return 0
}

func (src Source) Usage() (uint64, error) {
	switch src {
	case sourceCgroupV2:
		// cgroup v2
		data, err := fs.ReadFile(root, "sys/fs/cgroup/memory.current")
		if err != nil {
			return 0, err
		}
		usage, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
		if err != nil {
			return 0, err
		}
		inactiveFile, err := readStatField("sys/fs/cgroup/memory.stat", "inactive_file")
		if err != nil {
			return 0, err
		}
		return workingSet(usage, inactiveFile), nil
	case sourceCgroupV1:
		// cgroup v1
		data, err := fs.ReadFile(root, "sys/fs/cgroup/memory/memory.usage_in_bytes")
		if err != nil {
			return 0, err
		}
		usage, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
		if err != nil {
			return 0, err
		}
		inactiveFile, err := readStatField(
			"sys/fs/cgroup/memory/memory.stat",
			"total_inactive_file",
		)
		if err != nil {
			return 0, err
		}
		return workingSet(usage, inactiveFile), nil
	case sourceProcMeminfo:
		return residentBytes()
	default:
		return 0, errors.New("unknown memory source")
	}
}

// LimitBytes returns the memory ceiling this process should watermark
// against, along with which tier produced it.
func LimitBytes() (uint64, Source, error) {
	for _, src := range []Source{sourceCgroupV2, sourceCgroupV1} {
		lim, err := src.Limit()
		if err == nil {
			return lim, src, nil
		} else if errors.Is(err, ErrUnbounded) {
			break
		}
	}
	lim, err := sourceProcMeminfo.Limit()
	if err != nil {
		return 0, -1, errors.New("unable to detect memory limit")
	}
	return lim, sourceProcMeminfo, nil
}

func (src Source) Limit() (uint64, error) {
	switch src {
	case sourceCgroupV2:
		data, err := fs.ReadFile(root, "sys/fs/cgroup/memory.max")
		if os.IsNotExist(err) {
			return 0, err
		}
		s := strings.TrimSpace(string(data))
		if s == "max" {
			return 0, ErrUnbounded
		}
		limit, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return 0, err
		}
		return limit, nil
	case sourceCgroupV1:
		data, err := fs.ReadFile(root, "sys/fs/cgroup/memory/memory.limit_in_bytes")
		if err != nil {
			return 0, err
		}
		limit, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
		if err != nil {
			return 0, err
		}
		if limit >= uint64(math.MaxInt-os.Getpagesize()) {
			// if limit is LONG_MAX rounded to page size means unbounded memory
			return 0, ErrUnbounded
		}
		return limit, nil

	case sourceProcMeminfo:
		fd, err := root.Open("proc/meminfo")
		if err != nil {
			return 0, err
		}
		s := bufio.NewScanner(fd)
		for s.Scan() {
			fields := strings.Fields(s.Text())
			if len(fields) < 2 {
				continue
			}
			if fields[0] == "MemTotal:" {
				kb, err := strconv.ParseUint(fields[1], 10, 64)
				if err != nil {
					return 0, err
				}
				return kb * 1024, nil
			}
		}
		if err := s.Err(); err != nil {
			return 0, err
		}
		return 0, errors.New("MemTotal not found in /proc/meminfo")
	}
	return 0, fmt.Errorf("unknown memory source")
}
