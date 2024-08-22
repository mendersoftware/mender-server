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
import React, { useEffect, useMemo, useState } from 'react';

// material ui
import { Close as CloseIcon, InfoOutlined as InfoOutlinedIcon, WarningAmber as WarningIcon } from '@mui/icons-material';
import {
  Box,
  Button,
  Checkbox,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Divider,
  Drawer,
  FormControl,
  FormHelperText,
  IconButton,
  InputLabel,
  MenuItem,
  Select,
  TextField,
  Tooltip
} from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import validator from 'validator';

import { ALL_DEVICES } from '../../constants/deviceConstants';
import { ALL_RELEASES } from '../../constants/releaseConstants';
import { emptyRole, emptyUiPermissions, itemUiPermissionsReducer, rolesById, uiPermissionsByArea, uiPermissionsById } from '../../constants/userConstants';
import { deepCompare, isEmpty, toggle } from '../../helpers';

const menuProps = {
  anchorOrigin: {
    vertical: 'bottom',
    horizontal: 'left'
  },
  transformOrigin: {
    vertical: 'top',
    horizontal: 'left'
  }
};

const permissionEnabledDisabled = (uiPermission, values, permissionsArea, unscoped) => {
  const { permissionLevel, value: permissionValue, unscopedOnly = {} } = uiPermission;
  const disabled = values.some(permission => uiPermissionsById[permission].permissionLevel > permissionLevel);
  const enabled = values.some(permission => permission === permissionValue) || disabled;
  const skip = unscopedOnly[permissionsArea] && !unscoped;
  return { enabled, disabled, skip };
};

const useStyles = makeStyles()(theme => ({
  buttons: { '&.flexbox.centered': { justifyContent: 'flex-end' } },
  roleDeletion: { marginRight: theme.spacing(2) },
  permissionSelect: { marginLeft: theme.spacing(-1.5) },
  permissionsTitle: { marginBottom: theme.spacing(-1), minHeight: theme.spacing(3) }
}));

const PermissionsSelect = ({ disabled, label, onChange, options, permissionsArea, unscoped, values }) => {
  const { classes } = useStyles();

  const onInputChange = ({ target: { value } }) => {
    if (value.includes('')) {
      return onChange([]);
    }
    return onChange(value);
  };

  const { editablePermissions, selectedValues } = useMemo(
    () =>
      options.reduce(
        (accu, uiPermission) => {
          const { enabled, disabled, skip } = permissionEnabledDisabled(uiPermission, values, permissionsArea, unscoped);
          if (skip) {
            return accu;
          }
          accu.editablePermissions.push({ enabled, disabled, ...uiPermission });
          if (enabled) {
            accu.selectedValues.push(uiPermission.value);
          }
          return accu;
        },
        { editablePermissions: [], selectedValues: [] }
      ),
    [options, permissionsArea, unscoped, values]
  );

  return (
    <FormControl>
      <InputLabel id="permission-selection-label">{label && !values.length ? label : ''}</InputLabel>
      <Select
        labelId="permission-selection-label"
        disabled={disabled}
        displayEmpty={!label}
        fullWidth
        MenuProps={menuProps}
        multiple
        onChange={onInputChange}
        renderValue={() => (selectedValues.length ? selectedValues.map(value => uiPermissionsById[value].title).join(', ') : 'None')}
        value={values}
      >
        {editablePermissions.map(uiPermission => (
          <MenuItem disabled={uiPermission.disabled} key={uiPermission.value} value={uiPermission.value}>
            <Checkbox className={classes.permissionSelect} checked={uiPermission.enabled} disabled={uiPermission.disabled} />
            <div className={uiPermission.disabled ? 'text-muted' : ''}>{uiPermission.title}</div>
          </MenuItem>
        ))}
        <MenuItem value="">None</MenuItem>
      </Select>
    </FormControl>
  );
};

const PermissionsAreaTitle = ({ className = '', explanation, title }) => (
  <div className={`flexbox center-aligned margin-top ${className}`}>
    {title}
    <Tooltip arrow placement="bottom" title={explanation}>
      <InfoOutlinedIcon className="margin-left-small muted" fontSize="small" />
    </Tooltip>
  </div>
);

const PermissionsItem = ({ area, ...remainder }) => (
  <>
    <PermissionsAreaTitle title={area.title} explanation={area.explanation} />
    <PermissionsSelect options={area.uiPermissions} {...remainder} />
  </>
);

const groupsFilter = stateGroups =>
  Object.entries(stateGroups).reduce(
    (accu, [name, groupInfo]) => {
      if (!groupInfo.filters.length) {
        accu.push(name);
      }
      return accu;
    },
    [ALL_DEVICES]
  );

const releasesFilter = stateReleaseTags => [ALL_RELEASES, ...Object.keys(stateReleaseTags)];

const scopedPermissionAreas = {
  groups: {
    key: 'groups',
    filter: groupsFilter,
    placeholder: 'Search groups',
    excessiveAccessConfig: {
      selector: ALL_DEVICES,
      warning: `For 'All devices', users with the Manage permission may also create, edit and delete devices groups.`
    }
  },
  releases: {
    key: 'releases',
    filter: releasesFilter,
    placeholder: 'Search release tags',
    excessiveAccessConfig: {
      selector: ALL_RELEASES,
      warning: `For 'All releases', users with the Manage permission may also upload and delete releases.`
    }
  }
};

const maybeExtendPermissionSelection = (changedGroupSelection, currentGroup, items) => {
  if (items.every(item => changedGroupSelection.some(selectionItem => selectionItem.item === item))) {
    return changedGroupSelection;
  }
  if (changedGroupSelection.every(selection => selection.item && selection.uiPermissions.length)) {
    changedGroupSelection.push(emptyItemSelection);
    return changedGroupSelection;
  }
  // the following is horrible, but I couldn't come up with a better solution that ensures only a single partly defined definition exists
  const filtered = changedGroupSelection.filter(selection => !((!selection.item || !selection.uiPermissions.length) && selection !== currentGroup));
  if (!filtered.some(selection => !selection.item || !selection.uiPermissions.length)) {
    filtered.push(emptyItemSelection);
  }
  return filtered;
};

const ItemSelection = ({
  disableEdit,
  items,
  itemsSelection,
  setter,
  permissionsArea,
  placeholder,
  excessiveAccessConfig: { selector: excessiveAccessSelector, warning: excessiveAccessWarning }
}) => {
  const onItemSelect = (index, { target: { value } }, currentSelection) => {
    let changedSelection = [...currentSelection];
    changedSelection[index] = { ...changedSelection[index], item: value };
    changedSelection = maybeExtendPermissionSelection(changedSelection, changedSelection[index], items);
    setter(changedSelection);
  };

  // find missing items in selected items
  itemsSelection.forEach((selection, key) => {
    if (selection.item && !items.includes(selection.item)) {
      items.push(selection.item);
      itemsSelection[key].notFound = true;
    }
  });

  const onItemPermissionSelect = (index, selectedPermissions, currentSelection) => {
    let changedSelection = [...currentSelection];
    changedSelection[index] = { ...changedSelection[index], uiPermissions: selectedPermissions };
    changedSelection = maybeExtendPermissionSelection(changedSelection, changedSelection[index], items);
    setter(changedSelection);
  };

  const isItemNotFound = item => !!itemsSelection.filter(selection => selection.item === item && selection.notFound === true).length;

  const { title, uiPermissions, explanation } = uiPermissionsByArea[permissionsArea];
  return (
    <>
      <PermissionsAreaTitle className="margin-left-small" explanation={explanation} title={title} />
      {itemsSelection.map((itemSelection, index) => {
        const disabled = disableEdit || itemSelection.disableEdit;
        return (
          <div className="flexbox center-aligned margin-left" key={`${itemSelection.item}-${index}`}>
            <div className="two-columns center-aligned" style={{ maxWidth: 500 }}>
              {disabled ? (
                // empty label as a shortcut to align the layout with the select path
                <TextField disabled defaultValue={itemSelection.item} label=" " />
              ) : (
                <FormControl>
                  <InputLabel id="permissions-group-selection-label">{!itemSelection.item ? placeholder : ''}</InputLabel>
                  <Select labelId="permissions-group-selection-label" onChange={e => onItemSelect(index, e, itemsSelection)} value={itemSelection.item}>
                    {items.map(item => (
                      <MenuItem disabled={isItemNotFound(item)} key={item} value={item}>
                        <Box title={isItemNotFound(item) ? 'This item was removed' : ''} className="flexbox center-aligned">
                          {isItemNotFound(item) && <WarningIcon style={{ marginRight: 4 }} />}
                          {item}
                        </Box>
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
              )}
              <PermissionsSelect
                disabled={disabled}
                label="Select"
                onChange={selectedPermissions => onItemPermissionSelect(index, selectedPermissions, itemsSelection)}
                options={uiPermissions}
                permissionsArea={permissionsArea}
                unscoped={itemSelection.item === excessiveAccessSelector}
                values={itemSelection.uiPermissions}
              />
            </div>
            {itemSelection.item === excessiveAccessSelector && (
              <div className="margin-left text-muted" style={{ alignSelf: 'flex-end' }}>
                {excessiveAccessWarning}
              </div>
            )}
          </div>
        );
      })}
    </>
  );
};

const emptyItemSelection = { item: '', uiPermissions: [], disableEdit: false };

const permissionMapper = uiPermission => uiPermissionsById[uiPermission].value;

const uiPermissionCompare = (existingPermissions, changedPermissions) => deepCompare(existingPermissions, changedPermissions);

const deriveItemsAndPermissions = (stateItems, roleItems, options = {}) => {
  const { adding, disableEdit, filter } = options;
  let filteredStateItems = filter(stateItems);
  let itemSelections = Object.entries(roleItems).map(([group, permissions]) => ({
    ...emptyItemSelection,
    item: group,
    uiPermissions: permissions.map(permissionMapper)
  }));
  if (!adding && !itemSelections.length && filteredStateItems.length !== Object.keys(stateItems).length && !isEmpty(roleItems)) {
    filteredStateItems = Object.keys(roleItems);
    itemSelections = Object.keys(roleItems).map(group => ({
      item: group,
      uiPermissions: [uiPermissionsById.read.value, uiPermissionsById.manage.value],
      disableEdit: true
    }));
  } else if (!disableEdit) {
    itemSelections.push(emptyItemSelection);
  }
  return { filtered: filteredStateItems, selections: itemSelections };
};

const DeleteRoleDialog = ({ dismiss, open, submit, name }) => (
  <Dialog open={open}>
    <DialogTitle>Delete role?</DialogTitle>
    <DialogContent style={{ overflow: 'hidden' }}>
      Are you sure you want to delete the role{' '}
      <b>
        <i>{name}</i>
      </b>
      ?
    </DialogContent>
    <DialogActions>
      <Button style={{ marginRight: 10 }} onClick={dismiss}>
        Cancel
      </Button>
      <Button variant="contained" color="primary" onClick={submit}>
        Delete role
      </Button>
    </DialogActions>
  </Dialog>
);

export const RoleDefinition = ({ adding, editing, stateGroups, stateReleaseTags, onCancel, onSubmit, removeRole, selectedRole = { ...emptyRole } }) => {
  const [description, setDescription] = useState(selectedRole.description);
  const [groups, setGroups] = useState([]);
  const [releases, setReleases] = useState([]);
  const [name, setName] = useState(selectedRole.name);
  const [nameError, setNameError] = useState(false);
  const [auditlogPermissions, setAuditlogPermissions] = useState([]);
  const [groupSelections, setGroupSelections] = useState([]);
  const [releasesPermissions, setReleasesPermissions] = useState([]);
  const [releaseTagSelections, setReleaseTagSelections] = useState([]);
  const [removeDialog, setRemoveDialog] = useState(false);
  const [userManagementPermissions, setUserManagementPermissions] = useState([]);
  const { classes } = useStyles();

  useEffect(() => {
    const { name: roleName = '', description: roleDescription = '' } = selectedRole;
    const { auditlog, groups: roleGroups = {}, releases: roleReleases = {}, userManagement } = { ...emptyUiPermissions, ...selectedRole.uiPermissions };
    const disableEdit = editing && Boolean(rolesById[roleName] || !selectedRole.editable);
    setName(roleName);
    setDescription(roleDescription);
    setUserManagementPermissions(userManagement.map(permissionMapper));
    setAuditlogPermissions(auditlog.map(permissionMapper));
    const { filtered: filteredStateGroups, selections: groupSelections } = deriveItemsAndPermissions(stateGroups, roleGroups, {
      adding,
      disableEdit,
      filter: scopedPermissionAreas.groups.filter
    });
    setGroups(filteredStateGroups);
    setGroupSelections(groupSelections);
    const { filtered: filteredReleases, selections: releaseTagSelections } = deriveItemsAndPermissions(stateReleaseTags, roleReleases, {
      adding,
      disableEdit,
      filter: scopedPermissionAreas.releases.filter
    });
    setReleases(filteredReleases);
    setReleaseTagSelections(releaseTagSelections);
    setReleasesPermissions(
      releaseTagSelections.reduce((accu, { item, uiPermissions }) => {
        if (item === ALL_RELEASES) {
          return [...accu, ...uiPermissions];
        }
        return accu;
      }, [])
    );
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [adding, editing, JSON.stringify(selectedRole), JSON.stringify(stateGroups), JSON.stringify(stateReleaseTags)]);

  const validateNameChange = ({ target: { value } }) => {
    setNameError(!(value && validator.isWhitelisted(value, 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-')));
    setName(value);
  };

  const onSubmitClick = () => {
    const allowUserManagement = userManagementPermissions.includes(uiPermissionsById.manage.value);
    const role = {
      source: selectedRole,
      allowUserManagement,
      description,
      name,
      uiPermissions: {
        auditlog: auditlogPermissions,
        groups: groupSelections,
        releases: releaseTagSelections,
        userManagement: userManagementPermissions
      }
    };
    onSubmit(role);
  };

  const onRemoveRole = () => {
    setRemoveDialog(false);
    removeRole(name);
    onCancel();
  };

  const onToggleRemoveDialog = () => setRemoveDialog(toggle);

  const disableEdit = editing && Boolean(rolesById[selectedRole.id] || !selectedRole.editable);
  const isSubmitDisabled = useMemo(() => {
    const changedPermissions = {
      ...emptyUiPermissions,
      auditlog: auditlogPermissions,
      userManagement: userManagementPermissions,
      groups: groupSelections.reduce(itemUiPermissionsReducer, {}),
      releases: releaseTagSelections.reduce(itemUiPermissionsReducer, {})
    };
    const { hasPartiallyDefinedAreas, hasAreaPermissions } = [...groupSelections, ...releaseTagSelections].reduce(
      (accu, { item, uiPermissions }) => {
        accu.hasPartiallyDefinedAreas = accu.hasPartiallyDefinedAreas || (item && !uiPermissions.length) || (!item && uiPermissions.length);
        accu.hasAreaPermissions = accu.hasAreaPermissions || !!(item && uiPermissions.length);
        return accu;
      },
      { hasPartiallyDefinedAreas: false, hasAreaPermissions: false }
    );
    return Boolean(
      disableEdit ||
        !name ||
        nameError ||
        hasPartiallyDefinedAreas ||
        !(auditlogPermissions.length || hasAreaPermissions || releasesPermissions.length || userManagementPermissions.length) ||
        (Object.entries({ description, name }).every(([key, value]) => selectedRole[key] === value) &&
          uiPermissionCompare(selectedRole.uiPermissions, changedPermissions))
    );
  }, [
    auditlogPermissions,
    userManagementPermissions,
    groupSelections,
    releaseTagSelections,
    releasesPermissions,
    disableEdit,
    name,
    nameError,
    description,
    selectedRole
  ]);

  return (
    <Drawer anchor="right" open={adding || editing} PaperProps={{ style: { minWidth: 600, width: '50vw' } }}>
      <div className="flexbox margin-bottom-small space-between">
        <h3>{adding ? 'Add a' : 'Edit'} role</h3>
        <div className="flexbox center-aligned">
          {editing && !rolesById[selectedRole.id] && (
            <Button
              className={`flexbox center-aligned ${classes.roleDeletion}`}
              color="secondary"
              disabled={!!rolesById[selectedRole.id]}
              onClick={onToggleRemoveDialog}
            >
              delete role
            </Button>
          )}
          <IconButton onClick={onCancel} aria-label="close">
            <CloseIcon />
          </IconButton>
        </div>
      </div>
      <Divider />
      <div className="flexbox column" style={{ width: 500 }}>
        <FormControl>
          <TextField label="Name" id="role-name" value={name} disabled={disableEdit || editing} onChange={validateNameChange} />
          {nameError && <FormHelperText className="warning">Invalid character in role name. Valid characters are a-z, A-Z, 0-9, _ and -</FormHelperText>}
        </FormControl>
        <TextField
          disabled={disableEdit}
          label="Description"
          id="role-description"
          multiline
          placeholder="-"
          onChange={e => setDescription(e.target.value)}
          value={description}
        />
      </div>

      <InputLabel className={`margin-top ${classes.permissionsTitle}`} shrink>
        Permissions
      </InputLabel>
      <div className="two-columns center-aligned margin-left-small" style={{ maxWidth: 500 }}>
        <PermissionsItem
          area={uiPermissionsByArea.userManagement}
          disabled={disableEdit}
          onChange={setUserManagementPermissions}
          values={userManagementPermissions}
        />
        <PermissionsItem disabled={disableEdit} area={uiPermissionsByArea.auditlog} onChange={setAuditlogPermissions} values={auditlogPermissions} />
      </div>
      {(!disableEdit || !!releaseTagSelections.length) && (
        <ItemSelection
          disableEdit={disableEdit}
          excessiveAccessConfig={scopedPermissionAreas.releases.excessiveAccessConfig}
          items={releases}
          itemsSelection={releaseTagSelections}
          permissionsArea={scopedPermissionAreas.releases.key}
          placeholder={scopedPermissionAreas.releases.placeholder}
          setter={setReleaseTagSelections}
        />
      )}
      {(!disableEdit || !!groupSelections.length) && (
        <ItemSelection
          disableEdit={disableEdit}
          excessiveAccessConfig={scopedPermissionAreas.groups.excessiveAccessConfig}
          items={groups}
          itemsSelection={groupSelections}
          permissionsArea={scopedPermissionAreas.groups.key}
          placeholder={scopedPermissionAreas.groups.placeholder}
          setter={setGroupSelections}
        />
      )}
      <Divider className="margin-top-large" light />
      <div className={`flexbox centered margin-top ${classes.buttons}`}>
        <Button className="margin-right" onClick={onCancel}>
          Cancel
        </Button>
        <Button color="secondary" variant="contained" target="_blank" disabled={isSubmitDisabled} onClick={onSubmitClick}>
          Submit
        </Button>
      </div>
      <DeleteRoleDialog dismiss={onToggleRemoveDialog} open={removeDialog} submit={onRemoveRole} name={name} />
    </Drawer>
  );
};

export default RoleDefinition;
