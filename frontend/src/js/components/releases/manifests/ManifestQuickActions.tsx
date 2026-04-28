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
import { type ReactNode, useCallback, useMemo, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';

import {
  FileCopyOutlined as FileCopyOutlinedIcon,
  FileDownload,
  HighlightOffOutlined as HighlightOffOutlinedIcon,
  LabelOutlined as LabelOutlinedIcon,
  Replay as ReplayIcon
} from '@mui/icons-material';
import { ClickAwayListener, SpeedDial, SpeedDialAction, SpeedDialIcon, Typography, alpha, getOverlayAlpha } from '@mui/material';
import { speedDialActionClasses } from '@mui/material/SpeedDialAction';
import { makeStyles } from 'tss-react/mui';

import storeActions from '@northern.tech/store/actions';
import { DEPLOYMENT_ROUTES } from '@northern.tech/store/constants';
import { getManifestsListState, getSelectedManifest, getSelectedManifests, getUserCapabilities } from '@northern.tech/store/selectors';
import { isDarkMode } from '@northern.tech/store/utils';
import type { Manifest } from '@northern.tech/types/MenderTypes';
import { isEmpty, toggle } from '@northern.tech/utils/helpers';
import pluralize from 'pluralize';

const { setSnackbar } = storeActions;

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
    [`& .${speedDialActionClasses.staticTooltipLabel}`]: {
      minWidth: 'max-content'
    }
  },
  fab: { marginBottom: theme.spacing(2), marginRight: theme.spacing(2) },
  label: {
    background: isDarkMode(theme.palette.mode) ? alpha('#fff', getOverlayAlpha(6)) : theme.palette.common.white,
    boxShadow: isDarkMode(theme.palette.mode) ? 'none' : theme.shadows[6],
    padding: `${theme.spacing(1)} ${theme.spacing(2)}`,
    borderRadius: theme.spacing(0.5),
    marginRight: theme.spacing(1),
    marginBottom: theme.spacing(3)
  }
}));

interface ActionCallbacks {
  onCopyManifest: (selection: number[]) => void;
  onCreateDeployment: (selection: number[]) => void;
  onDeleteManifest: (selection: number[]) => void;
  onDownloadManifest: (selection: number[]) => void;
  onTagManifest: (selection: number[]) => void;
}

interface ManifestAction {
  action: (context: ActionCallbacks & { selection: number[] }) => void;
  icon: ReactNode;
  isApplicable: ({
    selectedManifest,
    selectedRows,
    userCapabilities
  }: {
    selectedManifest?: Manifest;
    selectedRows: number[];
    userCapabilities: Record<string, boolean | string[]>;
  }) => boolean;
  key: string;
  title: (pluralized: string) => string;
}

const defaultActions: ManifestAction[] = [
  {
    action: ({ onCreateDeployment, selection }) => onCreateDeployment(selection),
    icon: <ReplayIcon />,
    isApplicable: ({ userCapabilities: { canDeploy }, selectedRows, selectedManifest }) =>
      canDeploy && (!isEmpty(selectedManifest) || selectedRows.length === 1),
    key: 'deploy',
    title: () => 'Create a deployment for this Manifest'
  },
  {
    action: ({ onCopyManifest, selection }) => onCopyManifest(selection),
    icon: <FileCopyOutlinedIcon />,
    isApplicable: ({ selectedRows, selectedManifest }) => !isEmpty(selectedManifest) || selectedRows.length === 1,
    key: 'copy',
    title: () => 'Create a copy from this Manifest'
  },
  {
    action: ({ onDownloadManifest, selection }) => onDownloadManifest(selection),
    icon: <FileDownload />,
    isApplicable: ({ selectedRows, selectedManifest }) => !isEmpty(selectedManifest) || selectedRows.length === 1,
    key: 'download',
    title: () => 'Download Manifest (.mender file)'
  },
  {
    action: ({ onTagManifest, selection }) => onTagManifest(selection),
    icon: <LabelOutlinedIcon />,
    isApplicable: ({ userCapabilities: { canManageReleases }, selectedManifest }) => canManageReleases && isEmpty(selectedManifest),
    key: 'tag',
    title: (pluralized: string) => `Tag ${pluralized}`
  },
  {
    action: ({ onDeleteManifest, selection }) => onDeleteManifest(selection),
    icon: <HighlightOffOutlinedIcon className="red" />,
    isApplicable: ({ userCapabilities: { canManageReleases } }) => canManageReleases,
    key: 'delete',
    title: (pluralized: string) => `Delete ${pluralized}`
  }
];

export const ManifestQuickActions = () => {
  const [showActions, setShowActions] = useState(false);
  const { classes } = useStyles();
  const { selection: selectedRows } = useSelector(getManifestsListState);
  const selectedManifest = useSelector(getSelectedManifest);
  const selectedManifests = useSelector(getSelectedManifests);
  const userCapabilities = useSelector(getUserCapabilities);
  const dispatch = useDispatch();
  const navigate = useNavigate();

  const actions = useMemo(
    () => defaultActions.filter(action => action.isApplicable({ userCapabilities, selectedRows, selectedManifest })),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [JSON.stringify(userCapabilities), selectedRows, selectedManifest]
  );

  const onCreateDeployment = useCallback(
    (selection: number[]) => {
      if (selection.length !== 1) {
        return;
      }
      const { name: manifestName } = selectedManifests[0];
      navigate(`${DEPLOYMENT_ROUTES.active.route}?open=true&release=${encodeURIComponent(manifestName)}`, { state: { internal: true } });
    },
    [navigate, selectedManifests]
  );

  const onCopyManifest = useCallback(() => {
    dispatch(setSnackbar('Creating a copy from a Manifest is not yet supported'));
  }, [dispatch]);

  const onTagManifest = useCallback(() => {
    dispatch(setSnackbar('Tagging Manifests is not yet supported'));
  }, [dispatch]);

  const onDeleteManifest = useCallback(() => {
    dispatch(setSnackbar('Deleting Manifests is not yet supported'));
  }, [dispatch]);

  const onDownloadManifest = useCallback(() => {
    dispatch(setSnackbar('Downloading Manifests is not yet supported'));
  }, [dispatch]);

  const actionCallbacks: ActionCallbacks = { onCreateDeployment, onCopyManifest, onTagManifest, onDeleteManifest, onDownloadManifest };

  const handleShowActions = () => setShowActions(toggle);

  const handleClickAway = () => setShowActions(false);

  const pluralized = pluralize('Manifest', !isEmpty(selectedManifest) ? 1 : selectedRows.length);

  if (!actions.length) {
    return null;
  }
  return (
    <div className={classes.container}>
      {isEmpty(selectedManifest) && (
        <Typography variant="body1" className={classes.label}>
          {`${selectedRows.length} ${pluralized} selected`}
        </Typography>
      )}
      <ClickAwayListener onClickAway={handleClickAway}>
        <SpeedDial className={classes.fab} ariaLabel="manifest-actions" icon={<SpeedDialIcon />} onClick={handleShowActions} open={Boolean(showActions)}>
          {actions.map(action => (
            <SpeedDialAction
              key={action.key}
              aria-label={action.key}
              icon={action.icon}
              slotProps={{ tooltip: { title: action.title(pluralized), open: true } }}
              onClick={() => action.action({ ...actionCallbacks, selection: selectedRows })}
            />
          ))}
        </SpeedDial>
      </ClickAwayListener>
    </div>
  );
};

export default ManifestQuickActions;
