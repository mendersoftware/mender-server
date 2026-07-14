// Copyright 2021 Northern.tech AS
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
import { useSelector } from 'react-redux';

import { Chip, Typography } from '@mui/material';

import { mdiFlaskOutline as TestIcon } from '@mdi/js';
import { ContentSection } from '@northern.tech/common-ui/ContentSection';
import DeviceNameInput from '@northern.tech/common-ui/DeviceNameInput';
import MaterialDesignIcon from '@northern.tech/common-ui/MaterialDesignIcon';
import Time from '@northern.tech/common-ui/Time';
import { TwoColumnData } from '@northern.tech/common-ui/TwoColumnData';
import { MenderTooltipClickable } from '@northern.tech/common-ui/helptips/MenderTooltip';
import { DEVICE_STATES } from '@northern.tech/store/constants';
import { getTestDeviceCount } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { setDeviceListState } from '@northern.tech/store/thunks';

import { TestDeviceLimit } from '../widgets/TestDeviceLimit';
import AuthStatus from './AuthStatus';
import DeviceTags from './DeviceTags';

const TestDeviceTooltip = ({ testDeviceUsed }: { testDeviceUsed: number }) => {
  const dispatch = useAppDispatch();
  const onNavigate = () => dispatch(setDeviceListState({ selectedId: undefined, detailsTab: '' }));
  return (
    <div style={{ maxWidth: 350 }}>
      <TestDeviceLimit testDeviceUsed={testDeviceUsed} onNavigate={onNavigate} />
      <Typography className="margin-top-small">
        Enable up to 10 test devices to bypass rate limits and check in more frequently. Set or remove the ‘test device’ status from the <b>Device actions</b>{' '}
        menu.
      </Typography>
    </div>
  );
};

export const DeviceIdentity = ({ device, setSnackbar }) => {
  const { created_ts, tier, id, identity_data = {}, status = DEVICE_STATES.accepted, flags = {} } = device;
  const isTestDevice = !!flags.test_device;
  const testDeviceCount = useSelector(getTestDeviceCount);
  const { mac, status: _status, ...remainingIdentity } = identity_data;

  const content = {
    ID: id || '-',
    ...(mac ? { mac } : {}),
    ...(tier ? { tier } : {}),
    ...remainingIdentity
  };

  if (created_ts) {
    const createdTime = <Time value={created_ts} />;
    content[status === DEVICE_STATES.preauth ? 'Date added' : 'First request'] = createdTime;
  }

  return (
    <ContentSection
      title="Device identity"
      postTitle={
        isTestDevice ? (
          <MenderTooltipClickable title={<TestDeviceTooltip testDeviceUsed={testDeviceCount} />} arrow>
            <Chip clickable label="Test device" size="small" icon={<MaterialDesignIcon path={TestIcon} />} />
          </MenderTooltipClickable>
        ) : null
      }
    >
      <TwoColumnData data={{ Name: <DeviceNameInput device={device} isHovered />, ...content }} setSnackbar={setSnackbar} />
    </ContentSection>
  );
};

export default DeviceIdentity;

export const IdentityTab = ({ device, setSnackbar, userCapabilities, onDecommissionDevice }) => (
  <>
    <DeviceIdentity device={device} setSnackbar={setSnackbar} />
    <AuthStatus device={device} decommission={onDecommissionDevice} />
    <DeviceTags device={device} setSnackbar={setSnackbar} userCapabilities={userCapabilities} />
  </>
);
