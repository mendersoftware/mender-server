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
import { FunctionComponent, useEffect, useMemo, useState } from 'react';
import { FieldValues, UseFormSetValue, useFormContext } from 'react-hook-form';

// material ui
import { Close as CloseIcon } from '@mui/icons-material';
import { Button, Divider, Drawer, IconButton, InputLabel } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import Form from '@northern.tech/common-ui/forms/form';
import TextInput from '@northern.tech/common-ui/forms/textinput';
import {
  ALL_DEVICES,
  ALL_RELEASES,
  PermissionsArea,
  UiPermission,
  emptyRole,
  emptyUiPermissions,
  itemUiPermissionsReducer,
  rolesById,
  uiPermissionsByArea,
  uiPermissionsById
} from '@northern.tech/store/constants';
import { deepCompare, toggle } from '@northern.tech/utils/helpers';

import { DeleteRoleDialog } from './DeleteRoleDialog';
import { ItemScope, ItemSelection, ItemSelectionType, PermissionsItem, ScopedUiPermissions, emptyItemSelection } from './PermissionsItems';
import { PermissionsSelectionBaseProps } from './PermissionsSelect';

const useStyles = makeStyles()(theme => ({
  buttons: { '&.flexbox.centered': { justifyContent: 'flex-end' } },
  roleDeletion: { marginRight: theme.spacing(2) },
  permissionSelect: { marginLeft: theme.spacing(-1.5) },
  permissionsTitle: { marginBottom: theme.spacing(-1), minHeight: theme.spacing(3) }
}));

type FormValues = FieldValues & {
  name: string;
  description: string;
  auditlog: UiPermission[];
  groups: ScopedUiPermissions[];
  releases: ScopedUiPermissions[];
  userManagement: UiPermission[];
};

const defaultValues: FormValues = {
  name: '',
  description: '',
  auditlog: [],
  groups: [],
  releases: [],
  userManagement: []
};

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

const scopedPermissionAreas: Record<string, PermissionsArea> = {
  groups: {
    ...uiPermissionsByArea.groups,
    filter: groupsFilter,
    placeholder: 'Search groups',
    excessiveAccessConfig: {
      selector: ALL_DEVICES,
      warning: `For 'All devices', users with the Manage permission may also create, edit and delete devices groups.`
    }
  },
  releases: {
    ...uiPermissionsByArea.releases,
    filter: releasesFilter,
    placeholder: 'Search release tags',
    excessiveAccessConfig: {
      selector: ALL_RELEASES,
      warning: `For 'All releases', users with the Manage permission may also upload and delete releases.`
    }
  }
};

const permissionMapper = uiPermission => uiPermissionsById[uiPermission].value;

const uiPermissionCompare = (existingPermissions, changedPermissions) => deepCompare(existingPermissions, changedPermissions);

type DeriveOptions = { disableEdit: boolean; filter: (itemsById: Record<string, object>) => string[] };

const deriveItemsAndPermissions = (
  stateItems,
  roleItems,
  options
): { stateItems: Record<string, object>; roleItems: Record<string, UiPermission[]>; options?: DeriveOptions } => {
  const { disableEdit, filter } = options;
  let filteredStateItems: ItemScope[] = filter(stateItems).map(item => ({ title: item, notFound: false }));
  let { itemSelections, deletedScopes } = Object.entries(roleItems).reduce<{ itemSelections: ItemSelectionType[]; deletedScopes: ItemScope[] }>(
    (accu, [scope, permissions]) => {
      const notFound = !filteredStateItems.some(({ title }) => title === scope);
      accu.itemSelections.push({
        ...emptyItemSelection,
        item: scope,
        notFound,
        uiPermissions: permissions.map(permissionMapper)
      });
      if (notFound) {
        accu.deletedScopes.push({ title: scope, notFound: true });
      }
      return accu;
    },
    { itemSelections: [], deletedScopes: [] }
  );
  filteredStateItems = [...filteredStateItems, ...deletedScopes];
  if (!disableEdit) {
    itemSelections.push(emptyItemSelection);
  }
  return { filtered: filteredStateItems, selections: itemSelections };
};

interface PermissionSelectionFormVariant extends PermissionsSelectionBaseProps {
  groups: object[];
  releases: object[];
  setValue: UseFormSetValue<FieldValues>;
}

const DefaultPermissionSelection: FunctionComponent<PermissionSelectionFormVariant> = ({ disabled, groups, releases, setValue }) => (
  <>
    <PermissionsItem area={uiPermissionsByArea.userManagement} disabled={disabled} />
    <PermissionsItem area={uiPermissionsByArea.auditlog} disabled={disabled} />
    <ItemSelection disabled={disabled} setValue={setValue} options={releases} permissionsArea={scopedPermissionAreas.releases} />
    <ItemSelection disabled={disabled} setValue={setValue} options={groups} permissionsArea={scopedPermissionAreas.groups} />
  </>
);

export const FormContent = ({ editing, groups: stateGroups, releases: stateReleases, onCancel, selectedRole }) => {
  const { classes } = useStyles();
  const { watch, setValue } = useFormContext();
  const watchedValues = watch();
  const { description, name, auditlog, groups, releases, userManagement } = watchedValues;

  const disableEdit = editing && Boolean(rolesById[selectedRole.id] || !selectedRole.editable);

  const isSubmitDisabled = useMemo(() => {
    const changedPermissions = {
      ...emptyUiPermissions,
      auditlog,
      userManagement,
      groups: groups.reduce(itemUiPermissionsReducer, {}),
      releases: releases.reduce(itemUiPermissionsReducer, {})
    };
    const { hasPartiallyDefinedAreas, hasAreaPermissions } = [...groups, ...releases].reduce(
      (accu, { item, uiPermissions = [] }) => {
        accu.hasPartiallyDefinedAreas = accu.hasPartiallyDefinedAreas || (item && !uiPermissions.length) || (!item && uiPermissions.length);
        accu.hasAreaPermissions = accu.hasAreaPermissions || !!(item && uiPermissions.length);
        return accu;
      },
      { hasPartiallyDefinedAreas: false, hasAreaPermissions: false }
    );
    return Boolean(
      disableEdit ||
        !name ||
        hasPartiallyDefinedAreas ||
        !(auditlog.length || hasAreaPermissions || userManagement.length) ||
        (Object.entries({ description, name }).every(([key, value]) => selectedRole[key] === value) &&
          uiPermissionCompare(selectedRole.uiPermissions, changedPermissions))
    );
  }, [auditlog, description, name, userManagement, groups, releases, disableEdit, selectedRole]);

  return (
    <>
      <div className="flexbox column" style={{ width: 500 }}>
        <TextInput label="Name" id="name" value={name} disabled={disableEdit || editing} validations="isAlphanumericLocator" required />
        <TextInput disabled={disableEdit} label="Description" id="description" InputProps={{ multiline: true }} hint="-" />
      </div>
      <InputLabel className={`margin-top ${classes.permissionsTitle}`} shrink>
        Permissions
      </InputLabel>
      <DefaultPermissionSelection disabled={disableEdit} groups={stateGroups} releases={stateReleases} setValue={setValue} />
      <Divider className="margin-top-large" light />
      <div className={`flexbox centered margin-top ${classes.buttons}`}>
        <Button className="margin-right" onClick={onCancel}>
          Cancel
        </Button>
        <Button color="secondary" variant="contained" type="submit" disabled={isSubmitDisabled}>
          Submit
        </Button>
      </div>
    </>
  );
};

export const RoleDefinition = ({ adding, editing, stateGroups, stateReleaseTags, onCancel, onSubmit, removeRole, selectedRole = { ...emptyRole } }) => {
  const [groups, setGroups] = useState([]);
  const [releases, setReleases] = useState([]);
  const [values, setValues] = useState(defaultValues);
  const [removeDialog, setRemoveDialog] = useState(false);
  const { classes } = useStyles();
  const { name: roleName } = selectedRole;

  useEffect(() => {
    const { name: roleName = '', description: roleDescription = '' } = selectedRole;
    const { auditlog, groups: roleGroups = {}, releases: roleReleases = {}, userManagement } = { ...emptyUiPermissions, ...selectedRole.uiPermissions };
    const disableEdit = editing && Boolean(rolesById[roleName] || !selectedRole.editable);
    const { filtered: filteredStateGroups, selections: groupSelections } = deriveItemsAndPermissions(stateGroups, roleGroups, {
      disableEdit,
      filter: scopedPermissionAreas.groups.filter
    });
    setGroups(filteredStateGroups);
    const { filtered: filteredReleases, selections: releaseTagSelections } = deriveItemsAndPermissions(stateReleaseTags, roleReleases, {
      disableEdit,
      filter: scopedPermissionAreas.releases.filter
    });
    setReleases(filteredReleases);
    setValues({
      name: roleName,
      description: roleDescription,
      auditlog: auditlog.map(permissionMapper),
      userManagement: userManagement.map(permissionMapper),
      groups: groupSelections,
      releases: releaseTagSelections
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [editing, JSON.stringify(selectedRole), JSON.stringify(stateGroups), JSON.stringify(stateReleaseTags)]);

  const onSubmitClick = values => {
    const allowUserManagement = values.userManagement.includes(uiPermissionsById.manage.value);
    const { description, name, auditlog, groups, releases, userManagement } = values;
    const role = {
      source: selectedRole,
      allowUserManagement,
      description,
      name,
      uiPermissions: {
        auditlog,
        groups: groups,
        releases: releases,
        userManagement
      }
    };
    onSubmit(role);
  };

  const onRemoveRole = () => {
    setRemoveDialog(false);
    removeRole(roleName);
    onCancel();
  };

  const onToggleRemoveDialog = () => setRemoveDialog(toggle);

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
      <Form onSubmit={onSubmitClick} showButtons={false} autocomplete="off" defaultValues={defaultValues} initialValues={values}>
        <FormContent editing={editing} groups={groups} releases={releases} onCancel={onCancel} selectedRole={selectedRole} />
      </Form>
      <DeleteRoleDialog dismiss={onToggleRemoveDialog} open={removeDialog} submit={onRemoveRole} name={selectedRole.name} />
    </Drawer>
  );
};

export default RoleDefinition;
