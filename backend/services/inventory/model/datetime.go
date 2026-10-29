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

// dateTimeLayouts are the accepted value formats with the period each covers.
// Values without a UTC offset are taken to be in UTC.
var dateTimeLayouts = []struct {
	layout string
	period time.Duration
}{
	{layout: "2006-01-02T15:04:05Z07:00", period: time.Second},
	{layout: "2006-01-02T15:04:05", period: time.Second},
	{layout: "2006-01-02T15:04Z07:00", period: time.Minute},
	{layout: "2006-01-02T15:04", period: time.Minute},
	{layout: "2006-01-02", period: 24 * time.Hour},
}

// TimeRange is the half-open interval [Start, End).
type TimeRange struct {
	Start time.Time
	End   time.Time
}

func IsDateTimeAttribute(scope, name string) bool {
	return slices.Contains(dateTimeAttributes[scope], name)
}

// ParseTimeRange interprets a date or timestamp as the period its precision
// implies: a date covers the whole day, a timestamp without seconds the whole
// minute, and so on. Fractional seconds are matched to the millisecond, the
// resolution dates are stored with.
func ParseTimeRange(value string) (TimeRange, error) {
	for _, l := range dateTimeLayouts {
		t, err := time.Parse(l.layout, value)
		if err != nil {
			continue
		}
		period := l.period
		// time.Parse accepts fractional seconds even if the layout has none
		if period == time.Second && strings.Contains(value, ".") {
			period = time.Millisecond
		}
		start := t.Truncate(time.Millisecond).UTC()
		return TimeRange{Start: start, End: start.Add(period)}, nil
	}
	return TimeRange{}, errors.Errorf(
		"%q is not a date or timestamp (e.g. 2025-01-15 or 2025-01-15T10:00:00Z)", value,
	)
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
