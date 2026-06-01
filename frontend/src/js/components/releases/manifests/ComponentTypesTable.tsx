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
import { useMemo, useState } from 'react';

import { ExpandMore as ExpandMoreIcon } from '@mui/icons-material';
import { Button, Typography, formHelperTextClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { ContentSection } from '@northern.tech/common-ui/ContentSection';
import type { ColumnDefinition } from '@northern.tech/common-ui/DetailsTable';
import DetailsTable from '@northern.tech/common-ui/DetailsTable';
import { Link } from '@northern.tech/common-ui/Link';
import Pagination from '@northern.tech/common-ui/Pagination';
import { NumberInput } from '@northern.tech/common-ui/forms/NumberInput';
import storeActions from '@northern.tech/store/actions';
import { SORTING_OPTIONS } from '@northern.tech/store/constants';
import { useAppDispatch } from '@northern.tech/store/store';
import { selectManifest, selectRelease } from '@northern.tech/store/thunks';
import type { ManifestComponent, Software } from '@northern.tech/types/MenderTypes';
import { customSort, toggle } from '@northern.tech/utils/helpers';

import { SoftwareArtifactFilter } from '../../deployments/deployment-wizard/ReleaseArtifactFilter';

const { setActiveTab } = storeActions;

const useStyles = makeStyles()(() => ({
  orderInput: {
    [`& .${formHelperTextClasses.root}`]: {
      alignSelf: 'flex-end',
      width: 'max-content',
      whiteSpace: 'nowrap'
    }
  }
}));

interface ComponentTypesTableProps {
  componentTypes: Record<string, ManifestComponent>;
  existingReleases?: Record<string, boolean>;
  isCreation?: boolean;
  isEditable?: boolean;
  onChange?: (componentTypes: Record<string, ManifestComponent>) => void;
}

type ColumnExtras = {
  classes: Record<string, string>;
  existingReleases?: Record<string, boolean>;
  isCreation?: boolean;
  isEditable: boolean;
  onOrderChange: (type: string, order: number) => void;
  onReleaseClick: (name: string) => void;
  onReleaseEdit: (type: string) => void;
};

type ManifestColumnDefinition = Omit<ColumnDefinition, 'render'> & {
  render: (item: ManifestComponent & { type: string }, extras: ColumnExtras) => ReactNode | string;
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
    render: ({ type, artifact_name, artifact_path }, { isEditable, onReleaseClick, existingReleases, isCreation, onReleaseEdit }) => {
      if (isEditable) {
        return (
          <Button size="large" color="neutral" variant="outlined" endIcon={<ExpandMoreIcon />} onClick={() => onReleaseEdit(type)}>
            <span className={`text-overflow`}>{artifact_name || artifact_path || '-'}</span>
          </Button>
        );
      }
      if (artifact_name) {
        if (existingReleases && existingReleases[artifact_name]) {
          return <Link onClick={() => onReleaseClick(artifact_name)}>{artifact_name}</Link>;
        } else if (isCreation) {
          return (
            <>
              <Typography>{artifact_name}</Typography>
              <Typography color="warning" variant="caption">
                This Release is not available. You need to upload it to Releases before it can be deployed.
              </Typography>
            </>
          );
        }
        return artifact_name;
      }
      return artifact_path || '-';
    }
  },
  {
    key: 'order',
    title: 'Order',
    sortable: true,
    sortProp: 'update_strategy.order',
    cellProps: { align: 'right' },
    render: ({ type, update_strategy }, { isEditable, onOrderChange, classes }) =>
      isEditable ? (
        <NumberInput
          className={classes.orderInput}
          width={72}
          showSteps
          id={type}
          size="small"
          defaultValue={update_strategy?.order ?? null}
          rules={{
            required: 'Order is required',
            min: { value: 1, message: 'Values must be from 1 to 100' },
            max: { value: 100, message: 'Values must be from 1 to 100' }
          }}
          onBlur={value => value && onOrderChange(type, value)}
        />
      ) : (
        (update_strategy?.order ?? '-')
      )
  }
];

export const ComponentTypesTable = ({ componentTypes, existingReleases, isCreation, isEditable = false, onChange }: ComponentTypesTableProps) => {
  const [sortCol, setSortCol] = useState('');
  const [sortDown, setSortDown] = useState(false);
  const [page, setPage] = useState(1);
  const [perPage, setPerPage] = useState(10);
  const [editingType, setEditingType] = useState<string | null>(null);
  const dispatch = useAppDispatch();
  const { classes } = useStyles();

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

  const onOrderChange = (type: string, order: number) => {
    const component = componentTypes[type];
    onChange?.({ ...componentTypes, [type]: { ...component, update_strategy: { ...component.update_strategy, order } } });
  };

  const onReleaseSelect = (item: Software) => {
    if (editingType) {
      const { artifact_path: _artifact_path, ...component } = componentTypes[editingType];
      onChange?.({ ...componentTypes, [editingType]: { ...component, artifact_name: item.name } });
    }
    setEditingType(null);
  };

  const mappedColumns = columns.map(column => ({
    ...column,
    extras: { isEditable, onReleaseClick, onReleaseEdit: setEditingType, onOrderChange, existingReleases, isCreation, classes }
  }));

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
      <SoftwareArtifactFilter
        kind="release"
        open={!!editingType}
        selectedSoftware={editingType ? componentTypes[editingType]?.artifact_name : undefined}
        onClose={() => setEditingType(null)}
        onSelect={onReleaseSelect}
      />
    </ContentSection>
  );
};
