// Copyright 2026 Northern.tech AS
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
import { useMemo, useState } from 'react';

import { Table, TableBody, TableCell, TableHead, TableRow, TableSortLabel } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { SortCriteria } from '@northern.tech/types/MenderTypes';
import { SORTING_OPTIONS } from '@northern.tech/utils/constants';
import copy from 'copy-to-clipboard';

import { CopyableText } from './CopyableText';

const useStyles = makeStyles()(() => ({
  attributeColumn: {
    width: '30%'
  },
  valueColumn: {
    width: '70%'
  }
}));

type SortColumn = 'attribute' | 'value';

const columns: Record<SortColumn, SortColumn> = {
  attribute: 'attribute',
  value: 'value'
};

interface InventoryTableProps {
  config: Record<string, string>;
  setSnackbar?: (message: string) => void;
}

export const InventoryTable = ({ config, setSnackbar }: InventoryTableProps) => {
  const { classes } = useStyles();
  const [sortColumn, setSortColumn] = useState<SortColumn>(columns.attribute);
  const [sortDirection, setSortDirection] = useState<SortCriteria['order']>(SORTING_OPTIONS.asc);

  const isSortedByAttribute = sortColumn === columns.attribute;

  const onSort = (column: SortColumn) => {
    if (sortColumn === column) {
      setSortDirection(sortDirection === SORTING_OPTIONS.asc ? SORTING_OPTIONS.desc : SORTING_OPTIONS.asc);
    } else {
      setSortColumn(column);
      setSortDirection(SORTING_OPTIONS.asc);
    }
  };

  const onCopy = (value: string) => {
    if (setSnackbar) {
      copy(value);
      setSnackbar('Value copied to clipboard');
    }
  };

  const sortedEntries = useMemo(
    () =>
      Object.entries(config).sort((a, b) => {
        const aValue = isSortedByAttribute ? a[0] : a[1];
        const bValue = isSortedByAttribute ? b[0] : b[1];
        const comparison = aValue.localeCompare(bValue);
        return sortDirection === SORTING_OPTIONS.asc ? comparison : -comparison;
      }),
    [config, isSortedByAttribute, sortDirection]
  );

  return (
    <Table>
      <TableHead>
        <TableRow>
          <TableCell className={`bold ${classes.attributeColumn}`}>
            <TableSortLabel active={isSortedByAttribute} direction={sortDirection} onClick={() => onSort(columns.attribute)}>
              Attribute
            </TableSortLabel>
          </TableCell>
          <TableCell className={`bold ${classes.valueColumn}`}>
            <TableSortLabel active={!isSortedByAttribute} direction={sortDirection} onClick={() => onSort(columns.value)}>
              Value
            </TableSortLabel>
          </TableCell>
        </TableRow>
      </TableHead>
      <TableBody>
        {sortedEntries.map(([attribute, value]) => (
          <TableRow key={attribute}>
            <TableCell className={`bold ${classes.attributeColumn} ${setSnackbar ? 'clickable' : ''}`}>
              <CopyableText onCopy={() => onCopy(attribute)} textClasses="bold" title={attribute}>
                {attribute}
              </CopyableText>
            </TableCell>
            <TableCell className={setSnackbar ? 'clickable' : ''}>
              <CopyableText onCopy={() => onCopy(value)} title={value}>
                {value}
              </CopyableText>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
};
