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
import { defaultState, render } from '@/testUtils';
import { adminUserCapabilities, undefineds } from '@northern.tech/testing/mockData';
import { prettyDOM } from '@testing-library/react';
import { vi } from 'vitest';

import AuditLogsList from './AuditLogsList';
import { ActionDescriptor, ChangeDescriptor, ChangeDetailsDescriptor, TimeWrapper, TypeDescriptor, UserDescriptor } from './ColumnComponents';

describe('AuditlogsList Component', () => {
  it('renders correctly', async () => {
    const state = { ...defaultState };
    const { baseElement } = render(
      <AuditLogsList
        items={defaultState.organization.auditlog.events}
        loading={false}
        onChangeRowsPerPage={vi.fn}
        onChangePage={vi.fn}
        onChangeSorting={vi.fn}
        selectionState={defaultState.organization.auditlog.selectionState}
        setAuditlogsState={vi.fn}
        userCapabilities={adminUserCapabilities}
        auditLogColumns={[
          { title: 'Performed by', sortable: false, render: UserDescriptor },
          { title: 'Action', sortable: false, render: ActionDescriptor },
          { title: 'Type', sortable: false, render: TypeDescriptor },
          { title: 'Changed', sortable: false, render: ChangeDescriptor },
          { title: 'More details', sortable: false, render: ChangeDetailsDescriptor },
          { title: 'Time', sortable: true, render: TimeWrapper }
        ]}
      />,
      { state }
    );

    const view = prettyDOM(baseElement.firstChild, 100000, { highlight: false })
      .replace(/(:?aria-labelledby|id)=":.*:"/g, '')
      .replace(/\\/g, '');
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});
