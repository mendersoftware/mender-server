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
import React, { useCallback, useMemo, useRef, useState } from 'react';
import Dropzone from 'react-dropzone';
import { useDispatch, useSelector } from 'react-redux';

import { makeStyles } from 'tss-react/mui';

import DetailsTable from '@northern.tech/common-ui/DetailsTable';
import Loader from '@northern.tech/common-ui/Loader';
import Pagination from '@northern.tech/common-ui/Pagination';
import { RelativeTime } from '@northern.tech/common-ui/Time';
import storeActions from '@northern.tech/store/actions';
import { DEVICE_LIST_DEFAULTS, SORTING_OPTIONS, canAccess as canShow } from '@northern.tech/store/constants';
import { getFeatures, getHasReleases, getReleaseListState, getReleasesList, getSelectedReleases, getUserCapabilities } from '@northern.tech/store/selectors';
import { removeReleases, selectRelease, setReleasesListState } from '@northern.tech/store/thunks';

import { DeleteReleasesConfirmationDialog, ReleaseQuickActions } from './ReleaseDetails';
import AddTagsDialog from './dialogs/AddTags';

const { setSnackbar } = storeActions;

const columns = [
  {
    key: 'name',
    title: 'Name',
    render: ({ name }) => name,
    sortable: true,
    defaultSortDirection: SORTING_OPTIONS.asc,
    canShow
  },
  {
    key: 'artifacts-count',
    title: 'Number of artifacts',
    render: ({ artifacts = [] }) => artifacts.length,
    canShow
  },
  {
    key: 'tags',
    title: 'Tags',
    render: ({ tags = [] }) => tags.join(', ') || '-',
    defaultSortDirection: SORTING_OPTIONS.asc,
    sortable: true,
    canShow
  },
  {
    key: 'modified',
    title: 'Last modified',
    render: ({ modified }) => <RelativeTime updateTime={modified} />,
    defaultSortDirection: SORTING_OPTIONS.desc,
    sortable: true,
    canShow
  }
];

const useStyles = makeStyles()(() => ({
  empty: { margin: '8vh auto' }
}));

const { page: defaultPage, perPage: defaultPerPage } = DEVICE_LIST_DEFAULTS;

const EmptyState = ({ canUpload, className = '', dropzoneRef, uploading, onDrop, onUpload }) => (
  <div className={`dashboard-placeholder fadeIn ${className}`} ref={dropzoneRef}>
    <Dropzone activeClassName="active" disabled={uploading} multiple={false} noClick={true} onDrop={onDrop} rejectClassName="active">
      {({ getRootProps, getInputProps }) => (
        <div {...getRootProps({ className: uploading ? 'dropzone disabled muted' : 'dropzone' })} onClick={() => onUpload()}>
          <input {...getInputProps()} disabled={uploading} />
          <p>
            There are no Releases yet.{' '}
            {canUpload && (
              <>
                <a>Upload an Artifact</a> to create a new Release
              </>
            )}
          </p>
        </div>
      )}
    </Dropzone>
  </div>
);

export const ReleasesList = ({ className = '', onFileUploadClick }) => {
  const repoRef = useRef();
  const dropzoneRef = useRef();
  const uploading = useSelector(state => state.app.uploading);
  const releasesListState = useSelector(getReleaseListState);
  const {
    isLoading,
    page = defaultPage,
    perPage = defaultPerPage,
    searchTerm,
    sort = {},
    searchTotal,
    selection: selectedRows,
    selectedTags = [],
    total,
    type
  } = releasesListState;
  const hasReleases = useSelector(getHasReleases);
  const features = useSelector(getFeatures);
  const releases = useSelector(getReleasesList);
  const userCapabilities = useSelector(getUserCapabilities);
  const selectedReleases = useSelector(getSelectedReleases);
  const dispatch = useDispatch();
  const { classes } = useStyles();
  const [addTagsDialog, setAddTagsDialog] = useState(false);
  const [deleteDialogConfirmation, setDeleteDialogConfirmation] = useState(false);

  const { canUploadReleases } = userCapabilities;
  const { key: attribute, direction } = sort;

  const onSelect = useCallback(id => dispatch(selectRelease(id)), [dispatch]);

  const onChangeSorting = sortKey => {
    let sort = { key: sortKey, direction: direction === SORTING_OPTIONS.asc ? SORTING_OPTIONS.desc : SORTING_OPTIONS.asc };
    if (sortKey !== attribute) {
      sort = { ...sort, direction: columns.find(({ key }) => key === sortKey)?.defaultSortDirection ?? SORTING_OPTIONS.desc };
    }
    dispatch(setReleasesListState({ page: 1, sort }));
  };

  const onChangePagination = (page, currentPerPage = perPage) => dispatch(setReleasesListState({ page, perPage: currentPerPage }));

  const onDrop = (acceptedFiles, rejectedFiles) => {
    if (acceptedFiles.length) {
      onFileUploadClick(acceptedFiles[0]);
    }
    if (rejectedFiles.length) {
      dispatch(setSnackbar(`File '${rejectedFiles[0].name}' was rejected. File should be of type .mender`, null));
    }
  };

  const applicableColumns = useMemo(
    () =>
      columns.reduce((accu, column) => {
        if (column.canShow({ features })) {
          accu.push(column);
        }
        return accu;
      }, []),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [JSON.stringify(features)]
  );

  const onDeleteRelease = releases => {
    onSelectionChange(releases);
    setDeleteDialogConfirmation(true);
  };

  const onSelectionChange = useCallback((selection: number[] = []) => dispatch(setReleasesListState({ selection })), [dispatch]);

  const deleteReleases = useCallback(() => {
    dispatch(removeReleases(selectedReleases.map(({ name }) => name))).then(() => {
      setDeleteDialogConfirmation(false);
      onSelectionChange([]);
    });
  }, [dispatch, onSelectionChange, selectedReleases]);

  const onTagRelease = releases => {
    onSelectionChange(releases);
    setAddTagsDialog(true);
  };

  const actionCallbacks = {
    onDeleteRelease,
    onTagRelease
  };

  const isFiltering = !!(selectedTags.length || type || searchTerm);
  const potentialTotal = isFiltering ? searchTotal : total;
  if (!hasReleases) {
    return (
      <EmptyState
        canUpload={canUploadReleases}
        className={classes.empty}
        dropzoneRef={dropzoneRef}
        uploading={uploading}
        onDrop={onDrop}
        onUpload={onFileUploadClick}
      />
    );
  }

  return (
    <div className={className}>
      {isLoading === undefined ? (
        <Loader show />
      ) : !potentialTotal ? (
        <p className="margin-top muted align-center margin-right">There are no Releases {isFiltering ? 'for the filter selection' : 'yet'}</p>
      ) : (
        <>
          <DetailsTable
            columns={applicableColumns}
            items={releases}
            onItemClick={onSelect}
            sort={sort}
            onChangeSorting={onChangeSorting}
            tableRef={repoRef}
            onRowSelected={onSelectionChange}
            selectedRows={selectedRows}
          />
          <div className="flexbox">
            <Pagination
              className="margin-top-none"
              count={potentialTotal}
              rowsPerPage={perPage}
              onChangePage={onChangePagination}
              onChangeRowsPerPage={newPerPage => onChangePagination(1, newPerPage)}
              page={page}
            />
            <Loader show={isLoading} small />
          </div>
          {selectedReleases?.length > 0 && <ReleaseQuickActions actionCallbacks={actionCallbacks} />}
          {addTagsDialog && <AddTagsDialog selectedReleases={selectedReleases} onClose={() => setAddTagsDialog(false)} />}
          {deleteDialogConfirmation && <DeleteReleasesConfirmationDialog onClose={() => setDeleteDialogConfirmation(false)} onSubmit={deleteReleases} />}
        </>
      )}
    </div>
  );
};

export default ReleasesList;
