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
import { ControlledSearch } from '@northern.tech/common-ui/Search';
import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import { ControlledAutoComplete } from '@northern.tech/common-ui/forms/Autocomplete';
import { Release } from '@northern.tech/store/releasesSlice';
import { getReleaseListState, getReleaseTags, getReleasesById, getUpdateTypes } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { getReleases } from '@northern.tech/store/thunks';
import { useDebounce } from '@northern.tech/utils/debouncehook';

import { ReleaseItem } from './ReleaseItem';

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

export const ReleaseArtifactFilter = props => {
  const { open, onClose, onSelect } = props;
  const { classes } = useStyles();
  const [initialValues] = useState({ tags: [], type: null, searchTerm: '' });
  const [filterCount, setFilterCount] = useState(0);
  const dispatch = useAppDispatch();
  const [showAdvancedFilters, setShowAdvancedFilters] = useState(false);
  const existingTags = useSelector(getReleaseTags);
  const updateTypes = useSelector(getUpdateTypes);
  const { searchedIds: releases } = useSelector(getReleaseListState);
  const releasesById = useSelector(getReleasesById);
  const methods = useForm({ mode: 'onChange', defaultValues: initialValues });
  const { formState, reset, setFocus } = methods;
  const filterValues = useWatch({ control: methods.control, name: ['tags', 'type', 'searchTerm'] });
  const debouncedFilters = useDebounce(filterValues, 500);
  const timer = useRef<ReturnType<typeof setTimeout> | undefined>();

  const releaseItems = releases.map(rel => releasesById[rel]);

  const onSelectRelease = (release: Release) => {
    onSelect(release);
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
    const [tags, type, searchTerm] = debouncedFilters;
    setFilterCount(tags.length + !!type);
    dispatch(getReleases({ page: 1, perPage: 100, searchTerm, searchOnly: true, selectedTags: tags, type }));
  }, [debouncedFilters, dispatch]);
  const { isDirty } = formState;

  return (
    <BaseDialog open={open} title="Select release" onClose={() => onCloseModal()}>
      <DialogContent className={`${classes.dialogContainer} flexbox column`}>
        <Typography variant="body2" className="margin-bottom-small">
          Filter and browse all available releases. Use the filters below to narrow down your search.
        </Typography>
        <div className="flexbox space-between">
          <div className="flexbox center-aligned">
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
              Advanced filter ({filterCount})
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
                <div className="flexbox column">
                  <Typography variant="subtitle2" className="margin-bottom-x-small">
                    Tags
                  </Typography>
                  <ChipSelect label="" name="tags" placeholder="Type or Select tags..." options={existingTags} />
                </div>
                <div className="flexbox column">
                  <Typography variant="subtitle2" className="margin-bottom-x-small">
                    Contains Artifact type
                  </Typography>
                  <ControlledAutoComplete
                    name="type"
                    options={updateTypes}
                    renderInput={params => <TextField {...params} label="Select type..." placeholder="Select a type" />}
                  />
                </div>
              </div>
            )}
            <div className="flexbox column margin-top-small margin-bottom-small">
              <Typography variant="subtitle2" className="margin-bottom-x-small">
                Release name
              </Typography>
              <ControlledSearch asFormField name="searchTerm" placeholder="Search releases..." />
            </div>
          </form>
        </FormProvider>
        <Divider />
        <Typography className="margin-top-small" variant="body1">
          Results ({releaseItems.length})
        </Typography>
        <div className={classes.resultsContainer} id="deployment-release-container">
          {releaseItems.map(item => (
            <ReleaseItem key={item.name + item.modified} release={item} onClick={onSelectRelease} />
          ))}
        </div>
      </DialogContent>
    </BaseDialog>
  );
};
