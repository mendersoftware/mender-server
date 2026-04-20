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
import { useCallback } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { TextField, Typography } from '@mui/material';

import ChipSelect from '@northern.tech/common-ui/ChipSelect';
import { ControlledSearch } from '@northern.tech/common-ui/Search';
import { ControlledAutoComplete } from '@northern.tech/common-ui/forms/Autocomplete';
import { Filters } from '@northern.tech/common-ui/forms/Filters';
import { getReleaseListState, getReleaseTags, getUpdateTypes as getUpdateTypesSelector } from '@northern.tech/store/selectors';
import { setReleasesListState } from '@northern.tech/store/thunks';
import pluralize from 'pluralize';

export const ReleasesFilters = ({ classes }: { classes: Record<string, string> }) => {
  const { selectedTags = [], searchTerm = '', searchTotal, total, type } = useSelector(getReleaseListState);
  const existingTags = useSelector(getReleaseTags);
  const updateTypes = useSelector(getUpdateTypesSelector);
  const dispatch = useDispatch();

  const searchUpdated = useCallback(searchTerm => dispatch(setReleasesListState({ searchTerm })), [dispatch]);

  const onFiltersChange = useCallback(({ name, tags, type }) => dispatch(setReleasesListState({ selectedTags: tags, searchTerm: name, type })), [dispatch]);

  return (
    <>
      <Filters
        className={classes.container}
        onChange={onFiltersChange}
        initialValues={{ name: searchTerm, tags: selectedTags, type }}
        defaultValues={{ name: '', tags: [], type: '' }}
        filters={[
          {
            key: 'name',
            title: 'Release name',
            Component: ControlledSearch,
            componentProps: {
              onSearch: searchUpdated,
              placeholder: 'Starts with'
            }
          },
          {
            key: 'tags',
            title: 'Tags',
            Component: ChipSelect,
            componentProps: {
              options: existingTags,
              placeholder: 'Select tags',
              selection: selectedTags
            }
          },
          {
            key: 'type',
            title: 'Contains Artifact type',
            Component: ControlledAutoComplete,
            componentProps: {
              autoHighlight: true,
              filterSelectedOptions: true,
              freeSolo: true,
              handleHomeEndKeys: true,
              options: updateTypes,
              renderInput: params => <TextField {...params} placeholder="Any" />
            }
          }
        ]}
      />
      <Typography variant="caption" className={classes.searchNote}>
        {searchTerm && searchTotal !== total ? `Filtered from ${total} ${pluralize('Release', total)}` : ''}
      </Typography>
    </>
  );
};

export default ReleasesFilters;
