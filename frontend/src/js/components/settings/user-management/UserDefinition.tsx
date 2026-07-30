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
import { Alert, Button, Divider, TextField, Typography, textFieldClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import BaseDrawer from '@northern.tech/common-ui/BaseDrawer';
import { ConfirmModal } from '@northern.tech/common-ui/ConfirmModal';
import { CopyTextToClipboard } from '@northern.tech/common-ui/CopyText';
import { ColumnWidthProvider, SynchronizedTwoColumnData, TwoColumnData } from '@northern.tech/common-ui/TwoColumnData';
import actions from '@northern.tech/store/actions';
import { rolesByName, uiPermissionsByArea, uiPermissionsById } from '@northern.tech/store/constants';
import { useAppDispatch } from '@northern.tech/store/store';
import { passwordResetStart } from '@northern.tech/store/thunks';
import { mapUserRolesToUiPermissions } from '@northern.tech/store/utils';
import type { User } from '@northern.tech/types/MenderTypes';
import { isEmpty } from '@northern.tech/utils/helpers';

import { OAuth2Providers, genericProvider } from '../../login/OAuth2Providers';
import { EmailVerificationWarning } from '../EmailVerificationWarning';
import { SETTINGS_FORM_MAX_WIDTH, SETTINGS_INPUT_WIDTH, SETTINGS_INPUT_WIDTH_ROLES_AND_USERS_ONLY } from '../constants';
import { UserRolesSelect } from './UserForm';

const { setSnackbar } = actions;

const useStyles = makeStyles()(theme => ({
  oauthIcon: { fontSize: 36, marginRight: 10 },
  userIdWrapper: {
    '.copy-button': { marginTop: theme.spacing(0.25), whiteSpace: 'nowrap' },
    [`&.profile-settings .${textFieldClasses.root}`]: { minWidth: SETTINGS_INPUT_WIDTH },
    maxWidth: SETTINGS_INPUT_WIDTH_ROLES_AND_USERS_ONLY
  },
  widthLimit: { maxWidth: SETTINGS_FORM_MAX_WIDTH, [`.${textFieldClasses.root}`]: { width: SETTINGS_INPUT_WIDTH_ROLES_AND_USERS_ONLY } }
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

const mapPermissions = permissions => permissions.map(permission => uiPermissionsById[permission].title).join(', ') || 'None';

const scopedPermissionAreas = {
  groups: 'Device groups',
  releases: 'Releases'
};

export const UserId = ({ className = '', userId }) => {
  const { classes } = useStyles();
  return (
    <div className={`flexbox margin-bottom-small ${classes.userIdWrapper} ${className}`}>
      <TextField className="margin-right-small" label="User ID" key={userId} disabled defaultValue={userId} />
      <div className="copy-button">
        <CopyTextToClipboard notify={false} token={userId} />
      </div>
    </div>
  );
};

interface UserDefinitionProps {
  currentUser: User & { verified?: boolean };
  hasMultitenancy: boolean;
  isEnterprise: boolean;
  onCancel: () => void;
  onRemove: (user: User) => void;
  onSubmit: (userData: (User & { roles?: string[] }) | null, type: string, id: string) => void;
  roles: { name: string; value?: string }[];
  selectedUser: User & { roles?: string[] };
}

export const UserDefinition = ({ currentUser, hasMultitenancy, isEnterprise, onCancel, onSubmit, onRemove, roles, selectedUser }: UserDefinitionProps) => {
  const { id } = selectedUser;

  const { classes } = useStyles();
  const dispatch = useAppDispatch();

  const [hadRoleChanges, setHadRoleChanges] = useState(false);
  const [isEditingRoles, setIsEditingRoles] = useState(false);
  const [selectedRoles, setSelectedRoles] = useState([]);
  const [showResetConfirmation, setShowResetConfirmation] = useState(false);
  const rolesById = useMemo(
    () => roles.reduce((accu, role) => ({ ...accu, [role.value ?? role.name]: { ...role, value: role.value ?? role.name } }), {}),
    [roles]
  );

  useEffect(() => {
    setSelectedRoles(selectedUser.roles || []);
    setHadRoleChanges(false);
    setIsEditingRoles(false);
  }, [selectedUser.roles]);

  const onRemoveClick = () => {
    onRemove(selectedUser);
  };

  const onRolesSelect = (newlySelectedRoles, hadRoleChanges) => {
    setSelectedRoles(newlySelectedRoles);
    setHadRoleChanges(hadRoleChanges);
  };

  const onCancelRoleChanges = () => {
    setSelectedRoles(selectedUser.roles || []);
    setHadRoleChanges(false);
    setIsEditingRoles(false);
  };

  const onPasswordResetConfirmed = () =>
    dispatch(passwordResetStart(selectedUser.email))
      .unwrap()
      .then(() => dispatch(setSnackbar(`A password reset email was sent to ${selectedUser.email}.`)));

  const onSubmitClick = () => onSubmit({ ...selectedUser, roles: selectedRoles }, 'edit', id);

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
  const userNotVerified = !currentUser.verified;
  const isSubmitDisabled = !selectedRoles.length || !hadRoleChanges;

  const { isOAuth2, provider } = getUserSSOState(selectedUser);
  const rolesClasses = isEnterprise ? '' : 'muted';
  return (
    <BaseDrawer
      onClose={onCancel}
      open={!!id}
      size="md"
      slotProps={{
        header: {
          title: `User information for ${selectedUser.email}`
        }
      }}
    >
      {hasMultitenancy && userNotVerified && <EmailVerificationWarning className="margin-top-small" action="change another user’s email" />}
      <Typography className="margin-top-x-small" variant="subtitle1">
        User ID
      </Typography>
      <TwoColumnData className="margin-top-small" setSnackbar={setSnackbar} data={{ Email: selectedUser.email, 'User ID': selectedUser.id }} />
      {!isOAuth2 && (
        <>
          <div>
            <Button className="margin-top-small" onClick={() => setShowResetConfirmation(true)}>
              Send password reset link
            </Button>
          </div>
          <ConfirmModal
            header="Send password reset link?"
            description={
              <>
                We&rsquo;ll send an email to <b>{selectedUser.email}</b> with instructions for resetting their Mender account password.
              </>
            }
            isDanger={false}
            open={showResetConfirmation}
            close={() => setShowResetConfirmation(false)}
            onConfirm={onPasswordResetConfirmed}
          />
        </>
      )}
      {isOAuth2 && (
        <div className="flexbox margin-top-small margin-bottom">
          <div className={classes.oauthIcon}>{provider.icon}</div>
          <div className="info">
            This user logs in using their <strong>{provider.name}</strong> account.
            <br />
            They can connect to {provider.name} to update their login settings.
          </div>
        </div>
      )}
      <Typography className="margin-top" variant="subtitle1">
        Roles
      </Typography>
      <UserRolesSelect
        key={`roles-select-${isEditingRoles}`}
        disabled={!isEnterprise || !isEditingRoles}
        currentUser={currentUser}
        onSelect={onRolesSelect}
        roles={roles}
        user={selectedUser}
      />
      {isEnterprise && (
        <div className="flexbox margin-top-small">
          {isEditingRoles ? (
            <>
              <Button color="info" variant="outlined" className="margin-right-x-small" onClick={onCancelRoleChanges}>
                Cancel
              </Button>
              <Button variant="contained" disabled={isSubmitDisabled} onClick={onSubmitClick}>
                Save changes
              </Button>
            </>
          ) : (
            <Button onClick={() => setIsEditingRoles(true)}>Change roles</Button>
          )}
        </div>
      )}
      {!isEnterprise && (
        <Alert className={`margin-top-small ${classes.widthLimit}`} severity="warning">
          Role-base access control (RBAC) is not available in your current plan. All users will have full administrative access
          {selectedRoles.includes(rolesByName.admin) ? ', and the permissions shown below apply to all users' : ''}.
        </Alert>
      )}
      <ColumnWidthProvider>
        {!!(hasScopedPermissionsDefined || !isEmpty(areas)) && (
          <Typography className="margin-top margin-bottom-x-small" variant="subtitle1">
            Role permissions
          </Typography>
        )}
        <SynchronizedTwoColumnData className={rolesClasses} data={areas} />
        {Object.entries(scopedAreas).reduce((accu, [area, areaPermissions]) => {
          if (isEmpty(areaPermissions)) {
            return accu;
          }
          accu.push(
            <Fragment key={area}>
              <Typography className="margin-top margin-bottom-x-small" variant="subtitle1">
                {scopedPermissionAreas[area]}
              </Typography>
              <SynchronizedTwoColumnData className={rolesClasses} data={areaPermissions} />
            </Fragment>
          );
          return accu;
        }, [])}
      </ColumnWidthProvider>
      {currentUser.id !== id && (
        <>
          <Divider className="margin-top-large" />
          <div className="margin-top-medium">
            <Button color="error" onClick={onRemoveClick} variant="outlined">
              Delete user
            </Button>
          </div>
        </>
      )}
    </BaseDrawer>
  );
};

export default UserDefinition;
