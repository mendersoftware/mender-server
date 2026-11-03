// Copyright 2026 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.
package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func utc(s string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		panic(err)
	}
	return t.UTC()
}

func TestParseTimeRange(t *testing.T) {
	testCases := map[string]struct {
		value    string
		expected TimeRange
		err      bool
	}{
		"date": {
			value:    "2025-01-15",
			expected: TimeRange{utc("2025-01-15T00:00:00Z"), utc("2025-01-16T00:00:00Z")},
		},
		"minute": {
			value:    "2025-01-15T10:30",
			expected: TimeRange{utc("2025-01-15T10:30:00Z"), utc("2025-01-15T10:31:00Z")},
		},
		"minute with offset": {
			value:    "2025-01-15T10:30+02:00",
			expected: TimeRange{utc("2025-01-15T08:30:00Z"), utc("2025-01-15T08:31:00Z")},
		},
		"second": {
			value:    "2025-01-15T10:30:15Z",
			expected: TimeRange{utc("2025-01-15T10:30:15Z"), utc("2025-01-15T10:30:16Z")},
		},
		"second without offset": {
			value:    "2025-01-15T10:30:15",
			expected: TimeRange{utc("2025-01-15T10:30:15Z"), utc("2025-01-15T10:30:16Z")},
		},
		"milliseconds": {
			value:    "2025-01-15T10:30:15.000Z",
			expected: TimeRange{utc("2025-01-15T10:30:15Z"), utc("2025-01-15T10:30:15.001Z")},
		},
		"nanoseconds are truncated": {
			value: "2025-01-15T10:30:15.123456789-01:00",
			expected: TimeRange{
				utc("2025-01-15T11:30:15.123Z"), utc("2025-01-15T11:30:15.124Z"),
			},
		},
		"error, not a date": {
			value: "yesterday",
			err:   true,
		},
		"error, invalid date": {
			value: "2025-02-30",
			err:   true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			r, err := ParseTimeRange(tc.value)
			if tc.err {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.True(t, tc.expected.Start.Equal(r.Start), "start: %s", r.Start)
			assert.True(t, tc.expected.End.Equal(r.End), "end: %s", r.End)
		})
	}
}

func TestFilterPredicateTimeRanges(t *testing.T) {
	day := TimeRange{utc("2025-01-15T00:00:00Z"), utc("2025-01-16T00:00:00Z")}
	nextDay := TimeRange{utc("2025-01-16T00:00:00Z"), utc("2025-01-17T00:00:00Z")}

	testCases := map[string]struct {
		predicate FilterPredicate
		expected  []TimeRange
		err       string
	}{
		"datetime attribute": {
			predicate: FilterPredicate{
				Scope: "system", Attribute: "check_in_time", Type: "$gt", Value: "2025-01-15",
			},
			expected: []TimeRange{day},
		},
		"datetime attribute, $in from JSON": {
			predicate: FilterPredicate{
				Scope: "system", Attribute: "created_ts", Type: "$in",
				Value: []interface{}{"2025-01-15", "2025-01-16"},
			},
			expected: []TimeRange{day, nextDay},
		},
		"datetime attribute, $nin": {
			predicate: FilterPredicate{
				Scope: "system", Attribute: "updated_ts", Type: "$nin",
				Value: []string{"2025-01-15"},
			},
			expected: []TimeRange{day},
		},
		"datetime attribute, empty $in": {
			predicate: FilterPredicate{
				Scope: "system", Attribute: "updated_ts", Type: "$in", Value: []interface{}{},
			},
		},
		"datetime attribute, $exists is left alone": {
			predicate: FilterPredicate{
				Scope: "system", Attribute: "check_in_time", Type: "$exists", Value: true,
			},
		},
		"datetime attribute, $regex is left alone": {
			predicate: FilterPredicate{
				Scope: "system", Attribute: "check_in_time", Type: "$regex", Value: "2025",
			},
		},
		"same name in another scope is not a datetime": {
			predicate: FilterPredicate{
				Scope: "inventory", Attribute: "check_in_time", Type: "$eq", Value: "2025-01-15",
			},
		},
		"other attribute": {
			predicate: FilterPredicate{
				Scope: "inventory", Attribute: "build_date", Type: "$eq", Value: "2025-01-15",
			},
		},
		"error, not a string": {
			predicate: FilterPredicate{
				Scope: "system", Attribute: "check_in_time", Type: "$lt",
				Value: float64(1736899200),
			},
			err: "system/check_in_time expects dates or timestamps as strings",
		},
		"error, $in without array": {
			predicate: FilterPredicate{
				Scope: "system", Attribute: "check_in_time", Type: "$in", Value: "2025-01-15",
			},
			err: "$in expects an array of dates or timestamps",
		},
		"error, invalid array element": {
			predicate: FilterPredicate{
				Scope: "system", Attribute: "check_in_time", Type: "$in",
				Value: []interface{}{"2025-01-15", "soon"},
			},
			err: `"soon" is not a date or timestamp (e.g. 2025-01-15 or 2025-01-15T10:00:00Z)`,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ranges, err := tc.predicate.TimeRanges()
			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expected, ranges)
		})
	}
}
