// Copyright 2022 Northern.tech AS
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
import React, { useEffect, useState } from 'react';

// material ui
import { Checkbox, Table, TableBody, TableCell, TableHead, TableRow } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { SORTING_OPTIONS, SORT_DIRECTIONS, TIMEOUTS } from '@northern.tech/store/constants';

import { useDebounce } from '../../utils/debouncehook';
import SortIcon from './sorticon';

const useStyles = makeStyles()(() => ({
  header: {
    '.columnHeader': {
      display: 'table-cell'
    },
    '.columnHeader .sortIcon': {
      marginBottom: -6
    },
    '.nonSortable': { cursor: 'initial' }
  }
}));

const HeaderItem = ({ columnKey, hasMultiSort, extras, renderTitle, sortable, onSort, sortOptions, title }) => {
  const { direction, key: sortKey } = sortOptions.find(({ key: sortKey }) => columnKey === sortKey) ?? {};
  const [sortState, setSortState] = useState({ disabled: !sortKey, direction });
  const [resetDirection] = useState(hasMultiSort ? '' : SORT_DIRECTIONS[0]);

  const debouncedSortState = useDebounce(sortState, TIMEOUTS.debounceShort);

  useEffect(() => {
    if (!onSort) {
      return;
    }
    onSort({ key: columnKey, direction: debouncedSortState.direction, disabled: debouncedSortState.disabled });
  }, [columnKey, debouncedSortState.direction, debouncedSortState.disabled, onSort]);

  const onSortClick = () => {
    if (!sortable) {
      return;
    }
    const nextDirectionIndex = SORT_DIRECTIONS.indexOf(sortState.direction) + 1;
    const direction = SORT_DIRECTIONS[nextDirectionIndex] ?? resetDirection;
    setSortState({ direction, disabled: !direction });
  };

  const sortDown = sortKey && direction === SORTING_OPTIONS.desc;

  return (
    <TableCell className={`columnHeader ${sortable ? '' : 'nonSortable'}`} onClick={onSortClick}>
      {renderTitle ? renderTitle(extras) : title}
      {sortable && <SortIcon columnKey={sortKey} disabled={sortState.disabled} sortDown={sortDown} />}
    </TableCell>
  );
};

export const DetailsTable = ({
  className = '',
  columns,
  hasMultiSort = false,
  items,
  onChangeSorting,
  onItemClick,
  sort = [],
  style = {},
  tableRef,
  onRowSelected,
  selectedRows = []
}) => {
  const { classes } = useStyles();

  const onRowSelection = selectedRow => {
    let updatedSelection = [...selectedRows];
    const selectedIndex = updatedSelection.indexOf(selectedRow);
    if (selectedIndex === -1) {
      updatedSelection.push(selectedRow);
    } else {
      updatedSelection.splice(selectedIndex, 1);
    }
    onRowSelected(updatedSelection);
  };

  const onSelectAllClick = () => {
    let newSelectedRows = Array.from({ length: items.length }, (_, index) => index);
    if (selectedRows.length && selectedRows.length <= items.length) {
      newSelectedRows = [];
    }
    onRowSelected(newSelectedRows);
  };

  return (
    <Table className={`margin-bottom ${className}`} style={style} ref={tableRef}>
      <TableHead className={classes.header}>
        <TableRow>
          {!!onRowSelected && (
            <TableCell>
              <Checkbox indeterminate={false} checked={selectedRows.length === items.length} onChange={onSelectAllClick} />
            </TableCell>
          )}
          {columns.map(column => (
            <HeaderItem key={column.key} columnKey={column.key} hasMultiSort={hasMultiSort} onSort={onChangeSorting} sortOptions={sort} {...column} />
          ))}
        </TableRow>
      </TableHead>
      <TableBody>
        {items.map((item, index) => (
          <TableRow className={onItemClick ? 'clickable' : ''} hover key={item.id || index}>
            {onRowSelected && (
              <TableCell>
                <Checkbox checked={selectedRows.includes(index)} onChange={() => onRowSelection(index)} />
              </TableCell>
            )}
            {columns.map(column => (
              <TableCell className="relative" key={column.key} onClick={() => (onItemClick ? onItemClick(item) : null)}>
                {column.render(item, column.extras)}
              </TableCell>
            ))}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
};

export default DetailsTable;
