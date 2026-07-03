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
import { useState } from 'react';
import { useSelector } from 'react-redux';

// material ui
import { Button } from '@mui/material';
import { alpha } from '@mui/material/styles';
import { makeStyles } from 'tss-react/mui';

import Confirm from '@northern.tech/common-ui/Confirm';
import { DEVICE_DISMISSAL_STATE, DEVICE_STATES, onboardingSteps } from '@northern.tech/store/constants';
import { getCombinedLimit, getDeviceCountsByStatus, getLimitMaxed, getUserCapabilities } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { advanceOnboarding, deleteAuthset, updateDeviceAuth } from '@northern.tech/store/thunks';
import pluralize from 'pluralize';

import { HELPTOOLTIPS } from '../../../helptips/HelpTooltips';
import { MenderHelpTooltip } from '../../../helptips/MenderTooltip';
import { DeviceLimitWarning } from '../../dialogs/PreauthDialog';
import Authsetlist from './AuthSetList';

const useStyles = makeStyles()(theme => ({
  wrapper: {
    backgroundColor:
      theme.palette.mode === 'dark'
        ? alpha(theme.palette.grey[300], theme.palette.action.selectedOpacity)
        : alpha(theme.palette.grey[400], theme.palette.action.hoverOpacity),
    minWidth: 700
  }
}));

export const Authsets = ({ decommission, device, listRef }) => {
  const [confirmDecommission, setConfirmDecommission] = useState(false);
  const [loading, setLoading] = useState(false);
  const dispatch = useAppDispatch();
  const { accepted: acceptedDevices } = useSelector(getDeviceCountsByStatus);
  const deviceLimit = useSelector(getCombinedLimit);
  const limitMaxed = useSelector(getLimitMaxed);
  const userCapabilities = useSelector(getUserCapabilities);
  const { classes } = useStyles();
  const { auth_sets = [], status = DEVICE_STATES.accepted } = device;
  const { canManageDevices } = userCapabilities;

  const updateDeviceAuthStatus = (deviceId, authId, status) => {
    setLoading(authId);
    // call API to update authset
    const request =
      status === DEVICE_DISMISSAL_STATE ? dispatch(deleteAuthset({ deviceId, authId })) : dispatch(updateDeviceAuth({ deviceId, authId, status }));
    // on finish, change "loading" back to null
    return request.then(() => dispatch(advanceOnboarding(onboardingSteps.DEVICES_PENDING_ACCEPTING_ONBOARDING))).finally(() => setLoading(false));
  };

  return (
    <div className={`${classes.wrapper} padding-medium`}>
      <div className="margin-bottom-small flexbox space-between">
        {status === DEVICE_STATES.pending ? `Authentication ${pluralize('request', auth_sets.length)}` : 'Authentication sets'}
        <MenderHelpTooltip id={HELPTOOLTIPS.authExplainButton.id} small className="margin-left-small" />
      </div>
      <Authsetlist
        limitMaxed={limitMaxed}
        listRef={listRef}
        total={auth_sets.length}
        confirm={updateDeviceAuthStatus}
        loading={loading}
        device={device}
        userCapabilities={userCapabilities}
      />
      {limitMaxed && <DeviceLimitWarning acceptedDevices={acceptedDevices} deviceLimit={deviceLimit} hasContactInfo />}
      {![DEVICE_STATES.preauth, DEVICE_STATES.pending].includes(device.status) && canManageDevices && (
        <div className="margin-top-small flexbox relative">
          {confirmDecommission ? (
            <Confirm
              action={() => decommission(device.id)}
              cancel={() => setConfirmDecommission(false)}
              classes="margin-top-x-small margin-bottom-x-small"
              type="decommissioning"
            />
          ) : (
            <Button color="error" variant="outlined" onClick={() => setConfirmDecommission(true)}>
              Decommission device
            </Button>
          )}
        </div>
      )}
    </div>
  );
};

export default Authsets;
