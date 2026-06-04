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
import { type ReactNode, useCallback, useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router';

import {
  FileCopyOutlined as FileCopyOutlinedIcon,
  FileDownload,
  HighlightOffOutlined as HighlightOffOutlinedIcon,
  LabelOutlined as LabelOutlinedIcon,
  SyncOutlined as SyncOutlinedIcon
} from '@mui/icons-material';

import { ConfirmModal } from '@northern.tech/common-ui/ConfirmModal';
import { BaseQuickActions, type QuickAction } from '@northern.tech/common-ui/QuickActions';
import storeActions from '@northern.tech/store/actions';
import { DEPLOYMENT_ROUTES } from '@northern.tech/store/constants';
import { getManifestsListState, getSelectedManifest, getSelectedManifests, getUserCapabilities } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { removeManifests } from '@northern.tech/store/thunks';
import type { Manifest } from '@northern.tech/types/MenderTypes';
import { isEmpty } from '@northern.tech/utils/helpers';
import pluralize from 'pluralize';

const { setSnackbar } = storeActions;

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
    icon: <SyncOutlinedIcon />,
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
  const [confirmManifestDeletion, setConfirmManifestDeletion] = useState(false);
  const { selection: selectedRows } = useSelector(getManifestsListState);
  const selectedManifest = useSelector(getSelectedManifest);
  const selectedManifests = useSelector(getSelectedManifests);
  const userCapabilities = useSelector(getUserCapabilities);
  const dispatch = useAppDispatch();
  const navigate = useNavigate();

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

  const onCopyManifest = useCallback(() => dispatch(setSnackbar('Creating a copy from a Manifest is not yet supported')), [dispatch]);

  const onTagManifest = useCallback(() => dispatch(setSnackbar('Tagging Manifests is not yet supported')), [dispatch]);

  const onDeleteManifest = () => setConfirmManifestDeletion(true);

  const onCancelDeletion = () => setConfirmManifestDeletion(false);

  const onConfirmDeletion = useCallback(() => {
    const names = !isEmpty(selectedManifest) ? [selectedManifest.name] : selectedManifests.map(({ name }) => name);
    dispatch(removeManifests(names));
    setConfirmManifestDeletion(false);
  }, [dispatch, selectedManifest, selectedManifests]);

  const onDownloadManifest = useCallback(() => {
    dispatch(setSnackbar('Downloading Manifests is not yet supported'));
  }, [dispatch]);

  const actionCallbacks: ActionCallbacks = { onCreateDeployment, onCopyManifest, onTagManifest, onDeleteManifest, onDownloadManifest };

  const selectedSingleManifest = !isEmpty(selectedManifest) || selectedRows.length === 1;
  const pluralized = pluralize('Manifest', selectedSingleManifest ? 1 : selectedRows.length);

  const actions: QuickAction[] = defaultActions
    .filter(action => action.isApplicable({ userCapabilities, selectedRows, selectedManifest }))
    .map(({ action, key, icon, title }) => ({
      key,
      icon,
      title: title(pluralized),
      onClick: () => action({ ...actionCallbacks, selection: selectedRows })
    }));

  return (
    <>
      <BaseQuickActions
        actions={actions}
        ariaLabel="manifest-actions"
        label={selectedSingleManifest ? 'Manifest actions' : `${selectedRows.length} ${pluralized} selected`}
      />
      <ConfirmModal
        header={`Remove${selectedSingleManifest ? '' : ` ${selectedRows.length}`} ${pluralized}?`}
        description={`Are you sure you want to remove the ${selectedSingleManifest ? '' : `${selectedRows.length} `}selected ${pluralized}?`}
        open={confirmManifestDeletion}
        confirmButtonText="Remove"
        close={onCancelDeletion}
        onConfirm={onConfirmDeletion}
      />
    </>
  );
};

export default ManifestQuickActions;
