// Copyright 2024 Northern.tech AS
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
import React, { useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { Button, Dialog, DialogActions, DialogContent, DialogTitle, Divider } from '@mui/material';

import storeActions from '@northern.tech/store/actions';
import { DEVICE_ONLINE_CUTOFF, TIMEOUTS } from '@northern.tech/store/constants';
import { getIsDarkMode } from '@northern.tech/store/selectors';
import { saveGlobalSettings } from '@northern.tech/store/thunks';
import { useDebounce } from '@northern.tech/utils/debouncehook';

import logo from '../../../../assets/img/headerlogo.png';
import whiteLogo from '../../../../assets/img/whiteheaderlogo.png';

const { setShowStartupNotification } = storeActions;

const OfflineThresholdContent = () => (
  <>
    In our continuous efforts to enhance performance and to ensure the stability of our service, we have made adjustments to how granular device connectivity
    can be checked.
    <h4>What&apos;s changing:</h4>
    <ul>
      <li>
        <i>Offline threshold</i> setting: if a device no longer connects with the server, it may take up to 1 day until it is shown as an offline device in the
        UI.
      </li>
    </ul>
    More granular connectivity information will still be reflected in the device details.
    <h4>Why are we changing this:</h4>
    <ul>
      <li>Performance Improvements: Less frequent updates allow us to better optimize the platform and improve overall performance.</li>
      <li>
        Stability: This change helps ensure that our backend will remain stable while serving an ever growing number of devices and potential disruptions are
        minimized.
      </li>
    </ul>
    We appreciate your understanding and cooperation as we implement these improvements. Our commitment to providing you with a reliable and high-performing
    platform remains our top priority.
    <Divider className="margin-top-small margin-bottom-small" />
    If you have any questions or concerns regarding this change, please do not hesitate to{' '}
    <a href="mailto:support@mender.io" target="_blank" rel="noopener noreferrer">
      contact our team
    </a>
    .
  </>
);

const notifications = {
  offlineThreshold: {
    Content: OfflineThresholdContent,
    action: ({ dispatch }) => dispatch(saveGlobalSettings({ offlineThreshold: { interval: 1, intervalUnit: DEVICE_ONLINE_CUTOFF.intervalName } }))
  }
};

export const StartupNotificationDialog = () => {
  const [isAllowedToClose] = useState(false);
  const dispatch = useDispatch();
  const isDarkMode = useSelector(getIsDarkMode);

  const { action, Content } = notifications.offlineThreshold;

  const debouncedCloseGuard = useDebounce(isAllowedToClose, TIMEOUTS.fiveSeconds);

  const onClose = () => {
    action({ dispatch });
    dispatch(setShowStartupNotification(false));
  };
  const headerLogo = isDarkMode ? whiteLogo : logo;
  return (
    <Dialog open PaperProps={{ className: 'padding-small', sx: { maxWidth: 720 } }}>
      <DialogTitle className="flexbox center-aligned">
        <img src={headerLogo} style={{ maxHeight: 75 }} />
        <div className="margin-left-small">Welcome back!</div>
      </DialogTitle>
      <DialogContent className="margin-left-small margin-right-small">
        <Content />
      </DialogContent>
      <DialogActions>
        <Button variant="contained" disabled={debouncedCloseGuard} onClick={onClose}>
          Close
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default StartupNotificationDialog;
