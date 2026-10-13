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
	"fmt"
	"io/fs"
	"maps"
	"math"
	"os"
	"strconv"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withRoot points the package-level `root` filesystem at a temporary in-memory
// directory populated with the given relative paths/contents, and restores
// the previous root once the test finishes.
func withRoot(t *testing.T, files map[string]string) {
	t.Helper()
	orig := root
	t.Cleanup(func() { root = orig })
	root = fstest.MapFS(maps.Collect(
		func(yield func(string, *fstest.MapFile) bool) {
			for k, v := range files {
				if !yield(k, &fstest.MapFile{
					Data:    []byte(v),
					Mode:    fs.FileMode(0600),
					ModTime: time.Now(),
				}) {
					return
				}
			}
		},
	))
}

func TestSourceString(t *testing.T) {
	testCases := []struct {
		src      Source
		expected string
	}{
		{sourceCgroupV2, "cgroup v2 (memory.max)"},
		{sourceCgroupV1, "cgroup v1 (memory.limit_in_bytes)"},
		{sourceProcMeminfo, "/proc/meminfo (node-wide, unbounded)"},
		{Source(0), "unknown"},
		{Source(99), "unknown"},
	}
	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.src.String())
		})
	}
}

func TestSourceUsage(t *testing.T) {
	t.Run("cgroup v2", func(t *testing.T) {
		withRoot(t, map[string]string{
			"sys/fs/cgroup/memory.current": "104857600\n",
			"sys/fs/cgroup/memory.stat":    "active_file 1000\ninactive_file 10485760\n",
		})
		usage, err := sourceCgroupV2.Usage()
		require.NoError(t, err)
		assert.Equal(t, uint64(104857600-10485760), usage)
	})

	t.Run("cgroup v2 missing memory.current", func(t *testing.T) {
		withRoot(t, map[string]string{
			"sys/fs/cgroup/memory.stat": "inactive_file 0\n",
		})
		_, err := sourceCgroupV2.Usage()
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("cgroup v2 missing memory.stat", func(t *testing.T) {
		withRoot(t, map[string]string{
			"sys/fs/cgroup/memory.current": "1000\n",
		})
		_, err := sourceCgroupV2.Usage()
		assert.Error(t, err)
	})

	t.Run("cgroup v1", func(t *testing.T) {
		withRoot(t, map[string]string{
			"sys/fs/cgroup/memory/memory.usage_in_bytes": "52428800\n",
			"sys/fs/cgroup/memory/memory.stat":           "total_inactive_file 2097152\n",
		})
		usage, err := sourceCgroupV1.Usage()
		require.NoError(t, err)
		assert.Equal(t, uint64(52428800-2097152), usage)
	})

	t.Run("cgroup v1 missing usage_in_bytes", func(t *testing.T) {
		withRoot(t, map[string]string{
			"sys/fs/cgroup/memory/memory.stat": "total_inactive_file 0\n",
		})
		_, err := sourceCgroupV1.Usage()
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("cgroup v1 missing memory.stat", func(t *testing.T) {
		withRoot(t, map[string]string{
			"sys/fs/cgroup/memory/memory.usage_in_bytes": "1000\n",
		})
		_, err := sourceCgroupV1.Usage()
		assert.Error(t, err)
	})

	t.Run("proc/meminfo (resident set size)", func(t *testing.T) {
		const pages = 1024
		withRoot(t, map[string]string{
			"proc/self/statm": fmt.Sprintf("2048 %d 100 0 0 0 0\n", pages),
		})
		usage, err := sourceProcMeminfo.Usage()
		require.NoError(t, err)
		assert.Equal(t, uint64(pages*os.Getpagesize()), usage)
	})

	t.Run("proc/meminfo missing statm", func(t *testing.T) {
		withRoot(t, map[string]string{})
		_, err := sourceProcMeminfo.Usage()
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("proc/meminfo malformed statm", func(t *testing.T) {
		withRoot(t, map[string]string{
			"proc/self/statm": "not-enough-fields\n",
		})
		_, err := sourceProcMeminfo.Usage()
		assert.EqualError(t, err, "unexpected statm format")
	})

	t.Run("unknown source", func(t *testing.T) {
		_, err := Source(99).Usage()
		assert.EqualError(t, err, "unknown memory source")
	})
}

func TestSourceLimit(t *testing.T) {
	t.Run("cgroup v2", func(t *testing.T) {
		withRoot(t, map[string]string{
			"sys/fs/cgroup/memory.max": "536870912\n",
		})
		limit, err := sourceCgroupV2.Limit()
		require.NoError(t, err)
		assert.Equal(t, uint64(536870912), limit)
	})

	t.Run("cgroup v2 unbounded", func(t *testing.T) {
		withRoot(t, map[string]string{
			"sys/fs/cgroup/memory.max": "max\n",
		})
		_, err := sourceCgroupV2.Limit()
		assert.ErrorIs(t, err, ErrUnbounded)
	})

	t.Run("cgroup v2 missing memory.max", func(t *testing.T) {
		withRoot(t, map[string]string{})
		_, err := sourceCgroupV2.Limit()
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("cgroup v1", func(t *testing.T) {
		withRoot(t, map[string]string{
			"sys/fs/cgroup/memory/memory.limit_in_bytes": "268435456\n",
		})
		limit, err := sourceCgroupV1.Limit()
		require.NoError(t, err)
		assert.Equal(t, uint64(268435456), limit)
	})

	t.Run("cgroup v1 unbounded (LONG_MAX rounded to page size)", func(t *testing.T) {
		threshold := uint64(math.MaxInt - os.Getpagesize())
		withRoot(t, map[string]string{
			"sys/fs/cgroup/memory/memory.limit_in_bytes": strconv.FormatUint(threshold, 10) + "\n",
		})
		_, err := sourceCgroupV1.Limit()
		assert.ErrorIs(t, err, ErrUnbounded)
	})

	t.Run("cgroup v1 missing memory.limit_in_bytes", func(t *testing.T) {
		withRoot(t, map[string]string{})
		_, err := sourceCgroupV1.Limit()
		assert.Error(t, err)
	})

	t.Run("proc/meminfo", func(t *testing.T) {
		withRoot(t, map[string]string{
			"proc/meminfo": "MemTotal:       16384000 kB\nMemFree:          100 kB\n",
		})
		limit, err := sourceProcMeminfo.Limit()
		require.NoError(t, err)
		assert.Equal(t, uint64(16384000*1024), limit)
	})

	t.Run("proc/meminfo missing MemTotal field", func(t *testing.T) {
		withRoot(t, map[string]string{
			"proc/meminfo": "MemFree: 100 kB\n",
		})
		_, err := sourceProcMeminfo.Limit()
		assert.EqualError(t, err, "MemTotal not found in /proc/meminfo")
	})

	t.Run("proc/meminfo missing file", func(t *testing.T) {
		withRoot(t, map[string]string{})
		_, err := sourceProcMeminfo.Limit()
		assert.Error(t, err)
	})

	t.Run("unknown source", func(t *testing.T) {
		_, err := Source(99).Limit()
		assert.EqualError(t, err, "unknown memory source")
	})
}

func TestLimitBytes(t *testing.T) {
	t.Run("uses cgroup v2 when bounded", func(t *testing.T) {
		withRoot(t, map[string]string{
			"sys/fs/cgroup/memory.max": "1000000\n",
		})
		limit, src, err := LimitBytes()
		require.NoError(t, err)
		assert.Equal(t, uint64(1000000), limit)
		assert.Equal(t, sourceCgroupV2, src)
	})

	t.Run("falls back to cgroup v1 when v2 is absent", func(t *testing.T) {
		withRoot(t, map[string]string{
			"sys/fs/cgroup/memory/memory.limit_in_bytes": "2000000\n",
		})
		limit, src, err := LimitBytes()
		require.NoError(t, err)
		assert.Equal(t, uint64(2000000), limit)
		assert.Equal(t, sourceCgroupV1, src)
	})

	t.Run("skips cgroup v1 and goes straight to meminfo when v2 is unbounded", func(t *testing.T) {
		withRoot(t, map[string]string{
			"sys/fs/cgroup/memory.max":                   "max\n",
			"sys/fs/cgroup/memory/memory.limit_in_bytes": "2000000\n",
			"proc/meminfo":                               "MemTotal:       8000000 kB\n",
		})
		limit, src, err := LimitBytes()
		require.NoError(t, err)
		assert.Equal(t, sourceProcMeminfo, src)
		assert.Equal(t, uint64(8000000*1024), limit)
	})

	t.Run("falls back to meminfo when neither cgroup is present", func(t *testing.T) {
		withRoot(t, map[string]string{
			"proc/meminfo": "MemTotal:       4000000 kB\n",
		})
		limit, src, err := LimitBytes()
		require.NoError(t, err)
		assert.Equal(t, sourceProcMeminfo, src)
		assert.Equal(t, uint64(4000000*1024), limit)
	})

	t.Run("errors when no source is available", func(t *testing.T) {
		withRoot(t, map[string]string{})
		_, _, err := LimitBytes()
		assert.EqualError(t, err, "unable to detect memory limit")
	})
}
