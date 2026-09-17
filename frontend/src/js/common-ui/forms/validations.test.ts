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
import { getGroupNameError, groupNameValidationRules, invalidCharactersError, runValidations } from './validations';

describe('validations', () => {
  it('validates maxLength correctly', () => {
    const longValue = 'a'.repeat(257);
    const result = runValidations({
      required: false,
      value: longValue,
      id: 'name',
      validations: 'isAlphanumericLocator,isLength:1:256',
      wasMaybeTouched: true
    });
    expect(result.isValid).toBeFalsy();
    expect(result.errortext).toBe('Must be between 1 and 256 characters long');
  });

  it('accepts values within maxLength', () => {
    const value = 'a'.repeat(256);
    const result = runValidations({ required: false, value, id: 'name', validations: 'isAlphanumericLocator,isLength:1:256', wasMaybeTouched: true });
    expect(result.isValid).toBeTruthy();
  });

  it('rejects group names with invalid characters', () => {
    expect(getGroupNameError('tæst')).toBe(invalidCharactersError);
    expect(getGroupNameError('valid.name_1-2')).toBe('');
  });

  it('rejects group names longer than 256 characters', () => {
    expect(getGroupNameError('a'.repeat(257))).toBe('Name must be at most 256 characters long');
    expect(getGroupNameError('a'.repeat(256))).toBe('');
  });

  it('rejects the automatically created group name', () => {
    expect(getGroupNameError('Unassigned')).toBe('A group with the name Unassigned is created automatically');
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
