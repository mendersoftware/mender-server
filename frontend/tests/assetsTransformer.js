// Copyright 2020 Northern.tech AS
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
import path from 'path';

const processSvg = filename => {
  let componentName = path.parse(filename).name.replace(/[^a-z]/gi, '');
  componentName = componentName[0].toUpperCase() + componentName.slice(1);
  return `const React = require('react');
const Component = React.forwardRef((props, ref) => ({
  $$typeof: Symbol.for('react.element'),
  type: 'svg',
  ref,
  props: { ...props, id: '${componentName}' }
}));
module.exports = { __esModule: true, default: Component, Component };`;
};

export default {
  process(src, filename) {
    const assetFilename = JSON.stringify(path.basename(filename));
    let code = `module.exports = ${assetFilename};`;
    if (filename.endsWith('.svg')) {
      code = processSvg(filename);
    }

    return { code };
  }
};
