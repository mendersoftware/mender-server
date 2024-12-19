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
import React, { useCallback, useEffect, useRef, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { makeStyles } from 'tss-react/mui';

import EnterpriseNotification, { DefaultUpgradeNotification } from '@northern.tech/common-ui/enterpriseNotification';
import { HELPTOOLTIPS, MenderHelpTooltip } from '@northern.tech/helptips/helptooltips';
import { AUDIT_LOGS_TYPES, BEGINNING_OF_TIME, BENEFITS, SORTING_OPTIONS, SP_AUDIT_LOGS_TYPES, TIMEOUTS } from '@northern.tech/store/constants';
import {
  getAuditLog,
  getAuditLogEntry,
  getAuditLogSelectionState,
  getCurrentSession,
  getGroupNames,
  getIsServiceProvider,
  getTenantCapabilities,
  getUserCapabilities
} from '@northern.tech/store/selectors';
import { getAuditLogs, getAuditLogsCsvLink, getUserList, setAuditlogsState } from '@northern.tech/store/thunks';
import { createDownload, getISOStringBoundaries } from '@northern.tech/utils/helpers';
import { useLocationParams } from '@northern.tech/utils/liststatehook';
import dayjs from 'dayjs';

import historyImage from '../../../assets/img/history.png';
import AuditLogsFilter from './AuditLogsFilter';
import AuditlogsView from './AuditlogsView';
import { ActionDescriptor, ChangeDescriptor, ChangeDetailsDescriptor, TimeWrapper, TypeDescriptor, UserDescriptor } from './ColumnComponents';
import EventDetailsDrawerContentMap from './EventDetailsDrawerContentMap';
import AuditLogsList from './auditlogslist';
import EventDetailsFallbackComponent from './eventdetails/FallbackComponent';
import EventDetailsDrawer from './eventdetailsdrawer';

const useStyles = makeStyles()(theme => ({
  filters: {
    backgroundColor: theme.palette.background.lightgrey,
    padding: '0px 25px 5px',
    display: 'grid',
    gridTemplateColumns: '400px 250px 250px 1fr',
    gridColumnGap: theme.spacing(2),
    gridRowGap: theme.spacing(2)
  },
  filterReset: { alignSelf: 'flex-end', marginBottom: 5 },
  timeframe: { gridColumnStart: 2, gridColumnEnd: 4, marginLeft: 7.5 },
  typeDetails: { marginRight: 15, marginTop: theme.spacing(2) },
  upgradeNote: { marginTop: '5vh', placeSelf: 'center' }
}));

const isUserOptionEqualToValue = ({ email, id }, value) => id === value || email === value || email === value?.email;

const locationDefaults = { sort: { direction: SORTING_OPTIONS.desc } };

export const AuditLogs = () => {
  const [csvLoading, setCsvLoading] = useState(false);

  const [date] = useState(getISOStringBoundaries(new Date()));
  const { start: today, end: tonight } = date;

  const isInitialized = useRef();
  const [locationParams, setLocationParams] = useLocationParams('auditlogs', { today, tonight, defaults: locationDefaults });
  const { classes } = useStyles();
  const dispatch = useDispatch();
  const events = useSelector(getAuditLog);
  const eventItem = useSelector(getAuditLogEntry);
  const groups = useSelector(getGroupNames);
  const selectionState = useSelector(getAuditLogSelectionState);
  const userCapabilities = useSelector(getUserCapabilities);
  const tenantCapabilities = useSelector(getTenantCapabilities);
  const users = useSelector(state => state.users.byId);
  const { canReadUsers } = userCapabilities;
  const { hasAuditlogs } = tenantCapabilities;
  const [detailsReset, setDetailsReset] = useState('');
  const [dirtyField, setDirtyField] = useState('');
  const { token } = useSelector(getCurrentSession);
  const isSP = useSelector(getIsServiceProvider);
  const { detail, perPage, endDate, user, sort, startDate, type, total, isLoading } = selectionState;
  const [auditLogsTypes, setAuditLogsTypes] = useState(AUDIT_LOGS_TYPES);
  const timers = useRef({ init: null, detailsReset: null, dirtyField: null });

  useEffect(() => {
    if (isSP) {
      setAuditLogsTypes(SP_AUDIT_LOGS_TYPES);
    }
  }, [isSP]);

  useEffect(() => {
    if (!hasAuditlogs || !isInitialized.current) {
      return;
    }
    setLocationParams({ pageState: selectionState });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [detail, endDate, hasAuditlogs, perPage, selectionState.page, selectionState.selectedId, setLocationParams, startDate, type, user]);

  useEffect(() => {
    if (!isInitialized.current) {
      return;
    }
    setDetailsReset('detail');
    clearTimeout(timers.current.detailsReset);
    timers.current.detailsReset = setTimeout(() => setDetailsReset(''), TIMEOUTS.debounceShort);
  }, [type?.value]);

  useEffect(() => {
    if (canReadUsers) {
      dispatch(getUserList());
    }
  }, [canReadUsers, dispatch]);

  const initAuditlogState = useCallback(
    (result, state) => {
      const { detail, endDate, startDate, type, user } = state;
      const resultList = result ? Object.values(result.events) : [];
      if (resultList.length && startDate === today) {
        let newStartDate = new Date(resultList[resultList.length - 1].time);
        const { start } = getISOStringBoundaries(newStartDate);
        state.startDate = start;
      }
      dispatch(setAuditlogsState(state));
      clearTimeout(timers.current.dirtyField);
      timers.current.dirtyField = setTimeout(() => {
        let field = Object.entries({ detail, type, user }).reduce((accu, [key, value]) => (accu || value ? key : accu), '');
        field = field || (endDate !== tonight ? 'endDate' : field);
        field = field || (state.startDate !== today ? 'startDate' : field);
        setDirtyField(field);
      }, TIMEOUTS.debounceDefault);
      // the timeout here is slightly longer than the debounce in the filter component, otherwise the population of the filters with the url state would trigger a reset to page 1
      clearTimeout(timers.current.init);
      timers.current.init = setTimeout(() => (isInitialized.current = true), TIMEOUTS.oneSecond + TIMEOUTS.debounceDefault);
    },
    [dispatch, today, tonight]
  );

  const updateState = useCallback(
    nextState => {
      let state = { ...nextState };
      if (state.id && Boolean(state.open)) {
        state.selectedId = state.id[0];
        const [eventAction, eventTime] = atob(state.selectedId).split('|');
        if (eventTime && !events.some(item => item.time === eventTime && item.action === eventAction)) {
          const { start, end } = getISOStringBoundaries(new Date(eventTime));
          state.endDate = end;
          state.startDate = start;
        }
        let field = endDate !== tonight ? 'endDate' : '';
        field = field || (startDate !== today ? 'startDate' : field);
        setDirtyField(field);
      }
      // the timeout here is slightly longer than the debounce in the filter component, otherwise the population of the filters with the url state would trigger a reset to page 1
      dispatch(setAuditlogsState(state)).then(() => {
        clearTimeout(timers.current.init);
        timers.current.init = setTimeout(() => (isInitialized.current = true), TIMEOUTS.oneSecond + TIMEOUTS.debounceDefault);
      });
      return;
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [dispatch, endDate, JSON.stringify(events), startDate, today, tonight]
  );

  useEffect(() => {
    if (!hasAuditlogs || isInitialized.current !== undefined) {
      return;
    }
    isInitialized.current = false;
    const { id, open, detail, endDate, startDate, type, user } = locationParams;
    let state = { ...locationParams };
    if (id && Boolean(open)) {
      updateState(state);
      return;
    }
    dispatch(getAuditLogs({ page: state.page ?? 1, perPage: 50, startDate: startDate !== today ? startDate : BEGINNING_OF_TIME, endDate, user, type, detail }))
      .unwrap()
      .then(({ payload: result }) => initAuditlogState(result, state));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dispatch, hasAuditlogs, JSON.stringify(events), JSON.stringify(locationParams), initAuditlogState, updateState, today, tonight]);

  useEffect(() => {
    const currentTimers = timers.current;
    return () => {
      Object.values(currentTimers).forEach(clearTimeout);
    };
  }, []);

  const createCsvDownload = () => {
    setCsvLoading(true);
    dispatch(getAuditLogsCsvLink())
      .unwrap()
      .then(address => {
        createDownload(encodeURI(address), `Mender-AuditLog-${dayjs(startDate).format('YYYY-MM-DD')}-${dayjs(endDate).format('YYYY-MM-DD')}.csv`, token);
        setCsvLoading(false);
      });
  };

  const onChangeSorting = () => {
    const currentSorting = sort.direction === SORTING_OPTIONS.desc ? SORTING_OPTIONS.asc : SORTING_OPTIONS.desc;
    dispatch(setAuditlogsState({ page: 1, sort: { direction: currentSorting } }));
  };

  const onChangePagination = (page, currentPerPage = perPage) => dispatch(setAuditlogsState({ page, perPage: currentPerPage }));

  const onIssueSelection = selectedIssue =>
    dispatch(setAuditlogsState({ selectedId: selectedIssue ? btoa(`${selectedIssue.action}|${selectedIssue.time}`) : undefined }));

  const onFiltersChange = useCallback(
    ({ endDate, detail, startDate, user, type }) => {
      if (!isInitialized.current) {
        return;
      }
      const selectedUser = Object.values(users).find(item => isUserOptionEqualToValue(item, user));
      dispatch(setAuditlogsState({ page: 1, detail, startDate, endDate, user: selectedUser, type }));
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [dispatch, JSON.stringify(users)]
  );

  return (
    <AuditlogsView
      createCsvDownload={createCsvDownload}
      hasAuditlogs={hasAuditlogs}
      total={total}
      csvLoading={csvLoading}
      infoHintComponent={<EnterpriseNotification id={BENEFITS.auditlog.id} />}
      auditLogsFilter={
        <AuditLogsFilter
          groups={groups}
          users={users}
          disabled={!hasAuditlogs}
          onFiltersChange={onFiltersChange}
          detailsReset={detailsReset}
          selectionState={selectionState}
          auditLogsTypes={auditLogsTypes}
          dirtyField={dirtyField}
          setDirtyField={setDirtyField}
        />
      }
    >
      {!!total && (
        <AuditLogsList
          items={events}
          onChangePage={onChangePagination}
          onChangeRowsPerPage={newPerPage => onChangePagination(1, newPerPage)}
          onChangeSorting={onChangeSorting}
          selectionState={selectionState}
          onIssueSelection={onIssueSelection}
          userCapabilities={userCapabilities}
          auditLogColumns={[
            { title: 'Performed by', sortable: false, render: UserDescriptor },
            { title: 'Action', sortable: false, render: ActionDescriptor },
            { title: 'Type', sortable: false, render: TypeDescriptor },
            { title: 'Changed', sortable: false, render: ChangeDescriptor },
            { title: 'More details', sortable: false, render: ChangeDetailsDescriptor },
            { title: 'Time', sortable: true, render: TimeWrapper }
          ]}
        />
      )}
      {!(isLoading || total) && hasAuditlogs && (
        <div className="dashboard-placeholder">
          <p>No log entries were found.</p>
          <p>Try adjusting the filters.</p>
          <img src={historyImage} alt="Past" />
        </div>
      )}
      {!hasAuditlogs && (
        <div className={`dashboard-placeholder flexbox ${classes.upgradeNote}`}>
          <DefaultUpgradeNotification className="margin-right-small" />
          <MenderHelpTooltip id={HELPTOOLTIPS.auditlogExplanation.id} />
        </div>
      )}
      <EventDetailsDrawer
        mapChangeToContent={EventDetailsDrawerContentMap}
        fallbackComponent={EventDetailsFallbackComponent}
        eventItem={eventItem}
        open={Boolean(eventItem)}
        onClose={() => onIssueSelection()}
      />
    </AuditlogsView>
  );
};

export default AuditLogs;
