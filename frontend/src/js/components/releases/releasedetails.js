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
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { FormProvider, useForm } from 'react-hook-form';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';

// material ui
import {
  Close as CloseIcon,
  HighlightOffOutlined as HighlightOffOutlinedIcon,
  LabelOutlined as LabelOutlinedIcon,
  Link as LinkIcon,
  Replay as ReplayIcon,
  Sort as SortIcon
} from '@mui/icons-material';
import {
  Button,
  ClickAwayListener,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Divider,
  Drawer,
  IconButton,
  SpeedDial,
  SpeedDialAction,
  SpeedDialIcon,
  TextField,
  Tooltip
} from '@mui/material';
import { speedDialActionClasses } from '@mui/material/SpeedDialAction';
import { makeStyles } from 'tss-react/mui';

import copy from 'copy-to-clipboard';
import pluralize from 'pluralize';

import { setSnackbar } from '../../actions/appActions';
import { removeArtifact, removeRelease, selectRelease, setReleaseTags, updateReleaseInfo } from '../../actions/releaseActions';
import { DEPLOYMENT_ROUTES } from '../../constants/deploymentConstants';
import { FileSize, customSort, formatTime, toggle } from '../../helpers';
import { getReleaseListState, getReleaseTags, getSelectedRelease, getUserCapabilities } from '../../selectors';
import { generateReleasesPath } from '../../utils/locationutils';
import useWindowSize from '../../utils/resizehook';
import ChipSelect from '../common/chipselect';
import { ConfirmationButtons, EditButton } from '../common/confirm';
import ExpandableAttribute from '../common/expandable-attribute';
import { RelativeTime } from '../common/time';
import { HELPTOOLTIPS, MenderHelpTooltip } from '../helptips/helptooltips';
import Artifact from './artifact';
import RemoveArtifactDialog from './dialogs/removeartifact';

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
    icon: <ReplayIcon />,
    isApplicable: ({ userCapabilities: { canDeploy }, selectedSingleRelease }) => canDeploy && selectedSingleRelease,
    key: 'deploy',
    title: () => 'Create a deployment for this release'
  },
  {
    action: ({ onTagRelease, selectedReleases }) => onTagRelease(selectedReleases),
    icon: <LabelOutlinedIcon />,
    isApplicable: ({ userCapabilities: { canManageReleases } }) => canManageReleases,
    key: 'tag',
    title: pluralized => `Tag ${pluralized}`
  },
  {
    action: ({ onDeleteRelease, selection, selectedReleases }) => onDeleteRelease(selection || selectedReleases),
    icon: <HighlightOffOutlinedIcon className="red" />,
    isApplicable: ({ userCapabilities: { canManageReleases } }) => canManageReleases,
    key: 'delete',
    title: pluralized => `Delete ${pluralized}`
  }
];

const useStyles = makeStyles()(theme => ({
  container: {
    display: 'flex',
    position: 'fixed',
    bottom: theme.spacing(6.5),
    right: theme.spacing(6.5),
    zIndex: 10,
    minWidth: 'max-content',
    alignItems: 'flex-end',
    justifyContent: 'flex-end',
    pointerEvents: 'none',
    [`.${speedDialActionClasses.staticTooltipLabel}`]: {
      minWidth: 'max-content'
    }
  },
  fab: { margin: theme.spacing(2) },
  tagSelect: { marginRight: theme.spacing(2), maxWidth: 350 },
  label: {
    marginRight: theme.spacing(2),
    marginBottom: theme.spacing(4)
  },
  notes: { display: 'block', whiteSpace: 'pre-wrap' },
  notesWrapper: { minWidth: theme.components?.MuiFormControl?.styleOverrides?.root?.minWidth }
}));

export const ReleaseQuickActions = ({ actionCallbacks, innerRef, selectedRelease, userCapabilities, releases }) => {
  const [showActions, setShowActions] = useState(false);
  const [selectedReleases, setSelectedReleases] = useState([]);
  const { classes } = useStyles();
  const { selection: selectedRows } = useSelector(getReleaseListState);

  useEffect(() => {
    if (releases) {
      setSelectedReleases(selectedRows.map(row => releases[row]));
    }
  }, [releases, selectedRows, setSelectedReleases]);

  const actions = useMemo(() => {
    return Object.values(defaultActions).reduce((accu, action) => {
      if (action.isApplicable({ userCapabilities, selectedSingleRelease: !!selectedRelease })) {
        accu.push(action);
      }
      return accu;
    }, []);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(userCapabilities), selectedRelease]);

  const handleShowActions = () => {
    setShowActions(!showActions);
  };

  const handleClickAway = () => {
    setShowActions(false);
  };

  const pluralized = pluralize('releases', selectedRows.length);

  return (
    <div className={classes.container} ref={innerRef}>
      <div className={classes.label}>{selectedRelease ? 'Release actions' : `${selectedRows.length} ${pluralized} selected`}</div>
      <ClickAwayListener onClickAway={handleClickAway}>
        <SpeedDial className={classes.fab} ariaLabel="device-actions" icon={<SpeedDialIcon />} onClick={handleShowActions} open={Boolean(showActions)}>
          {actions.map(action => (
            <SpeedDialAction
              key={action.key}
              aria-label={action.key}
              icon={action.icon}
              tooltipTitle={action.title(pluralized)}
              tooltipOpen
              onClick={() => action.action({ ...actionCallbacks, selection: selectedRelease, selectedReleases })}
            />
          ))}
        </SpeedDial>
      </ClickAwayListener>
    </div>
  );
};

export const EditableLongText = ({ contentFallback = '', fullWidth, original, onChange, placeholder = '-' }) => {
  const [isEditing, setIsEditing] = useState(false);
  const [value, setValue] = useState(original);
  const { classes } = useStyles();

  useEffect(() => {
    setValue(original);
  }, [original]);

  const onCancelClick = () => {
    setValue(original);
    setIsEditing(false);
  };

  const onEdit = ({ target: { value } }) => setValue(value);

  const onEditClick = () => setIsEditing(true);

  const onToggleEditing = useCallback(
    event => {
      event.stopPropagation();
      if (event.key && (event.key !== 'Enter' || event.shiftKey)) {
        return;
      }
      if (isEditing) {
        // save change
        onChange(value);
      }
      setIsEditing(toggle);
    },
    [isEditing, onChange, value]
  );

  const fullWidthClass = fullWidth ? 'full-width' : '';

  return (
    <div className="flexbox" style={{ alignItems: 'end' }}>
      {isEditing ? (
        <>
          <TextField
            className={`margin-right ${fullWidthClass}`}
            multiline
            onChange={onEdit}
            onKeyDown={onToggleEditing}
            placeholder={placeholder}
            value={value}
          />
          <ConfirmationButtons onCancel={onCancelClick} onConfirm={onToggleEditing} />
        </>
      ) : (
        <>
          <ExpandableAttribute
            className={`${fullWidthClass} margin-right ${classes.notesWrapper}`}
            component="div"
            dense
            disableGutters
            primary=""
            secondary={original || value || contentFallback}
            textClasses={{ secondary: classes.notes }}
          />
          <EditButton onClick={onEditClick} />
        </>
      )}
    </div>
  );
};

const ReleaseNotes = ({ onChange, release: { notes = '' } }) => (
  <>
    <h4>Release notes</h4>
    <EditableLongText contentFallback="Add release notes here" original={notes} onChange={onChange} placeholder="Release notes" />
  </>
);

const ReleaseTags = ({ existingTags = [], release: { tags = [] }, onChange }) => {
  const [isEditing, setIsEditing] = useState(false);
  const [initialValues] = useState({ tags });
  const { classes } = useStyles();

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

  const onSave = () => {
    onChange(getValues('tags'));
    setIsEditing(false);
  };

  return (
    <div className="margin-bottom margin-top" style={{ maxWidth: 500 }}>
      <div className="flexbox center-aligned">
        <h4 className="margin-right">Tags</h4>
        {!isEditing && <EditButton onClick={onToggleEdit} />}
      </div>
      <div className="flexbox" style={{ alignItems: 'end' }}>
        <FormProvider {...methods}>
          <form noValidate>
            <ChipSelect
              className={classes.tagSelect}
              disabled={!isEditing}
              label=""
              name="tags"
              options={existingTags}
              placeholder={isEditing ? 'Enter release tags' : 'Click edit to add release tags'}
            />
          </form>
        </FormProvider>
        {isEditing && <ConfirmationButtons onConfirm={onSave} onCancel={onToggleEdit} />}
      </div>
    </div>
  );
};

const ArtifactsList = ({ artifacts, selectedArtifact, setSelectedArtifact, setShowRemoveArtifactDialog }) => {
  const [sortCol, setSortCol] = useState('modified');
  const [sortDown, setSortDown] = useState(true);

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

  if (!artifacts.length) {
    return null;
  }

  const items = artifacts.sort(customSort(sortDown, sortCol));

  return (
    <>
      <h4>Artifacts in this Release:</h4>
      <div>
        <div className="release-repo-item repo-item repo-header">
          {columns.map(item => (
            <div className="columnHeader" key={item.name} onClick={() => sortColumn(item)}>
              <Tooltip title={item.title} placement="top-start">
                <>{item.title}</>
              </Tooltip>
              {item.sortable ? <SortIcon className={`sortIcon ${sortCol === item.name ? 'selected' : ''} ${sortDown.toString()}`} /> : null}
              {item.tooltip}
            </div>
          ))}
          <div style={{ width: 48 }} />
        </div>
        {items.map((artifact, index) => {
          const expanded = selectedArtifact?.id === artifact.id;
          return (
            <Artifact
              key={`repository-item-${index}`}
              artifact={artifact}
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
      </div>
    </>
  );
};

export const ReleaseDetails = () => {
  const [showRemoveDialog, setShowRemoveArtifactDialog] = useState(false);
  const [confirmReleaseDeletion, setConfirmReleaseDeletion] = useState(false);
  const [selectedArtifact, setSelectedArtifact] = useState();

  // eslint-disable-next-line no-unused-vars
  const windowSize = useWindowSize();
  const creationRef = useRef();
  const drawerRef = useRef();
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const release = useSelector(getSelectedRelease);
  const existingTags = useSelector(getReleaseTags);
  const userCapabilities = useSelector(getUserCapabilities);

  const { name: releaseName, artifacts = [] } = release;

  const onRemoveArtifact = artifact => dispatch(removeArtifact(artifact.id)).finally(() => setShowRemoveArtifactDialog(false));

  const copyLinkToClipboard = () => {
    const location = window.location.href.substring(0, window.location.href.indexOf('/releases'));
    copy(`${location}${generateReleasesPath({ pageState: { selectedRelease: releaseName } })}`);
    dispatch(setSnackbar('Link copied to clipboard'));
  };

  const onCloseClick = () => dispatch(selectRelease());

  const onCreateDeployment = () => navigate(`${DEPLOYMENT_ROUTES.active.route}?open=true&release=${encodeURIComponent(releaseName)}`);

  const onToggleReleaseDeletion = () => setConfirmReleaseDeletion(toggle);

  const onDeleteRelease = () => dispatch(removeRelease(releaseName)).then(() => setConfirmReleaseDeletion(false));

  const onReleaseNotesChanged = useCallback(notes => dispatch(updateReleaseInfo(releaseName, { notes })), [dispatch, releaseName]);

  const onTagSelectionChanged = useCallback(tags => dispatch(setReleaseTags(releaseName, tags)), [dispatch, releaseName]);

  return (
    <Drawer anchor="right" open={!!releaseName} onClose={onCloseClick} PaperProps={{ style: { minWidth: '60vw' }, ref: drawerRef }}>
      <div className="flexbox center-aligned space-between">
        <div className="flexbox center-aligned">
          <b>
            Release information for <i>{releaseName}</i>
          </b>
          <IconButton onClick={copyLinkToClipboard} size="large">
            <LinkIcon />
          </IconButton>
        </div>
        <div className="flexbox center-aligned">
          <div className="muted margin-right flexbox">
            <div className="margin-right-small">Last modified:</div>
            <RelativeTime updateTime={release.modified} />
          </div>
          <IconButton onClick={onCloseClick} aria-label="close" size="large">
            <CloseIcon />
          </IconButton>
        </div>
      </div>
      <Divider className="margin-bottom" />
      <ReleaseNotes onChange={onReleaseNotesChanged} release={release} />
      <ReleaseTags existingTags={existingTags} onChange={onTagSelectionChanged} release={release} />
      <ArtifactsList
        artifacts={artifacts}
        selectedArtifact={selectedArtifact}
        setSelectedArtifact={setSelectedArtifact}
        setShowRemoveArtifactDialog={setShowRemoveArtifactDialog}
      />
      <RemoveArtifactDialog
        artifact={selectedArtifact}
        open={!!showRemoveDialog}
        onCancel={() => setShowRemoveArtifactDialog(false)}
        onRemove={() => onRemoveArtifact(selectedArtifact)}
      />
      <RemoveArtifactDialog open={!!confirmReleaseDeletion} onRemove={onDeleteRelease} onCancel={onToggleReleaseDeletion} release={release} />
      <ReleaseQuickActions
        actionCallbacks={{ onCreateDeployment, onDeleteRelease: onToggleReleaseDeletion }}
        innerRef={creationRef}
        selectedRelease={release}
        userCapabilities={userCapabilities}
      />
    </Drawer>
  );
};

export default ReleaseDetails;

export const DeleteReleasesConfirmationDialog = ({ onClose, onSubmit }) => (
  <Dialog open={true}>
    <DialogTitle>Delete releases?</DialogTitle>
    <DialogContent style={{ overflow: 'hidden' }}>All releases artifacts will be deleted. Are you sure you want to delete these releases ?</DialogContent>
    <DialogActions>
      <Button style={{ marginRight: 10 }} onClick={onClose}>
        Cancel
      </Button>
      <Button variant="contained" color="primary" onClick={onSubmit}>
        Delete
      </Button>
    </DialogActions>
  </Dialog>
);
