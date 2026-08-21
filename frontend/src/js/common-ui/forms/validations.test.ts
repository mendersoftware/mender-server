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
import { runValidations } from './validations';

describe('runValidations', () => {
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
});
