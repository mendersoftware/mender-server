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
import React, { useState } from 'react';

import { Autocomplete, FormHelperText, TextField } from '@mui/material';
import { createFilterOptions } from '@mui/material/useAutocomplete';

import validator from 'validator';

import { UNGROUPED_GROUP } from '../../../constants/deviceConstants';
import { fullyDecodeURI } from '../../../helpers';
import DocsLink from '../../common/docslink';
import InfoText from '../../common/infotext';

const filter = createFilterOptions();

export const validateGroupName = (encodedName, groups = [], selectedDevices = [], isCreationDynamic) => {
  const name = fullyDecodeURI(encodedName);
  let invalid = false;
  let errortext = null;
  const isModification = name.length && groups.some(group => decodeURIComponent(group) === name);
  if (!name && !isModification) {
    invalid = true;
  } else if (!validator.isWhitelisted(name, 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-.')) {
    invalid = true;
    errortext = 'Valid characters are a-z, A-Z, 0-9, ., _ and -';
  } else if (selectedDevices.length && selectedDevices.every(({ group }) => group === name)) {
    invalid = true;
    errortext = `${name} is the same group the selected devices are already in`;
  } else if (isModification && isCreationDynamic) {
    invalid = true;
    errortext = 'A group with the same name already exists';
  } else if (name === UNGROUPED_GROUP.name) {
    invalid = true;
    errortext = `A group with the name ${name} is created automatically`;
  }
  return { errortext, invalid, isModification, name };
};

const GroupOption = (props, option) => <li {...props}>{option.title}</li>;

export const GroupDefinition = ({ isCreationDynamic, groups, newGroup, onInputChange, selectedDevices, selectedGroup }) => {
  const [errortext, setErrorText] = useState('');

  const validateName = encodedName => {
    const { errortext: error, invalid, isModification, name } = validateGroupName(encodedName, groups, selectedDevices, isCreationDynamic);
    setErrorText(error);
    onInputChange(invalid, name, isModification);
  };

  const filteredGroups = groups
    .filter(group => group !== selectedGroup)
    .map(group => ({
      value: group,
      title: group
    }));
  return (
    <>
      <Autocomplete
        id="group-creation-selection"
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
        options={filteredGroups}
        onInputChange={(e, newValue) => validateName(newValue)}
        renderInput={params => <TextField {...params} label="Select a group, or type to create new" InputProps={{ ...params.InputProps }} />}
        renderOption={GroupOption}
      />
      <FormHelperText>{errortext}</FormHelperText>
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
