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
import { DeveloperBoard as DeveloperBoardIcon, Warning as WarningIcon } from '@mui/icons-material';
import { LinearProgress } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

interface DeviceCountProps {
  current: number;
  max: number;
  variant: 'common' | 'detailed';
}

export const LIMIT_THRESHOLD = 0.8;
const useStyles = makeStyles()(theme => ({
  devLimitProgress: {
    width: '375px'
  },
  devLimitDetailed: {
    paddingBottom: '11px',
    borderBottom: `1px solid ${theme.palette.grey['600']}`
  },
  devIcon: {
    color: theme.palette.primary.main,
    margin: '0 7px 0 10px',
    fontSize: '20px'
  },
  devLimitText: {
    margin: '8px 0',
    color: theme.palette.grey['900']
  },
  devText: {
    margin: '11px 11px 11px 0'
  },
  devLeft: {
    background: theme.palette.grey['300'],
    color: theme.palette.grey['900'],
    borderRadius: '2px',
    padding: '4px 6px'
  },
  progressBar: {
    backgroundColor: theme.palette.grey['500']
  }
}));
export const DeviceCount = (props: DeviceCountProps) => {
  const { classes } = useStyles();
  const { current, max, variant } = props;
  return (
    <div className={`${classes.devLimitProgress} ${variant === 'detailed' ? classes.devLimitDetailed : ''}`}>
      {variant === 'common' && (
        <div className="flexbox centered">
          <DeveloperBoardIcon className={classes.devIcon} />
          <p className={classes.devLimitText}>
            {current} of {max} devices
          </p>
        </div>
      )}
      {variant === 'detailed' && (
        <div className="flexbox centered space-between">
          <div className="flexbox centered">
            <p className={classes.devText}>
              Devices: {current}/{max}
            </p>
            {current / max >= LIMIT_THRESHOLD && <WarningIcon />}
          </div>
          <div className={classes.devLeft}>{Math.max(0, max - current)} devices left</div>
        </div>
      )}
      <LinearProgress className={classes.progressBar} variant="determinate" value={Math.round((current / max) * 100)} />
    </div>
  );
};
