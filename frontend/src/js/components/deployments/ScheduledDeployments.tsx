// Copyright 2020 Northern.tech AS
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
import { useCallback, useEffect, useRef, useState } from 'react';
import { Calendar, dayjsLocalizer } from 'react-big-calendar';
import 'react-big-calendar/lib/css/react-big-calendar.css';
import { useSelector } from 'react-redux';

import { CalendarToday as CalendarTodayIcon, List as ListIcon, Refresh as RefreshIcon } from '@mui/icons-material';
import { ToggleButton, ToggleButtonGroup, alpha } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { DefaultUpgradeNotification } from '@northern.tech/common-ui/EnterpriseNotification';
import storeActions from '@northern.tech/store/actions';
import { DEPLOYMENT_STATES } from '@northern.tech/store/constants';
import {
  getDeploymentsByStatus as getDeploymentsByStatusSelector,
  getDeploymentsSelectionState,
  getDevicesById,
  getIdAttribute,
  getMappedDeploymentSelection,
  getTenantCapabilities,
  getUserCapabilities
} from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { getDeploymentsByStatus, setDeploymentsState } from '@northern.tech/store/thunks';
import { isDarkMode } from '@northern.tech/store/utils';
import { clearAllRetryTimers, clearRetryTimer, setRetryTimer } from '@northern.tech/utils/retrytimer';
import dayjs from 'dayjs';

import { DeploymentDeviceCount, DeploymentEndTime, DeploymentPhases, DeploymentStartTime } from './DeploymentItem';
import DeploymentsList, { defaultHeaders } from './DeploymentsList';
import { defaultRefreshDeploymentsLength as refreshDeploymentsLength } from './constants';

const { setSnackbar } = storeActions;

const useStyles = makeStyles()(theme => {
  const greyShade = isDarkMode(theme.palette.mode) ? theme.palette.grey[300] : theme.palette.grey[500];
  const defaultBg = isDarkMode(theme.palette.mode) ? alpha(theme.palette.grey[300], theme.palette.action.selectedOpacity) : theme.palette.grey[50];
  return {
    refreshIcon: { fill: theme.palette.grey[400], width: 111, height: 111 },
    tabSelect: { textTransform: 'none' },
    toggleButtonGroup: { display: 'flex', alignContent: 'flex-end' },
    calendarContainer: {
      '& .rbc-toolbar': {
        '& button': {
          background: defaultBg,
          color: theme.palette.info.contrastText,
          borderColor: alpha(greyShade, 0.5),

          '&:hover, &:focus, &:active': {
            background: alpha(greyShade, 0.16),
            color: theme.palette.info.contrastText
          },

          '&.rbc-active': {
            '&, &:hover, &:focus, &:active': {
              background: alpha(greyShade, 0.3),
              color: theme.palette.info.contrastText
            }
          }
        }
      },

      '& .rbc-month-view, & .rbc-time-view': {
        '& .rbc-time-content': { borderTopWidth: '1px' },
        '& .rbc-time-slot': { borderTop: 'none' },
        '& .rbc-off-range-bg': { background: defaultBg },
        '& .rbc-today': { backgroundColor: alpha(theme.palette.primary.main, theme.palette.action.selectedOpacity) },

        '& .rbc-event': {
          backgroundColor: theme.palette.primary.main,
          color: theme.palette.primary.contrastText,
          '&.rbc-selected': {
            backgroundColor: theme.palette.primary.dark
          }
        },

        '& .rbc-show-more': {
          color: theme.palette.primary.main,
          fontWeight: 400,
          background: 'inherit'
        }
      },

      '& .rbc-month-view': {
        '& .rbc-today': { backgroundColor: 'inherit' },
        '& .rbc-now': {
          paddingTop: '3px',
          paddingRight: theme.spacing(1.5),
          '& > .rbc-button-link': {
            color: theme.palette.primary.contrastText,
            backgroundColor: theme.palette.primary.main,
            borderRadius: '100px',
            padding: '3px 7px'
          }
        }
      }
    }
  };
});

const localizer = dayjsLocalizer(dayjs);

const headers = [
  ...defaultHeaders.slice(0, 2),
  { title: 'Start time', renderer: DeploymentStartTime, props: { direction: 'up' } },
  { title: `End time`, renderer: DeploymentEndTime },
  { title: '# devices', class: 'align-right column-defined', renderer: DeploymentDeviceCount },
  { title: 'Phases', renderer: DeploymentPhases }
];

const tabs = {
  list: {
    icon: <ListIcon />,
    index: 'list',
    title: 'List'
  },
  calendar: {
    icon: <CalendarTodayIcon />,
    index: 'calendar',
    title: 'Calendar'
  }
};

const type = DEPLOYMENT_STATES.scheduled;

export const Scheduled = ({ abort, createClick, openReport, ...remainder }) => {
  const [calendarEvents, setCalendarEvents] = useState([]);
  const [tabIndex, setTabIndex] = useState(tabs.list.index);
  const timer = useRef();
  const { canConfigure, canDeploy } = useSelector(getUserCapabilities);
  const {
    scheduled: { total: count }
  } = useSelector(getDeploymentsByStatusSelector);
  const idAttribute = useSelector(getIdAttribute);
  const devices = useSelector(getDevicesById);
  // TODO: isEnterprise is misleading here, but is passed down to the DeploymentListItem, this should be renamed
  const { canDelta: isEnterprise } = useSelector(getTenantCapabilities);
  const { scheduled: scheduledState } = useSelector(getDeploymentsSelectionState);
  const items = useSelector(state => getMappedDeploymentSelection(state, type));
  const dispatch = useAppDispatch();
  const dispatchedSetSnackbar = useCallback((...args) => dispatch(setSnackbar(...args)), [dispatch]);
  const { classes } = useStyles();

  const { page, perPage } = scheduledState;

  const refreshDeployments = useCallback(
    () =>
      dispatch(getDeploymentsByStatus({ status: DEPLOYMENT_STATES.scheduled, page, perPage }))
        .then(({ payload }) => {
          clearRetryTimer(type, dispatchedSetSnackbar);
          const { total, deploymentIds } = payload[payload.length - 1];
          if (total && !deploymentIds.length) {
            return refreshDeployments();
          }
        })
        .catch(err => setRetryTimer(err, 'deployments', `Couldn't load deployments.`, refreshDeploymentsLength, dispatchedSetSnackbar)),
    [dispatch, dispatchedSetSnackbar, page, perPage]
  );

  useEffect(() => {
    if (!isEnterprise) {
      return;
    }
    refreshDeployments();
    return () => {
      clearAllRetryTimers(dispatchedSetSnackbar);
    };
  }, [dispatchedSetSnackbar, isEnterprise, refreshDeployments]);

  useEffect(() => {
    if (!isEnterprise) {
      return;
    }
    clearInterval(timer.current);
    timer.current = setInterval(refreshDeployments, refreshDeploymentsLength);
    return () => {
      clearInterval(timer.current);
    };
  }, [isEnterprise, page, perPage, refreshDeployments]);

  useEffect(() => {
    if (tabIndex !== tabs.calendar.index) {
      return;
    }
    const calendarEvents = items.map(deployment => {
      const start = new Date(deployment.start_ts || deployment.phases ? deployment.phases[0].start_ts : deployment.created);
      let endDate = start;
      if (deployment.phases && deployment.phases.length && deployment.phases[deployment.phases.length - 1].end_ts) {
        endDate = new Date(deployment.phases[deployment.phases.length - 1].end_ts);
      } else if (deployment.filter_id || deployment.filter) {
        // calendar doesn't support never ending events so we arbitrarly set one year
        endDate = dayjs(start).add(1, 'year').toDate();
      }
      return {
        allDay: !(deployment.filter_id || deployment.filter),
        id: deployment.id,
        title: `${deployment.name} ${deployment.artifact_name}`,
        start,
        end: endDate
      };
    });
    setCalendarEvents(calendarEvents);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(items), tabIndex]);

  const abortDeployment = id => abort(id).then(refreshDeployments);
  const handleToggleChange = (_, newMode: string) => {
    setTabIndex(newMode);
  };
  const props = {
    ...remainder,
    canDeploy,
    canConfigure,
    count,
    devices,
    idAttribute,
    isEnterprise,
    items,
    openReport,
    page
  };
  return (
    <div className={`fadeIn ${classes.calendarContainer}`}>
      {items.length ? (
        <>
          <div className="margin-top margin-bottom-medium margin-left-small">
            <ToggleButtonGroup size="small" exclusive value={tabIndex} onChange={handleToggleChange} className={classes.toggleButtonGroup}>
              {Object.entries(tabs).map(([currentIndex, tab]) => (
                <ToggleButton size="small" className={classes.tabSelect} key={currentIndex} value={tab.index}>
                  {tab.icon} <span className="margin-left-x-small">{tab.title}</span>
                </ToggleButton>
              ))}
            </ToggleButtonGroup>
          </div>
          {tabIndex === tabs.list.index && (
            <DeploymentsList
              {...props}
              abort={abortDeployment}
              headers={headers}
              type={type}
              onChangeRowsPerPage={perPage => dispatch(setDeploymentsState({ [DEPLOYMENT_STATES.scheduled]: { page: 1, perPage } }))}
              onChangePage={page => dispatch(setDeploymentsState({ [DEPLOYMENT_STATES.scheduled]: { page } }))}
            />
          )}
          {tabIndex === tabs.calendar.index && (
            <Calendar
              localizer={localizer}
              className="margin-left-small margin-bottom"
              events={calendarEvents}
              startAccessor="start"
              endAccessor="end"
              style={{ height: 700 }}
              onSelectEvent={calendarEvent => openReport(type, calendarEvent.id)}
            />
          )}
        </>
      ) : (
        <div className="dashboard-placeholder margin-top">
          {isEnterprise ? (
            <>
              <p>Scheduled deployments will appear here. </p>
              {canDeploy && (
                <p>
                  <a onClick={createClick}>Create a deployment</a> to get started
                </p>
              )}
            </>
          ) : (
            <div className="flexbox centered">
              <DefaultUpgradeNotification />
            </div>
          )}
          <RefreshIcon className={`flip-horizontal ${classes.refreshIcon}`} />
        </div>
      )}
    </div>
  );
};

export default Scheduled;
