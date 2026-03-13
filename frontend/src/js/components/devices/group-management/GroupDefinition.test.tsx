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
import { render } from '@/testUtils';
import { undefineds } from '@northern.tech/testing/mockData';

import GroupDefinition, { validateGroupName } from './GroupDefinition';

const selectedDevices = [{ id: 'test' }];

describe('GroupDefinition Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<GroupDefinition groups={[]} isCreationDynamic={true} />);
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('validates group names correctly', async () => {
    expect(validateGroupName('test', undefined, [{ ...selectedDevices[0], group: 'test' }], false)).toEqual({
      errorText: 'test is the same group the selected devices are already in',
      invalid: true,
      isModification: false,
      name: 'test'
    });
    expect(validateGroupName('tæst', undefined, selectedDevices, false).invalid).toBeTruthy();
    expect(validateGroupName('false', undefined, selectedDevices, false).invalid).toBeFalsy();
    expect(validateGroupName('', undefined, selectedDevices, false).invalid).toBeTruthy();
    expect(validateGroupName('test', ['test'], [], true).invalid).toBeTruthy();
  });

  it('rejects group names longer than 256 characters', () => {
    const longName = 'a'.repeat(257);
    const result = validateGroupName(longName, undefined, selectedDevices, false);
    expect(result.invalid).toBeTruthy();
    expect(result.errorText).toBe('Name must be at most 256 characters long');
  });

  it('accepts group names with exactly 256 characters', () => {
    const name256 = 'a'.repeat(256);
    expect(validateGroupName(name256, undefined, selectedDevices, false).invalid).toBeFalsy();
  });
});
