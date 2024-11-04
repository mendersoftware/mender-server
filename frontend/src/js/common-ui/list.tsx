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
import { CSSProperties, ComponentType, MutableRefObject, ReactElement, useCallback, useEffect, useRef, useState } from 'react';

import { Settings as SettingsIcon, Sort as SortIcon } from '@mui/icons-material';
import { Checkbox } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { DEVICE_LIST_DEFAULTS, SORTING_OPTIONS, TIMEOUTS } from '@northern.tech/store/commonConstants';
import { isDarkMode } from '@northern.tech/store/utils';
import { toggle } from '@northern.tech/utils/helpers';
import useWindowSize from '@northern.tech/utils/resizehook';

import Loader from './loader';
import MenderTooltip from './mendertooltip';
import Pagination from './pagination';

const { page: defaultPage, perPage: defaultPerPage } = DEVICE_LIST_DEFAULTS;

interface Attribute {
  name: string;
  scope: string;
}
export interface RendererProp<T> {
  column: ColumnHeader<T>;
  item?: T;
  [key: string]: any;
}

export interface ClassesOverrides {
  classes: Record<string, string>;
}
export interface ColumnHeader<T> {
  classes?: ClassesOverrides;
  component: ComponentType<RendererProp<T> & ClassesOverrides>;
  title: string;
  attribute: Attribute;
  sortable: boolean;
  customize?: () => void;
  style?: CSSProperties;
  textRender?: (props: RendererProp<T>) => string | ReactElement;
}

interface SortOptions {
  direction?: string;
  key?: string;
}

interface ListState {
  page?: number;
  perPage?: number;
  selection?: number[];
  sort?: SortOptions;
  // selectedAttributes: unknown[];
  // selectedIssues: unknown[];
  // state: string;
  total: number;
  // setOnly: boolean;
  // refreshTrigger: boolean;
  // detailsTab: string;
  // isLoading: boolean;
}

interface IdAttribute {
  attribute: string;
  scope: string;
}
type wID = { id: string };

interface CommonListProps<T extends wID> {
  listItems: T[];
  columnHeaders: ColumnHeader<T>[];
  customColumnSizes?: Attribute[];
  onExpandClick: (item: T) => void;
  onResizeColumns: ((columns: { size: number; attribute: Attribute }) => void) | false;
  onPageChange: (event: MouseEvent | null, page: number) => void;
  onSelect: ((rows: number[]) => void) | false;
  onSort?: (attr: Attribute | object) => void;
  onChangeRowsPerPage: (perPage: number) => void;
  pageLoading: boolean;
  PaginationProps?: object;
  idAttribute?: IdAttribute;
  listState: ListState;
  sortingNotes?: { [key: string]: string };
  ListItemComponent: ComponentType<ListItemComponentProps<T>>;
}
export interface ListItemComponentProps<T> {
  columnHeaders: ColumnHeader<T>[];
  listItem: T;
  listState: ListState;
  idAttribute?: IdAttribute;
  index: number;
  key: string;
  onClick: (item: T) => void;
  onRowSelect: (selectedRow: T) => void;
  selectable: boolean;
  selected: boolean;
}

const useStyles = makeStyles()(theme => ({
  header: {
    // @ts-ignore
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

export const minCellWidth = 150;

export const calculateResizeChange = ({ columnElements, columnHeaders, e, index, prev, selectable }) => {
  const isShrinkage = prev > e.clientX ? -1 : 1;
  const columnDelta = Math.abs(e.clientX - prev) * isShrinkage;
  const relevantColumns = getRelevantColumns(columnElements, selectable);
  const canModifyNextColumn = index + 1 < columnHeaders.length - 1;

  return relevantColumns.reduce((accu, element, columnIndex) => {
    const currentWidth = element.offsetWidth;
    const column = { attribute: columnHeaders[columnIndex + 1].attribute, size: currentWidth };
    if (canModifyNextColumn && index === columnIndex) {
      column.size = currentWidth + columnDelta;
    } else if (canModifyNextColumn && index + 1 === columnIndex) {
      column.size = currentWidth - columnDelta;
    }
    accu.push(column);
    return accu;
  }, []);
};
const getRelevantColumns = (columnElements, selectable) => [...columnElements].slice(selectable ? 1 : 0, columnElements.length - 1);
const getTemplateColumns = (columns, selectable) =>
  selectable ? `52px ${columns} minmax(${minCellWidth}px, 1fr)` : `${columns} minmax(${minCellWidth}px, 1fr)`;

const getColumnsStyle = (columns, defaultSize, selectable) => {
  const template = columns.map(({ size }) => `minmax(${minCellWidth}px, ${size ? `${size}px` : defaultSize})`);
  // applying styles via state changes would lead to less smooth changes, so we set the style directly on the components
  return getTemplateColumns(template.join(' '), selectable);
};
export const CommonList = <T extends wID>(props: CommonListProps<T>) => {
  const {
    columnHeaders,
    customColumnSizes = [],
    listItems,
    listState,
    idAttribute,
    onChangeRowsPerPage,
    PaginationProps = {},
    onExpandClick,
    onResizeColumns,
    onPageChange,
    onSelect,
    onSort = () => {},
    pageLoading,
    sortingNotes,
    ListItemComponent
  } = props;
  const { page: pageNo = defaultPage, perPage: pageLength = defaultPerPage, selection: selectedRows = [], sort = {}, total: pageTotal = 1 } = listState;
  const { direction: sortDown = SORTING_OPTIONS.desc, key: sortCol } = sort;
  const listRef = useRef<HTMLDivElement | null>(null);
  const selectedRowsRef = useRef(selectedRows);
  const initRef = useRef<number | null>(null);
  const [resizeTrigger, setResizeTrigger] = useState(false);

  const size = useWindowSize();
  const selectable = !!onSelect;
  const { classes } = useStyles();

  useEffect(() => {
    selectedRowsRef.current = selectedRows;
  }, [selectedRows]);

  useEffect(() => {
    if (!listRef.current) {
      return;
    }
    const relevantColumns = getRelevantColumns(listRef.current?.querySelector('.deviceListRow')?.children, selectable);
    listRef.current.style.gridTemplateColumns = getColumnsStyle(
      customColumnSizes.length && customColumnSizes.length === relevantColumns.length ? customColumnSizes : relevantColumns,
      '1.5fr',
      selectable
    );
  }, [customColumnSizes, columnHeaders, selectable, resizeTrigger, size.width]);

  useEffect(() => {
    clearTimeout(initRef.current || undefined);
    initRef.current = setTimeout(() => setResizeTrigger(toggle), TIMEOUTS.debounceDefault) as unknown as number;
    return () => {
      clearTimeout(initRef?.current || undefined);
    };
  }, [customColumnSizes.length]);

  const onRowSelection = selectedRow => {
    const updatedSelection = [...selectedRowsRef.current];
    const selectedIndex = updatedSelection.indexOf(selectedRow);
    if (selectedIndex === -1) {
      updatedSelection.push(selectedRow);
    } else {
      updatedSelection.splice(selectedIndex, 1);
    }
    if (onSelect) {
      onSelect(updatedSelection);
    }
  };

  const onSelectAllClick = () => {
    let newSelectedRows: number[] = Array.from({ length: listItems.length }, (_, i) => i);
    if (selectedRows.length && selectedRows.length <= listItems.length) {
      newSelectedRows = [];
    }
    if (onSelect) {
      onSelect(newSelectedRows);
    }
  };

  const handleResizeChange = useCallback(
    (e, { index, prev, ref }) => {
      const changedColumns = calculateResizeChange({
        columnElements: [...ref.current.parentElement.children],
        columnHeaders,
        e,
        index,
        prev,
        selectable
      });
      // applying styles via state changes would lead to less smooth changes, so we set the style directly on the components
      if (listRef.current) listRef.current.style.gridTemplateColumns = getColumnsStyle(changedColumns, undefined, selectable);
    },
    [columnHeaders, selectable]
  );

  const handleResizeFinish = useCallback(
    (e, { index, prev, ref }) => {
      const changedColumns = calculateResizeChange({
        columnElements: ref.current.parentElement.children,
        columnHeaders,
        e,
        index,
        prev,
        selectable
      });
      if (onResizeColumns) {
        onResizeColumns(changedColumns);
      }
    },
    [columnHeaders, onResizeColumns, selectable]
  );

  const numSelected = (selectedRows || []).length;
  return (
    <div className={`deviceList ${selectable ? 'selectable' : ''}`} ref={listRef}>
      <div className={`header ${classes.header}`}>
        <div className="deviceListRow">
          {selectable && (
            <div>
              <Checkbox
                indeterminate={numSelected > 0 && numSelected < listItems.length}
                checked={numSelected === listItems.length}
                onChange={onSelectAllClick}
              />
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
              sortingNotes={sortingNotes}
            />
          ))}
        </div>
      </div>
      <div className="body">
        {listItems.map((item, index) => (
          <ListItemComponent
            columnHeaders={columnHeaders}
            listItem={item}
            listState={listState}
            idAttribute={idAttribute}
            index={index}
            key={item.id}
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

interface HeaderItemProps<T> {
  sortingNotes?: { [key: string]: string };
  column: ColumnHeader<T>;
  columnCount: number;
  index: number;
  resizable: boolean;
  onSort: (attr: Attribute | object) => void;
  onResizeChange: (
    e: MouseEvent,
    eventData: {
      index: number;
      prev: number;
      ref: MutableRefObject<HTMLDivElement | null>;
    }
  ) => void;
  onResizeFinish: (
    e: MouseEvent,
    eventData: {
      index: number;
      prev: number;
      ref: MutableRefObject<HTMLDivElement | null>;
    }
  ) => void;
  sortCol?: string;
  sortDown?: string;
}

const HeaderItem = <T extends wID>(props: HeaderItemProps<T>) => {
  const { sortingNotes, column, columnCount, index, sortCol, sortDown = undefined, onSort, onResizeChange, onResizeFinish, resizable } = props;
  const [isHovering, setIsHovering] = useState(false);
  const [shouldRemoveListeners, setShouldRemoveListeners] = useState(false);
  const resizeRef = useRef<null | number>(null);
  const ref = useRef<HTMLDivElement | null>(null);
  const { classes } = useStyles();

  const onMouseOut = () => setIsHovering(false);

  const onMouseOver = () => setIsHovering(true);

  const mouseMove = useCallback(
    (e: MouseEvent) => {
      if (resizable && resizeRef.current) {
        onResizeChange(e, { index, prev: resizeRef.current, ref });
        resizeRef.current = e.clientX;
      }
    },
    [index, onResizeChange, resizable]
  );

  const mouseUp = useCallback(
    (e: MouseEvent) => {
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
        {index < columnCount - 2 && resizable && (
          <div onMouseDown={mouseDown} className={`${classes.resizer} full-height`}>
            <div className={`full-height ${classes.resizeHandle} ${resizeHandleClassName}`} />
          </div>
        )}
      </div>
    </div>
  );
  return column.sortable && sortingNotes && sortingNotes[column.attribute.name] ? (
    <MenderTooltip title={sortingNotes[column.attribute.name]} placement="top-start">
      {header}
    </MenderTooltip>
  ) : (
    header
  );
};
