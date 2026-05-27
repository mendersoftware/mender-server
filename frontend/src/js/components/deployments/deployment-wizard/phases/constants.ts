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

export const rolloutModes = {
  percentage: { key: 'percentage', title: 'By percentage of total devices', batchKey: 'batch_size' },
  device_count: { key: 'device_count', title: 'By number of devices', batchKey: 'batch_size_devices' }
};

export type RolloutMode = keyof typeof rolloutModes;

export const rolloutPatterns = {
  custom: {
    key: 'custom',
    title: 'Custom',
    tip: 'Define each deployment phase individually'
  },
  uniform: {
    key: 'uniform',
    title: 'Uniform (repeat until all devices deployed)',
    tip: 'Repeat all phases until all devices are deployed'
  }
};

export type RolloutPattern = keyof typeof rolloutPatterns;

export const delayUnits = {
  minutes: 'minutes',
  hours: 'hours',
  days: 'days'
};

export const phaseDefaults = {
  batchSize: 10,
  delay: 7200
};

export const delayDefaults = {
  delay: 2,
  delayUnit: delayUnits.hours
};
