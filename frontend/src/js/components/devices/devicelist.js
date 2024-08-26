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
import React, { memo, useCallback, useEffect, useRef, useState } from 'react';

// material ui
import { Settings as SettingsIcon, Sort as SortIcon } from '@mui/icons-material';
import { Checkbox } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { SORTING_OPTIONS, TIMEOUTS } from '../../constants/appConstants';
import { DEVICE_LIST_DEFAULTS } from '../../constants/deviceConstants';
import { deepCompare, isDarkMode, toggle } from '../../helpers';
import useWindowSize from '../../utils/resizehook';
import Loader from '../common/loader';
import MenderTooltip from '../common/mendertooltip';
import Pagination from '../common/pagination';
import DeviceListItem from './devicelistitem';

const { page: defaultPage, perPage: defaultPerPage } = DEVICE_LIST_DEFAULTS;

const sortingNotes = {
  name: 'Sorting by Name will only work properly with devices that already have a device name defined'
};

const useStyles = makeStyles()(theme => ({
  header: {
    color: theme.palette.text.hint
  },
  resizer: {
    cursor: 'col-resize',
    paddingLeft: 5,
    paddingRight: 5
  },
  resizeHandle: {
    width: 4,
    background: 'initial',
    ['&.hovering']: {
      background: theme.palette.grey[600]
    },
    ['&.resizing']: {
      background: isDarkMode(theme.palette.mode) ? theme.palette.grey[200] : theme.palette.grey[900]
    }
  }
}));

const HeaderItem = ({ column, columnCount, index, sortCol, sortDown, onSort, onResizeChange, onResizeFinish, resizable }) => {
  const [isHovering, setIsHovering] = useState(false);
  const [shouldRemoveListeners, setShouldRemoveListeners] = useState(false);
  const resizeRef = useRef();
  const ref = useRef();
  const { classes } = useStyles();

  const onMouseOut = () => setIsHovering(false);

  const onMouseOver = () => setIsHovering(true);

  const mouseMove = useCallback(
    e => {
      if (resizable && resizeRef.current) {
        onResizeChange(e, { index, prev: resizeRef.current, ref });
        resizeRef.current = e.clientX;
      }
    },
    [index, onResizeChange, resizable]
  );

  const mouseUp = useCallback(
    e => {
      if (resizeRef.current) {
        onResizeFinish(e, { index, prev: resizeRef.current, ref });
        resizeRef.current = null;
        setShouldRemoveListeners(true);
      }
    },
    [index, onResizeFinish]
  );

  const mouseDown = e => (resizeRef.current = e.clientX);

  useEffect(() => {
    window.addEventListener('mousemove', mouseMove);
    window.addEventListener('mouseup', mouseUp);
    return () => {
      setShouldRemoveListeners(!!resizeRef.current);
    };
  }, [mouseMove, mouseUp]);

  useEffect(() => {
    if (shouldRemoveListeners) {
      window.removeEventListener('mousemove', mouseMove);
      window.removeEventListener('mouseup', mouseUp);
      setShouldRemoveListeners(false);
    }
  }, [shouldRemoveListeners, mouseMove, mouseUp]);

  let resizeHandleClassName = resizable && isHovering ? 'hovering' : '';
  resizeHandleClassName = resizeRef.current ? 'resizing' : resizeHandleClassName;

  const header = (
    <div className="columnHeader flexbox space-between relative" style={column.style} onMouseEnter={onMouseOver} onMouseLeave={onMouseOut} ref={ref}>
      <div className="flexbox center-aligned" onClick={() => onSort(column.attribute ? column.attribute : {})}>
        {column.title}
        {column.sortable && (
          <SortIcon
            className={`sortIcon ${sortCol === column.attribute.name ? 'selected' : ''} ${(sortDown === SORTING_OPTIONS.desc).toString()}`}
            style={{ fontSize: 16 }}
          />
        )}
      </div>
      <div className="flexbox center-aligned full-height">
        {column.customize && <SettingsIcon onClick={column.customize} style={{ fontSize: 16 }} />}
        {index < columnCount - 2 && (
          <div onMouseDown={mouseDown} className={`${classes.resizer} full-height`}>
            <div className={`full-height ${classes.resizeHandle} ${resizeHandleClassName}`} />
          </div>
        )}
      </div>
    </div>
  );
  return column.sortable && sortingNotes[column.attribute.name] ? (
    <MenderTooltip title={sortingNotes[column.attribute.name]} placement="top-start">
      {header}
    </MenderTooltip>
  ) : (
    header
  );
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
const getTemplateColumns = (columns, selectable) =>
  selectable ? `52px ${columns} minmax(${minCellWidth}px, 1fr)` : `${columns} minmax(${minCellWidth}px, 1fr)`;

const getColumnsStyle = (columns, defaultSize, selectable) => {
  const template = columns.map(({ size }) => `minmax(${minCellWidth}px, ${size ? `${size}px` : defaultSize})`);
  // applying styles via state changes would lead to less smooth changes, so we set the style directly on the components
  return getTemplateColumns(template.join(' '), selectable);
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
  const { page: pageNo = defaultPage, perPage: pageLength = defaultPerPage, selection: selectedRows = [], sort = {} } = deviceListState;
  const { direction: sortDown = SORTING_OPTIONS.desc, key: sortCol } = sort;
  const deviceListRef = useRef();
  const selectedRowsRef = useRef(selectedRows);
  const initRef = useRef();
  const [resizeTrigger, setResizeTrigger] = useState(false);

  const size = useWindowSize();
  const selectable = !!onSelect;
  const { classes } = useStyles();

  useEffect(() => {
    selectedRowsRef.current = selectedRows;
  }, [selectedRows]);

  useEffect(() => {
    if (!deviceListRef.current) {
      return;
    }
    const relevantColumns = getRelevantColumns(deviceListRef.current.querySelector('.deviceListRow').children, selectable);
    deviceListRef.current.style.gridTemplateColumns = getColumnsStyle(
      customColumnSizes.length && customColumnSizes.length === relevantColumns.length ? customColumnSizes : relevantColumns,
      '1.5fr',
      selectable
    );
  }, [customColumnSizes, columnHeaders, selectable, resizeTrigger, size.width]);

  useEffect(() => {
    clearTimeout(initRef.current);
    initRef.current = setTimeout(() => setResizeTrigger(toggle), TIMEOUTS.debounceDefault);
    return () => {
      clearTimeout(initRef.current);
    };
  }, [customColumnSizes.length]);

  const onRowSelection = selectedRow => {
    let updatedSelection = [...selectedRowsRef.current];
    const selectedIndex = updatedSelection.indexOf(selectedRow);
    if (selectedIndex === -1) {
      updatedSelection.push(selectedRow);
    } else {
      updatedSelection.splice(selectedIndex, 1);
    }
    onSelect(updatedSelection);
  };

  const onSelectAllClick = () => {
    let newSelectedRows = Array.apply(null, { length: devices.length }).map(Number.call, Number);
    if (selectedRows.length && selectedRows.length <= devices.length) {
      newSelectedRows = [];
    }
    onSelect(newSelectedRows);
  };

  const handleResizeChange = useCallback(
    (e, { index, prev, ref }) => {
      const changedColumns = calculateResizeChange({ columnElements: [...ref.current.parentElement.children], columnHeaders, e, index, prev, selectable });
      // applying styles via state changes would lead to less smooth changes, so we set the style directly on the components
      deviceListRef.current.style.gridTemplateColumns = getColumnsStyle(changedColumns, undefined, selectable);
    },
    [columnHeaders, selectable]
  );

  const handleResizeFinish = useCallback(
    (e, { index, prev, ref }) => {
      const changedColumns = calculateResizeChange({ columnElements: ref.current.parentElement.children, columnHeaders, e, index, prev, selectable });
      onResizeColumns(changedColumns);
    },
    [columnHeaders, onResizeColumns, selectable]
  );

  const numSelected = (selectedRows || []).length;
  return (
    <div className={`deviceList ${selectable ? 'selectable' : ''}`} ref={deviceListRef}>
      <div className={`header ${classes.header}`}>
        <div className="deviceListRow">
          {selectable && (
            <div>
              <Checkbox indeterminate={numSelected > 0 && numSelected < devices.length} checked={numSelected === devices.length} onChange={onSelectAllClick} />
            </div>
          )}
          {columnHeaders.map((item, index) => (
            <HeaderItem
              column={item}
              columnCount={columnHeaders.length}
              index={index}
              key={`columnHeader-${index}`}
              onSort={onSort}
              resizable={!!onResizeColumns}
              sortCol={sortCol}
              sortDown={sortDown}
              onResizeChange={handleResizeChange}
              onResizeFinish={handleResizeFinish}
            />
          ))}
        </div>
      </div>
      <div className="body">
        {devices.map((device, index) => (
          <DeviceListItem
            columnHeaders={columnHeaders}
            device={device}
            deviceListState={deviceListState}
            idAttribute={idAttribute}
            index={index}
            key={device.id}
            onClick={onExpandClick}
            onRowSelect={onRowSelection}
            selectable={selectable}
            selected={selectedRows.indexOf(index) !== -1}
          />
        ))}
      </div>
      <div className="footer flexbox margin-top">
        <Pagination
          className="margin-top-none"
          count={pageTotal}
          rowsPerPage={pageLength}
          onChangeRowsPerPage={onChangeRowsPerPage}
          page={pageNo}
          onChangePage={onPageChange}
          {...PaginationProps}
        />
        <Loader show={pageLoading} small />
      </div>
    </div>
  );
};

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

export default memo(DeviceList, areEqual);
