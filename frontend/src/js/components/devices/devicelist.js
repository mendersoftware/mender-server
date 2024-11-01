// Copyright 2015 Northern.tech AS
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
import React, { memo } from 'react';

import { deepCompare } from '@northern.tech/utils/helpers';

import { CommonList } from '../common/list';
import DeviceListItem from './devicelistitem';

const sortingNotes = {
  name: 'Sorting by Name will only work properly with devices that already have a device name defined'
};

const getRelevantColumns = (columnElements, selectable) => [...columnElements].slice(selectable ? 1 : 0, columnElements.length - 1);

export const calculateResizeChange = ({ columnElements, columnHeaders, e, index, prev, selectable }) => {
  const isShrinkage = prev > e.clientX ? -1 : 1;
  const columnDelta = Math.abs(e.clientX - prev) * isShrinkage;
  const relevantColumns = getRelevantColumns(columnElements, selectable);
  const canModifyNextColumn = index + 1 < columnHeaders.length - 1;

  return relevantColumns.reduce((accu, element, columnIndex) => {
    const currentWidth = element.offsetWidth;
    let column = { attribute: columnHeaders[columnIndex + 1].attribute, size: currentWidth };
    if (canModifyNextColumn && index === columnIndex) {
      column.size = currentWidth + columnDelta;
    } else if (canModifyNextColumn && index + 1 === columnIndex) {
      column.size = currentWidth - columnDelta;
    }
    accu.push(column);
    return accu;
  }, []);
};

export const minCellWidth = 150;

const areEqual = (prevProps, nextProps) => {
  if (
    prevProps.pageTotal != nextProps.pageTotal ||
    prevProps.pageLoading != nextProps.pageLoading ||
    prevProps.idAttribute != nextProps.idAttribute ||
    !deepCompare(prevProps.columnHeaders, nextProps.columnHeaders) ||
    !deepCompare(prevProps.customColumnSizes, nextProps.customColumnSizes) ||
    !deepCompare(prevProps.devices, nextProps.devices)
  ) {
    return false;
  }
  return deepCompare(prevProps.deviceListState, nextProps.deviceListState);
};
export const DeviceList = ({
  columnHeaders,
  customColumnSizes,
  devices,
  deviceListState,
  idAttribute,
  onChangeRowsPerPage,
  PaginationProps = {},
  onExpandClick,
  onResizeColumns,
  onPageChange,
  onSelect,
  onSort,
  pageLoading,
  pageTotal
}) => {
  return (
    <CommonList
      columnHeaders={columnHeaders}
      customColumnSizes={customColumnSizes}
      listItems={devices}
      listState={deviceListState}
      idAttribute={idAttribute}
      onChangeRowsPerPage={onChangeRowsPerPage}
      onExpandClick={onExpandClick}
      onPageChange={onPageChange}
      onResizeColumns={onResizeColumns}
      onSelect={onSelect}
      onSort={onSort}
      pageLoading={pageLoading}
      pageTotal={pageTotal}
      PaginationProps={PaginationProps}
      sortingNotes={sortingNotes}
      ListItemComponent={DeviceListItem}
    ></CommonList>
  );
};

export default memo(DeviceList, areEqual);
