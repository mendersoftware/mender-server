// Copyright 2022 Northern.tech AS
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

package processor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessOptionString(t *testing.T) {
	var data string
	var setEncoding string
	var expectedEncoding Encoding
	var o Options

	setEncoding = encodingURL
	expectedEncoding = EncodingURL
	data = "encoding=" + setEncoding + ";"
	o = processOptionString(data)
	assert.Equal(t, Options{Encoding: expectedEncoding}, o)

	setEncoding = encodingURL
	expectedEncoding = EncodingURL
	data = "rightFlag=right;encoding=" + setEncoding + ";leftFlag=left;"
	o = processOptionString(data)
	assert.Equal(t, Options{Encoding: expectedEncoding}, o)

	setEncoding = "plain"
	expectedEncoding = EncodingPlain
	data = "encoding=" + setEncoding + ";"
	o = processOptionString(data)
	assert.Equal(t, Options{Encoding: expectedEncoding}, o)

	setEncoding = "plain"
	expectedEncoding = EncodingPlain
	data = "rightFlag=right;encoding=" + setEncoding + ";leftFlag=left;"
	o = processOptionString(data)
	assert.Equal(t, Options{Encoding: expectedEncoding}, o)
}

func TestHTMLEncoding(t *testing.T) {
	const (
		expected = `It escapes only five such characters: &lt;, &gt;, &amp;, &#39; and &#34;`
		input    = `It escapes only five such characters: <, >, &, ' and "`
	)
	actual := EncodingHTML.Apply(input)
	assert.Equal(t, expected, actual)
}

func TestVariableParse(t *testing.T) {
	var data string
	var setEncoding string
	var expectedVariableIdentifier string
	var expectedEncoding Encoding

	expectedEncoding = EncodingURL
	setEncoding = encodingURL
	expectedVariableIdentifier = "newVariable0"
	data = "${encoding=" + setEncoding + ";" + workflowInputVariable + expectedVariableIdentifier + "}"
	matches := reExpression.FindAllStringSubmatch(data, -1)

	for _, submatch := range matches {
		varName := submatch[reMatchIndexName]
		value := submatch[reMatchIndexDefault]
		options := processOptionString(submatch[reMatchIndexOptions])
		assert.Equal(t, varName, workflowInputVariable+expectedVariableIdentifier)
		assert.Equal(t, value, "")
		assert.Equal(t, Options{Encoding: expectedEncoding}, options)
	}
}
