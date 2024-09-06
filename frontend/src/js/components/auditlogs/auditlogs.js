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

import { Button, TextField } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import dayjs from 'dayjs';

import historyImage from '../../../assets/img/history.png';
import { getAuditLogs, getAuditLogsCsvLink, setAuditlogsState } from '../../actions/organizationActions';
import { getUserList } from '../../actions/userActions';
import { BEGINNING_OF_TIME, BENEFITS, SORTING_OPTIONS, TIMEOUTS } from '../../constants/appConstants';
import { AUDIT_LOGS_TYPES } from '../../constants/organizationConstants';
import { createDownload, getISOStringBoundaries } from '../../helpers';
import {
  getAuditLog,
  getAuditLogEntry,
  getAuditLogSelectionState,
  getCurrentSession,
  getGroupNames,
  getTenantCapabilities,
  getUserCapabilities
} from '../../selectors';
import { useLocationParams } from '../../utils/liststatehook';
import EnterpriseNotification, { DefaultUpgradeNotification } from '../common/enterpriseNotification';
import { ControlledAutoComplete } from '../common/forms/autocomplete';
import ClickFilter from '../common/forms/clickfilter';
import Filters from '../common/forms/filters';
import TimeframePicker from '../common/forms/timeframe-picker';
import { InfoHintContainer } from '../common/info-hint';
import Loader from '../common/loader';
import { HELPTOOLTIPS, MenderHelpTooltip } from '../helptips/helptooltips';
import AuditLogsList from './auditlogslist';

const detailsMap = {
  Deployment: 'to device group',
  User: 'email'
};

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

const getOptionLabel = option => option.title ?? option.email ?? option;

const renderOption = (props, option) => <li {...props}>{getOptionLabel(option)}</li>;

const isUserOptionEqualToValue = ({ email, id }, value) => id === value || email === value || email === value?.email;

const autoSelectProps = {
  autoSelect: true,
  filterSelectedOptions: true,
  getOptionLabel,
  handleHomeEndKeys: true,
  renderOption
};

const locationDefaults = { sort: { direction: SORTING_OPTIONS.desc } };

export const AuditLogs = props => {
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

  const { detail, isLoading, perPage, endDate, user, sort, startDate, total, type } = selectionState;

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
    setTimeout(() => setDetailsReset(''), TIMEOUTS.debounceShort);
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
      setTimeout(() => {
        let field = Object.entries({ detail, type, user }).reduce((accu, [key, value]) => (accu || value ? key : accu), '');
        field = field || (endDate !== tonight ? 'endDate' : field);
        field = field || (state.startDate !== today ? 'startDate' : field);
        setDirtyField(field);
      }, TIMEOUTS.debounceDefault);
      // the timeout here is slightly longer than the debounce in the filter component, otherwise the population of the filters with the url state would trigger a reset to page 1
      setTimeout(() => (isInitialized.current = true), TIMEOUTS.oneSecond + TIMEOUTS.debounceDefault);
    },
    [dispatch, today, tonight]
  );

  useEffect(() => {
    if (!hasAuditlogs || isInitialized.current !== undefined) {
      return;
    }
    isInitialized.current = false;
    const { id, open, detail, endDate, startDate, type, user } = locationParams;
    let state = { ...locationParams };
    if (id && Boolean(open)) {
      state.selectedId = id[0];
      const [eventAction, eventTime] = atob(state.selectedId).split('|');
      if (eventTime && !events.some(item => item.time === eventTime && item.action === eventAction)) {
        const { start, end } = getISOStringBoundaries(new Date(eventTime));
        state.endDate = end;
        state.startDate = start;
      }
      let field = endDate !== tonight ? 'endDate' : '';
      field = field || (startDate !== today ? 'startDate' : field);
      setDirtyField(field);
      // the timeout here is slightly longer than the debounce in the filter component, otherwise the population of the filters with the url state would trigger a reset to page 1
      dispatch(setAuditlogsState(state)).then(() => setTimeout(() => (isInitialized.current = true), TIMEOUTS.oneSecond + TIMEOUTS.debounceDefault));
      return;
    }
    dispatch(
      getAuditLogs({ page: state.page ?? 1, perPage: 50, startDate: startDate !== today ? startDate : BEGINNING_OF_TIME, endDate, user, type, detail })
    ).then(result => initAuditlogState(result, state));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dispatch, hasAuditlogs, JSON.stringify(events), JSON.stringify(locationParams), initAuditlogState, today, tonight]);

  const createCsvDownload = () => {
    setCsvLoading(true);
    dispatch(getAuditLogsCsvLink()).then(address => {
      createDownload(encodeURI(address), `Mender-AuditLog-${dayjs(startDate).format('YYYY-MM-DD')}-${dayjs(endDate).format('YYYY-MM-DD')}.csv`, token);
      setCsvLoading(false);
    });
  };

  const onChangeSorting = () => {
    const currentSorting = sort.direction === SORTING_OPTIONS.desc ? SORTING_OPTIONS.asc : SORTING_OPTIONS.desc;
    dispatch(setAuditlogsState({ page: 1, sort: { direction: currentSorting } }));
  };

  const onChangePagination = (page, currentPerPage = perPage) => dispatch(setAuditlogsState({ page, perPage: currentPerPage }));

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

  const typeOptionsMap = {
    Deployment: groups,
    User: Object.values(users)
  };
  const detailOptions = typeOptionsMap[type?.title] ?? [];

  return (
    <div className="fadeIn margin-left flexbox column" style={{ marginRight: '5%' }}>
      <div className="flexbox center-aligned">
        <h3 className="margin-right-small">Audit log</h3>
        <InfoHintContainer>
          <EnterpriseNotification id={BENEFITS.auditlog.id} />
        </InfoHintContainer>
      </div>
      <ClickFilter disabled={!hasAuditlogs}>
        <Filters
          initialValues={{ startDate, endDate, user, type, detail }}
          defaultValues={{ startDate: today, endDate: tonight, user: '', type: null, detail: '' }}
          fieldResetTrigger={detailsReset}
          dirtyField={dirtyField}
          clearDirty={setDirtyField}
          filters={[
            {
              key: 'user',
              title: 'Performed by',
              Component: ControlledAutoComplete,
              componentProps: {
                ...autoSelectProps,
                freeSolo: true,
                isOptionEqualToValue: isUserOptionEqualToValue,
                options: Object.values(users),
                renderInput: params => <TextField {...params} placeholder="Select a user" InputProps={{ ...params.InputProps }} />
              }
            },
            {
              key: 'type',
              title: 'Filter by changes',
              Component: ControlledAutoComplete,
              componentProps: {
                ...autoSelectProps,
                options: AUDIT_LOGS_TYPES,
                isOptionEqualToValue: (option, value) => option.value === value.value && option.object_type === value.object_type,
                renderInput: params => <TextField {...params} placeholder="Type" InputProps={{ ...params.InputProps }} />
              }
            },
            {
              key: 'detail',
              title: '',
              Component: ControlledAutoComplete,
              componentProps: {
                ...autoSelectProps,
                freeSolo: true,
                options: detailOptions,
                disabled: !type,
                renderInput: params => <TextField {...params} placeholder={detailsMap[type] || '-'} InputProps={{ ...params.InputProps }} />
              }
            },
            {
              key: 'timeframe',
              title: 'Start time',
              Component: TimeframePicker,
              componentProps: {
                tonight
              }
            }
          ]}
          onChange={onFiltersChange}
        />
      </ClickFilter>
      <div className="flexbox center-aligned" style={{ justifyContent: 'flex-end' }}>
        <Loader show={csvLoading} />
        <Button variant="contained" color="secondary" disabled={csvLoading || !total} onClick={createCsvDownload} style={{ marginLeft: 15 }}>
          Download results as csv
        </Button>
      </div>
      {!!total && (
        <AuditLogsList
          {...props}
          items={events}
          eventItem={eventItem}
          loading={isLoading}
          onChangePage={onChangePagination}
          onChangeRowsPerPage={newPerPage => onChangePagination(1, newPerPage)}
          onChangeSorting={onChangeSorting}
          selectionState={selectionState}
          setAuditlogsState={state => dispatch(setAuditlogsState(state))}
          userCapabilities={userCapabilities}
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
    </div>
  );
};

export default AuditLogs;
