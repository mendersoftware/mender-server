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
import React from 'react';

import { Sort as SortIcon } from '@mui/icons-material';
import { makeStyles } from 'tss-react/mui';

import DetailsIndicator from '@northern.tech/common-ui/detailsindicator';
import Loader from '@northern.tech/common-ui/loader';
import Pagination from '@northern.tech/common-ui/pagination';
import { SORTING_OPTIONS } from '@northern.tech/store/constants';

export const defaultRowsPerPage = 20;

const useStyles = makeStyles()(theme => ({
  auditlogsList: {
    '& .auditlogs-list-item': {
      display: 'grid',
      gridTemplateColumns: '2fr 1fr 1fr 2fr 2fr 1.75fr 120px',
      gridColumnGap: theme.spacing(4),
      padding: `5px ${theme.spacing(2)}`,
      borderBottom: `1px solid ${theme.palette.border.main}`,
      height: theme.spacing(6),
      minHeight: theme.spacing(6),
      maxHeight: theme.spacing(6),
      alignItems: 'center',
      '&:last-of-type': {
        borderBottom: 'transparent'
      },
      '& > *': {
        display: 'flex',
        alignItems: 'center',
        maxHeight: theme.spacing(6),
        overflow: 'hidden'
      },
      '&.auditlogs-list-item-header': {
        borderBottom: 'transparent',
        cursor: 'initial',
        padding: `10px ${theme.spacing(2)}`,
        position: 'relative'
      }
    }
  }
}));

export const AuditLogsList = ({
  items,
  onChangePage,
  onChangeRowsPerPage,
  onChangeSorting,
  selectionState,
  onIssueSelection,
  userCapabilities,
  auditLogColumns
}) => {
  const { page, perPage, sort = {}, total: count, isLoading } = selectionState;
  const { classes } = useStyles();

  return (
    !!items.length && (
      <div className={`fadeIn deploy-table-contain auditlogs-list ${classes.auditlogsList}`}>
        <div className="auditlogs-list-item auditlogs-list-item-header muted">
          {auditLogColumns.map((column, index) => (
            <div
              className="columnHeader"
              key={`columnHeader-${index}`}
              onClick={() => (column.sortable ? onChangeSorting() : null)}
              style={column.sortable ? {} : { cursor: 'initial' }}
            >
              {column.title}
              {column.sortable ? <SortIcon className={`sortIcon selected ${(sort.direction === SORTING_OPTIONS.desc).toString()}`} /> : null}
            </div>
          ))}
          <div />
        </div>
        <div className="auditlogs-list">
          {items.map(item => {
            const allowsExpansion = !!item.change || item.action.includes('terminal') || item.action.includes('portforward');
            return (
              <div
                className={`auditlogs-list-item ${allowsExpansion ? 'clickable' : ''}`}
                key={`event-${item.time}`}
                onClick={() => onIssueSelection(allowsExpansion ? item : undefined)}
              >
                {auditLogColumns.map((column, index) => column.render(item, index, userCapabilities))}
                {allowsExpansion ? <DetailsIndicator /> : <div />}
              </div>
            );
          })}
        </div>
        <div className="flexbox margin-top">
          <Pagination
            className="margin-top-none"
            count={count}
            rowsPerPage={perPage}
            onChangeRowsPerPage={onChangeRowsPerPage}
            page={page}
            onChangePage={onChangePage}
          />
          <Loader show={isLoading} small />
        </div>
      </div>
    )
  );
};

export default AuditLogsList;
