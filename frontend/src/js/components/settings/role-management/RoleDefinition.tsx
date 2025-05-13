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
import { Button, Divider, Drawer, InputLabel } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { ConfirmModal } from '@northern.tech/common-ui/ConfirmModal';
import { DrawerTitle } from '@northern.tech/common-ui/DrawerTitle';
import Form from '@northern.tech/common-ui/forms/Form';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import {
  ALL_DEVICES,
  ALL_RELEASES,
  PermissionsArea,
  UiPermission,
  UiRoleDefinition,
  emptyRole,
  emptyUiPermissions,
  itemUiPermissionsReducer,
  rolesById,
  uiPermissionsByArea,
  uiPermissionsById
} from '@northern.tech/store/constants';
import { deepCompare, toggle } from '@northern.tech/utils/helpers';
import { AsyncThunkAction } from '@reduxjs/toolkit';

import { ItemScope, ItemSelection, ItemSelectionType, PermissionsItem, ScopedUiPermissions, emptyItemSelection } from './PermissionsItems';
import { PermissionsSelectionBaseProps } from './PermissionsSelect';

const useStyles = makeStyles()(theme => ({
  buttons: { '&.flexbox.centered': { justifyContent: 'flex-end' } },
  roleDeletion: { marginRight: theme.spacing(2) },
  formWrapper: { display: 'flex', flexDirection: 'column', gap: theme.spacing(2), paddingTop: theme.spacing(4) },
  permissionSelect: { marginLeft: theme.spacing(-1.5) },
  permissionsTitle: { marginBottom: theme.spacing(-1), minHeight: theme.spacing(3) }
}));

type FormValues = FieldValues & {
  auditlog: UiPermission[];
  description: string;
  groups: ScopedUiPermissions[];
  name: string;
  releases: ScopedUiPermissions[];
  tenantManagement: UiPermission[];
  userManagement: UiPermission[];
};

const defaultValues: FormValues = {
  name: '',
  description: '',
  auditlog: [],
  groups: [],
  releases: [],
  tenantManagement: [],
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
): { options?: DeriveOptions; roleItems: Record<string, UiPermission[]>; stateItems: Record<string, object> } => {
  const { disableEdit, filter } = options;
  let filteredStateItems: ItemScope[] = filter(stateItems).map(item => ({ title: item, notFound: false }));
  let { itemSelections, deletedScopes } = Object.entries(roleItems).reduce<{ deletedScopes: ItemScope[]; itemSelections: ItemSelectionType[] }>(
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

const ServiceProviderPermissionSelection: FunctionComponent<PermissionsSelectionBaseProps> = ({ disabled }) => (
  <>
    <PermissionsItem area={uiPermissionsByArea.userManagement} disabled={disabled} />
    <PermissionsItem area={uiPermissionsByArea.tenantManagement} disabled={disabled} />
    <PermissionsItem area={uiPermissionsByArea.auditlog} disabled={disabled} />
  </>
);

interface RoleDefinitionFormProps {
  editing: boolean;
  groups: ItemScope[];
  isServiceProvider: boolean;
  onCancel: () => void;
  releases: ItemScope[];
  selectedRole: UiRoleDefinition;
}

export const FormContent: FunctionComponent<RoleDefinitionFormProps> = ({
  editing,
  groups: stateGroups,
  isServiceProvider,
  releases: stateReleases,
  onCancel,
  selectedRole
}) => {
  const { classes } = useStyles();
  const { watch, setValue } = useFormContext();
  const watchedValues = watch();
  const { description, name, auditlog, groups, releases, tenantManagement, userManagement } = watchedValues;

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
        !(auditlog.length || hasAreaPermissions || userManagement.length || tenantManagement.length) ||
        (Object.entries({ description, name }).every(([key, value]) => selectedRole[key] === value) &&
          uiPermissionCompare(selectedRole.uiPermissions, changedPermissions))
    );
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [auditlog, description, name, userManagement, JSON.stringify(groups), JSON.stringify(releases), tenantManagement, disableEdit, selectedRole]);

  return (
    <>
      <TextInput
        className="margin-top-none"
        disabled={disableEdit || editing}
        id="name"
        label="Name"
        required
        validations="isAlphanumericLocator"
        value={name}
      />
      <TextInput className="margin-top-none" disabled={disableEdit} label="Description" id="description" InputProps={{ multiline: true }} hint="-" />
      <InputLabel className={`margin-top ${classes.permissionsTitle}`} shrink>
        Permissions
      </InputLabel>
      {isServiceProvider ? (
        <ServiceProviderPermissionSelection disabled={disableEdit} />
      ) : (
        <DefaultPermissionSelection disabled={disableEdit} groups={stateGroups} releases={stateReleases} setValue={setValue} />
      )}
      <Divider className="margin-top-large" />
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

interface RoleDefinitionProps {
  adding: boolean;
  editing: boolean;
  isServiceProvider: boolean;
  onCancel: () => void;
  onSubmit: (role: UiRoleDefinition) => void;
  removeRole: () => AsyncThunkAction<void, string, object>;
  selectedRole: UiRoleDefinition;
  stateGroups: Record<string, object>;
  stateReleaseTags: Record<string, object>;
}

export const RoleDefinition: FunctionComponent<RoleDefinitionProps> = ({
  adding,
  editing,
  isServiceProvider,
  stateGroups,
  stateReleaseTags,
  onCancel,
  onSubmit,
  removeRole,
  selectedRole = { ...emptyRole }
}) => {
  const [groups, setGroups] = useState([]);
  const [releases, setReleases] = useState([]);
  const [values, setValues] = useState(defaultValues);
  const [removeDialog, setRemoveDialog] = useState(false);
  const { classes } = useStyles();
  const { name: roleName } = selectedRole;

  useEffect(() => {
    const { name: roleName = '', description: roleDescription = '' } = selectedRole;
    const {
      auditlog,
      groups: roleGroups = {},
      releases: roleReleases = {},
      tenantManagement,
      userManagement
    } = { ...emptyUiPermissions, ...selectedRole.uiPermissions };
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
      tenantManagement: tenantManagement.map(permissionMapper),
      userManagement: userManagement.map(permissionMapper),
      groups: groupSelections,
      releases: releaseTagSelections
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [editing, JSON.stringify(selectedRole), JSON.stringify(stateGroups), JSON.stringify(stateReleaseTags)]);

  const onSubmitClick = values => {
    const allowUserManagement = values.userManagement.includes(uiPermissionsById.manage.value);
    const { description, name, auditlog, groups, releases, tenantManagement, userManagement } = values;
    const role = {
      source: selectedRole,
      allowUserManagement,
      description,
      name,
      uiPermissions: {
        auditlog,
        groups: groups,
        releases: releases,
        tenantManagement,
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
    <Drawer anchor="right" open={adding || editing} onClose={onCancel} PaperProps={{ style: { minWidth: 600, width: '50vw' } }}>
      <DrawerTitle
        title={`${adding ? 'Add a' : 'Edit'} role`}
        onClose={onCancel}
        preCloser={
          editing &&
          !rolesById[selectedRole.value] && (
            <Button
              className={`flexbox center-aligned ${classes.roleDeletion}`}
              color="secondary"
              disabled={!!rolesById[selectedRole.value]}
              onClick={onToggleRemoveDialog}
            >
              delete role
            </Button>
          )
        }
      />
      <Divider />
      <Form
        className={classes.formWrapper}
        onSubmit={onSubmitClick}
        showButtons={false}
        autocomplete="off"
        defaultValues={defaultValues}
        initialValues={values}
      >
        <FormContent
          editing={editing}
          groups={groups}
          releases={releases}
          isServiceProvider={isServiceProvider}
          onCancel={onCancel}
          selectedRole={selectedRole}
        />
      </Form>
      <ConfirmModal
        header="Delete role?"
        description={`Are you sure you want to delete the role ${selectedRole.name}?`}
        toType={selectedRole.name}
        open={removeDialog}
        close={onToggleRemoveDialog}
        onConfirm={onRemoveRole}
      />
    </Drawer>
  );
};

export default RoleDefinition;
