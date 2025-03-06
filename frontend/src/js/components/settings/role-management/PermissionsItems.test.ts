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
import { shouldExtendPermissionSelection } from './PermissionsItems';

const emptyElement = {
  item: '',
  uiPermissions: [],
  disableEdit: false,
  notFound: false,
  id: '6792662d-65e9-41e0-8fdf-fc5dad58e10b'
};
const addedElement = {
  item: '',
  uiPermissions: ['read'],
  disableEdit: false,
  notFound: false,
  id: '14f15811-d979-4cc8-9ffb-15310aba25f5'
};
const currentState = [
  {
    item: '123',
    uiPermissions: ['read'],
    disableEdit: false,
    notFound: false,
    id: 'f9f56aa1-1a0f-403f-a9c0-95f473af0c72'
  }
];
const items = [
  {
    title: 'All releases',
    notFound: false
  },
  {
    title: '123',
    notFound: false
  }
];

describe('shouldExtendPermissionSelection works correctly', () => {
  it(`doesn't add a row when there is an empty row`, () => {
    expect(shouldExtendPermissionSelection([...currentState, addedElement, emptyElement], addedElement, items)).toBeFalsy();
  });
  it('adds a row when one is partly filled', () => {
    expect(shouldExtendPermissionSelection([...currentState, addedElement], addedElement, items)).toBeTruthy();
    const state = [...currentState, { ...emptyElement, item: '123' }];
    expect(shouldExtendPermissionSelection(state, state[1], items)).toBeTruthy();
  });
  it(`doesn't add a row when all items selected`, () => {
    const state = [...currentState, { ...emptyElement, item: items[0].title }];
    expect(shouldExtendPermissionSelection(state, state[1], items)).toBeFalsy();
  });
});
