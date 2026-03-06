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
import { PendingOutlined } from '@mui/icons-material';
import { Button, Typography } from '@mui/material';

import { onboardingSteps } from '../../../utils/onboardingManager';
import { BaseWidget } from './BaseWidget';

export const PendingDevices = props => {
  const { advanceOnboarding, innerRef, onboardingState, onClick, pendingDevicesCount } = props;

  const onWidgetClick = () => {
    if (!onboardingState.complete) {
      advanceOnboarding(onboardingSteps.DEVICES_PENDING_ONBOARDING);
    }
    onClick({ route: '/devices/pending' });
  };

  const header = (
    <>
      Pending devices
      <PendingOutlined fontSize="small" className="margin-left-small" />
    </>
  );

  const main = (
    <div className="flexbox column full-width" ref={ref => (innerRef ? (innerRef.current = ref) : null)}>
      <Typography variant="h5">{(pendingDevicesCount || 0).toLocaleString('en-US')}</Typography>
      <Button className="align-self-start margin-top-x-small" color="primary" size="small" variant="text">
        View devices
      </Button>
    </div>
  );

  return <BaseWidget {...props} header={header} main={main} onClick={onWidgetClick} />;
};

export default PendingDevices;
