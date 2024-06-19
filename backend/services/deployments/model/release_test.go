// Copyright 2023 Northern.tech AS
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
	b64 "encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReleaseTags(t *testing.T) {

	t.Run("Validate", func(t *testing.T) {
		tooLong := make(Tags, TagsMaxPerRelease+1)
		for i := range tooLong {
			tooLong[i] = Tag(strconv.Itoa(i))
		}
		err := tooLong.Validate()
		assert.ErrorIs(t, err, ErrTooManyTags)

		invalidTags := Tags{"Va1_id", "invælid"}
		err = invalidTags.Validate()
		var charErr *InvalidCharacterError
		assert.ErrorAs(t, err, &charErr)

		tooLongTag := Tag(strings.Repeat("veryLongTag", 100))
		err = tooLongTag.Validate()
		assert.ErrorIs(t, err, ErrTagTooLong)

		var emptyTag Tag
		err = emptyTag.Validate()
		assert.ErrorIs(t, err, ErrTagEmpty)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		var tags Tags
		b, _ := json.Marshal(tags)
		assert.Equal(t, []byte(`[]`), b)
		err := json.Unmarshal(
			[]byte(`["tag", "tag", "more-tags"]`),
			&tags,
		)
		assert.NoError(t, err)
		assert.ElementsMatch(t, Tags{"tag", "more-tags"}, tags)

		err = json.Unmarshal([]byte(`{}`), &tags)
		assert.Error(t, err)
	})
}

func TestReleaseNotesValidation(t *testing.T) {
	longNotes := make([]byte, NotesLengthMaximumCharacters+1)
	for i := range longNotes {
		longNotes[i] = '1'
	}
	var notes Notes
	notes = Notes(longNotes)
	err := notes.Validate()
	assert.ErrorIs(t, err, ErrReleaseNotesTooLong)

	randomBytesBase64 := "zT2AkjJAP34PTuqq2JyXKgOVlY7k9gcnvSivs5jUbz8Wt6tb7OgZY7zJi5VyJeVwaaJkmdzb6KeojVg5YKt4wTj6Miiys5DBw0//Q+XAnpJdsPQM/a/GZzQzYUYBj1coHt1nAFFQebLbQc22IQOYJ4VUwhhxOoOhb90f2KtHgdIBylQWPPfYoDMat9J1E2S4TZjkfwxUKX9qq7zNCtm+HLFUfZh7zvK69PTyfVSnid7Gk7rJEX778whBAr690YNwdmTnBHLWNFErX7nSbDH3mPK5XvJasWEH/gUy6RphTv4DjWk9Dn7Dps3+9ksUVz+8qoBQ/jwSOfSjuHICewlSRMnoQWHKEsqn6p9p/xksMtS5Bh2ftQUwn3cfs6JGQmfaThJo0bKHML9eDZl+i5sSE85dzhCv7ZRLd0G9+n5097cOHE02QP2t1K7u5UNZ+DQ4dU1gXu+2qIqQUpuK5OnwAgw4jc60mROOsDLBtfu9Vfp8kr9E7OBlKg2aY5OxigF1toWhWYWsPc1jNCAYf/uyO7eHboNXeX6nPjpsjmy/M6EHCtfGlZtdVH4p8Pd8kLIjp6sasLRmT0AMihV7PbkHqdO+EAvrFDZR9jQkYd6KzSEMWbRKh0AoDONj+RSyQC0esUl94XXQoXu0T4R52pUA3Gal7SYce2c9pV7rfl0M2o4="
	raw, err := b64.StdEncoding.DecodeString(randomBytesBase64)
	assert.NoError(t, err)
	notes = Notes(raw)
	err = notes.Validate()
	assert.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "invalid character '"))
}

func TestConvertReleasesToV1(t *testing.T) {
	now := time.Now()
	releases := []Release{
		{
			Name:     "r1",
			Modified: &now,
			Artifacts: []Image{
				{
					Id: "i1",
				},
			},
			ArtifactsCount: 1,
			Tags:           Tags{"t1", "t2"},
		},
	}
	expected := []ReleaseV1{
		{
			Name:     "r1",
			Modified: &now,
			Artifacts: []Image{
				{
					Id: "i1",
				},
			},
			ArtifactsCount: 1,
			Tags:           Tags{"t1", "t2"},
		},
	}
	releasesV1 := ConvertReleasesToV1(releases)
	assert.Equal(t, expected, releasesV1)
}
