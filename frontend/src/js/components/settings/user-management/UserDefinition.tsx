// Copyright 2021 Northern.tech AS
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
import { Fragment, useEffect, useMemo, useState } from 'react';

// material ui
import { Button, Checkbox, Divider, Drawer, FormControl, FormControlLabel, FormHelperText, InputLabel, TextField, textFieldClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { CopyTextToClipboard } from '@northern.tech/common-ui/CopyText';
import { DrawerTitle } from '@northern.tech/common-ui/DrawerTitle';
import { TwoColumnData } from '@northern.tech/common-ui/TwoColumnData';
import { uiPermissionsByArea, uiPermissionsById } from '@northern.tech/store/constants';
import { mapUserRolesToUiPermissions } from '@northern.tech/store/utils';
import { isEmpty, toggle } from '@northern.tech/utils/helpers';
import validator from 'validator';

import { OAuth2Providers, genericProvider } from '../../login/OAuth2Providers';
import { UserRolesSelect } from './UserForm';

const useStyles = makeStyles()(theme => ({
  actionButtons: { justifyContent: 'flex-end' },
  divider: { marginTop: theme.spacing(4) },
  leftButton: { marginRight: theme.spacing(2) },
  oauthIcon: { fontSize: 36, marginRight: 10 },
  userIdWrapper: {
    // the following 2 lines are required to align the CopyTextToClipboard with the tenant token without sacrificing consistent behaviour
    marginBottom: theme.spacing(-3),
    '.copy-button': { marginBottom: theme.spacing(-1) },
    [`.${textFieldClasses.root}`]: { width: 400 }
  },
  widthLimit: { marginTop: theme.spacing(3), maxWidth: 620, [`.${textFieldClasses.root}`]: { width: 400 } }
}));

export const getUserSSOState = user => {
  const { sso = [] } = user;
  const isOAuth2 = !!sso.length;
  let provider = null;
  if (isOAuth2) {
    provider = OAuth2Providers.find(provider => sso.some(({ kind }) => kind.includes(provider.id))) ?? genericProvider;
  }
  return { isOAuth2, provider };
};

const mapPermissions = permissions => permissions.map(permission => uiPermissionsById[permission].title).join(', ');

const scopedPermissionAreas = {
  groups: 'Device groups',
  releases: 'Releases'
};

export const UserId = ({ className = '', userId }) => {
  const { classes } = useStyles();
  return (
    <div className={`flexbox space-between ${classes.userIdWrapper} ${className}`}>
      <TextField label="User ID" key={userId} disabled defaultValue={userId} />
      <div className="flexbox center-aligned copy-button">
        <CopyTextToClipboard token={userId} />
      </div>
    </div>
  );
};

export const UserDefinition = ({ currentUser, isEnterprise, onCancel, onSubmit, onRemove, roles, selectedUser }) => {
  const { email = '', id } = selectedUser;

  const { classes } = useStyles();

  const [nameError, setNameError] = useState(false);
  const [hadRoleChanges, setHadRoleChanges] = useState(false);
  const [selectedRoles, setSelectedRoles] = useState([]);
  const [shouldResetPassword, setShouldResetPassword] = useState(false);
  const [currentEmail, setCurrentEmail] = useState('');
  const rolesById = useMemo(
    () => roles.reduce((accu, role) => ({ ...accu, [role.value ?? role.name]: { ...role, value: role.value ?? role.name } }), {}),
    [roles]
  );

  useEffect(() => {
    setCurrentEmail(email);
  }, [email]);

  useEffect(() => {
    setSelectedRoles(selectedUser.roles || []);
  }, [selectedUser.roles]);

  const validateNameChange = ({ target: { value } }) => {
    setNameError(!validator.isEmail(value) || validator.isEmpty(value));
    setCurrentEmail(value);
  };

  const onRemoveClick = () => {
    onRemove(selectedUser);
  };

  const onRolesSelect = (newlySelectedRoles, hadRoleChanges) => {
    setSelectedRoles(newlySelectedRoles);
    setHadRoleChanges(hadRoleChanges);
  };

  const onSubmitClick = () => {
    if (id && !hadRoleChanges && email === currentEmail) {
      return onSubmit(null, 'edit', id, shouldResetPassword ? email : null);
    }
    const changedRoles = hadRoleChanges ? { roles: selectedRoles } : {};
    const submissionData = { ...selectedUser, ...changedRoles, email: currentEmail };
    return onSubmit(submissionData, 'edit', id, shouldResetPassword ? currentEmail : null);
  };

  const togglePasswordReset = () => setShouldResetPassword(toggle);

  const { areas, ...scopedAreas } = useMemo(() => {
    const emptySelection = { areas: {}, groups: {}, releases: {} };
    if (!selectedRoles.length || isEmpty(rolesById)) {
      return emptySelection;
    }

    return Object.entries(mapUserRolesToUiPermissions(selectedRoles, rolesById)).reduce((accu, [key, values]) => {
      if (scopedPermissionAreas[key]) {
        accu[key] = Object.entries(values).reduce((groupsAccu, [name, uiPermissions]) => {
          groupsAccu[name] = mapPermissions(uiPermissions);
          return groupsAccu;
        }, {});
      } else {
        accu.areas[uiPermissionsByArea[key].title] = mapPermissions(values);
      }
      return accu;
    }, emptySelection);
  }, [selectedRoles, rolesById]);

  const hasScopedPermissionsDefined = Object.values(scopedAreas).some(permissions => !isEmpty(permissions));
  const isSubmitDisabled = !selectedRoles.length;

  const { isOAuth2, provider } = getUserSSOState(selectedUser);
  const rolesClasses = isEnterprise ? '' : 'muted';
  return (
    <Drawer anchor="right" onClose={onCancel} open={!!id} PaperProps={{ style: { minWidth: 600, width: '50vw' } }}>
      <DrawerTitle
        title="Edit user"
        onClose={onCancel}
        preCloser={
          currentUser.id !== id && (
            <Button className={`flexbox center-aligned ${classes.leftButton}`} color="secondary" onClick={onRemoveClick}>
              delete user
            </Button>
          )
        }
      />
      <Divider />
      <UserId className={classes.widthLimit} userId={id} />
      <FormControl className={classes.widthLimit}>
        <TextField label="Email" id="email" value={currentEmail} disabled={isOAuth2 || currentUser.id === id} error={nameError} onChange={validateNameChange} />
        {nameError && <FormHelperText className="warning">Please enter a valid email address</FormHelperText>}
      </FormControl>
      {isOAuth2 ? (
        <div className="flexbox margin-top-small margin-bottom">
          <div className={classes.oauthIcon}>{provider.icon}</div>
          <div className="info">
            This user logs in using their <strong>{provider.name}</strong> account.
            <br />
            They can connect to {provider.name} to update their login settings.
          </div>
        </div>
      ) : (
        <FormControlLabel
          control={<Checkbox checked={shouldResetPassword} onChange={togglePasswordReset} />}
          label="Send an email to the user containing a link to reset the password"
        />
      )}
      <UserRolesSelect disabled={!isEnterprise} currentUser={currentUser} onSelect={onRolesSelect} roles={roles} user={selectedUser} />
      {!!(hasScopedPermissionsDefined || !isEmpty(areas)) && (
        <InputLabel className="margin-top" shrink>
          Role permissions
        </InputLabel>
      )}
      <TwoColumnData className={rolesClasses} config={areas} />
      {Object.entries(scopedAreas).reduce((accu, [area, areaPermissions]) => {
        if (isEmpty(areaPermissions)) {
          return accu;
        }
        accu.push(
          <Fragment key={area}>
            <InputLabel className="margin-top-small" shrink>
              {scopedPermissionAreas[area]}
            </InputLabel>
            <TwoColumnData className={rolesClasses} config={areaPermissions} />
          </Fragment>
        );
        return accu;
      }, [])}
      <Divider className={classes.divider} light />
      <div className={`flexbox centered margin-top ${classes.actionButtons}`}>
        <Button className={classes.leftButton} onClick={onCancel}>
          Cancel
        </Button>
        <Button variant="contained" disabled={isSubmitDisabled} target="_blank" onClick={onSubmitClick}>
          Save
        </Button>
      </div>
    </Drawer>
  );
};

export default UserDefinition;
