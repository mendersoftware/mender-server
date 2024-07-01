// Copyright 2023 Northern.tech AS
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

package model

import (
	"encoding/json"
	"errors"
)

const (
	defaultPage    = 0
	defaultPerPage = 20

	attrDeviceID = "id"
)

type ArrayOpts int

const (
	ArrNotAllowed ArrayOpts = iota
	ArrAllowed
	ArrRequired
)

var (
	ErrArrayNotSupported = errors.New("filter doesn't support array values")
	ErrArrayRequired     = errors.New("filter supports only array values")
	ErrStrRequired       = errors.New("filter supports only string values")
	ErrNumRequired       = errors.New("filter supports only numeric values")
	ErrBoolRequired      = errors.New("filter supports only boolean values")
)

type M map[string]interface{}
type S []interface{}

// Query represents the ES query
// general form:
//
//	{
//	  "query": {
//	    "bool": {
//	      "must": [...conditions...],
//	      "must_not": [...conditions...],
//	    }
//	  "sort": [...],
//	  "from": ...,
//	  "size": ...,
//	}
//
// it exposes an API for query parts to insert themselves in the right place
type Query interface {
	Must(condition interface{}) Query
	MustNot(condition interface{}) Query
	WithSize(size int) Query
	WithSort(sort interface{}) Query
	WithPage(page, per_page int) Query
	With(parts map[string]interface{}) Query
	WithGeoFilters(df *GeoDistanceFilter, bf *GeoBoundingBoxFilter) Query

	MarshalJSON() ([]byte, error)
}

// QueryPart represents a bit of the ES query
// which knows how and where to insert itself
type QueryPart interface {
	AddTo(q Query) Query
}

type query struct {
	must                 []interface{}
	mustNot              []interface{}
	sort                 []interface{}
	geoDistanceFilter    *GeoDistanceFilter
	geoBoundingBoxFilter *GeoBoundingBoxFilter
	from                 int
	size                 int

	extra map[string]interface{}
}

func NewQuery() Query {
	return &query{
		from:  (defaultPage - 1) * defaultPerPage,
		size:  defaultPerPage,
		extra: make(map[string]interface{}),
	}
}

func (q *query) Must(condition interface{}) Query {
	q.must = append(q.must, condition)
	return q
}

func (q *query) MustNot(condition interface{}) Query {
	q.mustNot = append(q.mustNot, condition)
	return q
}

func (q *query) WithSize(size int) Query {
	q.size = size
	return q
}

func (q *query) WithSort(condition interface{}) Query {
	q.sort = append(q.sort, condition)
	return q
}

func (q *query) WithPage(page, perPage int) Query {
	q.from = (page - 1) * perPage
	q.size = perPage
	return q
}

func (q *query) With(parts map[string]interface{}) Query {
	if len(parts) == 0 {
		return q
	}

	for k, v := range parts {
		q.extra[k] = v
	}

	return q
}

func (q *query) WithGeoFilters(df *GeoDistanceFilter, bf *GeoBoundingBoxFilter) Query {
	q.geoDistanceFilter = df
	q.geoBoundingBoxFilter = bf

	return q
}

func (q *query) MarshalJSON() ([]byte, error) {
	qbool := M{}

	if q.must != nil {
		qbool["must"] = q.must
	}

	if q.mustNot != nil {
		qbool["must_not"] = q.mustNot
	}

	if q.geoDistanceFilter != nil {
		qbool["filter"] = q.geoDistanceFilter
	} else if q.geoBoundingBoxFilter != nil {
		qbool["filter"] = q.geoBoundingBoxFilter
	}

	qjson := M{
		"query": M{
			"bool": qbool,
		},
	}

	if q.sort != nil {
		qjson["sort"] = q.sort
	}

	qjson["from"] = q.from
	qjson["size"] = q.size

	if len(q.extra) > 0 {
		for k, v := range q.extra {
			qjson[k] = v
		}
	}

	return json.Marshal(qjson)
}

// filter factory
func getFilterPart(pred FilterPredicate) (QueryPart, error) {
	switch pred.Type {
	case "$eq":
		return NewFilterEq(pred)
	case "$ne":
		return NewFilterNe(pred)
	case "$gt":
		return NewFilterRange(pred, "gt")
	case "$gte":
		return NewFilterRange(pred, "gte")
	case "$lt":
		return NewFilterRange(pred, "lt")
	case "$lte":
		return NewFilterRange(pred, "lte")
	case "$in":
		return NewFilterIn(pred)
	case "$nin":
		return NewFilterNin(pred)
	case "$exists":
		return NewFilterExists(pred)
	case "$regex":
		return NewFilterRegex(pred)
	}

	return nil, errors.New("filter type not supported")
}

// filter query parts below
type filter struct {
	// computed attr name
	attr string
	val  interface{}
}

func NewFilter(fp FilterPredicate, arrOpts ArrayOpts, typeOpts Type) (*filter, error) {
	// inspect type to
	// a) compose attribute name
	// b) restrict inputs
	typ, isarr, err := fp.ValueType()

	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	if isarr && arrOpts == ArrNotAllowed {
		return nil, ErrArrayNotSupported
	}

	if !isarr && arrOpts == ArrRequired {
		return nil, ErrArrayRequired
	}

	if typeOpts != TypeAny && typeOpts != typ {
		switch typ {
		case TypeStr:
			return nil, ErrStrRequired
		case TypeNum:
			return nil, ErrNumRequired
		case TypeBool:
			return nil, ErrBoolRequired

		}
	}

	// some special attributes translate to non-scoped, predefined fields
	attr := parseSpecialAttr(fp.Attribute)
	if attr == "" {
		attr = ToAttr(fp.Scope, fp.Attribute, typ)
	}

	return &filter{
		attr: attr,
		val:  fp.Value,
	}, nil
}

type filterEq struct {
	*filter
}

func NewFilterEq(fp FilterPredicate) (*filterEq, error) {
	f, err := NewFilter(fp, ArrNotAllowed, TypeAny)
	if err != nil {
		return nil, err
	}

	return &filterEq{
		filter: f,
	}, nil
}

func (f *filterEq) AddTo(q Query) Query {
	return q.Must(M{
		"match": M{
			f.attr: f.val,
		},
	})
}

type filterNe struct {
	*filter
}

func NewFilterNe(fp FilterPredicate) (*filterNe, error) {
	f, err := NewFilter(fp, ArrNotAllowed, TypeAny)
	if err != nil {
		return nil, err
	}

	return &filterNe{
		filter: f,
	}, nil
}

func (f *filterNe) AddTo(q Query) Query {
	return q.MustNot(M{
		"match": M{
			f.attr: f.val,
		},
	})
}

type filterRegex struct {
	*filter
}

func NewFilterRegex(fp FilterPredicate) (*filterRegex, error) {
	f, err := NewFilter(fp, ArrNotAllowed, TypeStr)
	if err != nil {
		return nil, err
	}
	return &filterRegex{
		filter: f,
	}, nil
}

func (f *filterRegex) AddTo(q Query) Query {
	return q.Must(M{
		"regexp": M{
			f.attr: f.val,
		},
	})
}

type filterIn struct {
	*filter
}

func NewFilterIn(fp FilterPredicate) (*filterIn, error) {
	f, err := NewFilter(fp, ArrRequired, TypeAny)
	if err != nil {
		return nil, err
	}
	return &filterIn{
		filter: f,
	}, nil
}

func (f *filterIn) AddTo(q Query) Query {
	return q.Must(M{
		"terms": M{
			f.attr: f.val,
		},
	})
}

type filterNin struct {
	*filter
}

func NewFilterNin(fp FilterPredicate) (*filterNin, error) {
	f, err := NewFilter(fp, ArrRequired, TypeAny)
	if err != nil {
		return nil, err
	}

	return &filterNin{
		filter: f,
	}, nil
}

func (f *filterNin) AddTo(q Query) Query {
	return q.MustNot(M{
		"terms": M{
			f.attr: f.val,
		},
	})
}

type filterExists struct {
	*filter
	fp FilterPredicate
}

func NewFilterExists(fp FilterPredicate) (*filterExists, error) {
	f, err := NewFilter(fp, ArrNotAllowed, TypeBool)
	if err != nil {
		return nil, err
	}
	return &filterExists{
		filter: f,
		fp:     fp,
	}, nil
}

func (f *filterExists) AddTo(q Query) Query {
	exists := f.fp.Value.(bool)
	astr := ToAttr(f.fp.Scope, f.fp.Attribute, TypeStr)
	anum := ToAttr(f.fp.Scope, f.fp.Attribute, TypeNum)
	abool := ToAttr(f.fp.Scope, f.fp.Attribute, TypeBool)

	if exists {
		return q.Must(M{
			"bool": M{
				"minimum_should_match": 1,
				"should": S{
					M{"exists": M{"field": astr}},
					M{"exists": M{"field": anum}},
					M{"exists": M{"field": abool}},
				},
			},
		})
	}

	return q.
		MustNot(M{"exists": M{"field": astr}}).
		MustNot(M{"exists": M{"field": anum}}).
		MustNot(M{"exists": M{"field": abool}})
}

// "$gt", "$gte", "$lt", "$lte"
type filterRange struct {
	*filter

	// internal ES range operator
	op string
}

func NewFilterRange(fp FilterPredicate, op string) (*filterRange, error) {
	f, err := NewFilter(fp, ArrNotAllowed, TypeAny)
	if err != nil {
		return nil, err
	}
	return &filterRange{
		filter: f,
		op:     op,
	}, nil
}

func (f *filterRange) AddTo(q Query) Query {
	return q.Must(M{
		"range": M{
			f.attr: M{
				f.op: f.val,
			},
		},
	})
}

type sort struct {
	attrStr  string
	attrNum  string
	attrBool string
	order    string
}

func NewSort(sc SortCriteria) *sort {
	order := sc.Order
	if order == "" {
		order = SortOrderAsc
	}
	return &sort{
		attrStr:  ToAttr(sc.Scope, sc.Attribute, TypeStr),
		attrNum:  ToAttr(sc.Scope, sc.Attribute, TypeNum),
		attrBool: ToAttr(sc.Scope, sc.Attribute, TypeBool),
		order:    order,
	}
}

func (s *sort) AddTo(q Query) Query {
	q = q.
		WithSort(
			M{
				s.attrStr: M{
					"order":         s.order,
					"unmapped_type": "keyword",
				},
			},
		).WithSort(
		M{
			s.attrNum: M{
				"order":         s.order,
				"unmapped_type": "double",
			},
		},
	)

	return q
}

type sel struct {
	attrs []SelectAttribute
}

func NewSelect(attrs []SelectAttribute) *sel {
	return &sel{
		attrs: attrs,
	}
}

func (s *sel) AddTo(q Query) Query {
	fields := []string{}

	for _, a := range s.attrs {
		fields = append(fields,
			ToAttr(a.Scope, a.Attribute, TypeStr),
			ToAttr(a.Scope, a.Attribute, TypeNum),
			ToAttr(a.Scope, a.Attribute, TypeBool),
		)
	}

	//always include a device id
	fields = append(fields, "id")

	//always include a check-in time
	fields = append(fields, FieldNameCheckIn)

	return q.With(map[string]interface{}{
		"fields":  fields,
		"_source": false,
	})

}

type geoFilters struct {
	df *GeoDistanceFilter
	bf *GeoBoundingBoxFilter
}

func NewGeoFilters(df *GeoDistanceFilter, bf *GeoBoundingBoxFilter) *geoFilters {
	return &geoFilters{
		df: df,
		bf: bf,
	}
}

func (f *geoFilters) AddTo(q Query) Query {
	return q.WithGeoFilters(f.df, f.bf)
}

type devIDsFilter struct {
	devIDs []string
}

func NewDevIDsFilter(ids []string) *devIDsFilter {
	return &devIDsFilter{
		devIDs: ids,
	}
}

func (f *devIDsFilter) AddTo(q Query) Query {
	return q.Must(M{
		"terms": M{
			attrDeviceID: f.devIDs,
		},
	})
}

func BuildQuery(params SearchParams) (Query, error) {
	query := NewQuery()

	for _, f := range params.Filters {
		fpart, err := getFilterPart(f)
		if err != nil {
			return nil, err
		}
		query = fpart.AddTo(query)
	}

	if len(params.Groups) > 0 {
		fp := FilterPredicate{
			Scope:     ScopeSystem,
			Attribute: AttrNameGroup,
			Type:      "$in",
			Value:     params.Groups,
		}
		fpart, err := NewFilterIn(fp)
		if err != nil {
			return nil, err
		}
		query = fpart.AddTo(query)
	}

	for _, s := range params.Sort {
		sort := NewSort(s)
		query = sort.AddTo(query)
	}

	query = query.WithPage(params.Page, params.PerPage)

	if len(params.Attributes) > 0 {
		sel := NewSelect(params.Attributes)
		query = sel.AddTo(query)
	}

	if len(params.DeviceIDs) > 0 {
		devs := NewDevIDsFilter(params.DeviceIDs)
		query = devs.AddTo(query)
	}

	if params.GeoDistanceFilter != nil || params.GeoBoundingBoxFilter != nil {
		f := NewGeoFilters(params.GeoDistanceFilter, params.GeoBoundingBoxFilter)
		query = f.AddTo(query)
	}

	return query, nil
}

// parseSpecialAttr detects attributes like `Device ID`, which
// translate to plain flat fields (e.g. 'id'), and not
// scoped attributes
func parseSpecialAttr(attr string) string {
	switch attr {
	case attrDeviceID:
		return "id"
	default:
		return ""
	}
}
