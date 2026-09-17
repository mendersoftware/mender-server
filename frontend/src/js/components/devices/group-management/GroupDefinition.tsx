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
import { useFormContext } from 'react-hook-form';

import { TextField, Typography } from '@mui/material';
import { createFilterOptions } from '@mui/material/useAutocomplete';

import DocsLink from '@northern.tech/common-ui/DocsLink';
import InfoText from '@northern.tech/common-ui/InfoText';
import { ControlledAutoComplete } from '@northern.tech/common-ui/forms/Autocomplete';
import { runValidations } from '@northern.tech/common-ui/forms/validations';
import { UNGROUPED_GROUP } from '@northern.tech/store/constants';

const filter = createFilterOptions();

const nameValidations = `isAlphanumericLocator,isLength:1:256,isNot:${UNGROUPED_GROUP.name}`;

type GroupNameValidationOptions = {
  existingGroups?: string[];
  isDynamic?: boolean;
  selectedDevices?: { group?: string }[];
};

export const getGroupNameError = (name: string, { existingGroups = [], isDynamic = false, selectedDevices = [] }: GroupNameValidationOptions = {}) => {
  const { isValid, errortext } = runValidations({ id: 'groupName', required: true, validations: nameValidations, value: name });
  if (!isValid) {
    return errortext;
  }
  // devices can be added to an existing static group, but a dynamic group needs a name of its own
  if (isDynamic && existingGroups.includes(name)) {
    return 'A group with the same name already exists';
  }
  if (selectedDevices.length && selectedDevices.every(({ group }) => group === name)) {
    return `${name} is the same group the selected devices are already in`;
  }
  return '';
};

export const groupNameValidationRules = (options: GroupNameValidationOptions = {}) => ({
  required: 'Group name is required',
  validate: (value: string) => getGroupNameError(value, options) || true
});

const GroupOption = ({ key, ...optionProps }, option) => (
  <li key={key} {...optionProps}>
    {option.title}
  </li>
);

export const GroupDefinition = ({ groups, name, selectedDevices }) => {
  const {
    formState: { errors }
  } = useFormContext();
  const error = errors[name]?.message;

  const mappedGroups = groups.map(group => ({ value: group, title: group }));
  return (
    <>
      <Typography variant="subtitle1">Name</Typography>
      <ControlledAutoComplete
        id="group-creation-selection"
        className="margin-top-x-small"
        autoSelect
        freeSolo
        filterSelectedOptions
        filterOptions={(options, params) => {
          const filtered = filter(options, params);
          if (params.inputValue !== '' && !groups.includes(params.inputValue) && (filtered.length !== 1 || filtered[0].title !== params.inputValue)) {
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
        name={name}
        options={mappedGroups}
        renderInput={params => (
          <TextField
            {...params}
            error={!!error}
            slotProps={{ ...params.slotProps, inputLabel: { ...params.slotProps.inputLabel, shrink: true } }}
            placeholder="Select a group, or type to create new"
            helperText={error}
          />
        )}
        renderOption={GroupOption}
        rules={groupNameValidationRules({ selectedDevices })}
      />
      <InfoText>
        Note: individual devices can&apos;t be added to dynamic groups.
        <br />
        <DocsLink path="overview/device-group" title="Learn more about static vs. dynamic groups" />
      </InfoText>
    </>
  );
};

export default GroupDefinition;
