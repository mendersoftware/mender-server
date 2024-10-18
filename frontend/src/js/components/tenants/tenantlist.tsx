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
import React, { useCallback, useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { ArrowForward as ArrowForwardIcon, Check as CheckIcon } from '@mui/icons-material';

import { getDeviceLimit, getTenantsList } from '@northern.tech/store/selectors';
import { AppDispatch } from '@northern.tech/store/store';
import { getTenants, setTenantsListState } from '@northern.tech/store/thunks';
import dayjs from 'dayjs';

import { ColumnHeader, CommonList, ListItemComponentProps, RendererProp } from '../common/list';
import { ExpandedTenant } from './expanded-tenant';
import { Tenant } from './types';

export const defaultTextRender = (props: RendererProp<Tenant>) => {
  const { column, item } = props;
  const attributeValue = item?.[column.attribute.name];
  return typeof attributeValue === 'object' ? JSON.stringify(attributeValue) : attributeValue;
};
export const DeviceLimitRender = (props: RendererProp<Tenant>) => {
  //TODO: use better alternative once backend is ready (MEN-7615)
  const deviceLimit = useSelector(getDeviceLimit);
  const { column, item } = props;
  const attributeValue = item?.[column.attribute.name];
  return `${attributeValue}/${deviceLimit}`;
};
export const BoolRender = (props: RendererProp<Tenant>) => {
  const { column, item } = props;
  return <div>{item?.[column.attribute.name] ? <CheckIcon /> : <div>-</div>}</div>;
};
const AttributeRenderer = ({ content, textContent }) => (
  <div title={typeof textContent === 'string' ? textContent : ''}>
    <div className="text-overflow">{content}</div>
  </div>
);
const DateRender = (props: RendererProp<Tenant>) => {
  const { column, item } = props;
  const attributeValue = dayjs(item?.[column.attribute.name]).format('YYYY-MM-DD HH:mm');
  return <AttributeRenderer content={attributeValue} textContent={item?.[column.attribute.name]}></AttributeRenderer>;
};
const moreDetailsRender = () => {
  return (
    <div className="link-color">
      View details <ArrowForwardIcon sx={{ fontSize: 12 }} />
    </div>
  );
};
const columnHeaders: ColumnHeader<Tenant>[] = [
  {
    title: 'Name',
    attribute: {
      name: 'name',
      scope: ''
    },
    sortable: false,
    textRender: defaultTextRender
  },
  {
    title: 'Devices',
    attribute: {
      name: 'device_limit',
      scope: ''
    },
    sortable: false,
    textRender: DeviceLimitRender
  },
  {
    title: 'Delta updates enabled ',
    attribute: {
      name: 'binary_delta',
      scope: ''
    },
    sortable: false,
    textRender: BoolRender
  },
  {
    title: 'Created',
    attribute: {
      name: 'created_at',
      scope: ''
    },
    sortable: false,
    textRender: DateRender
  },
  {
    title: 'More details',
    attribute: {
      name: '',
      scope: ''
    },
    sortable: false,
    textRender: moreDetailsRender
  }
];

const DefaultAttributeRenderer = (props: RendererProp<Tenant>) => {
  const { column, item } = props;
  const text = column.textRender({ item, column });
  return <AttributeRenderer content={text} textContent={text} />;
};

const TenantListItem = (props: ListItemComponentProps<Tenant>) => {
  const { listItem, columnHeaders, onClick } = props;
  const handleOnClick = useCallback(() => {
    onClick(listItem);
  }, [listItem.id, onClick]);

  return (
    <div onClick={handleOnClick} className={`deviceListRow deviceListItem clickable`}>
      {columnHeaders.map((column: ColumnHeader<Tenant>) => {
        const Component = column.component ? column.component : DefaultAttributeRenderer;
        return <Component column={column} item={listItem} key={column.title} />;
      })}
    </div>
  );
};
export const TenantList = () => {
  const tenantListState = useSelector(getTenantsList);
  const { tenants, perPage, selectedTenant } = tenantListState;
  const dispatch: AppDispatch = useDispatch();
  useEffect(() => {
    dispatch(getTenants());
  }, [dispatch]);
  const onExpandClick = useCallback(
    (tenant: Tenant) => {
      return dispatch(setTenantsListState({ selectedTenant: tenant }));
    },
    [dispatch]
  );
  const onCloseClick = useCallback(() => {
    return dispatch(setTenantsListState({ selectedTenant: null }));
  }, [dispatch]);
  const onChangePagination = useCallback(
    (page, currentPerPage = perPage) => {
      dispatch(setTenantsListState({ page, perPage: currentPerPage }));
    },
    [dispatch, perPage]
  );
  return (
    <div>
      <CommonList
        columnHeaders={columnHeaders}
        listItems={tenants}
        listState={tenantListState}
        onChangeRowsPerPage={newPerPage => onChangePagination(1, newPerPage)}
        onExpandClick={onExpandClick}
        onPageChange={onChangePagination}
        onResizeColumns={false}
        onSelect={false}
        pageLoading={false}
        ListItemComponent={TenantListItem}
      ></CommonList>
      {selectedTenant && <ExpandedTenant onCloseClick={onCloseClick} tenantId={selectedTenant.id}></ExpandedTenant>}
    </div>
  );
};
