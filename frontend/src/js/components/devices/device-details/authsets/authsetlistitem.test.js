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
import React from 'react';

import { adminUserCapabilities, undefineds } from '../../../../../../tests/mockData';
import { render } from '../../../../../../tests/setupTests';
import { DEVICE_STATES } from '../../../../constants/deviceConstants';
import { defaultColumns } from './authsetlist';
import AuthsetListItem, { getConfirmationMessage } from './authsetlistitem';

describe('AuthsetListItem Component', () => {
  it('renders correctly', async () => {
    const authset = {
      id: '123',
      identity_data: { mac: '24:7d:30:90:21:a8' },
      pubkey: `-----BEGIN PUBLIC KEY-----
MIIBojANBgkqhkiG9w0BAQEFAAOCAY8AMIIBigKCAYEA4XP347xOgC0PnBgplast
GmsSDWZE2nMCAkV9wx2J09hzqlEWK8tOOIS99IpAR4TtwIQi2GssNyGBBNRNMfLi
8E2JLalm/X2jFdutf4QVIvLOs1vT0FqpYH+B3BnocYC5TfBXAwVUoW8HxK0MuxBo
UTixFC4o2Wu3fQs+mMiVnV/jcYAV1O0N4+lgszObX8Buq8l817HB3WzUw/XxOyxC
yahN4skp1D9JFHZB3i6lnfSJNJJvABe/lLf2jnjeFzbBgJOGxzolBa3+UyAWJRLB
JxsSxbbnXS3vAwODZEBQ1VSs43en1o5IT/Z/79UC6wAKg+Z4VnkdcK0b9EsW9VQU
oIfVXZjmm5CWPMiV5f9gG2t36j2wydpDryYqEAE+n8N76JzD/ZKlmHo+FJBJNsSx
Hcoq3VRgR5v53BTLFZLLBaqLIqnQAUwn/RcWlEbS3dEGQDvKglNjwmSVf6Myub5w
gnr0OSIDwEL31l+12DbAQ9+ANv6TLpWNfLpX0E6IStkZAgMBAAE=
-----END PUBLIC KEY-----
`,
      ts: '2020-09-21T12:42:34.571Z',
      status: 'accepted'
    };
    const device = {
      id: 'a7503dcc-afb6-4926-897e-3b1346d1600f',
      attributes: { device_type: [], artifact_name: '' },
      updated_ts: '2020-09-21T12:42:38.751Z',
      identity_data: { mac: '24:7d:30:90:21:a8', status: 'accepted' },
      status: 'accepted',
      created_ts: '2020-09-21T12:42:34.567Z',
      auth_sets: [authset]
    };

    const { baseElement } = render(
      <AuthsetListItem
        authset={authset}
        classes={{ accordion: 'accordion', divider: 'divider' }}
        columns={defaultColumns}
        isActive={true}
        isExpanded={true}
        onExpand={jest.fn}
        confirm={jest.fn}
        device={device}
        limitMaxed={false}
        loading={false}
        total={10}
        userCapabilities={adminUserCapabilities}
      />
    );
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
  it('shows proper confirmation messages depending on device auth status', async () => {
    expect(
      getConfirmationMessage(DEVICE_STATES.accepted, { status: DEVICE_STATES.accepted, auth_sets: [1, 2, 3] }, { status: DEVICE_STATES.accepted })
    ).toEqual(
      'By accepting, the device with this identity data and public key will be granted authentication by the server. The previously accepted public key will be rejected automatically in favor of this new key.'
    );
    expect(getConfirmationMessage(DEVICE_STATES.accepted, { status: DEVICE_STATES.pending, auth_sets: [1] }, { status: DEVICE_STATES.pending })).toEqual(
      'By accepting, the device with this identity data and public key will be granted authentication by the server.'
    );
    expect(getConfirmationMessage(DEVICE_STATES.accepted, { status: DEVICE_STATES.accepted, auth_sets: [1] }, { status: DEVICE_STATES.rejected })).toEqual(
      'By accepting, the device with this identity data and public key will be granted authentication by the server. The previously accepted public key will be rejected automatically in favor of this new key.'
    );
    expect(getConfirmationMessage(DEVICE_STATES.rejected, { status: DEVICE_STATES.accepted, auth_sets: [1] }, { status: DEVICE_STATES.accepted })).toEqual(
      'The device with this identity data and public key will be rejected, and blocked from communicating with the Mender server.'
    );
    expect(getConfirmationMessage(DEVICE_STATES.rejected, { status: DEVICE_STATES.accepted, auth_sets: [1] }, { status: DEVICE_STATES.pending })).toEqual(
      'The device with this identity data and public key will be rejected, and blocked from communicating with the Mender server. Rejecting this request will not affect the device status as it is using a different key. '
    );
    expect(getConfirmationMessage('dismiss', { status: DEVICE_STATES.accepted, auth_sets: [1] }, { status: DEVICE_STATES.preauth })).toEqual(
      'The device authentication set will be removed from the preauthorization list.'
    );
    expect(getConfirmationMessage('dismiss', { status: DEVICE_STATES.accepted, auth_sets: [1] }, { status: DEVICE_STATES.accepted })).toEqual(
      'The device with this public key will no longer be accepted, and will be removed from the UI. If it makes another request in the future, it will show again as pending for you to accept or reject at that time.'
    );
    expect(getConfirmationMessage('dismiss', { status: DEVICE_STATES.accepted, auth_sets: [1, 2] }, { status: DEVICE_STATES.accepted })).toEqual(
      'The device with this public key will no longer be accepted, and this authorization request will be removed from the UI.'
    );
    expect(getConfirmationMessage('dismiss', { status: DEVICE_STATES.accepted, auth_sets: [1] }, { status: DEVICE_STATES.pending })).toEqual(
      'You can dismiss this authentication request for now. The device will be removed from the UI, but if the same device asks for authentication again in the future, it will show again as pending.'
    );
    expect(getConfirmationMessage('dismiss', { status: DEVICE_STATES.accepted, auth_sets: [1, 2] }, { status: DEVICE_STATES.pending })).toEqual(
      'You can dismiss this authentication request for now. This will remove this request from the UI, but won’t affect the device.'
    );
    expect(getConfirmationMessage('dismiss', { status: DEVICE_STATES.accepted, auth_sets: [1, 2] }, { status: DEVICE_STATES.rejected })).toEqual(
      'This request will be removed from the UI, but if the device asks for authentication again in the future, it will show as pending for you to accept or reject it at that time.'
    );
  });
});
