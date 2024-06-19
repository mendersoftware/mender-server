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

package app

var (
	//Any session that transfers at least one byte above this limit will be automatically
	//closed. This is set to 8MB, based on a research that a typical couple of minutes session
	//of `top -d1` (top refreshing every second) is 2MB large (at terminal height=135 width=33))
	//Note: this refers to the number of bytes seen on the terminal stdout, as opposed
	//      to the number of bytes saved in the db. The key difference being:
	//      the latter is compressed.
	//see: https://northerntech.atlassian.net/browse/MEN-4448
	MessageSizeLimit = 8 * 1024 * 1024
)
