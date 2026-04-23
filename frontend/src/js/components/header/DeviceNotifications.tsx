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
import type { MouseEvent } from 'react';
import { useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import { Link } from 'react-router-dom';

// material ui
import {
  ArrowDropDown as ArrowDropDownIcon,
  DeveloperBoard as DeveloperBoardIcon,
  ErrorOutline as ErrorOutlineIcon,
  WarningAmber as WarningAmberIcon
} from '@mui/icons-material';
import { Alert, Badge, Button, Divider, LinearProgress, Paper, Popover, Tooltip, Typography, alpha } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { getDeviceLimitStats, getUserRoles } from '@northern.tech/store/selectors';
import pluralize from 'pluralize';

const useStyles = makeStyles()(theme => ({
  primary: {
    color: theme.palette.text.secondary,
    '--pressed-color': alpha(theme.palette.text.secondary, 0.3)
  },
  error: {
    color: theme.palette.error.main,
    '&:hover': { backgroundColor: 'unset' },
    '--pressed-color': alpha(theme.palette.error.main, 0.3)
  },
  warning: {
    color: theme.palette.warning.main,
    '&:hover': { backgroundColor: 'unset' },
    '--pressed-color': alpha(theme.palette.warning.main, 0.3)
  },
  endIcon: {
    color: theme.palette.text.secondary
  },
  limitContainer: {
    width: 'calc(100% + 12px)',
    marginLeft: '-6px',
    borderRadius: 4
  },
  warningBg: {
    background: alpha(theme.palette.warning.light, 0.1)
  },
  errorBg: {
    background: alpha(theme.palette.error.light, 0.1)
  },
  popover: { minWidth: '330px', padding: theme.spacing(2) },
  pressed: {
    backgroundColor: 'var(--pressed-color)'
  }
}));
interface DeviceLimitProps {
  className?: string;
  compact?: boolean;
  disabled?: boolean;
  disablePadding?: boolean;
  limit: number;
  serviceProvider?: boolean;
  total: number;
  type: string;
}

const numberLocale = 'en-US';
export const getLimitStatus = (total: number, limit: number) => {
  const unlimited = limit === -1;
  const percentageUsed = unlimited || limit === 0 ? 0 : Math.floor((total / limit) * 100);
  const warning = !unlimited && percentageUsed > 80 && total < limit;
  const error = !unlimited && total >= limit && total !== 0;
  const color = error ? 'error' : warning ? 'warning' : 'primary';
  return { unlimited, percentageUsed, error, warning, color };
};
export const DeviceLimit = (props: DeviceLimitProps) => {
  const { className = '', type, limit, total, compact = false, serviceProvider = false, disabled = false, disablePadding = true } = props;
  const { warning, error, percentageUsed, color, unlimited } = getLimitStatus(total, limit);
  const { classes } = useStyles();
  const textColor = disabled ? 'text.disabled' : 'text.primary';
  return (
    <Paper
      variant={disablePadding ? 'elevation' : 'outlined'}
      elevation={disablePadding ? 0 : 1}
      className={`flexbox column ${disablePadding ? 'padding-x-small' : 'padding-small'} ${disabled ? 'disabled' : ''} ${classes.limitContainer} ${warning ? classes.warningBg : ''} ${error ? classes.errorBg : ''} ${className}`}
    >
      <div className="flexbox full-width space-between">
        <div className="flexbox">
          <Typography color={textColor} variant="subtitle2" className="capitalized-start">
            {type}
          </Typography>
          {warning && <WarningAmberIcon fontSize="small" color="warning" className="margin-left-x-small" />}
          {error && <ErrorOutlineIcon fontSize="small" color="error" className="margin-left-x-small" />}
        </div>
        <Typography color={textColor} variant="body2" data-testid={`device-limit-${type}`}>
          {total.toLocaleString(numberLocale)}
          {!unlimited && `/${limit.toLocaleString(numberLocale)}`}
        </Typography>
      </div>
      {!unlimited && (
        <>
          <div className="margin-top-x-small">
            <LinearProgress color={disabled ? 'inherit' : color} variant="determinate" value={percentageUsed} />
          </div>

          {!compact && (
            <div className="flexbox margin-top-x-small space-between">
              <div className="flexbox">
                {disabled ? (
                  <Typography variant="caption" color={textColor}>
                    Not enabled for this tenant
                  </Typography>
                ) : (
                  <Typography variant="caption">
                    {percentageUsed}% {serviceProvider ? 'allocated' : 'used'}
                  </Typography>
                )}
                {(warning || error) && (
                  <>
                    <Typography variant="caption" className="margin-left-x-small margin-right-x-small">
                      •
                    </Typography>{' '}
                    <Typography variant="caption" color={color}>
                      {warning ? 'Near limit' : 'Limit reached'}
                    </Typography>
                  </>
                )}
              </div>
              {serviceProvider && <Typography variant="caption">{limit - total} remaining</Typography>}
            </div>
          )}
        </>
      )}
    </Paper>
  );
};

const DeviceNotifications = ({ className = '', total, pending }) => {
  const { classes } = useStyles();
  const navigate = useNavigate();
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [tooltip, setTooltip] = useState(false);
  const { isAdmin } = useSelector(getUserRoles);
  const mappedLimits = useSelector(getDeviceLimitStats);

  const handleOpen = (event: MouseEvent<HTMLButtonElement>) => {
    setAnchorEl(event.currentTarget);
    setTooltip(false);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };
  const viewPendingClick = () => {
    handleClose();
    navigate('/devices/pending');
  };

  const severityMap = { 0: 'primary', 1: 'warning', 2: 'error' };

  const maxSeverityIndex = mappedLimits.reduce((maxIndex, { limit, total }) => {
    let severity = 0;

    if (limit !== -1) {
      if (total >= limit) {
        severity = 2;
      } else if (total / limit > 0.8) {
        severity = 1;
      }
    }
    return Math.max(maxIndex, severity);
  }, 0);

  const severity = severityMap[maxSeverityIndex];

  return (
    <div className={className}>
      <Tooltip
        title="View device limits"
        open={!anchorEl && tooltip}
        disableFocusListener
        disableTouchListener
        disableHoverListener={!!anchorEl}
        onMouseEnter={() => !anchorEl && setTooltip(true)}
        onMouseLeave={() => setTooltip(false)}
      >
        <div>
          <Badge color={severity} badgeContent={pending} anchorOrigin={{ horizontal: 'left' }}>
            <Button
              startIcon={<DeveloperBoardIcon className="margin-right-x-small margin-left-x-small" fontSize="small" />}
              endIcon={<ArrowDropDownIcon fontSize="small" className={classes.endIcon} />}
              className={`flexbox align-items-center ${classes[severity]} ${!!anchorEl && classes.pressed}`}
              onClick={handleOpen}
            >
              {total.toLocaleString(numberLocale)}
            </Button>
          </Badge>
          <Popover
            onClose={handleClose}
            anchorEl={anchorEl}
            open={Boolean(anchorEl)}
            anchorOrigin={{
              vertical: 'bottom',
              horizontal: 'right'
            }}
            transformOrigin={{
              vertical: 'top',
              horizontal: 'right'
            }}
            slotProps={{ paper: { className: classes.popover } }}
          >
            {!!pending && (
              <Alert
                severity="info"
                className="flexbox align-items-center margin-bottom-small"
                onClick={() => viewPendingClick()}
                slotProps={{ message: { className: 'flexbox align-items-center space-between full-width' } }}
              >
                <Typography variant="body2">
                  {pending} {pluralize('device', pending)} pending
                </Typography>
                <Button className="padding-bottom-none padding-top-none" size="small" variant="text">
                  View
                </Button>
              </Alert>
            )}
            <div className="margin-bottom-x-small">
              <div className="flexbox space-between full-width">
                <div className="flexbox centered">
                  <DeveloperBoardIcon className="margin-right-x-small" fontSize="small" />
                  <Typography variant="subtitle1">Accepted devices</Typography>
                </div>
                <Typography variant="subtitle1" color="text.secondary">
                  {total.toLocaleString(numberLocale)}
                </Typography>
              </div>
            </div>
            <Divider />
            <div className="margin-top-small">
              {mappedLimits.map(limit => (
                <div key={limit.type} className="margin-bottom-small">
                  <DeviceLimit {...limit} />
                </div>
              ))}
            </div>
            {isAdmin && (
              <Button color="primary" variant="text" component={Link} to="/settings/subscription" onClick={handleClose}>
                Manage device limit
              </Button>
            )}
          </Popover>
        </div>
      </Tooltip>
    </div>
  );
};
export default DeviceNotifications;
