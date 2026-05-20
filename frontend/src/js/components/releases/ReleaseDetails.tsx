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
import { useCallback, useEffect, useRef, useState } from 'react';
import { FormProvider, useForm } from 'react-hook-form';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';

// material ui
import {
  HighlightOffOutlined as HighlightOffOutlinedIcon,
  LabelOutlined as LabelOutlinedIcon,
  Sort as SortIcon,
  SyncOutlined as SyncOutlinedIcon
} from '@mui/icons-material';
import { Button, DialogActions, DialogContent, Tooltip, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import BaseDrawer from '@northern.tech/common-ui/BaseDrawer';
import ChipSelect from '@northern.tech/common-ui/ChipSelect';
import { ConfirmationButtons, EditButton } from '@northern.tech/common-ui/Confirm';
import { ContentSection } from '@northern.tech/common-ui/ContentSection';
import { EditableLongText } from '@northern.tech/common-ui/EditableLongText';
import FileSize from '@northern.tech/common-ui/FileSize';
import { BaseQuickActions, type QuickAction } from '@northern.tech/common-ui/QuickActions';
import { RelativeTime } from '@northern.tech/common-ui/Time';
import { ColumnWidthProvider, TwoColumnData } from '@northern.tech/common-ui/TwoColumnData';
import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import storeActions from '@northern.tech/store/actions';
import { DEPLOYMENT_ROUTES } from '@northern.tech/store/constants';
import { generateReleasesPath } from '@northern.tech/store/locationutils';
import { getReleaseListState, getReleaseTags, getSelectedRelease, getUserCapabilities } from '@northern.tech/store/selectors';
import { removeArtifact, removeRelease, selectRelease, setReleaseTags, updateReleaseInfo } from '@northern.tech/store/thunks';
import { customSort, formatTime, isEmpty, toggle } from '@northern.tech/utils/helpers';
import { useWindowSize } from '@northern.tech/utils/resizehook';
import copy from 'copy-to-clipboard';
import pluralize from 'pluralize';

import { HELPTOOLTIPS } from '../helptips/HelpTooltips';
import { MenderHelpTooltip } from '../helptips/MenderTooltip';
import Artifact from './Artifact';
import RemoveArtifactDialog from './dialogs/RemoveArtifact';

const { setSnackbar } = storeActions;

const DeviceTypeCompatibility = ({ artifact }) => {
  const compatible = artifact.artifact_depends ? artifact.artifact_depends.device_type.join(', ') : artifact.device_types_compatible.join(', ');
  return (
    <Tooltip title={compatible} placement="top-start">
      <div className="text-overflow">{compatible}</div>
    </Tooltip>
  );
};

export const columns = [
  {
    title: 'Device type compatibility',
    name: 'device_types',
    sortable: false,
    render: DeviceTypeCompatibility,
    tooltip: <MenderHelpTooltip id={HELPTOOLTIPS.expandArtifact.id} className="margin-left-small" />
  },
  {
    title: 'Type',
    name: 'type',
    sortable: false,
    render: ({ artifact }) => <div style={{ maxWidth: '100vw' }}>{artifact.updates.reduce((accu, item) => (accu ? accu : item.type_info.type), '')}</div>
  },
  { title: 'Size', name: 'size', sortable: true, render: ({ artifact }) => <FileSize fileSize={artifact.size} /> },
  { title: 'Last modified', name: 'modified', sortable: true, render: ({ artifact }) => <RelativeTime updateTime={formatTime(artifact.modified)} /> }
];

const defaultActions = [
  {
    action: ({ onCreateDeployment, selection }) => onCreateDeployment(selection),
    icon: <SyncOutlinedIcon />,
    isApplicable: ({ userCapabilities: { canDeploy }, selectedSingleRelease, selectedRows }) =>
      canDeploy && (selectedSingleRelease || selectedRows.length === 1),
    key: 'deploy',
    title: () => 'Create a deployment for this release'
  },
  {
    action: ({ onTagRelease, selection }) => onTagRelease(selection),
    icon: <LabelOutlinedIcon />,
    isApplicable: ({ userCapabilities: { canManageReleases }, selectedSingleRelease }) => canManageReleases && !selectedSingleRelease,
    key: 'tag',
    title: pluralized => `Tag ${pluralized}`
  },
  {
    action: ({ onDeleteRelease, selection }) => onDeleteRelease(selection),
    icon: <HighlightOffOutlinedIcon className="red" />,
    isApplicable: ({ userCapabilities: { canManageReleases } }) => canManageReleases,
    key: 'delete',
    title: pluralized => `Delete ${pluralized}`
  }
];

const useStyles = makeStyles()(theme => ({
  releaseRepoItem: {
    paddingBottom: theme.spacing(2),
    '&.repo-header': {
      paddingLeft: theme.spacing(6),
      paddingRight: theme.spacing(4),
      fontSize: '0.8rem'
    },
    '&.repo-item, .repo-item': {
      alignItems: 'center',
      display: 'grid',
      gridTemplateColumns: `2fr 1fr 1fr 1fr ${theme.spacing(6)}`,
      gridColumnGap: 20,
      margin: 0
    }
  },
  tagSelect: { marginRight: theme.spacing(2), maxWidth: 350 }
}));

export const ReleaseQuickActions = ({ actionCallbacks }) => {
  const { selection: selectedRows } = useSelector(getReleaseListState);
  const selectedRelease = useSelector(getSelectedRelease);
  const userCapabilities = useSelector(getUserCapabilities);

  const selectedSingleRelease = !isEmpty(selectedRelease) || selectedRows.length === 1;
  const pluralized = pluralize('Releases', selectedRows.length);

  const actions: QuickAction[] = defaultActions
    .filter(action => action.isApplicable({ userCapabilities, selectedSingleRelease, selectedRows }))
    .map(({ action, key, icon, title }) => ({
      key,
      icon,
      title: title(pluralized),
      onClick: () => action({ ...actionCallbacks, selection: selectedRows })
    }));

  return (
    <BaseQuickActions
      actions={actions}
      ariaLabel="release-actions"
      label={selectedSingleRelease ? 'Release actions' : `${selectedRows.length} ${pluralized} selected`}
    />
  );
};

const ReleaseNotes = ({ onChange, release: { notes = '' } }) => (
  <ContentSection title="Release notes">
    <EditableLongText contentFallback="Add release notes here" original={notes} onChange={onChange} placeholder="Release notes" />
  </ContentSection>
);

const ReleaseTags = ({ existingTags = [], release: { tags = [] }, onChange, userCapabilities }) => {
  const [isEditing, setIsEditing] = useState(false);
  const [initialValues] = useState({ tags });
  const { classes } = useStyles();
  const { canManageReleases } = userCapabilities;

  const methods = useForm({ mode: 'onChange', defaultValues: initialValues });
  const { setValue, getValues } = methods;

  useEffect(() => {
    if (!initialValues.tags.length) {
      setValue('tags', tags);
    }
  }, [initialValues.tags, setValue, tags]);

  const onToggleEdit = useCallback(() => {
    setValue('tags', tags);
    setIsEditing(toggle);
  }, [setValue, tags]);

  const onSave = () => onChange(getValues('tags')).then(() => setIsEditing(false));

  return (
    <ContentSection title="Tags" postTitle={!isEditing && canManageReleases && <EditButton onClick={onToggleEdit} />}>
      <div className="flexbox" style={{ alignItems: 'center' }}>
        <FormProvider {...methods}>
          <form noValidate>
            <ChipSelect
              className={classes.tagSelect}
              disabled={!isEditing}
              label=""
              name="tags"
              options={existingTags}
              placeholder={isEditing ? 'Enter release tags' : canManageReleases ? '' : 'No tags yet'}
            />
          </form>
        </FormProvider>
        {isEditing && <ConfirmationButtons onConfirm={onSave} onCancel={onToggleEdit} />}
      </div>
    </ContentSection>
  );
};

const ArtifactsList = ({ artifacts, selectedArtifact, setSelectedArtifact, setShowRemoveArtifactDialog }) => {
  const [sortCol, setSortCol] = useState('modified');
  const [sortDown, setSortDown] = useState(true);
  const [items, setItems] = useState([...artifacts]);
  const { classes } = useStyles();

  useEffect(() => {
    const items = [...artifacts].sort(customSort(sortDown, sortCol));
    setItems(items);
  }, [artifacts, sortCol, sortDown]);

  const onRowSelection = artifact => {
    if (artifact?.id === selectedArtifact?.id) {
      return setSelectedArtifact();
    }
    setSelectedArtifact(artifact);
  };

  const sortColumn = col => {
    if (!col.sortable) {
      return;
    }
    // sort table
    setSortDown(toggle);
    setSortCol(col);
  };

  if (!items.length) {
    return null;
  }

  return (
    <ContentSection title="Artifacts in this Release:">
      <div className={`${classes.releaseRepoItem} repo-item repo-header`}>
        {columns.map(item => (
          <div className="columnHeader" key={item.name} onClick={() => sortColumn(item)}>
            <Tooltip title={item.title} placement="top-start">
              {item.title}
            </Tooltip>
            {item.sortable ? <SortIcon className={`sortIcon ${sortCol === item.name ? 'selected' : ''} ${sortDown.toString()}`} /> : null}
            {item.tooltip}
          </div>
        ))}
        <div />
      </div>
      {items.map((artifact, index) => {
        const expanded = selectedArtifact?.id === artifact.id;
        return (
          <Artifact
            key={`repository-item-${index}`}
            artifact={artifact}
            className={classes.releaseRepoItem}
            columns={columns}
            expanded={expanded}
            index={index}
            onRowSelection={() => onRowSelection(artifact)}
            // this will be run after expansion + collapse and both need some time to fully settle
            // otherwise the measurements are off
            showRemoveArtifactDialog={setShowRemoveArtifactDialog}
          />
        );
      })}
    </ContentSection>
  );
};

export const ReleaseDetails = () => {
  const [showRemoveDialog, setShowRemoveArtifactDialog] = useState(false);
  const [confirmReleaseDeletion, setConfirmReleaseDeletion] = useState(false);
  const [selectedArtifact, setSelectedArtifact] = useState();

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const windowSize = useWindowSize();
  const drawerRef = useRef();
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const release = useSelector(getSelectedRelease);
  const existingTags = useSelector(getReleaseTags);
  const userCapabilities = useSelector(getUserCapabilities);

  const { name: releaseName, artifacts = [] } = release;

  const onRemoveArtifact = artifact => dispatch(removeArtifact(artifact.id)).finally(() => setShowRemoveArtifactDialog(false));

  const copyLinkToClipboard = () => {
    copy(`${window.location.origin}/ui${generateReleasesPath({ pageState: { selectedRelease: releaseName } })}`);
    dispatch(setSnackbar('Link copied to clipboard'));
  };

  const onCloseClick = () => dispatch(selectRelease(null));

  const onCreateDeployment = () =>
    navigate(`${DEPLOYMENT_ROUTES.active.route}?open=true&release=${encodeURIComponent(releaseName)}`, { state: { internal: true } });

  const onToggleReleaseDeletion = () => setConfirmReleaseDeletion(toggle);

  const onDeleteRelease = () => dispatch(removeRelease(releaseName)).then(() => setConfirmReleaseDeletion(false));

  const onReleaseNotesChanged = useCallback(notes => dispatch(updateReleaseInfo({ name: releaseName, info: { notes } })), [dispatch, releaseName]);

  const onTagSelectionChanged = useCallback(tags => dispatch(setReleaseTags({ name: releaseName, tags })).unwrap(), [dispatch, releaseName]);

  return (
    <BaseDrawer
      open={!!releaseName}
      onClose={onCloseClick}
      size="md"
      slotProps={{
        paper: { ref: drawerRef },
        header: {
          title: (
            <>
              Release information for <div className="margin-left-small margin-right-small">{releaseName}</div>
            </>
          ),
          onLinkCopy: copyLinkToClipboard,
          preCloser: (
            <>
              <Typography className="margin-right-x-small" variant="body2">
                Last modified:
              </Typography>
              <RelativeTime updateTime={release.modified} />
            </>
          )
        }
      }}
    >
      <ColumnWidthProvider>
        <TwoColumnData data={{ Name: release.name }} />
        <ReleaseNotes onChange={onReleaseNotesChanged} release={release} />
        <ReleaseTags existingTags={existingTags} onChange={onTagSelectionChanged} release={release} userCapabilities={userCapabilities} />
        <ArtifactsList
          artifacts={artifacts}
          selectedArtifact={selectedArtifact}
          setSelectedArtifact={setSelectedArtifact}
          setShowRemoveArtifactDialog={setShowRemoveArtifactDialog}
        />
      </ColumnWidthProvider>
      <RemoveArtifactDialog
        artifact={selectedArtifact}
        open={!!showRemoveDialog}
        onCancel={() => setShowRemoveArtifactDialog(false)}
        onRemove={() => onRemoveArtifact(selectedArtifact)}
      />
      <RemoveArtifactDialog open={!!confirmReleaseDeletion} onRemove={onDeleteRelease} onCancel={onToggleReleaseDeletion} release={release} />
      <ReleaseQuickActions actionCallbacks={{ onCreateDeployment, onDeleteRelease: onToggleReleaseDeletion }} />
    </BaseDrawer>
  );
};

export default ReleaseDetails;

export const DeleteReleasesConfirmationDialog = ({ onClose, onSubmit }) => (
  <BaseDialog open title="Delete releases?" onClose={onClose}>
    <DialogContent style={{ overflow: 'hidden' }}>All releases artifacts will be deleted. Are you sure you want to delete these releases ?</DialogContent>
    <DialogActions>
      <Button style={{ marginRight: 10 }} onClick={onClose}>
        Cancel
      </Button>
      <Button variant="contained" color="primary" onClick={onSubmit}>
        Delete
      </Button>
    </DialogActions>
  </BaseDialog>
);
