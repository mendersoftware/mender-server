// Copyright 2025 Northern.tech AS
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
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Link } from 'react-router-dom';

import { Typography, tableCellClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import DetailsTable from '@northern.tech/common-ui/DetailsTable';
import { DefaultUpgradeNotification } from '@northern.tech/common-ui/EnterpriseNotification';
import Loader from '@northern.tech/common-ui/Loader';
import Pagination from '@northern.tech/common-ui/Pagination';
import { MaybeTime } from '@northern.tech/common-ui/Time';
import storeActions from '@northern.tech/store/actions';
import { DEVICE_LIST_DEFAULTS, SORTING_OPTIONS, SortOptions } from '@northern.tech/store/constants';
import { getDeltaJobsById, getDeltaJobsListState, getIsEnterprise, getSelectedJob } from '@northern.tech/store/selectors';
import { getDeltaGenerationJobs } from '@northern.tech/store/thunks';
import { formatTime } from '@northern.tech/utils/helpers';

import DeltaGenerationDetailsDrawer, { StatusIndicator } from './DeltaGenerationDetailsDrawer';

const { setDeltaJobsListState, setSelectedJob } = storeActions;

const useStyles = makeStyles()(() => ({
  table: {
    [`.${tableCellClasses.body}, .${tableCellClasses.head}`]: {
      paddingLeft: 0
    }
  }
}));

const deltaJobColumns = [
  {
    key: 'to_release',
    title: 'To version',
    sortable: true,
    cellProps: { style: { width: '20%' } },
    render: ({ to_release, to_version }) => <Link to={`/releases/${to_release || to_version || '-'}`}>{to_release || to_version || '-'}</Link>
  },
  {
    key: 'from_release',
    title: 'From version',
    sortable: true,
    cellProps: { style: { width: '20%' } },
    render: ({ from_release, from_version }) => <Link to={`/releases/${from_release || from_version}`}>{from_release || from_version || '-'}</Link>
  },
  {
    key: 'device_types_compatible',
    title: 'Device types compatible',
    sortable: false,
    cellProps: { style: { width: '25%' } },
    render: ({ device_types_compatible, devices_types_compatible }) => (device_types_compatible || devices_types_compatible || []).join(', ') || '-'
  },
  {
    key: 'started',
    title: 'Started',
    sortable: true,
    cellProps: { style: { width: '10%' } },
    render: ({ started }) => <MaybeTime value={formatTime(started)} />
  },
  {
    key: 'status',
    title: 'Status',
    sortable: true,
    cellProps: { style: { width: '20%' } },
    render: StatusIndicator
  },
  {
    key: 'spacer',
    title: '',
    sortable: false,
    cellProps: { style: { width: '5%' } },
    render: () => ''
  }
];

const { page: defaultPage, perPage: defaultPerPage } = DEVICE_LIST_DEFAULTS;

export const DeltaProgress = ({ className = '' }) => {
  const dispatch = useDispatch();
  const isEnterprise = useSelector(getIsEnterprise);
  const { jobIds, total, sort = {} as SortOptions, page = defaultPage, perPage = defaultPerPage } = useSelector(getDeltaJobsListState);
  const byId = useSelector(getDeltaJobsById);
  const selectedJob = useSelector(getSelectedJob);
  const [isLoading, setIsLoading] = useState(false);
  const { classes } = useStyles();

  useEffect(() => {
    if (!isEnterprise) {
      return;
    }
    setIsLoading(true);
    dispatch(getDeltaGenerationJobs({ sort: { key: sort.key, direction: sort.direction }, page, perPage }))
      .unwrap()
      .finally(() => setIsLoading(false));
  }, [dispatch, isEnterprise, sort.key, sort.direction, page, perPage]);

  const jobsList = useMemo(() => jobIds.map(id => byId[id]).filter(Boolean), [byId, jobIds]);

  const onChangeSorting = useCallback(
    key => {
      const direction = sort.key === key && sort.direction === SORTING_OPTIONS.desc ? SORTING_OPTIONS.asc : SORTING_OPTIONS.desc;
      const newSort = { key, direction };
      dispatch(setDeltaJobsListState({ sort: newSort, page: 1 }));
    },
    [dispatch, sort]
  );

  const onChangePagination = useCallback(
    (newPage, currentPerPage = perPage) => {
      dispatch(setDeltaJobsListState({ page: newPage, perPage: currentPerPage }));
    },
    [dispatch, perPage]
  );

  const onJobSelect = useCallback(selection => dispatch(setSelectedJob(selection.id)), [dispatch]);

  const onCloseDetailsDrawer = () => onJobSelect({ id: null });

  if (!isEnterprise) {
    return (
      <div className={`dashboard-placeholder ${className}`} style={{ display: 'grid', placeContent: 'center' }}>
        <DefaultUpgradeNotification />
      </div>
    );
  }

  if (!total) {
    return (
      <div className={`dashboard-placeholder ${className}`} style={{ display: 'grid', placeContent: 'center' }}>
        No Delta Artifacts have been generated in the last 30 days.
      </div>
    );
  }

  return (
    <div className={className}>
      <Typography className="margin-bottom" variant="subtitle1">
        Generated Delta Artifacts
      </Typography>
      <DetailsTable
        className={classes.table}
        columns={deltaJobColumns}
        items={jobsList}
        onChangeSorting={onChangeSorting}
        onItemClick={onJobSelect}
        sort={sort}
      />
      <div className="flexbox">
        <Pagination
          className="margin-top-none"
          count={total}
          onChangePage={onChangePagination}
          onChangeRowsPerPage={newPerPage => onChangePagination(1, newPerPage)}
          page={page}
          rowsPerPage={perPage}
        />
        <Loader show={isLoading} small />
      </div>
      <DeltaGenerationDetailsDrawer jobId={selectedJob?.id} open={!!selectedJob?.id} onClose={onCloseDetailsDrawer} />
    </div>
  );
};
