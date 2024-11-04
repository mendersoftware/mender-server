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
import React from 'react';

// material ui
import { Sort as SortIcon } from '@mui/icons-material';
import { Checkbox, Table, TableBody, TableCell, TableHead, TableRow } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { SORTING_OPTIONS } from '@northern.tech/store/constants';

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

export const DetailsTable = ({
  className = '',
  columns,
  items,
  onChangeSorting,
  onItemClick,
  sort = {},
  style = {},
  tableRef,
  onRowSelected = undefined,
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
          {onRowSelected !== undefined && (
            <TableCell>
              <Checkbox indeterminate={false} checked={selectedRows.length === items.length} onChange={onSelectAllClick} />
            </TableCell>
          )}
          {columns.map(({ extras, key, renderTitle, sortable, title }) => (
            <TableCell key={key} className={`columnHeader ${sortable ? '' : 'nonSortable'}`} onClick={() => (sortable ? onChangeSorting(key) : null)}>
              {renderTitle ? renderTitle(extras) : title}
              {sortable && <SortIcon className={`sortIcon ${sort.key === key ? 'selected' : ''} ${(sort.direction === SORTING_OPTIONS.desc).toString()}`} />}
            </TableCell>
          ))}
        </TableRow>
      </TableHead>
      <TableBody>
        {items.map((item, index) => (
          <TableRow className={onItemClick ? 'clickable' : ''} hover key={item.id || index}>
            {onRowSelected !== undefined && (
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
