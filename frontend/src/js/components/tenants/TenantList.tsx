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
import { useCallback, useEffect, useRef } from 'react';
import { useSelector } from 'react-redux';
import { useLocation } from 'react-router';

import { Alert, Tooltip, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import DetailsTable from '@northern.tech/common-ui/DetailsTable';
import type { ColumnDefinition, ColumnRendererProps } from '@northern.tech/common-ui/DetailsTable';
import Pagination from '@northern.tech/common-ui/Pagination';
import Time from '@northern.tech/common-ui/Time';
import { SORTING_OPTIONS } from '@northern.tech/store/constants';
import { useLocationParams } from '@northern.tech/store/liststatehook';
import { getDisabledTiers, getTenantListWithLimits } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { setTenantsListState } from '@northern.tech/store/thunks';

import { getLimitStatus } from '../header/DeviceNotifications';
import { ExpandedTenant } from './ExpandedTenant';
import type { Tenant } from './types';

const useStyles = makeStyles()(theme => ({
  error: {
    color: theme.palette.error.light
  },
  warning: {
    color: theme.palette.warning.main
  },
  alertIcon: {
    padding: 0,
    fontSize: theme.spacing(2),
    marginRight: theme.spacing(0.5)
  },
  alert: {
    display: 'flex',
    alignItems: 'center',
    height: theme.spacing(4),
    padding: theme.spacing(0.75)
  },
  primary: {}
}));
const DeviceLimitNumbers = (props: { limit: number; total: number }) => {
  const { limit, total } = props;
  const { warning, error, percentageUsed, color } = getLimitStatus(total, limit);
  const { classes } = useStyles();
  if (limit === 0 && total === 0) {
    return <Typography variant="body2">-</Typography>;
  }
  return (
    <div className="flexbox align-items-center">
      {warning || error ? (
        <Tooltip title={`${percentageUsed}% used${error ? ' - limit reached' : ''}`}>
          <Alert severity={color} classes={{ root: classes.alert, message: `${classes[color]} padding-none`, icon: classes.alertIcon }}>
            {total}/{limit}
          </Alert>
        </Tooltip>
      ) : (
        <Typography variant="body2">
          {total}/{limit}
        </Typography>
      )}
    </div>
  );
};

const DeviceLimitRender = ({ column, item }: ColumnRendererProps<Tenant>) => {
  const deviceLimit = item.device_limits?.[column.key];
  if (!deviceLimit) {
    return null;
  }
  return <DeviceLimitNumbers limit={Number(deviceLimit.limit ?? 0)} total={Number(deviceLimit.current ?? 0)} />;
};

const columns: ColumnDefinition<Tenant>[] = [
  {
    key: 'name',
    title: 'Name',
    render: ({ name }) => (
      <div className="text-overflow" title={name}>
        {name}
      </div>
    )
  },
  { key: 'micro', title: 'Micro', component: DeviceLimitRender },
  { key: 'standard', title: 'Standard', component: DeviceLimitRender },
  { key: 'system', title: 'System', component: DeviceLimitRender },
  {
    key: 'created_at',
    title: 'Created',
    render: ({ created_at }) => <Time value={created_at} />
  }
];

export const TenantList = () => {
  const disabledTiers: string[] = useSelector(getDisabledTiers);
  const tenantListState = useSelector(getTenantListWithLimits);
  const { tenants, page = 1, perPage, selectedTenant, sort = {}, total } = tenantListState;
  const dispatch = useAppDispatch();
  const isInitialized = useRef(false);
  const location = useLocation();

  const [locationParams, setLocationParams, { shouldInitializeFromUrl }] = useLocationParams('tenants', {
    defaults: {
      direction: SORTING_OPTIONS.desc,
      key: 'name',
      sort: {}
    }
  });
  const enabledColumns = columns.filter(column => !disabledTiers.includes(column.key));
  useEffect(() => {
    if (shouldInitializeFromUrl) {
      isInitialized.current = false;
    }
  }, [shouldInitializeFromUrl, location.key]);

  useEffect(() => {
    if (isInitialized.current || !shouldInitializeFromUrl) {
      isInitialized.current = true;
      return;
    }
    const { selectedTenant: selectedTenantName } = locationParams;
    if (selectedTenantName) {
      dispatch(setTenantsListState({ selectedTenant: selectedTenantName }));
    }
    isInitialized.current = true;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dispatch, JSON.stringify(locationParams), shouldInitializeFromUrl]);

  useEffect(() => {
    if (!isInitialized.current || !selectedTenant) {
      return;
    }
    setLocationParams({ pageState: { ...tenantListState, selectedTenant } });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [setLocationParams, JSON.stringify(sort), selectedTenant]);

  const onExpandClick = useCallback((tenant: Tenant) => dispatch(setTenantsListState({ selectedTenant: tenant.id })), [dispatch]);

  const onCloseClick = useCallback(() => {
    setLocationParams({ pageState: { ...tenantListState, selectedTenant: '' } });
    return dispatch(setTenantsListState({ selectedTenant: null }));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dispatch, setLocationParams, JSON.stringify(tenantListState)]);

  const onChangePagination = useCallback(
    (page, currentPerPage = perPage) => {
      dispatch(setTenantsListState({ page, perPage: currentPerPage }));
    },
    [dispatch, perPage]
  );

  const tenant = selectedTenant && tenants.find((tenant: Tenant) => selectedTenant === tenant.id);
  return (
    <div className="margin-top-small">
      <DetailsTable columns={enabledColumns} items={tenants} onItemClick={onExpandClick} />
      <Pagination
        className="margin-top-none"
        count={total}
        rowsPerPage={perPage}
        onChangePage={onChangePagination}
        onChangeRowsPerPage={newPerPage => onChangePagination(1, newPerPage)}
        page={page}
      />
      {selectedTenant && tenant && <ExpandedTenant onCloseClick={onCloseClick} tenant={tenant} />}
    </div>
  );
};
