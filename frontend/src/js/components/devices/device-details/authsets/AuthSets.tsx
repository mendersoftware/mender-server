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
import { useDispatch, useSelector } from 'react-redux';

// material ui
import { Button } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import Confirm from '@northern.tech/common-ui/Confirm';
import { DEVICE_DISMISSAL_STATE, DEVICE_STATES, onboardingSteps } from '@northern.tech/store/constants';
import { getAcceptedDevices, getDeviceLimit, getLimitMaxed, getUserCapabilities } from '@northern.tech/store/selectors';
import { advanceOnboarding, deleteAuthset, updateDeviceAuth } from '@northern.tech/store/thunks';
import pluralize from 'pluralize';

import { HELPTOOLTIPS } from '../../../helptips/HelpTooltips';
import { MenderHelpTooltip } from '../../../helptips/MenderTooltip';
import { DeviceLimitWarning } from '../../dialogs/PreauthDialog';
import Authsetlist from './AuthSetList';

const useStyles = makeStyles()(theme => ({
  decommission: { justifyContent: 'flex-end', marginTop: theme.spacing(2) },
  wrapper: {
    backgroundColor: theme.palette.background.lightgrey ? theme.palette.grey[400] : theme.palette.info.light,
    marginBottom: theme.spacing(2),
    minWidth: 700,
    padding: theme.spacing(2)
  }
}));

export const Authsets = ({ decommission, device, listRef }) => {
  const [confirmDecommission, setConfirmDecomission] = useState(false);
  const [loading, setLoading] = useState(false);
  const dispatch = useDispatch();
  const { total: acceptedDevices = 0 } = useSelector(getAcceptedDevices);
  const deviceLimit = useSelector(getDeviceLimit);
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
    return request.then(() => dispatch(advanceOnboarding(onboardingSteps.DEVICES_PENDING_ACCEPTING_ONBOARDING))).finally(() => setLoading(null));
  };

  return (
    <div className={classes.wrapper}>
      <div className="margin-bottom-small flexbox space-between">
        {status === DEVICE_STATES.pending ? `Authorization ${pluralize('request', auth_sets.length)}` : 'Authorization sets'}
        <MenderHelpTooltip id={HELPTOOLTIPS.authExplainButton.id} className="margin-left-small" />
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
        <div className={`flexbox ${classes.decommission}`}>
          {confirmDecommission ? (
            <Confirm action={() => decommission(device.id)} cancel={() => setConfirmDecomission(false)} type="decommissioning" />
          ) : (
            <Button color="secondary" onClick={setConfirmDecomission}>
              Decommission device
            </Button>
          )}
        </div>
      )}
    </div>
  );
};

export default Authsets;
