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
import { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';

import { Close as CloseIcon } from '@mui/icons-material';
// material ui
import { Drawer, IconButton, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { ALL_DEVICE_STATES, DEVICE_STATES, SORTING_OPTIONS, TIMEOUTS } from '@northern.tech/store/constants';
import { getIdAttribute, getMappedDevicesList, getSearchState, getUserSettings } from '@northern.tech/store/selectors';
import { setDeviceListState, setSearchState } from '@northern.tech/store/thunks';
import pluralize from 'pluralize';

import { getHeaders } from './devices/AuthorizedDevices';
import { routes } from './devices/BaseDevices';
import Devicelist from './devices/DeviceList';

const useStyles = makeStyles()(theme => ({
  drawerOffset: {
    top: theme.mixins.toolbar.minHeight + 1,
    left: 200
  },
  paper: {
    maxWidth: '100vw',
    minHeight: '20vh',
    boxShadow: 'none'
  }
}));

const ResultTitle = ({ term, total }) => {
  const content = `${total ? total : 'No'} ${pluralize('device', total)} found for "${term}"`;
  return (
    <div className={`flexbox ${total ? 'align-items-center' : 'centered'}`}>
      <Typography variant="h6">{content}</Typography>
    </div>
  );
};

export const SearchResult = ({ onToggleSearchResult, open = true }) => {
  const navigate = useNavigate();
  const dispatch = useDispatch();

  const { columnSelection } = useSelector(getUserSettings);
  const customColumnSizes = useSelector(state => state.users.customColumns);
  const devices = useSelector(state => getMappedDevicesList(state, 'search'));
  const idAttribute = useSelector(getIdAttribute);
  const searchState = useSelector(getSearchState);

  const { classes } = useStyles();

  const [columnHeaders, setColumnHeaders] = useState(getHeaders(columnSelection, routes.devices.defaultHeaders, idAttribute));

  const { isSearching, searchTerm, searchTotal, sort = {}, page } = searchState;
  const { direction: sortDown = SORTING_OPTIONS.desc, key: sortCol } = sort;

  useEffect(() => {
    const columnHeaders = getHeaders(columnSelection, routes.devices.defaultHeaders, idAttribute);
    setColumnHeaders(columnHeaders);
  }, [columnSelection, idAttribute, idAttribute.attribute, idAttribute.scope]);

  useEffect(() => {
    if (!open && isSearching) {
      onToggleSearchResult();
    }
  }, [open, isSearching, onToggleSearchResult]);

  useEffect(() => {
    if (open && !searchTerm) {
      onToggleSearchResult();
    }
  }, [onToggleSearchResult, open, searchTerm]);

  const onDeviceSelect = device => {
    const deviceState = DEVICE_STATES[device.status] ?? ALL_DEVICE_STATES;
    dispatch(setDeviceListState({ selectedId: device.id, state: deviceState }));
    onToggleSearchResult();
    setTimeout(() => navigate(`/devices/${deviceState}?id=${device.id}`, { state: { internal: true } }), TIMEOUTS.debounceShort);
  };

  const handlePageChange = page => {
    dispatch(setSearchState({ page }));
  };

  const onSortChange = attribute => {
    const changedSortCol = attribute.name;
    let changedSortDown = sortDown === SORTING_OPTIONS.desc ? SORTING_OPTIONS.asc : SORTING_OPTIONS.desc;
    if (changedSortCol !== sortCol) {
      changedSortDown = SORTING_OPTIONS.desc;
    }
    dispatch(setSearchState({ page: 1, sort: { direction: changedSortDown, key: changedSortCol, scope: attribute.scope } }));
  };

  return (
    <Drawer
      anchor="top"
      classes={classes}
      disableAutoFocus
      disableEnforceFocus
      disableRestoreFocus
      open={open}
      ModalProps={{ className: classes.drawerOffset, BackdropProps: { className: classes.drawerOffset } }}
      slotProps={{ paper: { className: `${classes.drawerOffset} ${classes.paper}` }, transition: { direction: 'left' } }}
    >
      <div className="flexbox align-items-center margin-bottom-small space-between">
        <ResultTitle term={searchTerm} total={searchTotal} />
        <IconButton onClick={onToggleSearchResult} aria-label="close" size="large">
          <CloseIcon />
        </IconButton>
      </div>
      {!!searchTotal && (
        <Devicelist
          columnHeaders={columnHeaders}
          customColumnSizes={customColumnSizes}
          deviceListState={{ total: searchTotal, page, perPage: 10, sort: {} }}
          devices={devices}
          idAttribute={idAttribute}
          onSort={onSortChange}
          PaginationProps={{ rowsPerPageOptions: [10] }}
          pageTotal={searchTotal}
          onPageChange={handlePageChange}
          pageLoading={isSearching}
          onExpandClick={onDeviceSelect}
        />
      )}
    </Drawer>
  );
};

export default SearchResult;
