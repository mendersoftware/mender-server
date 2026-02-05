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
import { useEffect, useMemo, useState } from 'react';
import { useWatch } from 'react-hook-form';
import { useSelector } from 'react-redux';

import { InfoOutlined } from '@mui/icons-material';
import {
  Alert,
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
  Tooltip
} from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import DocsLink from '@northern.tech/common-ui/DocsLink';
import EnterpriseNotification from '@northern.tech/common-ui/EnterpriseNotification';
import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import Form from '@northern.tech/common-ui/forms/Form';
import FormCheckbox from '@northern.tech/common-ui/forms/FormCheckbox';
import PasswordInput from '@northern.tech/common-ui/forms/PasswordInput';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import { BENEFITS, rolesById, rolesByName, uiPermissionsById } from '@northern.tech/store/constants';
import { getIsEnterprise } from '@northern.tech/store/selectors';
import pluralize from 'pluralize';
import validator from 'validator';

const { isUUID } = validator;

const useStyles = makeStyles()(theme => ({
  formWrapper: { display: 'flex', flexDirection: 'column', gap: theme.spacing(2), paddingTop: theme.spacing(4) }
}));

export const UserRolesSelect = ({ currentUser, disabled, onSelect, roles, user }) => {
  const isEnterprise = useSelector(getIsEnterprise);
  const relevantRolesById = useMemo(
    () => roles.reduce((accu, role) => ({ ...accu, [role.value ?? role.name]: { ...role, value: role.value ?? role.name } }), {}),
    [roles]
  );
  const [selectedRoleIds, setSelectedRoleIds] = useState(
    (user.roles || [rolesByName.admin]).reduce((accu, roleId) => {
      const foundRole = relevantRolesById[roleId];
      if (foundRole) {
        accu.push(roleId);
      }
      return accu;
    }, [])
  );

  const onInputChange = ({ target: { value } }) => {
    const { roles = [] } = user;
    let newlySelectedRoles = value;
    if (value.includes('')) {
      newlySelectedRoles = [];
    }
    const hadRoleChanges =
      roles.length !== newlySelectedRoles.length || roles.some(currentRoleId => !newlySelectedRoles.some(roleId => currentRoleId === roleId));
    setSelectedRoleIds(newlySelectedRoles);
    onSelect(newlySelectedRoles, hadRoleChanges);
  };

  const { editableRoles, showRoleUsageNotification } = useMemo(() => {
    const editableRoles = Object.entries(relevantRolesById).map(([value, role]) => {
      const enabled = selectedRoleIds.some(roleId => value === roleId);
      return { enabled, value, ...role };
    });
    const showRoleUsageNotification = selectedRoleIds.reduce((accu, roleId) => {
      const { permissions, uiPermissions } = relevantRolesById[roleId];
      const hasUiApiAccess = [rolesByName.ci].includes(roleId)
        ? false
        : roleId === rolesByName.admin ||
          permissions.some(permission => ![rolesByName.deploymentCreation.action].includes(permission.action)) ||
          uiPermissions.userManagement.includes(uiPermissionsById.read.value);
      if (hasUiApiAccess) {
        return false;
      }
      return typeof accu !== 'undefined' ? accu : true;
    }, undefined);
    return { editableRoles, showRoleUsageNotification };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(relevantRolesById), selectedRoleIds]);

  return (
    <div className="flexbox column">
      <div className="flexbox margin-top-small" style={{ alignItems: 'flex-end' }}>
        <FormControl id="roles-form" style={{ maxWidth: 400 }}>
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
          {showRoleUsageNotification && (
            <FormHelperText className="info">
              The selected {pluralize('role', selectedRoleIds.length)} may prevent {currentUser.email === user.email ? 'you' : <i>{user.email}</i>} from using
              the Mender UI.
              <br />
              Consider adding the <i>{rolesById[rolesByName.readOnly].name}</i> role as well.
            </FormHelperText>
          )}
        </FormControl>
        <EnterpriseNotification className="margin-left-small" id={BENEFITS.rbac.id} />
      </div>
      {!isEnterprise && (
        <Alert className="margin-top-small" severity="warning">
          Role-based access control (RBAC) is not available in your current plan. All users will have full administrative access.
        </Alert>
      )}
    </div>
  );
};

export const PasswordLabel = () => (
  <div className="flexbox center-aligned">
    Optional
    <Tooltip
      title={
        <>
          <p>You can skip setting a password for now - you can opt to send the new user an email containing a password reset link by checking the box below.</p>
          <p>
            For Single Sign-On to work, you must create users with no password. Please{' '}
            <DocsLink path="server-integration/saml-federated-authentication" title="see the documentation" /> for more information.
          </p>
        </>
      }
    >
      <InfoOutlined fontSize="small" className="margin-left-small" />
    </Tooltip>
  </div>
);

const UserIdentifier = ({ userIdAllowed, onHasUserId }) => {
  const value = useWatch({ name: 'email', defaultValue: '' });

  useEffect(() => {
    if (userIdAllowed) {
      onHasUserId(isUUID(value));
    }
  }, [userIdAllowed, value, onHasUserId]);

  return (
    <TextInput
      hint="Email"
      label={userIdAllowed ? 'Email or User ID' : 'Email'}
      id="email"
      validations={userIdAllowed ? 'isLength:1,isUUID||isEmail,trim' : 'isLength:1,trim'}
      required
      autocomplete="off"
    />
  );
};

export const UserForm = ({ closeDialog, currentUser, canManageUsers, isEnterprise, roles, submit, isTrial }) => {
  const [hadRoleChanges, setHadRoleChanges] = useState(false);
  const [selectedRoles, setSelectedRoles] = useState();
  const [isAddingExistingUser, setIsAddingExistingUser] = useState(false);
  const { classes } = useStyles();

  const onSelect = (newlySelectedRoles, hadRoleChanges) => {
    setSelectedRoles(newlySelectedRoles);
    setHadRoleChanges(hadRoleChanges);
  };

  const onSubmit = data => {
    const { password, ...remainder } = data;
    const roleData = hadRoleChanges ? { roles: selectedRoles } : {};
    if (isAddingExistingUser) {
      const { email: userId } = data;
      return submit(userId, 'add');
    }
    return submit({ ...remainder, ...roleData, password }, 'create');
  };

  return (
    <BaseDialog open title="Add new user" fullWidth maxWidth="sm" onClose={closeDialog}>
      <DialogContent style={{ overflowY: 'initial' }}>
        <Form
          className={classes.formWrapper}
          onSubmit={onSubmit}
          handleCancel={closeDialog}
          submitLabel={`${isAddingExistingUser ? 'Add' : 'Create'} user`}
          showButtons={true}
          autocomplete="off"
        >
          <UserIdentifier userIdAllowed={isEnterprise && !isTrial} onHasUserId={setIsAddingExistingUser} />
          <Collapse in={!isAddingExistingUser}>
            <PasswordInput
              id="password"
              autocomplete="off"
              create
              edit={false}
              generate
              InputLabelProps={{ shrink: true }}
              label={<PasswordLabel />}
              placeholder="Password"
              validations="isLength:8"
            />
            <FormCheckbox id="shouldResetPassword" label="Send an email to the user containing a link to reset the password" />
            <UserRolesSelect currentUser={currentUser} disabled={!(canManageUsers && isEnterprise)} onSelect={onSelect} roles={roles} user={{}} />
          </Collapse>
        </Form>
      </DialogContent>
      <DialogActions />
    </BaseDialog>
  );
};

export default UserForm;
