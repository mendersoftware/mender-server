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
import { makeStyles } from 'tss-react/mui';

import DetailsIndicator from '@northern.tech/common-ui/DetailsIndicator';
import type { ColumnDefinition } from '@northern.tech/common-ui/DetailsTable';
import DetailsTable from '@northern.tech/common-ui/DetailsTable';
import Loader from '@northern.tech/common-ui/Loader';
import Pagination from '@northern.tech/common-ui/Pagination';
import type { AuditLog } from '@northern.tech/types/MenderTypes';

const allowsExpansion = (item: AuditLog) =>
  !!item.change || item.action.includes('terminal') || item.action.includes('portforward') || item.action.includes('file');

const detailsColumn: ColumnDefinition<AuditLog> = {
  key: 'details',
  title: '',
  render: item => (allowsExpansion(item) ? <DetailsIndicator /> : null),
  cellProps: { style: { minWidth: 120 } }
};

const useStyles = makeStyles()(() => ({
  table: { 'tr.clickable': { cursor: 'inherit' } }
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

  if (!items.length) {
    return null;
  }

  const columns: ColumnDefinition<AuditLog>[] = [...auditLogColumns.map(column => ({ ...column, extras: userCapabilities })), detailsColumn];

  return (
    <div className="fadeIn deploy-table-contain">
      <DetailsTable
        className={`auditlogs-list ${classes.table}`}
        columns={columns}
        items={items}
        onChangeSorting={onChangeSorting}
        onItemClick={item => onIssueSelection(allowsExpansion(item) ? item : undefined)}
        sort={{ ...sort, key: 'time' }}
      />
      <div>
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
  );
};

export default AuditLogsList;
