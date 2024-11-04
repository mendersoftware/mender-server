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
import React, { useEffect, useRef } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { CheckCircle as CheckCircleIcon } from '@mui/icons-material';
import { Button } from '@mui/material';
import { withStyles } from 'tss-react/mui';

import Loader from '@northern.tech/common-ui/loader';
import { MenderTooltipClickable } from '@northern.tech/common-ui/mendertooltip';
import { DEVICE_STATES, onboardingSteps } from '@northern.tech/store/constants';
import { getDemoDeviceAddress } from '@northern.tech/store/selectors';
import { getDeviceById, getDevicesByStatus, setOnboardingComplete } from '@northern.tech/store/thunks';

export const CompletionButton = withStyles(Button, ({ palette }) => ({
  root: {
    backgroundColor: palette.background.default,
    '&:hover': {
      backgroundColor: palette.background.default
    }
  }
}));

export const OnboardingCompleteTip = ({ anchor, targetUrl }) => {
  const timer = useRef();
  const dispatch = useDispatch();
  const url = useSelector(getDemoDeviceAddress) || targetUrl;

  useEffect(() => {
    dispatch(getDevicesByStatus({ status: DEVICE_STATES.accepted }))
      .unwrap()
      .then(tasks => {
        return Promise.all(tasks[tasks.length - 1].deviceAccu.ids.map(id => dispatch(getDeviceById(id))));
      })
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
      PopperProps={{ style: { marginLeft: -30, marginTop: -20 } }}
      title={
        <div className="content">
          <b>
            <p>Fantastic! You completed your first deployment!</p>
            <p>Your deployment is finished and your device is now running the updated software.</p>
          </b>
          <div className="margin-bottom-small margin-top-small">
            {!url ? (
              <Loader show={true} />
            ) : (
              <CompletionButton className="button" variant="text" href={`${url}/index.html?source=${encodeURIComponent(window.location)}`} target="_blank">
                {`Go to ${url}`}
              </CompletionButton>
            )}
            <br />
            and you should see the demo web application actually being run on the device.
          </div>
          <p>NOTE: if you have local network restrictions, you may need to check them if you have difficulty loading the page.</p>
          <div className="flexbox">
            <div style={{ flexGrow: 1 }} />
            <Button variant="contained" color="secondary" onClick={() => dispatch(setOnboardingComplete(true))}>
              Close
            </Button>
          </div>
        </div>
      }
    >
      <CheckCircleIcon />
    </MenderTooltipClickable>
  );
};

export default OnboardingCompleteTip;
