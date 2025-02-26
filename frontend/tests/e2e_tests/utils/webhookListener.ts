// Copyright 2025 Northern.tech AS
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
import express from 'express';

const app = express();
app.use(express.json());
app.post('/webhooks/all', (req, res) => {
  console.log('received webhook', req.body);
  res.sendStatus(200);
});
app.post('/webhooks/inventory', (req, res) => {
  console.log('received inventory webhook', req.body);
  res.sendStatus(200);
});

export const startServer = () => app.listen(9000, () => console.log('Node.js server started on port 9000.'));
