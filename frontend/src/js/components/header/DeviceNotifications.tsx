// Copyright 2017 Northern.tech AS
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
import { Link } from 'react-router-dom';

// material ui
import { DeveloperBoard as DeveloperBoardIcon } from '@mui/icons-material';
import { makeStyles } from 'tss-react/mui';

import { MenderTooltipClickable } from '@northern.tech/common-ui/helptips/MenderTooltip';
import pluralize from 'pluralize';

import { DeviceLimitContact } from '../devices/dialogs/PreauthDialog';

const useStyles = makeStyles()(theme => ({
  root: {
    '&:hover .warning span, &:hover .warning svg': {
      color: theme.palette.error.light
    },
    // no hover warning
    '.warning span, .warning svg': {
      color: theme.palette.error.main
    },
    // hover warning
    '&:hover .approaching span, &:hover .approaching svg': {
      color: theme.palette.warning.light
    },
    '.approaching span, .approaching svg': {
      color: theme.palette.warning.main
    }
  }
}));

const DeviceNotifications = ({ className = '', total, limit, pending }) => {
  const { classes } = useStyles();
  const approaching = limit && total / limit > 0.8;
  const warning = limit && limit <= total;
  const classNames = `flexbox center-aligned ${className} ${classes.root}`;
  const content = (
    <div className={classNames}>
      <Link to="/devices" className={`flexbox center-aligned ${warning ? 'warning' : approaching ? 'approaching' : ''}`}>
        <DeveloperBoardIcon className="margin-right-x-small" fontSize="small" />
        <div>{total.toLocaleString()}</div>
        {!!limit && <div id="limit">/{limit.toLocaleString()}</div>}
      </Link>
      {pending ? (
        <Link to="/devices/pending" className={limit && limit < pending + total ? 'warning margin-left-x-small' : 'margin-left-x-small'}>
          {pending.toLocaleString()} pending
        </Link>
      ) : null}
    </div>
  );
  if (!limit) {
    return content;
  }
  return (
    <MenderTooltipClickable
      className={classNames}
      disabled={!limit}
      disableHoverListener={false}
      enterDelay={500}
      title={
        <>
          <h3>Device limit</h3>
          {approaching || warning ? (
            <p>You {approaching ? <span>are nearing</span> : <span>have reached</span>} your device limit.</p>
          ) : (
            <p>
              You can still connect another {(limit - total).toLocaleString()} {pluralize('devices', limit - total)}.
            </p>
          )}
          <DeviceLimitContact />
          <p>
            Learn about the different plans available by visiting {/* eslint-disable-next-line react/jsx-no-target-blank */}
            <a href="https://mender.io/pricing" target="_blank" rel="noopener">
              mender.io/pricing
            </a>
          </p>
        </>
      }
    >
      {content}
    </MenderTooltipClickable>
  );
};
export default DeviceNotifications;
