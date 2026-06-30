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

import ChipSelect from '@northern.tech/common-ui/forms/ChipSelect';
import { ControlledSearch } from '@northern.tech/common-ui/Search';
import ClickFilter from '@northern.tech/common-ui/forms/ClickFilter';
import { Filters } from '@northern.tech/common-ui/forms/Filters';
import { getIsEnterprise, getManifestTags, getManifestsListState } from '@northern.tech/store/selectors';
import { setManifestsListState } from '@northern.tech/store/thunks';
import pluralize from 'pluralize';

export const ManifestsFilters = ({ classes }: { classes: Record<string, string> }) => {
  const { selectedTags = [], searchTerm = '', searchTotal, total } = useSelector(getManifestsListState);
  const existingTags = useSelector(getManifestTags);
  const isEnterprise = useSelector(getIsEnterprise);
  const dispatch = useDispatch();

  const manifestSearchUpdated = useCallback(searchTerm => dispatch(setManifestsListState({ searchTerm })), [dispatch]);

  const onFiltersChange = useCallback(({ name, tags }) => dispatch(setManifestsListState({ selectedTags: tags, searchTerm: name })), [dispatch]);

  return (
    <ClickFilter disabled={!isEnterprise}>
      <Filters
        className={classes.container}
        onChange={onFiltersChange}
        initialValues={{ name: searchTerm, tags: selectedTags }}
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
              selection: selectedTags
            }
          }
        ]}
      />
      <Typography variant="caption" className={classes.searchNote}>
        {(searchTerm || selectedTags.length > 0) && searchTotal !== total ? `Filtered from ${total} ${pluralize('Manifest', total)}` : ''}
      </Typography>
    </ClickFilter>
  );
};

export default ManifestsFilters;
