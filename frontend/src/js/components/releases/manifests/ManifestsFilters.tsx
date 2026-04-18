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

import { Typography } from '@mui/material';

import ChipSelect from '@northern.tech/common-ui/ChipSelect';
import { ControlledSearch } from '@northern.tech/common-ui/Search';
import { Filters } from '@northern.tech/common-ui/forms/Filters';
import { getManifestsListState, getReleaseTags } from '@northern.tech/store/selectors';
import { setManifestsListState } from '@northern.tech/store/thunks';
import pluralize from 'pluralize';

export const ManifestsFilters = ({ classes }: { classes: Record<string, string> }) => {
  const { searchTerm = '', searchTotal, total } = useSelector(getManifestsListState);
  const existingTags = useSelector(getReleaseTags);
  const dispatch = useDispatch();

  const manifestSearchUpdated = useCallback(searchTerm => dispatch(setManifestsListState({ searchTerm })), [dispatch]);
  const onFiltersChange = useCallback(({ name }) => dispatch(setManifestsListState({ searchTerm: name })), [dispatch]);

  return (
    <>
      <Filters
        className={classes.container}
        onChange={onFiltersChange}
        initialValues={{ name: searchTerm, tags: [] }}
        defaultValues={{ name: '', tags: [] }}
        filters={[
          {
            key: 'name',
            title: 'Manifest name',
            Component: ControlledSearch,
            componentProps: {
              onSearch: manifestSearchUpdated,
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
              selection: []
            }
          }
        ]}
      />
      <Typography variant="caption" className={classes.searchNote}>
        {searchTerm && searchTotal !== total ? `Filtered from ${total} ${pluralize('Manifest', total)}` : ''}
      </Typography>
    </>
  );
};

export default ManifestsFilters;
