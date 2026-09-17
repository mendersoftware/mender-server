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
import { formRenderWrapper } from '@/testUtils';
import { undefineds } from '@northern.tech/testing/mockData';

import GroupDefinition, { getGroupNameError, groupNameValidationRules } from './GroupDefinition';

const invalidCharactersError = 'This please only enter valid characters. Valid characters are a-z, A-Z, 0-9, _ and -';

describe('GroupDefinition Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = formRenderWrapper(<GroupDefinition groups={[]} name="group" selectedDevices={[]} />, { defaultValues: { group: '' } });
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('rejects group names with invalid characters', () => {
    expect(getGroupNameError('tæst')).toBe(invalidCharactersError);
    expect(getGroupNameError('no.dots.allowed')).toBe(invalidCharactersError);
    expect(getGroupNameError('valid_name-12')).toBe('');
  });

  it('rejects group names longer than 256 characters', () => {
    expect(getGroupNameError('a'.repeat(257))).toBe('Must be between 1 and 256 characters long');
    expect(getGroupNameError('a'.repeat(256))).toBe('');
  });

  it('rejects the automatically created group name', () => {
    expect(getGroupNameError('Unassigned')).toBe('This field should have a value other than Unassigned');
  });

  it('rejects existing group names only for dynamic groups', () => {
    expect(getGroupNameError('test', { existingGroups: ['test'] })).toBe('');
    expect(getGroupNameError('test', { existingGroups: ['test'], isDynamic: true })).toBe('A group with the same name already exists');
  });

  it('rejects the group the selected devices are already in', () => {
    expect(getGroupNameError('test', { selectedDevices: [{ group: 'test' }] })).toBe('test is the same group the selected devices are already in');
    expect(getGroupNameError('test', { selectedDevices: [{ group: 'test' }, { group: 'other' }] })).toBe('');
  });

  it('builds form rules for group names', () => {
    const { required, validate } = groupNameValidationRules({ existingGroups: ['a-b'], isDynamic: true });
    expect(required).toBe('Group name is required');
    expect(validate('a-b')).toBe('A group with the same name already exists');
    expect(validate('100%')).toBe(invalidCharactersError);
    expect(validate('a-c')).toBe(true);
  });
});
