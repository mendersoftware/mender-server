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
import { useDispatch, useSelector } from 'react-redux';
import { useLocation } from 'react-router-dom';

import { Alert, Tooltip, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import DetailsIndicator from '@northern.tech/common-ui/DetailsIndicator';
import type { ColumnHeader, ListItemComponentProps, RendererProp } from '@northern.tech/common-ui/List';
import { CommonList } from '@northern.tech/common-ui/List';
import { SORTING_OPTIONS } from '@northern.tech/store/constants';
import { useLocationParams } from '@northern.tech/store/liststatehook';
import { getDisabledTiers, getTenantListWithLimits } from '@northern.tech/store/selectors';
import type { AppDispatch } from '@northern.tech/store/store';
import { setTenantsListState } from '@northern.tech/store/thunks';
import dayjs from 'dayjs';

import { getLimitStatus } from '../header/DeviceNotifications';
import { ExpandedTenant } from './ExpandedTenant';
import type { Tenant } from './types';

const useStyles = makeStyles()(theme => ({
  container: {
    borderRadius: theme.spacing(0.5),
    padding: theme.spacing(0.5)
  },
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
export const defaultTextRender = (props: RendererProp<Tenant>) => {
  const { column, item } = props;
  const attributeValue = item?.[column.attribute.name];
  return typeof attributeValue === 'object' ? JSON.stringify(attributeValue) : attributeValue;
};

const DeviceLimitNumbers = (props: { limit: number; total: number }) => {
  const { limit, total } = props;
  const { warning, error, percentageUsed, color } = getLimitStatus(total, limit);
  const { classes } = useStyles();
  if (limit === 0 && total === 0) {
    return (
      <Typography variant="body2" className="padding-left-small">
        -
      </Typography>
    );
  }
  return (
    <div className={`${classes.container} flexbox align-items-center`}>
      {warning || error ? (
        <Tooltip title={`${percentageUsed}% used${error ? ' - limit reached' : ''}`}>
          <Alert severity={color} classes={{ root: classes.alert, message: `${classes[color]} padding-none`, icon: classes.alertIcon }}>
            {total}/{limit}
          </Alert>
        </Tooltip>
      ) : (
        <Typography variant="body2" className="padding-left-x-small">
          {total}/{limit}
        </Typography>
      )}
    </div>
  );
};

export const DeviceLimitRender = (props: RendererProp<Tenant>) => {
  const { column, item } = props;
  if (!item?.device_limits[column.attribute.name]) {
    return null;
  }
  const attributeValue = item?.device_limits[column.attribute.name].limit ?? 0;
  const deviceCount = item?.device_limits[column.attribute.name].current ?? 0;
  return <DeviceLimitNumbers limit={Number(attributeValue)} total={Number(deviceCount)} />;
};

const AttributeRenderer = ({ content, textContent }) => (
  <div title={typeof textContent === 'string' ? textContent : ''}>
    <div className="text-overflow">{content}</div>
  </div>
);
const DetailsButtonRenderer = props => (
  <div className="padding-bottom-small padding-top-small">
    <DetailsIndicator {...props} />
  </div>
);

const DateRender = (props: RendererProp<Tenant>) => {
  const { column, item } = props;
  const attributeValue = dayjs(item?.[column.attribute.name]).format('YYYY-MM-DD HH:mm');
  return <AttributeRenderer content={attributeValue} textContent={item?.[column.attribute.name]} />;
};
export const columnHeaders: ColumnHeader<Tenant>[] = [
  {
    component: () => <></>,
    title: 'Name',
    attribute: {
      name: 'name',
      scope: ''
    },
    sortable: false,
    textRender: defaultTextRender
  },
  {
    title: 'Micro',
    attribute: {
      name: 'micro',
      scope: ''
    },
    sortable: false,
    component: DeviceLimitRender
  },
  {
    title: 'Standard',
    attribute: {
      name: 'standard',
      scope: ''
    },
    sortable: false,
    component: DeviceLimitRender
  },
  {
    title: 'System',
    attribute: {
      name: 'system',
      scope: ''
    },
    sortable: false,
    component: DeviceLimitRender
  },
  {
    title: 'Created',
    attribute: {
      name: 'created_at',
      scope: ''
    },
    sortable: false,
    component: DateRender
  },
  {
    title: '',
    attribute: {
      name: '',
      scope: ''
    },
    sortable: false,
    component: DetailsButtonRenderer
  }
];

export const TenantListItem = (props: ListItemComponentProps<Tenant>) => {
  const { listItem, columnHeaders, onClick } = props;
  const handleOnClick = useCallback(() => {
    onClick(listItem);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [listItem.id, onClick]);

  return (
    <div onClick={handleOnClick} className="deviceListRow deviceListItem clickable">
      {columnHeaders.map((column: ColumnHeader<Tenant>) => {
        const { classes = {}, component: Component, textRender } = column;
        if (textRender) {
          return <AttributeRenderer content={textRender({ item: listItem, column })} key={column.title} textContent={textRender({ item: listItem, column })} />;
        }
        return <Component classes={classes} column={column} item={listItem} key={column.title} />;
      })}
    </div>
  );
};
export const TenantList = () => {
  const disabledTiers: string[] = useSelector(getDisabledTiers);
  const tenantListState = useSelector(getTenantListWithLimits);
  const { tenants, perPage, selectedTenant, sort = {} } = tenantListState;
  const dispatch: AppDispatch = useDispatch();
  const isInitialized = useRef(false);
  const location = useLocation();

  const [locationParams, setLocationParams, { shouldInitializeFromUrl }] = useLocationParams('tenants', {
    defaults: {
      direction: SORTING_OPTIONS.desc,
      key: 'name',
      sort: {}
    }
  });
  const enabledHeaders = columnHeaders.filter(column => !disabledTiers.includes(column.attribute.name));
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
      <CommonList
        columnHeaders={enabledHeaders}
        listItems={tenants}
        listState={tenantListState}
        onChangeRowsPerPage={newPerPage => onChangePagination(1, newPerPage)}
        onExpandClick={onExpandClick}
        onPageChange={onChangePagination}
        onResizeColumns={false}
        onSelect={false}
        pageLoading={false}
        ListItemComponent={TenantListItem}
      />
      {selectedTenant && tenant && <ExpandedTenant onCloseClick={onCloseClick} tenant={tenant} />}
    </div>
  );
};
