// Copyright 2024 Northern.tech AS
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
import { FunctionComponent, useCallback } from 'react';
import { Controller, FieldValues, UseFormSetValue, useFieldArray, useFormContext } from 'react-hook-form';

import { InfoOutlined as InfoOutlinedIcon, WarningAmber as WarningIcon } from '@mui/icons-material';
import { FormControl, InputLabel, MenuItem, Select, TextField, Tooltip } from '@mui/material';

import { PermissionsArea, UiPermission, uiPermissionsByArea } from '@northern.tech/store/constants';

import { PermissionsSelect, PermissionsSelectionBaseProps } from './PermissionsSelect';

export type ScopedUiPermissions = {
  item: string;
  uiPermissions: UiPermission[];
};

export type ItemSelectionType = ScopedUiPermissions & {
  disableEdit: boolean;
  notFound: boolean;
};

export type ItemScope = {
  title: string;
  notFound: boolean;
};

export const emptyItemSelection: ItemSelectionType = { item: '', uiPermissions: [], disableEdit: false, notFound: false };

const PermissionsAreaTitle: FunctionComponent<{ className?: string; explanation: string; title: string }> = ({ className = '', explanation, title }) => (
  <div className={`flexbox center-aligned margin-top ${className}`}>
    {title}
    <Tooltip arrow placement="bottom" title={explanation}>
      <InfoOutlinedIcon className="margin-left-small muted" fontSize="small" />
    </Tooltip>
  </div>
);

interface IPermissionsItem extends PermissionsSelectionBaseProps {
  area: PermissionsArea;
}

export const PermissionsItem: FunctionComponent<IPermissionsItem> = ({ area, disabled }) => (
  <div className="two-columns center-aligned margin-left-small" style={{ maxWidth: 500 }}>
    <PermissionsAreaTitle title={area.title} explanation={area.explanation} />
    <PermissionsSelect disabled={disabled} options={area.uiPermissions} permissionsArea={area} />
  </div>
);

const shouldExtendPermissionSelection = (changedSelection, currentItem, items) => {
  if (items.every(({ title }) => changedSelection.some(selectionItem => selectionItem.item === title))) {
    return false;
  }
  if (changedSelection.every(selection => selection.item && selection.uiPermissions.length)) {
    return true;
  }
  // the following is horrible, but I couldn't come up with a better solution that ensures only a single partly defined definition exists
  const filtered = changedSelection.filter(selection => {
    const isDifferentThanCurrent = selection !== currentItem;
    const isPartiallyDefined = selection.item || selection.uiPermissions.length;
    return isPartiallyDefined && isDifferentThanCurrent;
  });
  return filtered.length === 1;
};

interface IScopedPermissionSelect extends PermissionsSelectionBaseProps {
  index: number;
  permissionsArea: PermissionsArea;
  options: ItemScope[];
  itemSelection: ItemSelectionType;
  name: string;
  onChange: (index: number, change: { attribute: string; [change: string]: string }) => void;
}

const ScopeSelect: FunctionComponent<IScopedPermissionSelect> = ({ disabled, permissionsArea, index, options, itemSelection, name = '', onChange }) => {
  const { control } = useFormContext();
  const { key, placeholder } = permissionsArea;
  return disabled ? (
    <TextField disabled defaultValue={itemSelection.item} />
  ) : (
    <FormControl>
      <InputLabel id={`${key}-scope-selection-select-label`}>{!itemSelection.item ? placeholder : ''}</InputLabel>
      <Controller
        name={name || `${key}.${index}.item`}
        control={control}
        render={({ field }) => (
          <Select labelId={`${key}-scope-selection-select-label`} disabled={disabled} {...field} onChange={({ target: { value } }) => onChange(value)}>
            {options.map(option => (
              <MenuItem disabled={option.notFound} key={option.title} value={option.title}>
                <div title={option.notFound ? 'This item was removed' : ''} className="flexbox center-aligned">
                  {option.notFound && <WarningIcon style={{ marginRight: 4 }} />}
                  {option.title}
                </div>
              </MenuItem>
            ))}
          </Select>
        )}
      />
    </FormControl>
  );
};

const ScopedPermissionsItem: FunctionComponent<Omit<IScopedPermissionSelect, 'name'>> = ({
  permissionsArea,
  disabled: disableEdit,
  index,
  itemSelection,
  options,
  onChange
}) => {
  const { excessiveAccessConfig, key } = permissionsArea;
  const { selector: excessiveAccessSelector, warning: excessiveAccessWarning } = excessiveAccessConfig;
  const { uiPermissions } = uiPermissionsByArea[key];
  const { item } = itemSelection;

  const disabled = disableEdit || itemSelection.disableEdit;
  return (
    <div className="flexbox center-aligned margin-left">
      <div className="two-columns center-aligned" style={{ maxWidth: 500 }}>
        <ScopeSelect
          disabled={disabled}
          permissionsArea={permissionsArea}
          index={index}
          options={options}
          itemSelection={itemSelection}
          onChange={item => onChange(index, { item, attribute: 'item' })}
          name={`${key}.${index}.item`}
        />
        <PermissionsSelect
          disabled={disabled}
          name={`${key}.${index}.uiPermissions`}
          label="Select"
          onChange={uiPermissions => onChange(index, { uiPermissions, attribute: 'uiPermissions' })}
          options={uiPermissions}
          permissionsArea={permissionsArea}
          unscoped={item === excessiveAccessSelector}
        />
      </div>
      {item === excessiveAccessSelector && (
        <div className="margin-left text-muted" style={{ alignSelf: 'flex-end' }}>
          {excessiveAccessWarning}
        </div>
      )}
    </div>
  );
};

interface IItemSelection extends PermissionsSelectionBaseProps {
  options: ItemScope[];
  permissionsArea: PermissionsArea;
  setValue: UseFormSetValue<FieldValues>;
}

export const ItemSelection: FunctionComponent<IItemSelection> = ({ disabled, options, permissionsArea, setValue }) => {
  const { control, watch } = useFormContext();
  const { key } = permissionsArea;
  const { title, explanation } = uiPermissionsByArea[key];
  const { fields, append } = useFieldArray({ control, name: permissionsArea.key });
  const watchFieldArray = watch(permissionsArea.key);
  const controlledFields = fields.map((field, index) => ({ ...field, ...watchFieldArray[index] }));

  const onItemPermissionSelectChange = useCallback(
    (index, { attribute, ...change }) => {
      let changedSelection = [...controlledFields];
      changedSelection[index] = { ...changedSelection[index], ...change };
      if (shouldExtendPermissionSelection(changedSelection, changedSelection[index], options)) {
        append(emptyItemSelection);
      }
      setValue(`${key}.${index}.${attribute}`, change[attribute]);
    },
    [append, setValue, key, controlledFields, options]
  );

  return (
    <>
      <PermissionsAreaTitle className="margin-left-small" explanation={explanation} title={title} />
      {controlledFields.map((field, index) => (
        <ScopedPermissionsItem
          key={field.id}
          disabled={disabled}
          permissionsArea={permissionsArea}
          itemSelection={field}
          index={index}
          options={options}
          onChange={onItemPermissionSelectChange}
        />
      ))}
    </>
  );
};
