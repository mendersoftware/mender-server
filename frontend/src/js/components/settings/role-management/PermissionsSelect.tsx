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
import { FunctionComponent, useMemo } from 'react';
import { Controller, useFormContext } from 'react-hook-form';

import { Checkbox, FormControl, InputLabel, MenuItem, PopoverProps, Select } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { PermissionsArea, UiPermission, uiPermissionsById } from '@northern.tech/store/constants';

const menuProps: Partial<PopoverProps> = {
  anchorOrigin: {
    vertical: 'bottom',
    horizontal: 'left'
  },
  transformOrigin: {
    vertical: 'top',
    horizontal: 'left'
  }
};

const useStyles = makeStyles()(theme => ({
  permissionSelect: { marginLeft: theme.spacing(-1.5) }
}));

const permissionEnabledDisabled = (uiPermission, selectedUiPermissions, permissionsArea, unscoped) => {
  const { permissionLevel, value: permissionValue, unscopedOnly = {} } = uiPermission;
  const disabled = selectedUiPermissions.some(permission => uiPermissionsById[permission].permissionLevel > permissionLevel);
  const enabled = selectedUiPermissions.some(permission => permission === permissionValue) || disabled;
  const skip = unscopedOnly[permissionsArea] && !unscoped;
  return { enabled, disabled, skip };
};

const renderSelectionValues = (options, selectedValues, permissionsArea, unscoped) => {
  if (!selectedValues.length) {
    return 'None';
  }
  return options
    .reduce((accu, uiPermission) => {
      const { enabled } = permissionEnabledDisabled(uiPermission, selectedValues, permissionsArea, unscoped);
      if (enabled) {
        accu.push(uiPermission.title);
      }
      return accu;
    }, [])
    .join(', ');
};

export interface PermissionsSelectionBaseProps {
  disabled: boolean;
}

interface IPermissionsSelect extends PermissionsSelectionBaseProps {
  label?: string;
  onChange?: (string) => void;
  options: UiPermission[];
  name?: string;
  permissionsArea: PermissionsArea;
  unscoped?: boolean;
}

type EditableUiPermission = UiPermission & {
  enabled: boolean;
  disabled: boolean;
};

export const PermissionsSelect: FunctionComponent<IPermissionsSelect> = ({
  disabled,
  label = '',
  onChange,
  options,
  name = '',
  permissionsArea,
  unscoped = false
}) => {
  const { control, getValues } = useFormContext();
  const { classes } = useStyles();
  const selectedUiPermissions = name ? getValues(name) : getValues(permissionsArea.key);

  const onInputChange =
    setter =>
    ({ target: { value } }) => {
      if (value.includes('')) {
        return setter([]);
      }
      if (onChange) {
        return onChange(value);
      }
      return setter(value);
    };

  const editablePermissions = useMemo(
    () =>
      options.reduce<EditableUiPermission[]>((accu, uiPermission) => {
        const { enabled, disabled, skip } = permissionEnabledDisabled(uiPermission, selectedUiPermissions, permissionsArea, unscoped);
        if (skip) {
          return accu;
        }
        accu.push({ enabled, disabled, ...uiPermission });
        return accu;
      }, []),
    [options, permissionsArea, unscoped, selectedUiPermissions]
  );

  return (
    <FormControl>
      <InputLabel id="permission-selection-label">{label && !selectedUiPermissions.length ? label : ''}</InputLabel>
      <Controller
        name={name || permissionsArea.key}
        control={control}
        render={({ field }) => (
          <Select
            labelId="permission-selection-label"
            disabled={disabled}
            displayEmpty={!label}
            fullWidth
            MenuProps={menuProps}
            multiple
            renderValue={selection => renderSelectionValues(options, selection, permissionsArea, unscoped)}
            {...field}
            onChange={onInputChange(field.onChange)}
          >
            {editablePermissions.map(uiPermission => (
              <MenuItem disabled={uiPermission.disabled} key={uiPermission.value} value={uiPermission.value}>
                <Checkbox className={classes.permissionSelect} checked={uiPermission.enabled} disabled={uiPermission.disabled} />
                <div className={uiPermission.disabled ? 'text-muted' : ''}>{uiPermission.title}</div>
              </MenuItem>
            ))}
            <MenuItem value="">None</MenuItem>
          </Select>
        )}
      />
    </FormControl>
  );
};
