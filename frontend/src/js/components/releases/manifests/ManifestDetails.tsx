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
import { useCallback, useEffect, useState } from 'react';
import { FormProvider, useForm } from 'react-hook-form';

import { Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import BaseDrawer from '@northern.tech/common-ui/BaseDrawer';
import ChipSelect from '@northern.tech/common-ui/ChipSelect';
import { ConfirmationButtons, EditButton } from '@northern.tech/common-ui/Confirm';
import { ContentSection } from '@northern.tech/common-ui/ContentSection';
import { EditableLongText } from '@northern.tech/common-ui/EditableLongText';
import { RelativeTime } from '@northern.tech/common-ui/Time';
import { ColumnWidthProvider, TwoColumnData } from '@northern.tech/common-ui/TwoColumnData';
import storeActions from '@northern.tech/store/actions';
import { ATTRIBUTE_SCOPES, DEVICE_FILTERING_OPTIONS } from '@northern.tech/store/constants';
import { formatReleases, generateReleasesPath } from '@northern.tech/store/locationutils';
import { getManifestTags, getSelectedManifest, getUserCapabilities } from '@northern.tech/store/selectors';
import { useAppDispatch, useAppSelector } from '@northern.tech/store/store';
import { checkReleasesExistence, getDevicesByStatus, selectManifest, updateManifestInfo } from '@northern.tech/store/thunks';
import type { Manifest } from '@northern.tech/types/MenderTypes';
import { toggle } from '@northern.tech/utils/helpers';
import copy from 'copy-to-clipboard';
import pluralize from 'pluralize';

import { SignatureSign } from '../utils';
import { ComponentTypesTable } from './ComponentTypesTable';
import { ManifestQuickActions } from './ManifestQuickActions';

const { setSnackbar } = storeActions;

const useStyles = makeStyles()(theme => ({
  tagSelect: { marginRight: theme.spacing(2), maxWidth: 350 }
}));

const ManifestInfo = ({ manifest }: { manifest: Manifest }) => {
  const [installCount, setInstallCount] = useState(0);
  const dispatch = useAppDispatch();

  const { artifact, manifest: manifestContent } = manifest;
  const { device_types_compatible, id = '-', signed } = artifact ?? {};
  const compatibleTypes = manifestContent?.system_types_compatible?.join(', ') || device_types_compatible?.join(', ') || '-';

  useEffect(() => {
    if (!manifest.name) {
      return;
    }
    dispatch(
      getDevicesByStatus({
        fetchAuth: false,
        page: 1,
        perPage: 1,
        filterSelection: [
          {
            key: 'rootfs-image.update-module.mender-orchestrator-manifest.version',
            operator: DEVICE_FILTERING_OPTIONS.$eq.key,
            scope: ATTRIBUTE_SCOPES.inventory,
            value: manifest.name
          }
        ]
      })
    )
      .unwrap()
      .then(({ total }) => setInstallCount(total));
  }, [dispatch, manifest.name]);

  return (
    <TwoColumnData
      data={{
        ID: id || '-',
        Name: manifest.name,
        'Compatible types': compatibleTypes,
        'Installed on': `${installCount} ${pluralize('device', installCount)}`,
        Signature: <SignatureSign isSigned={!!signed} />
      }}
    />
  );
};

const ManifestNotes = ({ notes, onSave }: { notes: string; onSave: (notes: string) => Promise<void> }) => {
  const [isEditing, setIsEditing] = useState(false);
  return (
    <ContentSection title="Notes" postTitle={!isEditing && <EditButton onClick={() => setIsEditing(true)} />}>
      <EditableLongText original={notes} onChange={onSave} placeholder="Notes" isEditing={isEditing} onEditToggle={setIsEditing} />
    </ContentSection>
  );
};

interface ManifestTagsProps {
  canManageReleases: boolean;
  existingTags: string[];
  onSave: (tags: string[]) => Promise<void>;
  tags: string[];
}

const ManifestTags = ({ existingTags, tags, canManageReleases, onSave }: ManifestTagsProps) => {
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

  const onSubmit = () => onSave(getValues('tags')).then(() => setIsEditing(false));

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
              placeholder={isEditing ? 'Enter manifest tags' : canManageReleases ? '' : 'No tags yet'}
            />
          </form>
        </FormProvider>
        {isEditing && <ConfirmationButtons onConfirm={onSubmit} onCancel={onToggleEdit} />}
      </div>
    </ContentSection>
  );
};

export const ManifestDetails = ({ onCopy }: { onCopy?: (name: string) => void }) => {
  const dispatch = useAppDispatch();
  const manifest = useAppSelector(getSelectedManifest) as Manifest;
  const existingTags = useAppSelector(getManifestTags);
  const userCapabilities = useAppSelector(getUserCapabilities);

  const { name: manifestName, manifest: manifestContent, notes = '', tags = [] } = manifest;
  const [existingReleases, setExistingReleases] = useState<Record<string, boolean>>({});

  useEffect(() => {
    if (!manifestContent?.component_types) {
      return;
    }
    const artifactNames = Object.values(manifestContent.component_types)
      .map(c => c.artifact_name)
      .filter((name): name is string => !!name);
    const uniqueNames = [...new Set(artifactNames)];
    dispatch(checkReleasesExistence(uniqueNames)).unwrap().then(setExistingReleases);
  }, [dispatch, manifestContent?.component_types]);

  const copyLinkToClipboard = () => {
    copy(
      `${window.location.origin}/ui${generateReleasesPath({ pageState: { selectedRelease: manifestName } })}?${formatReleases({ pageState: { tab: 'manifests' } })}`
    );
    dispatch(setSnackbar('Link copied to clipboard'));
  };

  const onCloseClick = () => dispatch(selectManifest(null));

  const onNotesChanged = useCallback(
    (notes: string) => dispatch(updateManifestInfo({ name: manifestName, info: { notes } })).unwrap(),
    [dispatch, manifestName]
  );

  const onTagSelectionChanged = useCallback(
    (tags: string[]) => dispatch(updateManifestInfo({ name: manifestName, info: { tags } })).unwrap(),
    [dispatch, manifestName]
  );

  return (
    <BaseDrawer
      open={!!manifestName}
      onClose={onCloseClick}
      size="md"
      slotProps={{
        header: {
          title: (
            <>
              Manifest information for <span className="margin-left-small">{manifestName}</span>
            </>
          ),
          onLinkCopy: copyLinkToClipboard,
          preCloser: (
            <>
              <Typography className="margin-right-x-small" variant="body2">
                Last modified:
              </Typography>
              <RelativeTime updateTime={manifest.modified} />
            </>
          )
        }
      }}
    >
      <ColumnWidthProvider>
        <ManifestInfo manifest={manifest} />
        <ManifestNotes notes={notes} onSave={onNotesChanged} />
        <ManifestTags existingTags={existingTags} tags={tags} canManageReleases={userCapabilities.canManageReleases} onSave={onTagSelectionChanged} />
      </ColumnWidthProvider>
      {manifestContent?.component_types && <ComponentTypesTable componentTypes={manifestContent.component_types} existingReleases={existingReleases} />}
      <ManifestQuickActions onCopy={onCopy} />
    </BaseDrawer>
  );
};
