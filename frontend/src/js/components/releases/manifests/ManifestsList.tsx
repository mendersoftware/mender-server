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
import { useCallback, useEffect, useRef } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { Typography } from '@mui/material';

import DetailsTable from '@northern.tech/common-ui/DetailsTable';
import { DefaultUpgradeNotification } from '@northern.tech/common-ui/EnterpriseNotification';
import Loader from '@northern.tech/common-ui/Loader';
import Pagination from '@northern.tech/common-ui/Pagination';
import { RelativeTime } from '@northern.tech/common-ui/Time';
import { DEVICE_LIST_DEFAULTS, SORTING_OPTIONS } from '@northern.tech/store/constants';
import { getHasManifests, getIsEnterprise, getManifestsList, getManifestsListState, getSelectedManifests } from '@northern.tech/store/selectors';
import { getManifests, selectManifest, setManifestsListState } from '@northern.tech/store/thunks';
import type { Manifest } from '@northern.tech/types/MenderTypes';

import { ManifestDetails } from './ManifestDetails';
import { ManifestQuickActions } from './ManifestQuickActions';

const columns = [
  {
    key: 'name',
    title: 'Name',
    render: ({ name }: Manifest) => name,
    sortable: true,
    defaultSortDirection: SORTING_OPTIONS.asc
  },
  {
    key: 'tags',
    title: 'Tags',
    render: ({ tags = [] }: Manifest) => tags.join(', ') || '-',
    sortable: true,
    defaultSortDirection: SORTING_OPTIONS.asc
  },
  {
    key: 'modified',
    title: 'Latest modified',
    render: ({ modified }: Manifest) => <RelativeTime updateTime={modified} />,
    sortable: true,
    defaultSortDirection: SORTING_OPTIONS.desc
  }
];

const { page: defaultPage, perPage: defaultPerPage } = DEVICE_LIST_DEFAULTS;

const EmptyState = ({ className = '', isFiltering, onFileUploadClick }: { className?: string; isFiltering?: boolean; onFileUploadClick: () => void }) => (
  <div className={`dashboard-placeholder fadeIn ${className}`}>
    <Typography className="margin-top align-center">
      There are no Manifests{' '}
      {isFiltering ? (
        'for the filter selection'
      ) : (
        <>
          to display. <a onClick={onFileUploadClick}>Create a Manifest</a>
        </>
      )}
    </Typography>
  </div>
);

export const ManifestsList = ({ className = '', onFileUploadClick }: { className?: string; onFileUploadClick: () => void }) => {
  const tableRef = useRef<HTMLDivElement>(null);
  const manifestsListState = useSelector(getManifestsListState);
  const { isLoading, page = defaultPage, perPage = defaultPerPage, searchTerm, sort = {}, searchTotal, selection: selectedRows, total } = manifestsListState;
  const hasManifests = useSelector(getHasManifests);
  const manifests = useSelector(getManifestsList);
  const selectedManifests = useSelector(getSelectedManifests);
  const isEnterprise = useSelector(getIsEnterprise);
  const dispatch = useDispatch();

  const { key: attribute, direction } = sort;

  useEffect(() => {
    if (!isEnterprise) {
      return;
    }
    dispatch(getManifests())
      .unwrap()
      .then(() => dispatch(setManifestsListState({ isLoading: false })));
  }, [dispatch, isEnterprise]);

  const onSelect = useCallback((id: string) => dispatch(selectManifest(id)), [dispatch]);

  const onChangeSorting = (sortKey: string) => {
    let sort = { key: sortKey, direction: direction === SORTING_OPTIONS.asc ? SORTING_OPTIONS.desc : SORTING_OPTIONS.asc };
    if (sortKey !== attribute) {
      sort = { ...sort, direction: columns.find(({ key }) => key === sortKey)?.defaultSortDirection ?? SORTING_OPTIONS.desc };
    }
    dispatch(setManifestsListState({ page: 1, sort }));
  };

  const onChangePagination = (page: number, currentPerPage: number = perPage) => dispatch(setManifestsListState({ page, perPage: currentPerPage }));

  const onSelectionChange = useCallback((selection: number[] = []) => dispatch(setManifestsListState({ selection })), [dispatch]);

  const isFiltering = !!searchTerm;
  const potentialTotal = isFiltering ? searchTotal : total;

  if (!isEnterprise) {
    return <DefaultUpgradeNotification className={`dashboard-placeholder ${className}`} style={{ display: 'grid', placeContent: 'center' }} />;
  }

  if (!hasManifests) {
    return <EmptyState className={className} onFileUploadClick={onFileUploadClick} />;
  }

  return (
    <div className={className}>
      {isLoading === undefined ? (
        <Loader show />
      ) : !potentialTotal ? (
        <EmptyState isFiltering={isFiltering} onFileUploadClick={onFileUploadClick} />
      ) : (
        <>
          <DetailsTable
            columns={columns}
            items={manifests}
            onItemClick={onSelect}
            sort={sort}
            onChangeSorting={onChangeSorting}
            tableRef={tableRef}
            onRowSelected={onSelectionChange}
            selectedRows={selectedRows}
          />
          <div className="flexbox">
            <Pagination
              className="margin-top-none"
              count={potentialTotal}
              rowsPerPage={perPage}
              onChangePage={onChangePagination}
              onChangeRowsPerPage={(newPerPage: number) => onChangePagination(1, newPerPage)}
              page={page}
            />
            <Loader show={isLoading} small />
          </div>
          {selectedManifests?.length > 0 && <ManifestQuickActions />}
        </>
      )}
      <ManifestDetails />
    </div>
  );
};

export default ManifestsList;
