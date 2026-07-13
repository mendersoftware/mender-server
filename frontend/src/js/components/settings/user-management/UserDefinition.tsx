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
import { Alert, Button, Chip, TextField, Typography, textFieldClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import BaseDrawer from '@northern.tech/common-ui/BaseDrawer';
import { ContentSection } from '@northern.tech/common-ui/ContentSection';
import { CopyTextToClipboard } from '@northern.tech/common-ui/CopyText';
import { ColumnWidthProvider, SynchronizedTwoColumnData, TwoColumnData } from '@northern.tech/common-ui/TwoColumnData';
import { rolesByName, twoFAStates, uiPermissionsByArea, uiPermissionsById } from '@northern.tech/store/constants';
import { mapUserRolesToUiPermissions } from '@northern.tech/store/utils';
import type { User } from '@northern.tech/types/MenderTypes';
import { isEmpty, toggle } from '@northern.tech/utils/helpers';

import { OAuth2Providers, genericProvider } from '../../login/OAuth2Providers';
import { EmailVerificationWarning } from '../EmailVerificationWarning';
import { SETTINGS_FORM_MAX_WIDTH, SETTINGS_INPUT_WIDTH, SETTINGS_INPUT_WIDTH_ROLES_AND_USERS_ONLY } from '../constants';
import { UserRolesSelect } from './UserForm';

const useStyles = makeStyles()(theme => ({
  divider: { marginTop: theme.spacing(4) },
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
  isEnterprise: boolean;
  onCancel: () => void;
  onRemove: (user: User) => void;
  onSubmit: (userData: (User & { roles?: string[] }) | null, type: string, id: string) => void;
  roles: { name: string; value?: string }[];
  selectedUser: User & { roles?: string[] };
}

const authChipProps = {
  size: 'small',
  variant: 'outlined',
  color: 'warning'
};

export const UserDefinition = ({ currentUser, isEnterprise, onCancel, onSubmit, onRemove, roles, selectedUser = {} }: UserDefinitionProps) => {
  const { email = '', id } = selectedUser;

  const { classes } = useStyles();

  const [hadRoleChanges, setHadRoleChanges] = useState(false);
  const [selectedRoles, setSelectedRoles] = useState([]);
  const [isEditingRoles, setIsEditingRoles] = useState(false);
  const rolesById = useMemo(
    () => roles.reduce((accu, role) => ({ ...accu, [role.value ?? role.name]: { ...role, value: role.value ?? role.name } }), {}),
    [roles]
  );

  useEffect(() => {
    setSelectedRoles(selectedUser.roles || []);
  }, [selectedUser.roles]);

  const onRemoveClick = () => {
    onRemove(selectedUser);
  };

  const onRolesSelect = (newlySelectedRoles, hadRoleChanges) => {
    setSelectedRoles(newlySelectedRoles);
    setHadRoleChanges(hadRoleChanges);
  };

  const onSubmitClick = () => {
    onSubmit({ ...selectedUser, roles: selectedRoles }, 'edit', id);
    setIsEditingRoles(false);
  };

  const onRoleEditToggle = () => setIsEditingRoles(toggle);

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

  const verificationChip = selectedUser.verified ? (
    <Chip {...authChipProps} label="Verified" color="success" />
  ) : (
    <Chip {...authChipProps} label="Not verified" />
  );

  const tfaStatus = isOAuth2 ? (
    'SSO managed'
  ) : selectedUser.tfa_status === twoFAStates.enabled ? (
    <Chip {...authChipProps} label="Enabled" color="success" />
  ) : (
    <Chip {...authChipProps} label="Not enabled" />
  );

  const signInData = {
    Email: (
      <>
        {email} {verificationChip}
      </>
    ),
    'User ID': id,
    'Two-factor authentication': tfaStatus
  };

  return (
    <BaseDrawer
      onClose={onCancel}
      open={!!id}
      size="md"
      slotProps={{
        header: { title: `User information for ${email}` }
      }}
    >
      {userNotVerified && <EmailVerificationWarning className="margin-top-small" action="change another user's email" />}
      <ContentSection title="Sign-in & security">
        <TwoColumnData data={signInData} />
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
      </ContentSection>
      <ContentSection title="Roles">
        <UserRolesSelect
          disabled={!(isEnterprise && isEditingRoles)}
          currentUser={currentUser}
          key={isEditingRoles}
          onSelect={onRolesSelect}
          roles={roles}
          user={selectedUser}
        />
        <div className="flexbox margin-top-small">
          {isEditingRoles ? (
            <>
              <Button className="margin-right-small" color="neutral" onClick={onRoleEditToggle} variant="outlined">
                Cancel
              </Button>
              <Button color="secondary" variant="contained" disabled={isSubmitDisabled} onClick={onSubmitClick}>
                Save changes
              </Button>
            </>
          ) : (
            <Button color="secondary" onClick={onRoleEditToggle}>
              Change roles
            </Button>
          )}
        </div>
        {!isEnterprise && (
          <Alert className={`margin-top-small ${classes.widthLimit}`} severity="warning">
            Role-base access control (RBAC) is not available in your current plan. All users will have full administrative access
            {selectedRoles.includes(rolesByName.admin) ? ', and the permissions shown below apply to all users' : ''}.
          </Alert>
        )}
        <ColumnWidthProvider>
          {!!(hasScopedPermissionsDefined || !isEmpty(areas)) && (
            <Typography className="margin-top margin-bottom-small" variant="subtitle1">
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
                <Typography className="margin-top-medium margin-bottom-small" variant="subtitle1">
                  {scopedPermissionAreas[area]}
                </Typography>
                <SynchronizedTwoColumnData className={rolesClasses} data={areaPermissions} />
              </Fragment>
            );
            return accu;
          }, [])}
        </ColumnWidthProvider>
      </ContentSection>
      <div className="flexbox margin-top-small">
        {currentUser.id !== id && (
          <Button color="error" onClick={onRemoveClick} variant="outlined">
            Delete user
          </Button>
        )}
      </div>
    </BaseDrawer>
  );
};

export default UserDefinition;
