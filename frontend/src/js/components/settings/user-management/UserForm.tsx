// Copyright 2017 Northern.tech AS
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
import { useEffect, useMemo, useRef, useState } from 'react';
import { useWatch } from 'react-hook-form';
import { useSelector } from 'react-redux';

import {
  Alert,
  Button,
  Checkbox,
  Collapse,
  DialogActions,
  DialogContent,
  FormControl,
  FormHelperText,
  InputLabel,
  ListItemText,
  MenuItem,
  Select,
  Typography
} from '@mui/material';

import EnterpriseNotification from '@northern.tech/common-ui/EnterpriseNotification';
import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import Form from '@northern.tech/common-ui/forms/Form';
import PasswordInput from '@northern.tech/common-ui/forms/PasswordInput';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import { BENEFITS, rolesById, rolesByName, uiPermissionsById } from '@northern.tech/store/constants';
import { getIsEnterprise } from '@northern.tech/store/selectors';
import pluralize from 'pluralize';
import validator from 'validator';

import { SETTINGS_SELECT_WIDTH } from '../constants';

const { isUUID } = validator;

export const UserRolesSelect = ({ currentUser, disabled, error = '', maxWidth = SETTINGS_SELECT_WIDTH, onSelect, roles, user }) => {
  const isEnterprise = useSelector(getIsEnterprise);
  const relevantRolesById = useMemo(
    () =>
      Object.fromEntries(
        roles.map(role => {
          const value = role.value ?? role.name;
          return [value, { ...role, value }];
        })
      ),
    [roles]
  );
  const [selectedRoleIds, setSelectedRoleIds] = useState<string[]>([]);

  useEffect(() => {
    setSelectedRoleIds((user.roles || [rolesByName.admin]).filter(roleId => relevantRolesById[roleId]));
  }, [user.roles, relevantRolesById]);

  const onInputChange = ({ target: { value: newlySelectedRoles } }) => {
    const { roles: assignedRoles = [] } = user;
    const hadRoleChanges = assignedRoles.length !== newlySelectedRoles.length || assignedRoles.some(roleId => !newlySelectedRoles.includes(roleId));
    setSelectedRoleIds(newlySelectedRoles);
    onSelect(newlySelectedRoles, hadRoleChanges);
  };

  const hasUiApiAccess = ({ value, permissions, uiPermissions }) => {
    if (value === rolesByName.ci) {
      return false;
    }
    return (
      value === rolesByName.admin ||
      permissions.some(({ action }) => action !== rolesByName.deploymentCreation.action) ||
      uiPermissions.userManagement.includes(uiPermissionsById.read.value)
    );
  };

  const { editableRoles, showRoleUsageNotification } = useMemo(() => {
    const editableRoles = Object.values(relevantRolesById).map(role => ({ ...role, enabled: selectedRoleIds.includes(role.value) }));
    const showRoleUsageNotification = selectedRoleIds.length ? !selectedRoleIds.some(roleId => hasUiApiAccess(relevantRolesById[roleId])) : undefined;
    return { editableRoles, showRoleUsageNotification };
  }, [relevantRolesById, selectedRoleIds]);

  return (
    <div className="flexbox column">
      <FormControl error={!!error} id="roles-form" style={{ maxWidth }}>
        <InputLabel id="roles-selection-label">Roles</InputLabel>
        <Select
          label="Roles"
          labelId="roles-selection-label"
          id={`roles-selector-${selectedRoleIds.length}`}
          disabled={disabled}
          multiple
          value={selectedRoleIds}
          required
          onChange={onInputChange}
          renderValue={selected => selected.map(role => relevantRolesById[role].name).join(', ')}
        >
          {editableRoles.map(role => (
            <MenuItem id={role.value} key={role.value} value={role.value}>
              <Checkbox id={`${role.value}-checkbox`} checked={role.enabled} />
              <ListItemText id={`${role.value}-text`} primary={role.name} />
            </MenuItem>
          ))}
        </Select>
        {!!error && <FormHelperText>{error}</FormHelperText>}
      </FormControl>
      {showRoleUsageNotification && (
        <Alert className="margin-top-small" severity="warning">
          The selected {pluralize('role', selectedRoleIds.length)} may prevent {currentUser.email === user.email ? 'you' : <i>{user.email}</i>} from using the
          Mender UI. Consider adding the <i>{rolesById[rolesByName.readOnly].name}</i> role as well.
        </Alert>
      )}
      {!isEnterprise && (
        <Alert className="margin-top-small" severity="warning">
          Role-based access control (RBAC) is not available in your current plan. All users will have full administrative access.
        </Alert>
      )}
    </div>
  );
};

const UserIdentifier = ({ userIdAllowed, onHasUserId, hasMultitenancy }) => {
  const value = useWatch({ name: 'email', defaultValue: '' });

  useEffect(() => {
    if (userIdAllowed) {
      onHasUserId(isUUID(value));
    }
  }, [userIdAllowed, value, onHasUserId]);

  return (
    <div className="flexbox column">
      <Typography className="margin-bottom-x-small" variant="subtitle1">
        {userIdAllowed ? 'Email or User ID' : 'Email'}
      </Typography>
      <TextInput
        hint={userIdAllowed ? 'email@example.com or User ID' : 'email@example.com'}
        InputLabelProps={{ shrink: true }}
        helperText={
          userIdAllowed
            ? 'Use an email address to invite a new user, or a User ID to add an existing one.'
            : `Use an email address to ${hasMultitenancy ? 'invite' : 'create'} a new user.`
        }
        id="email"
        validations={userIdAllowed ? 'isLength:1,isUUID||isEmail,trim' : 'isLength:1,isEmail,trim'}
        required
        requiredRendered={false}
        autocomplete="off"
        width="100%"
      />
    </div>
  );
};

export const UserForm = ({ closeDialog, currentUser, canManageUsers, hasMultitenancy, isEnterprise, roles, submit, isTrial }) => {
  const [hadRoleChanges, setHadRoleChanges] = useState(false);
  const [selectedRoles, setSelectedRoles] = useState<string[] | undefined>();
  const [isAddingExistingUser, setIsAddingExistingUser] = useState(false);
  const submitRef = useRef<() => void>(undefined);
  const onSelect = (newlySelectedRoles, hadRoleChanges) => {
    setSelectedRoles(newlySelectedRoles);
    setHadRoleChanges(hadRoleChanges);
  };

  const hasEmptyRoleSelection = !isAddingExistingUser && !!selectedRoles && !selectedRoles.length;

  const onSubmit = async data => {
    if (hasEmptyRoleSelection) {
      return;
    }
    const { password, ...remainder } = data;
    const roleData = hadRoleChanges ? { roles: selectedRoles } : {};
    // Add via id / invite via email / OS create
    const [payload, type] = isAddingExistingUser
      ? [remainder.email, 'add']
      : hasMultitenancy
        ? [{ ...remainder, ...roleData }, 'createV2']
        : [{ ...remainder, ...roleData, password, shouldResetPassword: false }, 'create'];
    if (await submit(payload, type)) {
      closeDialog();
    }
  };

  return (
    <BaseDialog open title="Add user" fullWidth maxWidth="sm" onClose={closeDialog}>
      <DialogContent>
        <Form className="flexbox column" onSubmit={onSubmit} showButtons={false} submitRef={submitRef} autocomplete="off">
          <UserIdentifier userIdAllowed={isEnterprise && !isTrial} hasMultitenancy={hasMultitenancy} onHasUserId={setIsAddingExistingUser} />
          <Collapse in={!isAddingExistingUser}>
            {!hasMultitenancy && (
              <PasswordInput
                id="password"
                autocomplete="off"
                create
                InputLabelProps={{ shrink: false }}
                label="Password"
                placeholder="Password"
                required
                validations="isLength:8:256"
              />
            )}
            <div className="flexbox column margin-top-small">
              <div className="flexbox align-items-center margin-bottom-x-small">
                <Typography variant="subtitle1">Roles</Typography>
                <EnterpriseNotification className="margin-left-small" id={BENEFITS.rbac.id} />
              </div>
              <UserRolesSelect
                currentUser={currentUser}
                disabled={!(canManageUsers && isEnterprise)}
                error={hasEmptyRoleSelection ? 'Select at least one role.' : ''}
                maxWidth={SETTINGS_SELECT_WIDTH}
                onSelect={onSelect}
                roles={roles}
                user={{}}
              />
            </div>
          </Collapse>
        </Form>
      </DialogContent>
      <DialogActions>
        <Button onClick={closeDialog}>Cancel</Button>
        <Button variant="contained" disabled={hasEmptyRoleSelection} onClick={() => submitRef.current?.()}>
          Add user
        </Button>
      </DialogActions>
    </BaseDialog>
  );
};

export default UserForm;
