// Copyright 2015 Northern.tech AS
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
import React from 'react';
import ReactDOM from 'react-dom/client';

import './../less/main.less';
import { AppProviders } from './components/app';

const welcomeMessage = `Welcome to the Mender project!

Does this page need fixes or improvements?

Open an issue, or contribute a fix to:

- https://github.com/mendersoftware/gui

🤝 Contribute to Mender: https://github.com/mendersoftware/mender/blob/master/CONTRIBUTING.md
🔎 Ask about problems, and report issues: https://hub.mender.io
🚀 We like your curiosity! Help us improve Mender by joining the team: https://northern.tech/careers
`;

console.log(welcomeMessage);
const root = ReactDOM.createRoot(document.getElementById('main'));
root.render(<AppProviders />);
