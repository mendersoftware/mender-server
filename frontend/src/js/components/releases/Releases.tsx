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
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useLocation } from 'react-router-dom';

import { CloudUpload } from '@mui/icons-material';
import { Button, Tab, Tabs } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import EnterpriseNotification from '@northern.tech/common-ui/EnterpriseNotification';
import storeActions from '@northern.tech/store/actions';
import { BENEFITS, SORTING_OPTIONS, TIMEOUTS } from '@northern.tech/store/constants';
import { useLocationParams } from '@northern.tech/store/liststatehook';
import { getActiveTab, getReleaseListState, getReleasesList, getSelectedRelease, getUserCapabilities } from '@northern.tech/store/selectors';
import { selectRelease, setReleasesListState } from '@northern.tech/store/thunks';
import { useDebounce } from '@northern.tech/utils/debouncehook';

import { HELPTOOLTIPS } from '../helptips/HelpTooltips';
import { MenderHelpTooltip } from '../helptips/MenderTooltip';
import { DeltaProgress } from './DeltaGeneration';
import { ReleasesFilters } from './ReleasesFilters';
import ReleasesList from './ReleasesList';
import AddArtifactDialog from './dialogs/AddArtifact';

const { setActiveTab, setSelectedJob } = storeActions;

const UploadRelease = ({ classes, onUploadClick }) => (
  <div className="flexbox align-items-center">
    <Button className={classes.uploadButton} onClick={onUploadClick} startIcon={<CloudUpload fontSize="small" />} variant="contained">
      Upload an artifact
    </Button>
    <MenderHelpTooltip id={HELPTOOLTIPS.artifactUpload.id} style={{ marginTop: 8 }} />
  </div>
);

const DeltaTitle = () => (
  <div className="flexbox align-items-center">
    <div>Delta Artifacts generation</div>
    <EnterpriseNotification className="margin-left-small" id={BENEFITS.deltaGeneration.id} />
  </div>
);

const tabs = [
  { key: 'releases', Title: () => 'Releases', component: ReleasesList },
  { key: 'delta', Title: DeltaTitle, component: DeltaProgress }
];

const tabbedComponents = {
  releases: { Filters: ReleasesFilters, Upload: UploadRelease }
};

const useStyles = makeStyles()(theme => ({
  container: { maxWidth: 1600 },
  searchNote: { minHeight: '1.8rem' },
  tabContainer: { alignSelf: 'flex-start' },
  uploadButton: { minWidth: 164, marginRight: theme.spacing(2) }
}));

const Header = ({ canUpload, tab, onTabChanged, onUploadClick, tabs }) => {
  const { classes } = useStyles();

  const { Filters: FilterComponent, Upload: UploadComponent } = tabbedComponents[tab] ?? {};

  return (
    <div>
      <div className="flexbox space-between align-items-center">
        <Tabs className={classes.tabContainer} value={tab} onChange={onTabChanged} textColor="primary">
          {tabs.map(({ key, Title }) => (
            <Tab key={key} label={<Title />} value={key} />
          ))}
        </Tabs>
        {canUpload && UploadComponent && <UploadComponent classes={classes} onUploadClick={onUploadClick} />}
      </div>
      {FilterComponent && <FilterComponent classes={classes} />}
    </div>
  );
};

export const Releases = () => {
  const releasesListState = useSelector(getReleaseListState);
  const { searchTerm, sort = {}, page, perPage, selectedTags, type } = releasesListState;
  const tab = useSelector(getActiveTab);
  const releases = useSelector(getReleasesList);
  const selectedRelease = useSelector(getSelectedRelease);
  const { canUploadReleases } = useSelector(getUserCapabilities);
  const dispatch = useDispatch();
  const { classes } = useStyles();

  const [selectedFile, setSelectedFile] = useState();
  const [showAddArtifactDialog, setShowAddArtifactDialog] = useState(false);
  const isInitialized = useRef(false);
  const location = useLocation();
  const [locationParams, setLocationParams, { shouldInitializeFromUrl }] = useLocationParams('releases', {
    defaults: { direction: SORTING_OPTIONS.desc, key: 'modified' }
  });
  const debouncedSearchTerm = useDebounce(searchTerm, TIMEOUTS.debounceDefault);
  const debouncedTypeFilter = useDebounce(type, TIMEOUTS.debounceDefault);

  useEffect(() => {
    // if an upload is ongoing and another upload is being prepared, prioritize the configuration of the new upload instead of showing the newly created release
    if (showAddArtifactDialog && selectedRelease) {
      dispatch(selectRelease(null));
    }
  }, [dispatch, selectedRelease, showAddArtifactDialog]);

  useEffect(() => {
    if (shouldInitializeFromUrl) {
      isInitialized.current = false;
    }
  }, [shouldInitializeFromUrl, location.key]);

  useEffect(() => {
    if (!isInitialized.current) {
      return;
    }
    setLocationParams({ pageState: { ...releasesListState, tab, selectedRelease: selectedRelease.name } });
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
    if (isInitialized.current || !shouldInitializeFromUrl) {
      isInitialized.current = true;
      return;
    }
    const { selectedRelease, selectedJob, tab: urlTab, tags, ...remainder } = locationParams;
    if (selectedRelease) {
      dispatch(selectRelease(selectedRelease));
    }
    if (selectedJob) {
      dispatch(setSelectedJob(selectedJob));
    }
    if (urlTab) {
      dispatch(setActiveTab(urlTab));
    }
    dispatch(setReleasesListState({ ...remainder, selectedTags: tags }));
    isInitialized.current = true;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dispatch, JSON.stringify(locationParams), shouldInitializeFromUrl]);

  const onUploadClick = () => setShowAddArtifactDialog(true);

  const onFileUploadClick = selectedFile => {
    setSelectedFile(selectedFile);
    setShowAddArtifactDialog(true);
  };

  const onHideAddArtifactDialog = () => setShowAddArtifactDialog(false);

  const onTabChanged = useCallback((_, changedTab: 'releases' | 'delta' | 'manifests') => dispatch(setActiveTab(changedTab)), [dispatch]);

  const ContentComponent = useMemo(() => tabs.find(({ key }) => key === tab).component, [tab]);
  return (
    <div className="margin">
      <div>
        <Header canUpload={canUploadReleases} tab={tab} onTabChanged={onTabChanged} onUploadClick={onUploadClick} tabs={tabs} />
        <ContentComponent className={classes.container} onFileUploadClick={onFileUploadClick} />
      </div>
      {showAddArtifactDialog && (
        <AddArtifactDialog releases={releases} onCancel={onHideAddArtifactDialog} onUploadStarted={onHideAddArtifactDialog} selectedFile={selectedFile} />
      )}
    </div>
  );
};

export default Releases;
