// Copyright 2019 Northern.tech AS
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

import { KeyboardArrowLeft, KeyboardArrowRight } from '@mui/icons-material';
import { IconButton, TablePagination } from '@mui/material';

import { DEVICE_LIST_DEFAULTS, DEVICE_LIST_MAXIMUM_LENGTH, TIMEOUTS } from '@northern.tech/store/constants';
import { useDebounce } from '@northern.tech/utils/debouncehook';

import MenderTooltip from './mendertooltip';

const defaultRowsPerPageOptions = [10, 20, DEVICE_LIST_MAXIMUM_LENGTH];
const { perPage: defaultPerPage } = DEVICE_LIST_DEFAULTS;
const paginationIndex = 1;
const paginationLimit = 10000;

const MaybeWrapper = ({ children, disabled }) =>
  disabled ? (
    <MenderTooltip arrow placement="top" title="Please refine your filter criteria first in order to proceed.">
      <div>{children}</div>
    </MenderTooltip>
  ) : (
    <div>{children}</div>
  );

export const TablePaginationActions = ({ count, page = 0, onPageChange, rowsPerPage = defaultPerPage, showCountInfo = true }) => {
  const [pageNo, setPageNo] = useState(page + paginationIndex);

  useEffect(() => {
    setPageNo(page + paginationIndex);
  }, [page, rowsPerPage, count]);

  const debouncedPage = useDebounce(pageNo, TIMEOUTS.debounceShort);

  useEffect(() => {
    const newPage = Math.min(Math.max(paginationIndex, debouncedPage), Math.max(paginationIndex, Math.ceil(count / rowsPerPage)));
    if (newPage !== page + paginationIndex) {
      onPageChange(newPage);
    }
  }, [count, debouncedPage, onPageChange, page, rowsPerPage]);

  const pages = Math.ceil(count / rowsPerPage);

  const isAtPaginationLimit = pageNo >= paginationLimit / rowsPerPage;
  return (
    <div className="flexbox center-aligned">
      {showCountInfo && <div>{`${(pageNo - paginationIndex) * rowsPerPage + 1}-${Math.min(pageNo * rowsPerPage, count)} of ${count}`}</div>}
      <IconButton onClick={() => setPageNo(pageNo - 1)} disabled={pageNo === paginationIndex} size="large">
        <KeyboardArrowLeft />
      </IconButton>
      <MaybeWrapper disabled={isAtPaginationLimit}>
        <IconButton onClick={() => setPageNo(pageNo + 1)} disabled={pageNo >= pages || isAtPaginationLimit} size="large">
          <KeyboardArrowRight />
        </IconButton>
      </MaybeWrapper>
    </div>
  );
};

const Pagination = props => {
  const { className, onChangeRowsPerPage, onChangePage, page = 0, rowsPerPageOptions = defaultRowsPerPageOptions, showCountInfo, ...remainingProps } = props;
  // this is required due to the MUI tablepagination being 0-indexed, whereas we work with 1-indexed apis
  // running it without adjustment will lead to warnings from MUI
  const propsPage = Math.max(page - paginationIndex, 0);
  return (
    <TablePagination
      className={`flexbox margin-top ${className || ''}`}
      classes={{ spacer: 'flexbox no-basis' }}
      component="div"
      labelDisplayedRows={() => ''}
      labelRowsPerPage="Rows"
      SelectProps={{ name: 'pagination' }}
      rowsPerPageOptions={rowsPerPageOptions}
      onRowsPerPageChange={e => onChangeRowsPerPage(e.target.value)}
      page={propsPage}
      onPageChange={onChangePage}
      ActionsComponent={actionProps => <TablePaginationActions {...actionProps} showCountInfo={showCountInfo} />}
      {...remainingProps}
    />
  );
};

export default Pagination;
