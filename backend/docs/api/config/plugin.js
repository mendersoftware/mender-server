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
/** @type {import('@redocly/cli').OasDecorator} */
const addServers = () => ({
  Root: {
    leave: (root) => {
      // could be read from the root.openapi.yaml
      root.servers = [{ url: "https://hosted.mender.io/" }];
    },
  },
});

/** @type {import('@redocly/cli').OasDecorator} */
const removeServers = () => ({
  PathItem: {
    leave: (pathItem) => {
      pathItem.servers = undefined;
    },
  },
});

// the default implementation takes the server from the joined files and adds it to all paths, so remove these and add the relevant one
export default function plugin() {
  return {
    id: "server-adjustments",
    decorators: {
      oas3: {
        "add-servers": addServers,
        "remove-servers": removeServers,
      },
    },
  };
}
