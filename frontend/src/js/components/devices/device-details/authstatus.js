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
import React, { useRef } from 'react';
import { useSelector } from 'react-redux';

import { Block as BlockIcon, CheckCircle as CheckCircleIcon, Check as CheckIcon } from '@mui/icons-material';
import { Chip, Icon } from '@mui/material';

import pendingIcon from '../../../../assets/img/pending_status.png';
import { DEVICE_STATES } from '../../../constants/deviceConstants';
import { onboardingSteps } from '../../../constants/onboardingConstants';
import { getOnboardingState } from '../../../selectors';
import { getOnboardingComponentFor } from '../../../utils/onboardingmanager';
import { HELPTOOLTIPS, MenderHelpTooltip } from '../../helptips/helptooltips';
import Authsets from './authsets/authsets';
import DeviceDataCollapse from './devicedatacollapse';

const iconStyle = { margin: 12 };

const states = {
  default: <Icon style={iconStyle} component="img" src={pendingIcon} />,
  pending: <Icon style={iconStyle} component="img" src={pendingIcon} />,
  accepted: <CheckCircleIcon className="green" style={iconStyle} />,
  rejected: <BlockIcon className="red" style={iconStyle} />,
  preauthorized: <CheckIcon style={iconStyle} />
};

export const AuthStatus = ({ decommission, device }) => {
  const listRef = useRef();
  const statusRef = useRef();
  const onboardingState = useSelector(getOnboardingState);
  const { auth_sets = [], status = DEVICE_STATES.accepted } = device;

  let hasPending = '';
  if (status === DEVICE_STATES.accepted && auth_sets.length > 1) {
    hasPending = auth_sets.reduce((accu, set) => {
      return set.status === DEVICE_STATES.pending ? 'This device has a pending authentication set' : accu;
    }, hasPending);
  }

  const statusIcon = states[status] ? states[status] : states.default;
  const requestNotification = !!hasPending && <Chip size="small" label="new request" color="primary" />;
  let onboardingComponent;
  if (listRef.current?.querySelector('.action-buttons')) {
    const anchor = {
      top: listRef.current.offsetTop + listRef.current.offsetHeight,
      left: listRef.current.offsetLeft + listRef.current.querySelector('.action-buttons').offsetLeft + 15
    };
    onboardingComponent = getOnboardingComponentFor(
      onboardingSteps.DEVICES_PENDING_ACCEPTING_ONBOARDING,
      onboardingState,
      { anchor, place: 'top' },
      onboardingComponent
    );
  }
  if (statusRef.current) {
    const anchor = {
      top: statusRef.current.offsetTop + statusRef.current.offsetHeight,
      left: statusRef.current.offsetLeft + statusRef.current.offsetWidth / 2
    };
    onboardingComponent = getOnboardingComponentFor(
      onboardingSteps.DEVICES_ACCEPTED_ONBOARDING,
      onboardingState,
      { anchor, place: 'top' },
      onboardingComponent
    );
  }
  return (
    <DeviceDataCollapse
      title={
        <div className="flexbox center-aligned">
          <h4>Authentication status</h4>
          <div className="flexbox center-aligned margin-left margin-right" ref={statusRef}>
            <div className="capitalized">{status}</div>
            {statusIcon}
          </div>
          {requestNotification}
          {status === DEVICE_STATES.pending && <MenderHelpTooltip id={HELPTOOLTIPS.authButton.id} style={{ marginTop: 5 }} />}
        </div>
      }
    >
      <Authsets decommission={decommission} device={device} listRef={listRef} />
      {onboardingComponent}
    </DeviceDataCollapse>
  );
};

export default AuthStatus;
