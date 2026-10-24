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
import { prettyDOM, screen, within } from '@testing-library/react';
import { vi } from 'vitest';

import AuditLogsList from './AuditLogsList';
import { ActionDescriptor, ChangeDescriptor, ChangeDetailsDescriptor, TimeWrapper, TypeDescriptor, UserDescriptor } from './ColumnComponents';

const actor = { id: 'string', type: 'user', email: 'string@example.com' };
const events = [
  ...defaultState.organization.auditlog.events,
  { actor, time: '2019-01-01T12:20:22.667Z', action: 'upload', object: { id: 'art1', type: 'artifact', artifact: { name: 'some artifact' } } },
  { actor, time: '2019-01-01T12:21:22.667Z', action: 'upload', object: { id: 'my release', type: 'software', software: { kind: 'release' } } },
  { actor, time: '2019-01-01T12:22:22.667Z', action: 'upload', object: { id: 'my manifest', type: 'software', software: { kind: 'manifest' } } },
  { actor, time: '2019-01-01T12:23:22.667Z', action: 'delete', object: { id: 'art2', type: 'artifact', artifact: { name: 'gone artifact' } } },
  { actor, time: '2019-01-01T12:24:22.667Z', action: 'delete', object: { id: 'gone release', type: 'software', software: { kind: 'release' } } },
  {
    actor,
    time: '2019-01-01T12:25:22.667Z',
    action: 'create',
    change: 'Created PAT my token (t1).',
    object: { id: 't1', type: 'user_access_token', user_access_token: { name: 'my token', expires_in: 3600, owner: 'string' } }
  }
];
const state = { ...defaultState, organization: { ...defaultState.organization, auditlog: { ...defaultState.organization.auditlog, events } } };

const renderList = (userCapabilities = adminUserCapabilities) =>
  render(
    <AuditLogsList
      items={state.organization.auditlog.events}
      loading={false}
      onChangeRowsPerPage={vi.fn}
      onChangePage={vi.fn}
      onChangeSorting={vi.fn}
      selectionState={state.organization.auditlog.selectionState}
      setAuditlogsState={vi.fn}
      userCapabilities={userCapabilities}
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

const getRow = (text: string) => within(screen.getByText(text).closest('.auditlogs-list-item') as HTMLElement);

describe('AuditlogsList Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = renderList();
    const view = prettyDOM(baseElement.firstChild, 100000, { highlight: false })
      .replace(/(:?aria-labelledby|id)=":.*:"/g, '')
      .replace(/\\/g, '');
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
  it('renders software related entries', async () => {
    renderList();
    expect(getRow('some artifact').getByRole('link', { name: /view software/i })).toHaveAttribute('href', '/software/some%20artifact?tab=releases');
    expect(getRow('my release').getByRole('link', { name: /view software/i })).toHaveAttribute('href', '/software/my%20release?tab=releases');
    expect(getRow('my manifest').getByRole('link', { name: /view software/i })).toHaveAttribute('href', '/software?tab=manifests&id=my%20manifest');
  });
  it('does not link software entries without release access', async () => {
    renderList({ ...adminUserCapabilities, canReadReleases: false });
    expect(getRow('my release').queryByRole('link')).not.toBeInTheDocument();
  });
  it('does not link deleted software entries', async () => {
    renderList();
    expect(getRow('gone artifact').queryByRole('link')).not.toBeInTheDocument();
    expect(getRow('gone release').queryByRole('link')).not.toBeInTheDocument();
  });
  it('renders personal access token entries', async () => {
    renderList();
    expect(getRow('my token').getByText('Created PAT my token (t1).')).toBeVisible();
    expect(screen.queryByText(/expires_in/)).not.toBeInTheDocument();
  });
});
