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
import { useEffect, useRef } from 'react';
import { useSelector } from 'react-redux';

import { CheckCircle as CheckCircleIcon } from '@mui/icons-material';
import { Button, Typography } from '@mui/material';

import Loader from '@northern.tech/common-ui/Loader';
import { MenderTooltipClickable } from '@northern.tech/common-ui/helptips/MenderTooltip';
import { DEVICE_STATES, onboardingSteps } from '@northern.tech/store/constants';
import { getDemoDeviceAddress } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { getDeviceById, getDevicesByStatus, setOnboardingComplete } from '@northern.tech/store/thunks';
import type { PositioningStrategy } from '@popperjs/core';

export const OnboardingCompleteTip = ({ anchor, targetUrl }) => {
  const timer = useRef();
  const dispatch = useAppDispatch();
  const url = useSelector(getDemoDeviceAddress) || targetUrl;

  useEffect(() => {
    dispatch(getDevicesByStatus({ status: DEVICE_STATES.accepted }))
      .unwrap()
      .then(({ deviceAccu }) => Promise.all(deviceAccu.ids.map(id => dispatch(getDeviceById(id)))))
      .finally(() => {
        timer.current = setTimeout(() => dispatch(setOnboardingComplete(true)), 120000);
      });
    return () => {
      clearTimeout(timer.current);
    };
  }, [dispatch]);

  return (
    <MenderTooltipClickable
      className="tooltip onboard-icon onboard-tip"
      id={onboardingSteps.DEPLOYMENTS_PAST_COMPLETED}
      onboarding
      startOpen
      style={anchor}
      slotProps={{
        popper: {
          disablePortal: true,
          popperOptions: {
            strategy: 'fixed' as PositioningStrategy,
            modifiers: [
              { name: 'flip', enabled: false },
              { name: 'preventOverflow', enabled: true, options: { boundary: window, altBoundary: false } }
            ]
          }
        }
      }}
      title={
        <>
          <Typography variant="subtitle2">
            <p>Fantastic! You completed your first deployment!</p>
          </Typography>
          <Typography className="flexbox column margin-bottom-small margin-top-small gap-small" component="div" variant="body2">
            Your deployment is finished and your device is now running the updated software.
            {!url ? (
              <Loader show={true} />
            ) : (
              <Button
                className="align-self-start"
                href={`${url}/index.html?source=${encodeURIComponent(window.location)}`}
                target="_blank"
                variant="text"
                color="inherit"
              >
                {`Go to ${url}`}
              </Button>
            )}
            and you should see the demo web application actually being run on the device.
            <Typography variant="body2" color="textSecondary">
              NOTE: if you have local network restrictions, you may need to check them if you have difficulty loading the page.
            </Typography>
            <Button className="align-self-end" variant="contained" color="secondary" onClick={() => dispatch(setOnboardingComplete(true))}>
              Close
            </Button>
          </Typography>
        </>
      }
    >
      <CheckCircleIcon />
    </MenderTooltipClickable>
  );
};

export default OnboardingCompleteTip;
