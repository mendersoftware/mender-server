// Copyright 2015 Northern.tech AS
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
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { CloudUpload } from '@mui/icons-material';
import { Button, Tab, Tabs, TextField } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import pluralize from 'pluralize';

import { getExistingReleaseTags, getReleases, getUpdateTypes, selectRelease, setReleasesListState } from '../../actions/releaseActions';
import { BENEFITS, SORTING_OPTIONS, TIMEOUTS } from '../../constants/appConstants';
import {
  getHasReleases,
  getIsEnterprise,
  getReleaseListState,
  getReleaseTags,
  getReleasesList,
  getSelectedRelease,
  getUpdateTypes as getUpdateTypesSelector,
  getUserCapabilities
} from '../../selectors';
import { useDebounce } from '../../utils/debouncehook';
import { useLocationParams } from '../../utils/liststatehook';
import ChipSelect from '../common/chipselect';
import EnterpriseNotification, { DefaultUpgradeNotification } from '../common/enterpriseNotification';
import { ControlledAutoComplete } from '../common/forms/autocomplete';
import { Filters } from '../common/forms/filters';
import { ControlledSearch } from '../common/search';
import { HELPTOOLTIPS, MenderHelpTooltip } from '../helptips/helptooltips';
import AddArtifactDialog from './dialogs/addartifact';
import ReleaseDetails from './releasedetails';
import ReleasesList from './releaseslist';

const refreshArtifactsLength = 60000;

const DeltaProgress = ({ className = '' }) => {
  const isEnterprise = useSelector(getIsEnterprise);
  return (
    <div className={`dashboard-placeholder ${className}`} style={{ display: 'grid', placeContent: 'center' }}>
      {isEnterprise ? 'There is no automatic delta artifacts generation running.' : <DefaultUpgradeNotification />}
    </div>
  );
};

const DeltaTitle = () => (
  <div className="flexbox center-aligned">
    <div>Delta Artifacts generation</div>
    <EnterpriseNotification className="margin-left-small" id={BENEFITS.deltaGeneration.id} />
  </div>
);

const tabs = [
  { key: 'releases', Title: () => 'Releases', component: ReleasesList },
  { key: 'delta', Title: DeltaTitle, component: DeltaProgress }
];

const useStyles = makeStyles()(theme => ({
  container: { maxWidth: 1600 },
  searchNote: { minHeight: '1.8rem' },
  tabContainer: { alignSelf: 'flex-start' },
  uploadButton: { minWidth: 164, marginRight: theme.spacing(2) }
}));

const Header = ({ canUpload, releasesListState, setReleasesListState, onUploadClick }) => {
  const { selectedTags = [], searchTerm = '', searchTotal, tab = tabs[0].key, total, type } = releasesListState;
  const { classes } = useStyles();
  const hasReleases = useSelector(getHasReleases);
  const existingTags = useSelector(getReleaseTags);
  const updateTypes = useSelector(getUpdateTypesSelector);

  const searchUpdated = useCallback(searchTerm => setReleasesListState({ searchTerm }), [setReleasesListState]);

  const onTabChanged = (e, tab) => setReleasesListState({ tab });

  const onFiltersChange = useCallback(({ name, tags, type }) => setReleasesListState({ selectedTags: tags, searchTerm: name, type }), [setReleasesListState]);

  return (
    <div>
      <div className="flexbox space-between center-aligned">
        <Tabs className={classes.tabContainer} value={tab} onChange={onTabChanged} textColor="primary">
          {tabs.map(({ key, Title }) => (
            <Tab key={key} label={<Title />} value={key} />
          ))}
        </Tabs>
        {canUpload && (
          <div className="flexbox center-aligned">
            <Button color="secondary" className={classes.uploadButton} onClick={onUploadClick} startIcon={<CloudUpload fontSize="small" />} variant="contained">
              Upload
            </Button>
            <MenderHelpTooltip id={HELPTOOLTIPS.artifactUpload.id} style={{ marginTop: 8 }} />
          </div>
        )}
      </div>
      {hasReleases && tab === tabs[0].key && (
        <Filters
          className={classes.container}
          onChange={onFiltersChange}
          initialValues={{ name: searchTerm, tags: selectedTags, type }}
          defaultValues={{ name: '', tags: [], type: '' }}
          filters={[
            {
              key: 'name',
              title: 'Release name',
              Component: ControlledSearch,
              componentProps: {
                onSearch: searchUpdated,
                placeholder: 'Starts with'
              }
            },
            {
              key: 'tags',
              title: 'Tags',
              Component: ChipSelect,
              componentProps: {
                options: existingTags,
                placeholder: 'Select tags',
                selection: selectedTags
              }
            },
            {
              key: 'type',
              title: 'Contains Artifact type',
              Component: ControlledAutoComplete,
              componentProps: {
                autoHighlight: true,
                autoSelect: true,
                filterSelectedOptions: true,
                freeSolo: true,
                handleHomeEndKeys: true,
                options: updateTypes,
                renderInput: params => <TextField {...params} placeholder="Any" InputProps={{ ...params.InputProps }} />
              }
            }
          ]}
        />
      )}
      <p className={`muted ${classes.searchNote}`}>{searchTerm && searchTotal !== total ? `Filtered from ${total} ${pluralize('Release', total)}` : ''}</p>
    </div>
  );
};

export const Releases = () => {
  const releasesListState = useSelector(getReleaseListState);
  const { searchTerm, sort = {}, page, perPage, tab = tabs[0].key, selectedTags, type } = releasesListState;
  const releases = useSelector(getReleasesList);
  const selectedRelease = useSelector(getSelectedRelease);
  const { canUploadReleases } = useSelector(getUserCapabilities);
  const dispatch = useDispatch();
  const { classes } = useStyles();

  const [selectedFile, setSelectedFile] = useState();
  const [showAddArtifactDialog, setShowAddArtifactDialog] = useState(false);
  const artifactTimer = useRef();
  const [locationParams, setLocationParams] = useLocationParams('releases', { defaults: { direction: SORTING_OPTIONS.desc, key: 'modified' } });
  const debouncedSearchTerm = useDebounce(searchTerm, TIMEOUTS.debounceDefault);
  const debouncedTypeFilter = useDebounce(type, TIMEOUTS.debounceDefault);

  useEffect(() => {
    if (!artifactTimer.current) {
      return;
    }
    setLocationParams({ pageState: { ...releasesListState, selectedRelease: selectedRelease.name } });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    debouncedSearchTerm,
    debouncedTypeFilter,
    // eslint-disable-next-line react-hooks/exhaustive-deps
    JSON.stringify(sort),
    page,
    perPage,
    selectedRelease.name,
    setLocationParams,
    tab,
    // eslint-disable-next-line react-hooks/exhaustive-deps
    JSON.stringify(selectedTags)
  ]);

  useEffect(() => {
    const { selectedRelease, tags, ...remainder } = locationParams;
    if (selectedRelease) {
      dispatch(selectRelease(selectedRelease));
    }
    dispatch(setReleasesListState({ ...remainder, selectedTags: tags }));
    clearInterval(artifactTimer.current);
    artifactTimer.current = setInterval(() => dispatch(getReleases()), refreshArtifactsLength);
    return () => {
      clearInterval(artifactTimer.current);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dispatch, JSON.stringify(locationParams)]);

  useEffect(() => {
    dispatch(getReleases({ searchTerm: '', searchOnly: true, page: 1, perPage: 1, selectedTags: [], type: '' }));
    dispatch(getExistingReleaseTags());
    dispatch(getUpdateTypes());
  }, [dispatch]);

  const onUploadClick = () => setShowAddArtifactDialog(true);

  const onFileUploadClick = selectedFile => {
    setSelectedFile(selectedFile);
    setShowAddArtifactDialog(true);
  };

  const onHideAddArtifactDialog = () => setShowAddArtifactDialog(false);

  const onSetReleasesListState = useCallback(state => dispatch(setReleasesListState(state)), [dispatch]);

  const ContentComponent = useMemo(() => tabs.find(({ key }) => key === tab).component, [tab]);
  return (
    <div className="margin">
      <div>
        <Header
          canUpload={canUploadReleases}
          onUploadClick={onUploadClick}
          releasesListState={releasesListState}
          setReleasesListState={onSetReleasesListState}
        />
        <ContentComponent className={classes.container} onFileUploadClick={onFileUploadClick} />
      </div>
      <ReleaseDetails />
      {showAddArtifactDialog && (
        <AddArtifactDialog releases={releases} onCancel={onHideAddArtifactDialog} onUploadStarted={onHideAddArtifactDialog} selectedFile={selectedFile} />
      )}
    </div>
  );
};

export default Releases;
