// Copyright 2019 Northern.tech AS
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
import React from 'react';

import { screen } from '@testing-library/react';

import { defaultState, undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import { ALL_DEVICES } from '../../constants/deviceConstants.js';
import { emptyRole } from '../../constants/userConstants';
import RoleDefinition from './roledefinition';

describe('Roles Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(
      <RoleDefinition
        adding
        editing={false}
        onCancel={jest.fn}
        onSubmit={jest.fn}
        removeRole={jest.fn}
        stateGroups={defaultState.devices.groups.byId}
        stateReleaseTags={{}}
        selectedRole={undefined}
      />
    );
    const view = baseElement.lastElementChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('displays missing groups', async () => {
    const groupPermissions = ['read'];
    const nonExistingGroupName = 'nonExistingGroup';
    // removes read-only
    const copiedEmptyRole = JSON.parse(JSON.stringify(emptyRole));
    let selectedRole = {
      ...copiedEmptyRole,
      editable: true,
      uiPermissions: {
        ...copiedEmptyRole.uiPermissions,
        groups: {
          [nonExistingGroupName]: groupPermissions,
          [ALL_DEVICES]: groupPermissions
        }
      }
    };

    const { rerender } = render(
      <RoleDefinition
        adding={false}
        editing
        onCancel={jest.fn}
        onSubmit={jest.fn}
        removeRole={jest.fn}
        stateGroups={defaultState.devices.groups.byId}
        stateReleaseTags={{}}
        selectedRole={selectedRole}
      />
    );
    expect(screen.getByDisplayValue(nonExistingGroupName)).toBeInTheDocument();
    expect(screen.getByTitle(/This item was removed/)).toBeInTheDocument();

    delete selectedRole.uiPermissions.groups[nonExistingGroupName];
    rerender(
      <RoleDefinition
        adding={false}
        editing
        onCancel={jest.fn}
        onSubmit={jest.fn}
        removeRole={jest.fn}
        stateGroups={defaultState.devices.groups.byId}
        stateReleaseTags={{}}
        selectedRole={selectedRole}
      />
    );
    expect(screen.queryByDisplayValue(nonExistingGroupName)).not.toBeInTheDocument();
    expect(screen.queryByTitle(/This item was removed/)).not.toBeInTheDocument();
  });
});
