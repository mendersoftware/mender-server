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

const productName = 'Mender';
const separator = ' | ';

export const getPageTitle = (segments: (string | undefined)[]) => [...segments.filter(Boolean), productName].join(separator);

// segments are ordered from most to least specific, so the distinguishing part stays visible in tabs & the history list
// only a single PageTitle should be rendered at any time, as React hoists every <title> into the document head
export const PageTitle = ({ segments }: { segments: (string | undefined)[] }) => <title>{getPageTitle(segments)}</title>;

export default PageTitle;
