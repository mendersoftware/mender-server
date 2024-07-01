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
import React, { useEffect, useRef, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import {
  ArrowBack as ArrowBackIcon,
  ArrowDownward as ArrowDownwardIcon,
  ArrowForward as ArrowForwardIcon,
  ArrowUpward as ArrowUpwardIcon
} from '@mui/icons-material';
import { buttonBaseClasses, buttonClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { setShowDismissOnboardingTipsDialog } from '../../actions/onboardingActions';
import { TIMEOUTS } from '../../constants/appConstants';
import { toggle } from '../../helpers';
import Tracking from '../../tracking';
import { OnboardingTooltip } from '../common/mendertooltip';

const iconWidth = 30;

export const orientations = {
  top: {
    arrow: <ArrowUpwardIcon />,
    placement: 'bottom',
    offsetStyle: style => {
      style.left = style.left - iconWidth / 2;
      return style;
    }
  },
  right: {
    arrow: <ArrowBackIcon />,
    placement: 'right',
    offsetStyle: style => {
      style.top = style.top - iconWidth / 2;
      style.left = style.left + iconWidth / 2;
      return style;
    }
  },
  bottom: {
    arrow: <ArrowDownwardIcon />,
    placement: 'top',
    offsetStyle: style => {
      style.left = style.left - iconWidth / 2;
      return style;
    }
  },
  left: {
    arrow: <ArrowForwardIcon />,
    placement: 'left',
    offsetStyle: style => {
      style.top = style.top - iconWidth / 2;
      return style;
    }
  }
};

const useStyles = makeStyles()(theme => ({
  root: {
    [`.${buttonBaseClasses.root}.${buttonClasses.text}`]: { color: theme.palette.background.paper, paddingRight: 0 }
  }
}));

export const BaseOnboardingTooltip = ({ anchor = { left: 0, top: 0 }, icon, children, id = '1', place = 'top', ...props }) => {
  const [open, setOpen] = useState(false);
  const showDismissHelptipsDialog = useSelector(state => state.onboarding.showTipsDialog);
  const showDeviceConnectionDialog = useSelector(state => state.users.showConnectDeviceDialog);
  const delayTimer = useRef();

  const { classes } = useStyles();

  useEffect(() => {
    Tracking.event({ category: 'onboarding', action: id });
    clearTimeout(delayTimer.current);
    delayTimer.current = setTimeout(() => setOpen(true), TIMEOUTS.debounceShort);
    return () => {
      clearTimeout(delayTimer.current);
    };
  }, [id]);

  const toggleVisibility = () => setOpen(toggle);

  const hide = () => setOpen(false);

  if (showDismissHelptipsDialog || showDeviceConnectionDialog) {
    return null;
  }

  const orientation = orientations[place];
  const style = orientation.offsetStyle({ left: anchor.left, top: anchor.top, overflow: 'initial' });
  return (
    <OnboardingTooltip
      disableFocusListener
      disableHoverListener
      disableTouchListener
      id={id}
      onClose={hide}
      open={open}
      placement={orientation.placement}
      PopperProps={{
        className: classes.root,
        disablePortal: true,
        popperOptions: {
          strategy: 'fixed',
          modifiers: [
            { name: 'flip', enabled: false },
            { name: 'preventOverflow', enabled: true, options: { boundary: window, altBoundary: false } }
          ]
        }
      }}
      title={children}
      className={classes.root}
      {...props}
    >
      <div className="onboard-tip" onClick={toggleVisibility} style={style}>
        <div className={`tooltip onboard-icon ${orientation.placement}`}>{icon ?? orientation.arrow}</div>
      </div>
    </OnboardingTooltip>
  );
};

export const BaseOnboardingTip = ({ anchor, icon, component, place, id, ...others }) => {
  const dispatch = useDispatch();
  return (
    <BaseOnboardingTooltip anchor={anchor} icon={icon} place={place} id={id}>
      <div className="content">
        {React.cloneElement(component, others)}
        <div>
          <b className="clickable" onClick={() => dispatch(setShowDismissOnboardingTipsDialog(true))}>
            Dismiss the tutorial
          </b>
        </div>
      </div>
    </BaseOnboardingTooltip>
  );
};

export default BaseOnboardingTip;
