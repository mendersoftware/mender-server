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
import { Link } from 'react-router-dom';

// material ui
import { Add as ContentAddIcon } from '@mui/icons-material';
import { Fab } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import pluralize from 'pluralize';

import { onboardingSteps } from '../../../utils/onboardingmanager';
import { BaseWidget } from './baseWidget';

const useStyles = makeStyles()(theme => ({
  fab: { top: '-28px', right: theme.spacing(2), zIndex: 1 }
}));

export const PendingDevices = props => {
  const { advanceOnboarding, isActive: hasPending, innerRef, onboardingState, onClick, pendingDevicesCount } = props;

  const { classes } = useStyles();

  const onWidgetClick = () => {
    if (!onboardingState.complete) {
      advanceOnboarding(onboardingSteps.DEVICES_PENDING_ONBOARDING);
    }
    onClick({ route: '/devices/pending' });
  };

  const pendingNotification = `Pending ${pluralize('devices', hasPending)}`;

  const widgetMain = {
    counter: pendingDevicesCount,
    targetLabel: 'View details'
  };

  return (
    <div className="relative" ref={ref => (innerRef ? (innerRef.current = ref) : null)}>
      <Fab className={`absolute ${classes.fab}`} color="primary" component={Link} to="/devices/pending">
        <ContentAddIcon />
      </Fab>
      <BaseWidget {...props} header={pendingNotification} main={widgetMain} onClick={onWidgetClick} />
    </div>
  );
};

export default PendingDevices;
