// Copyright 2022 Northern.tech AS
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
import React, { useMemo } from 'react';
import { Link } from 'react-router-dom';

// material ui
import { CancelOutlined as FailureIcon, VpnKeyOutlined as KeyIcon, WifiOff as OfflineIcon, WarningAmber as WarningIcon } from '@mui/icons-material';
import { makeStyles } from 'tss-react/mui';

import { DEVICE_ISSUE_OPTIONS } from '../../../constants/deviceConstants';
import { BaseWidget } from './baseWidget';

const issueTypes = [
  {
    key: DEVICE_ISSUE_OPTIONS.offline.key,
    title: 'Offline',
    icon: OfflineIcon,
    target: `/devices?issues=${DEVICE_ISSUE_OPTIONS.offline.key}`,
    alwaysShown: true
  },
  {
    key: DEVICE_ISSUE_OPTIONS.failedLastUpdate.key,
    title: 'Failed last update',
    icon: FailureIcon,
    target: `/devices?issues=${DEVICE_ISSUE_OPTIONS.failedLastUpdate.key}`,
    alwaysShown: false
  },
  {
    key: DEVICE_ISSUE_OPTIONS.monitoring.key,
    title: 'Monitoring',
    icon: WarningIcon,
    target: `/devices?issues=${DEVICE_ISSUE_OPTIONS.monitoring.key}`,
    alwaysShown: false
  },
  {
    key: DEVICE_ISSUE_OPTIONS.authRequests.key,
    title: 'Authentication requests',
    icon: KeyIcon,
    target: `/devices?issues=${DEVICE_ISSUE_OPTIONS.authRequests.key}`,
    alwaysShown: false
  }
];

const useStyles = makeStyles()(theme => ({
  issueContainer: {
    border: `1px solid currentColor`,
    borderRadius: '50%',
    fontSize: '2rem',
    height: '5vh',
    width: '5vh',
    maxWidth: '3em',
    maxHeight: '3em'
  },
  issueType: {
    maxWidth: '6vh',
    ['&.red']: { color: theme.palette.secondary.main },
    ['&.green']: { color: theme.palette.primary.main }
  },
  issueCounter: {
    display: 'grid',
    placeContent: 'center',
    width: '100%',
    height: '100%'
  },
  issueWrapper: {
    bottom: theme.spacing(-1),
    height: 20,
    width: '100%',
    '> div': {
      display: 'grid',
      placeItems: 'center',
      height: '100%',
      width: '100%'
    },
    '> div > div': {
      background: 'currentColor',
      borderRadius: theme.spacing(0.75),
      height: '100%',
      width: theme.spacing(3)
    },
    'svg': {
      fill: theme.palette.background.paper
    }
  },
  widget: {
    '&.widget': { maxWidth: 'initial !important' },
    '.widgetMainContent': {
      columnGap: theme.spacing(2),
      display: 'flex',
      justifyContent: 'space-evenly'
    }
  }
}));

export const ActionableDevices = props => {
  const { classes } = useStyles();
  const { issues } = props;

  const relevantIssues = useMemo(
    () =>
      issueTypes.reduce((accu, item) => {
        if (issues[item.key] || item.alwaysShown) {
          accu.push({ ...item, count: issues[item.key]?.count || 0 });
        }
        return accu;
      }, []),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [JSON.stringify(issues)]
  );

  const widgetMain = (
    <>
      {relevantIssues.map(type => {
        const value = type.count;
        const Icon = type.icon;
        const length = `${value}`.length;
        const counterStyle = length > 2 ? { fontSize: `calc(5vh / ${Math.max(1, length - 1)}.25)` } : {};
        return (
          <Link key={type.key} className={`flexbox center-aligned column ${classes.issueType} ${value ? 'red' : 'green'}`} to={type.target}>
            <div className={`relative ${classes.issueContainer}`}>
              <div className={classes.issueCounter} style={counterStyle}>
                {value}
              </div>
              <div className={`absolute ${classes.issueWrapper}`}>
                <div className="relative">
                  <div className="absolute"></div>
                  <Icon className="absolute" fontSize="small" />
                </div>
              </div>
            </div>
            <p className="align-center slightly-smaller">{type.title}</p>
          </Link>
        );
      })}
    </>
  );
  return <BaseWidget {...props} className={classes.widget} header="Devices with issues" main={widgetMain} />;
};

export default ActionableDevices;
