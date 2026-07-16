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
import type { ZxcvbnFactory } from '@zxcvbn-ts/core';

let instancePromise: Promise<ZxcvbnFactory> | undefined;
const PASSWORD_CHARSET = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';

export const generatePassword = (length = 16) =>
  Array.from(crypto.getRandomValues(new Uint32Array(length)), n => PASSWORD_CHARSET[n % PASSWORD_CHARSET.length]).join('');

const initInstance = async () => {
  const [{ ZxcvbnFactory }, { adjacencyGraphs, dictionary: commonDictionary }, { dictionary: enDictionary, translations }] = await Promise.all([
    import(/* webpackChunkName: "zxcvbn" */ '@zxcvbn-ts/core'),
    import(/* webpackChunkName: "zxcvbn" */ '@zxcvbn-ts/language-common'),
    import(/* webpackChunkName: "zxcvbn" */ '@zxcvbn-ts/language-en')
  ]);
  return new ZxcvbnFactory({ translations, graphs: adjacencyGraphs, dictionary: { ...commonDictionary, ...enDictionary } });
};

export const checkPasswordStrength = (password: string) => {
  instancePromise = instancePromise ?? initInstance();
  return instancePromise.then(zxcvbn => zxcvbn.check(password));
};
