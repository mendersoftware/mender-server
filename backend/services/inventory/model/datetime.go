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
	"slices"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// dateTimeAttributes are the attributes stored as dates: the timestamps
// maintained by inventory and the check-in time synchronized from deviceauth.
var dateTimeAttributes = map[string][]string{
	AttrScopeSystem: {AttrNameCreated, AttrNameUpdated, AttrNameCheckInTime},
}

// dateTimeOperators are the filter operators comparing an attribute to a value.
var dateTimeOperators = []string{
	"$eq", "$ne", "$gt", "$gte", "$lt", "$lte", "$ltne", "$in", "$nin",
}

// TimeRange is the half-open interval [Start, End).
type TimeRange struct {
	Start time.Time
	End   time.Time
}

func IsDateTimeAttribute(scope, name string) bool {
	return slices.Contains(dateTimeAttributes[scope], name)
}

// periodEnd returns the end of the period starting at start.
type periodEnd func(start time.Time) time.Time

var (
	nextMonth periodEnd = func(t time.Time) time.Time { return t.AddDate(0, 1, 0) }
	nextDay   periodEnd = func(t time.Time) time.Time { return t.AddDate(0, 0, 1) }
)

func after(d time.Duration) periodEnd {
	return func(t time.Time) time.Time { return t.Add(d) }
}

// padDateTime expands a month, a date or a truncated RFC 3339 timestamp into
// a full RFC 3339 timestamp and returns the end of the period the precision
// of the value implies. Accepted forms, each optionally followed by "Z" or
// "±HH:MM" - "Z" (UTC) is assumed if neither is given:
//
//	2006-01                        a month
//	2006-01-02                     a day
//	2006-01-02T15:04               a minute
//	2006-01-02T15:04:05            a second
//	2006-01-02T15:04:05.000        a millisecond
//	2006-01-02T15:04:05.000000     a millisecond, dates are stored at
//	2006-01-02T15:04:05.000000000  millisecond resolution
func padDateTime(value string) (string, periodEnd, error) {
	// split off the zone designator, defaulting to UTC; dates and months
	// always end in "-DD" or "-MM", which can't be mistaken for an offset
	rest, zone := value, "Z"
	switch n := len(value); {
	case n > 0 && value[n-1] == 'Z':
		rest = value[:n-1]
	case n >= 6 && (value[n-6] == '+' || value[n-6] == '-') && value[n-3] == ':':
		rest, zone = value[:n-6], value[n-6:]
	}

	date, clock, hasClock := strings.Cut(rest, "T")
	if !hasClock {
		switch len(date) {
		case len("2006-01"):
			return date + "-01T00:00:00" + zone, nextMonth, nil
		case len(time.DateOnly):
			return date + "T00:00:00" + zone, nextDay, nil
		}
		return "", nil, errors.New("expected a month, a date or 'T' after a date")
	}
	if len(date) != len(time.DateOnly) {
		return "", nil, errors.New("expected a date before 'T'")
	}

	var end periodEnd
	switch len(clock) {
	case len("15:04"):
		clock, end = clock+":00", after(time.Minute)
	case len("15:04:05"):
		end = after(time.Second)
	case len("15:04:05.000"), len("15:04:05.000000"), len("15:04:05.000000000"):
		// the length doesn't prove the fraction separator is there, time.Parse
		// only validates the digits
		if clock[len("15:04:05")] != '.' {
			return "", nil, errors.New("expected '.' before the fraction of a second")
		}
		end = after(time.Millisecond)
	default:
		return "", nil, errors.New("expected a time with minutes, seconds " +
			"or fractions of a second in 3, 6 or 9 digits")
	}
	return date + "T" + clock + zone, end, nil
}

// ParseTimeRange interprets a filter value as the period its precision
// implies, e.g. a date covers the whole day and a timestamp with minutes the
// whole minute.
func ParseTimeRange(value string) (TimeRange, error) {
	padded, end, err := padDateTime(value)
	var t time.Time
	if err == nil {
		t, err = time.Parse(time.RFC3339Nano, padded)
	}
	if err != nil {
		return TimeRange{}, errors.Errorf(
			"%q is not a month, date or timestamp "+
				"(e.g. 2025-01, 2025-01-15+02:00 or 2025-01-15T10:30Z)", value,
		)
	}
	start := t.Truncate(time.Millisecond)
	// the end is computed before converting to UTC so that it falls on
	// the next day or month in the given timezone
	return TimeRange{Start: start.UTC(), End: end(start).UTC()}, nil
}

// TimeRanges returns the periods the predicate value covers if the predicate
// compares a datetime attribute, nil otherwise. For $in and $nin there is one
// range per array element; an empty array yields nil as there is nothing to
// interpret.
func (f FilterPredicate) TimeRanges() ([]TimeRange, error) {
	if !IsDateTimeAttribute(string(f.Scope), f.Attribute) ||
		!slices.Contains(dateTimeOperators, f.Type) {
		return nil, nil
	}

	var values []interface{}
	if f.Type == "$in" || f.Type == "$nin" {
		switch v := f.Value.(type) {
		case []interface{}:
			values = v
		case []string:
			for _, s := range v {
				values = append(values, s)
			}
		default:
			return nil, errors.Errorf("%s expects an array of dates or timestamps", f.Type)
		}
	} else {
		values = []interface{}{f.Value}
	}

	var ranges []TimeRange
	for _, v := range values {
		s, ok := v.(string)
		if !ok {
			return nil, errors.Errorf(
				"%s/%s expects dates or timestamps as strings", f.Scope, f.Attribute,
			)
		}
		r, err := ParseTimeRange(s)
		if err != nil {
			return nil, err
		}
		ranges = append(ranges, r)
	}
	return ranges, nil
}
