// Copyright 2022 Northern.tech AS
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

import { defaultState, undefineds } from '../../../../../tests/mockData';
import { render } from '../../../../../tests/setupTests';
import InstalledSoftware, { extractSoftwareInformation } from './installedsoftware';

const rootfs = 'stablev1-beta-final-v0';

describe('DeviceInventory Component', () => {
  it('renders correctly', async () => {
    const attributes = {
      ...defaultState.devices.byId.a1.attributes,
      artifact_name: 'myapp',
      'rootfs-image.version': rootfs,
      'rootfs-image.checksum': '12341143',
      'test.version': 'test-2',
      'a.whole.lot.of.dots.version': 'test-3',
      'a.whole.lot.of.dots.more': 'test-4',
      'even.more.dots.than.before.version': 'test-5',
      'even.more.dots.than.before.more': 'test-6',
      'rootfs-image.update-module.single-files.version': '13',
      'rootfs-image.update-module.single-files.mender_update_module': 'single-file-multi',
      'rootfs-image.update-module.delta-module.version': '13',
      'rootfs-image.update-module.delta-module.mender_update_module': 'custom-delta-updater',
      'rootfs-image.directory.version': 'mender-demo-artifact-3.4.0',
      'uefi-firmware.GUID-edk2.name': '4123rfsad',
      'uefi-firmware.GUID-edk2.version': 'v5',
      'uefi-firmware.GUID.edk2.checksum': '1232415512',
      'uefi-firmware.GUID.edk2.name': 'a2124',
      'uefi-firmware.GUID.edk2.version': 'v1'
    };
    const { baseElement } = render(<InstalledSoftware device={{ attributes, id: 'a1' }} setSnackbar={jest.fn} />);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});

describe('extractSoftwareInformation function', () => {
  it('works as expected', async () => {
    expect(extractSoftwareInformation(defaultState.releases.byId.r1.artifacts[0].artifact_provides)).toEqual([
      { children: [], content: { version: 'v2020.10' }, key: 'data-partition.myapp', priority: 0, title: 'data-partition.myapp' }
    ]);
    expect(extractSoftwareInformation(defaultState.devices.byId.a1.attributes)).toEqual([]);
    expect(
      extractSoftwareInformation({
        artifact_name: 'myapp',
        'rootfs-image.version': rootfs,
        'rootfs-image.checksum': '12341143',
        'test.version': 'test-2',
        'a.whole.lot.of.dots.version': 'test-3',
        'a.whole.lot.of.dots.more': 'test-4',
        'even.more.dots.than.before.version': 'test-5',
        'even.more.dots.than.before.more': 'test-6'
      })
    ).toEqual([
      { children: [], content: { checksum: '12341143', version: rootfs }, key: 'rootfs-image', priority: 0, title: 'Root filesystem' },
      { children: [], content: { version: 'test-2' }, key: 'test', priority: 2, title: 'test' },
      { children: [], content: { more: 'test-4', version: 'test-3' }, key: 'a.whole.lot.of.dots', priority: 3, title: 'a.whole.lot.of.dots' },
      { children: [], content: { more: 'test-6', version: 'test-5' }, key: 'even.more.dots.than.before', priority: 5, title: 'even.more.dots.than.before' }
    ]);
  });
});
