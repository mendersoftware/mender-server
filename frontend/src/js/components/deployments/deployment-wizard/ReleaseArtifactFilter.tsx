// Copyright 2025 Northern.tech AS
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
import { useEffect, useRef, useState } from 'react';
import { FormProvider, useForm, useWatch } from 'react-hook-form';
import { useSelector } from 'react-redux';

import { Close as CloseIcon, ExpandLess as ExpandLessIcon, ExpandMore as ExpandMoreIcon, FilterList as FilterListIcon } from '@mui/icons-material';
import { Button, DialogContent, Divider, TextField, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import ChipSelect from '@northern.tech/common-ui/ChipSelect';
import EnterpriseNotification from '@northern.tech/common-ui/EnterpriseNotification';
import { Link } from '@northern.tech/common-ui/Link';
import { ControlledSearch } from '@northern.tech/common-ui/Search';
import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import { ControlledAutoComplete } from '@northern.tech/common-ui/forms/Autocomplete';
import { BENEFITS } from '@northern.tech/store/constants';
import type { SoftwareKind } from '@northern.tech/store/releasesSlice';
import { getFeatures, getIsEnterprise, getSoftwareById, getSoftwareListState, getSoftwareTags, getUpdateTypes } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { getSoftware } from '@northern.tech/store/thunks';
import type { Software } from '@northern.tech/types/MenderTypes';
import { useDebounce } from '@northern.tech/utils/debouncehook';

import { SoftwareItem } from './ReleaseItem';

const useStyles = makeStyles()(theme => ({
  resultsContainer: {
    paddingBottom: theme.spacing(2),
    overflowY: 'auto',
    flexGrow: 1,
    paddingRight: theme.spacing(1)
  },
  advancedFiltersOpened: {
    background: theme.alpha(theme.palette.info.main, 0.5)
  },
  dialogContainer: {
    height: '75vh'
  }
}));
interface SoftwareArtifactFilterProps {
  kind?: SoftwareKind;
  onClose: () => void;
  onSelect: (item: Software) => void;
  open: boolean;
  selectedSoftware?: string;
}

type SoftwareKindOption = { key: SoftwareKind; title: string };
const softwareKindOptions: Record<SoftwareKind, SoftwareKindOption> = {
  release: { key: 'release', title: 'Release' },
  manifest: { key: 'manifest', title: 'Manifest' }
};

export const SoftwareArtifactFilter = (props: SoftwareArtifactFilterProps) => {
  const { open, onClose, onSelect, selectedSoftware, kind } = props;
  const { classes } = useStyles();
  const [initialValues] = useState<{ kind: SoftwareKind | null; searchTerm: string; tags: string[]; type: string | null }>({
    tags: [],
    type: null,
    searchTerm: '',
    kind: kind ?? null
  });
  const [filterCount, setFilterCount] = useState(0);
  const dispatch = useAppDispatch();
  const [showAdvancedFilters, setShowAdvancedFilters] = useState(false);
  const existingTags = useSelector(getSoftwareTags);
  const updateTypes = useSelector(getUpdateTypes);
  const { softwareIds } = useSelector(getSoftwareListState);
  const softwareById = useSelector(getSoftwareById);
  const isEnterprise = useSelector(getIsEnterprise);
  const { hasManifestsEnabled } = useSelector(getFeatures);
  const methods = useForm({ mode: 'onChange', defaultValues: initialValues });
  const { formState, reset, setFocus } = methods;
  const filterValues = useWatch({ control: methods.control, name: ['tags', 'type', 'searchTerm', 'kind'] });
  const selectedKind = useWatch({ control: methods.control, name: 'kind' });
  const debouncedFilters = useDebounce(filterValues, 500);
  const timer = useRef<ReturnType<typeof setTimeout> | undefined>();
  const softwareItems: Software[] = softwareIds.map(id => softwareById[id]).filter((item): item is Software => Boolean(item));

  const onSelectSoftware = (item: Software) => {
    onSelect(item);
    onClose();
  };

  const onCloseModal = () => {
    setShowAdvancedFilters(false);
    onClose();
  };

  useEffect(() => {
    if (open) {
      timer.current = setTimeout(() => setFocus('searchTerm'), 100);
    }
    return () => clearTimeout(timer.current);
  }, [open, setFocus]);

  useEffect(() => {
    if (!open) {
      return;
    }
    const [tags, type, searchTerm, debouncedKind] = debouncedFilters;
    setFilterCount(tags.length + Number(debouncedKind !== softwareKindOptions.manifest.key && !!type) + Number(!!debouncedKind));
    dispatch(
      getSoftware({
        page: 1,
        perPage: 100,
        searchTerm,
        kind: debouncedKind ?? undefined,
        selectedTags: tags,
        type: debouncedKind === softwareKindOptions.manifest.key ? undefined : (type ?? undefined)
      })
    );
  }, [open, debouncedFilters, dispatch]);
  const { isDirty } = formState;

  return (
    <BaseDialog open={open} title="Select software" onClose={() => onCloseModal()}>
      <DialogContent className={`${classes.dialogContainer} flexbox column`}>
        <Typography variant="body2" className="margin-bottom-small">
          Filter and browse all available software. Use the filters below to narrow down your search.
        </Typography>
        <div className="flexbox space-between">
          <div className="flexbox align-items-center">
            <Typography variant="body1">Filters</Typography>
            <Button
              size="small"
              color="neutral"
              variant="outlined"
              className={`margin-left-small  ${showAdvancedFilters ? classes.advancedFiltersOpened : ''}`}
              startIcon={<FilterListIcon />}
              endIcon={showAdvancedFilters ? <ExpandLessIcon /> : <ExpandMoreIcon />}
              onClick={() => setShowAdvancedFilters(!showAdvancedFilters)}
            >
              Advanced filter {filterCount ? `(${filterCount})` : null}
            </Button>
          </div>
          {isDirty && (
            <Button onClick={() => reset()} startIcon={<CloseIcon />} variant="text" color="inherit" size="small">
              Clear all
            </Button>
          )}
        </div>
        <FormProvider {...methods}>
          <form noValidate>
            {showAdvancedFilters && (
              <div className="two-columns margin-top-small">
                {hasManifestsEnabled && (
                  <div className="flexbox column">
                    <div className="flexbox align-items-center margin-bottom-x-small">
                      <Typography variant="subtitle2">Software type</Typography>
                      <EnterpriseNotification className="margin-left-small" id={BENEFITS.manifests.id} />
                    </div>
                    <ControlledAutoComplete
                      name="kind"
                      disabled={!!kind}
                      options={Object.keys(softwareKindOptions) as SoftwareKind[]}
                      getOptionLabel={(value: SoftwareKind) => softwareKindOptions[value].title}
                      getOptionDisabled={(option: SoftwareKind) => option === softwareKindOptions.manifest.key && !isEnterprise}
                      renderInput={params => <TextField {...params} label="Select type" placeholder="Select type" />}
                    />
                  </div>
                )}
                <div className="flexbox column">
                  <Typography variant="subtitle2" className="margin-bottom-x-small">
                    Tags
                  </Typography>
                  <ChipSelect label="" name="tags" placeholder="Enter or Select tags..." options={existingTags} />
                </div>
                {selectedKind !== softwareKindOptions.manifest.key && (
                  <div className={`flexbox column ${hasManifestsEnabled ? 'margin-top-small' : ''}`}>
                    <Typography variant="subtitle2" className="margin-bottom-x-small">
                      Contains Artifact type
                    </Typography>
                    <ControlledAutoComplete
                      name="type"
                      options={updateTypes}
                      renderInput={params => <TextField {...params} label="Select Artifact type" placeholder="Select type" />}
                    />
                  </div>
                )}
              </div>
            )}
            <div className="flexbox column margin-top-small margin-bottom-small">
              <ControlledSearch asFormField clearButtonOnHover name="searchTerm" placeholder="Search software name..." />
            </div>
          </form>
        </FormProvider>
        <Divider />
        <Typography className="margin-top-small" variant="body1">
          Results ({softwareItems.length})
        </Typography>
        <div className={classes.resultsContainer} id="deployment-release-container">
          {softwareItems.length > 0 ? (
            softwareItems.map(item => (
              <SoftwareItem key={item.name + item.modified} selected={selectedSoftware === item.name} software={item} onClick={onSelectSoftware} />
            ))
          ) : (
            <div className="flexbox column align-items-center margin-top-small">
              <Typography>No results to display.</Typography>
              <Typography className="margin-top-small">
                Try adjusting the filters, or go to{' '}
                <Link to="/software" color="inherit">
                  Software
                </Link>{' '}
                to manage uploads
              </Typography>
            </div>
          )}
        </div>
      </DialogContent>
    </BaseDialog>
  );
};
