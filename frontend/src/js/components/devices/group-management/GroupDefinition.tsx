// Copyright 2020 Northern.tech AS
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
import { useState } from 'react';

import { Autocomplete, FormHelperText, TextField } from '@mui/material';
import { createFilterOptions } from '@mui/material/useAutocomplete';

import DocsLink from '@northern.tech/common-ui/DocsLink';
import InfoText from '@northern.tech/common-ui/InfoText';
import { getGroupNameError } from '@northern.tech/common-ui/forms/validations';

const filter = createFilterOptions();

const GroupOption = (props, option) => <li {...props}>{option.title}</li>;

export const GroupDefinition = ({ isCreationDynamic, groups, newGroup, onInputChange, selectedDevices }) => {
  const [errorText, setErrorText] = useState('');

  const validateName = (name: string) => {
    const isModification = !!name && groups.includes(name);
    const error = name ? getGroupNameError(name, { selectedDevices }) : '';
    setErrorText(error);
    onInputChange(!name || !!error, name, isModification);
  };

  const mappedGroups = groups.map(group => ({ value: group, title: group }));
  return (
    <>
      <Autocomplete
        id="group-creation-selection"
        className="margin-top-x-small"
        autoSelect
        freeSolo
        filterSelectedOptions
        filterOptions={(options, params) => {
          const filtered = filter(options, params);
          if (
            params.inputValue !== '' &&
            !groups.some(group => decodeURIComponent(group) === params.inputValue) &&
            (filtered.length !== 1 || (filtered.length === 1 && filtered[0].title !== params.inputValue))
          ) {
            filtered.push({
              inputValue: params.inputValue,
              title: `Create "${params.inputValue}" group`
            });
          }
          return filtered;
        }}
        getOptionLabel={option => {
          if (typeof option === 'string') {
            return option;
          }
          if (option.inputValue) {
            return option.inputValue;
          }
          return option.title;
        }}
        handleHomeEndKeys
        inputValue={newGroup}
        options={mappedGroups}
        onInputChange={(e, newValue) => validateName(newValue)}
        renderInput={params => <TextField {...params} label="Select a group, or type to create new" />}
        renderOption={GroupOption}
      />
      <FormHelperText>{errorText}</FormHelperText>
      {isCreationDynamic && (
        <InfoText>
          Note: individual devices can&apos;t be added to dynamic groups.
          <br />
          <DocsLink path="overview/device-group" title="Learn more about static vs. dynamic groups" />
        </InfoText>
      )}
    </>
  );
};

export default GroupDefinition;
