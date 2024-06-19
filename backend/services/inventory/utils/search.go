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
package utils

import (
	"strings"
	"unicode"

	"github.com/mendersoftware/mender-server/services/inventory/model"
)

// Check if string is presnt in an array. Would use interface{} but
// whatever.
func ContainsString(val string, vals []string) bool {
	for _, v := range vals {
		if val == v {
			return true
		}
	}
	return false
}

// TextToKeywords converts text to keywords for full-text indexing
func TextToKeywords(in string) string {
	s := []rune(in)
	j := 0
	for _, b := range s {
		if unicode.IsLetter(b) || unicode.IsDigit(b) {
			s[j] = b
			j++
		} else if j > 0 && s[j-1] != ' ' {
			s[j] = ' '
			j++
		}
	}
	return string(s[:j])
}

// GetTextField returns the text field for the given device
func GetTextField(device *model.Device) string {
	var text strings.Builder
	text.WriteString(TextToKeywords(device.ID.String()))
	if device.Group.String() != "" {
		text.WriteString(" " + TextToKeywords(device.Group.String()))
	}
	for _, attr := range device.Attributes {
		if IsScopeFullText(attr.Scope) {
			if val, _ := attr.Value.(string); val != "" {
				text.WriteString(" " + TextToKeywords(val))
			}
		}
	}
	return text.String()
}

func IsScopeFullText(scope string) bool {
	return scope == model.AttrScopeIdentity ||
		scope == model.AttrScopeInventory ||
		scope == model.AttrScopeTags
}
