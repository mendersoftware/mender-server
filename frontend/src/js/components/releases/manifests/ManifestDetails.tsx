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
import type { ReactNode } from 'react';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { FormProvider, useForm } from 'react-hook-form';

import { Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import BaseDrawer from '@northern.tech/common-ui/BaseDrawer';
import ChipSelect from '@northern.tech/common-ui/ChipSelect';
import { ConfirmationButtons, EditButton } from '@northern.tech/common-ui/Confirm';
import { ContentSection } from '@northern.tech/common-ui/ContentSection';
import type { ColumnDefinition } from '@northern.tech/common-ui/DetailsTable';
import DetailsTable from '@northern.tech/common-ui/DetailsTable';
import { EditableLongText } from '@northern.tech/common-ui/EditableLongText';
import { Link } from '@northern.tech/common-ui/Link';
import Pagination from '@northern.tech/common-ui/Pagination';
import { RelativeTime } from '@northern.tech/common-ui/Time';
import { ColumnWidthProvider, TwoColumnData } from '@northern.tech/common-ui/TwoColumnData';
import storeActions from '@northern.tech/store/actions';
import { ATTRIBUTE_SCOPES, DEVICE_FILTERING_OPTIONS, SORTING_OPTIONS } from '@northern.tech/store/constants';
import { formatReleases, generateReleasesPath } from '@northern.tech/store/locationutils';
import { getReleaseTags, getSelectedManifest, getUserCapabilities } from '@northern.tech/store/selectors';
import { useAppDispatch, useAppSelector } from '@northern.tech/store/store';
import { getDevicesByStatus, selectManifest, selectRelease } from '@northern.tech/store/thunks';
import type { Manifest, ManifestComponent } from '@northern.tech/types/MenderTypes';
import { customSort, toggle } from '@northern.tech/utils/helpers';
import copy from 'copy-to-clipboard';
import pluralize from 'pluralize';

import { SignatureSign } from '../utils';
import { ManifestQuickActions } from './ManifestQuickActions';

const { setActiveTab, setSnackbar } = storeActions;

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

const ManifestNotes = ({ notes, onSave }: { notes: string; onSave: (notes: string) => void }) => (
  <ContentSection title="Notes">
    <EditableLongText contentFallback="Add notes here" original={notes} onChange={onSave} placeholder="Notes" />
  </ContentSection>
);

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

interface ComponentTypesTableProps {
  componentTypes: Record<string, ManifestComponent>;
}

type ManifestColumnDefinition = Omit<ColumnDefinition, 'render'> & {
  render: (item: ManifestComponent & { type: string }, { onReleaseClick }: { onReleaseClick: (name: string) => void }) => ReactNode | string;
  sortProp: string;
};

const columns: ManifestColumnDefinition[] = [
  {
    key: 'type',
    title: 'Type',
    sortable: true,
    sortProp: 'type',
    render: ({ type }) => type || '-'
  },
  {
    key: 'release',
    title: 'Release',
    sortable: true,
    sortProp: 'artifact_name',
    render: ({ artifact_name, artifact_path }, { onReleaseClick }) =>
      artifact_name ? <Link onClick={() => onReleaseClick(artifact_name)}>{artifact_name}</Link> : artifact_path || '-'
  },
  {
    key: 'order',
    title: 'Order',
    sortable: true,
    sortProp: 'update_strategy.order',
    cellProps: { align: 'right' },
    render: ({ update_strategy }) => update_strategy?.order ?? '-'
  }
];

const ComponentTypesTable = ({ componentTypes }: ComponentTypesTableProps) => {
  const [sortCol, setSortCol] = useState('');
  const [sortDown, setSortDown] = useState(false);
  const [page, setPage] = useState(1);
  const [perPage, setPerPage] = useState(10);
  const dispatch = useAppDispatch();

  const { items, total } = useMemo(() => {
    const entries = Object.entries(componentTypes).map(([type, content]) => ({ type, ...content }));
    const sorted = sortCol ? [...entries].sort(customSort(sortDown, columns.find(({ key }) => key === sortCol)!.sortProp, true)) : entries;
    const start = (page - 1) * perPage;
    return { items: sorted.slice(start, start + perPage), total: entries.length };
  }, [componentTypes, sortCol, sortDown, page, perPage]);

  const onChangeSorting = key => {
    setSortDown(toggle);
    setSortCol(key);
    setPage(1);
  };

  const onChangePagination = (newPage, newPerPage = perPage) => {
    setPage(newPage);
    setPerPage(newPerPage);
  };

  const onReleaseClick = release => {
    dispatch(setActiveTab('releases'));
    dispatch(selectRelease(release));
    dispatch(selectManifest(null));
  };

  const mappedColumns = columns.map(column => ({ ...column, extras: { onReleaseClick } }));

  if (!total) {
    return (
      <ContentSection title="Component types:">
        <Typography variant="body2" className="margin-top-small">
          No component types defined
        </Typography>
      </ContentSection>
    );
  }

  return (
    <ContentSection title="Component types:">
      <DetailsTable
        columns={mappedColumns}
        items={items}
        onChangeSorting={onChangeSorting}
        sort={{ key: sortCol, direction: sortDown ? SORTING_OPTIONS.desc : SORTING_OPTIONS.asc }}
      />
      <div className="flexbox">
        <Pagination
          className="margin-top-none"
          count={total}
          rowsPerPage={perPage}
          onChangePage={onChangePagination}
          onChangeRowsPerPage={newPerPage => onChangePagination(1, newPerPage)}
          page={page}
        />
      </div>
    </ContentSection>
  );
};

export const ManifestDetails = () => {
  const dispatch = useAppDispatch();
  const manifest = useAppSelector(getSelectedManifest) as Manifest;
  const existingTags = useAppSelector(getReleaseTags);
  const userCapabilities = useAppSelector(getUserCapabilities);

  const { name: manifestName, manifest: manifestContent, notes = '', tags = [] } = manifest;

  const copyLinkToClipboard = () => {
    copy(
      `${window.location.origin}/ui${generateReleasesPath({ pageState: { selectedRelease: manifestName } })}?${formatReleases({ pageState: { tab: 'manifests' } })}`
    );
    dispatch(setSnackbar('Link copied to clipboard'));
  };

  const onCloseClick = () => dispatch(selectManifest(null));

  const onNotesChanged = useCallback((_: string) => dispatch(setSnackbar('Updating manifest notes is not yet supported')), [dispatch]);

  const onTagSelectionChanged = useCallback(
    (_: string[]): Promise<void> => {
      dispatch(setSnackbar('Updating manifest tags is not yet supported'));
      return Promise.resolve();
    },
    [dispatch]
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
      {manifestContent?.component_types && <ComponentTypesTable componentTypes={manifestContent.component_types} />}
      <ManifestQuickActions />
    </BaseDrawer>
  );
};
