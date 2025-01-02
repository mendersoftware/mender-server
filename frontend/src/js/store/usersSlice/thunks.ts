'use strict';

// Copyright 2019 Northern.tech AS
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
// @ts-nocheck
import { HELPTOOLTIPS } from '@northern.tech/helptips/HelpTooltips';
import storeActions from '@northern.tech/store/actions';
import GeneralApi from '@northern.tech/store/api/general-api';
import UsersApi from '@northern.tech/store/api/users-api';
import { cleanUp, getSessionInfo, maxSessionAge, setSessionInfo } from '@northern.tech/store/auth';
import {
  ALL_RELEASES,
  APPLICATION_JSON_CONTENT_TYPE,
  APPLICATION_JWT_CONTENT_TYPE,
  SSO_TYPES,
  TIMEOUTS,
  apiRoot,
  emptyRole,
  emptyUiPermissions,
  tenantadmApiUrlv2
} from '@northern.tech/store/constants';
import { getOnboardingState, getOrganization, getTooltipsState, getUserSettings as getUserSettingsSelector } from '@northern.tech/store/selectors';
import { commonErrorFallback, commonErrorHandler } from '@northern.tech/store/store';
import { setOfflineThreshold } from '@northern.tech/store/thunks';
import { extractErrorMessage, mergePermissions, preformatWithRequestID } from '@northern.tech/store/utils';
import { duplicateFilter, isEmpty } from '@northern.tech/utils/helpers';
import { clearAllRetryTimers } from '@northern.tech/utils/retrytimer';
import { createAsyncThunk } from '@reduxjs/toolkit';
import hashString from 'md5';
import Cookies from 'universal-cookie';

import { actions, sliceName } from '.';
import {
  OWN_USER_ID,
  PermissionTypes,
  READ_STATES,
  USER_LOGOUT,
  defaultPermissionSets,
  rolesById as defaultRolesById,
  itemUiPermissionsReducer,
  scopedPermissionAreas,
  settingsKeys,
  twoFAStates,
  uiPermissionsByArea,
  uiPermissionsById,
  useradmApiUrl,
  useradmApiUrlv2
} from './constants';
import { getCurrentUser, getRolesById, getUsersById } from './selectors';

const cookies = new Cookies();

const { setAnnouncement, setSnackbar } = storeActions;

const handleLoginError =
  (err, { token2fa: has2FA, password }, rejectWithValue) =>
  dispatch => {
    const errorText = extractErrorMessage(err);
    const is2FABackend = errorText.includes('2fa');
    if (is2FABackend && !has2FA) {
      return rejectWithValue({ error: '2fa code missing' });
    }
    if (password === undefined) {
      // Enterprise supports two-steps login. On the first step you can enter only email
      // and in case of SSO set up you will receive a redirect URL
      // otherwise you will receive 401 status code and password field will be shown.
      return Promise.reject();
    }
    const twoFAError = is2FABackend ? ' and verification code' : '';
    const errorMessage = `There was a problem logging in. Please check your email${
      twoFAError ? ',' : ' and'
    } password${twoFAError}. If you still have problems, contact an administrator.`;
    return Promise.reject(dispatch(setSnackbar({ message: preformatWithRequestID(err.response, errorMessage), action: 'Copy to clipboard' })));
  };

/*
  User management
*/
export const loginUser = createAsyncThunk(`${sliceName}/loginUser`, ({ stayLoggedIn, ...userData }, { dispatch, rejectWithValue }) =>
  UsersApi.postLogin(`${useradmApiUrl}/auth/login`, { ...userData, no_expiry: stayLoggedIn })
    .catch(err => {
      cleanUp();
      return Promise.reject(dispatch(handleLoginError(err, userData, rejectWithValue)));
    })
    .then(({ text: response, contentType }) => {
      // If the content type is application/json then backend returned SSO configuration.
      // user should be redirected to the start sso url to finish login process.
      if (contentType.includes(APPLICATION_JSON_CONTENT_TYPE)) {
        const { id, kind } = response;
        const type = kind.split('/')[1];
        const ssoLoginUrl = SSO_TYPES[type].getStartUrl(id);
        window.location.replace(ssoLoginUrl);
        return;
      }

      const token = response;
      if (contentType !== APPLICATION_JWT_CONTENT_TYPE || !token) {
        return;
      }
      // save token to local storage & set maxAge if noexpiry checkbox not checked
      let now = new Date();
      now.setSeconds(now.getSeconds() + maxSessionAge);
      const expiresAt = stayLoggedIn ? undefined : now.toISOString();
      setSessionInfo({ token, expiresAt });
      cookies.remove('JWT', { path: '/' });
      return dispatch(getUser(OWN_USER_ID))
        .unwrap()
        .catch(e => {
          cleanUp();
          return Promise.reject(dispatch(setSnackbar(extractErrorMessage(e))));
        })
        .then(() => {
          window.sessionStorage.removeItem('pendings-redirect');
          if (window.location.pathname !== '/ui/') {
            window.location.replace('/ui/');
          }
          return Promise.resolve(dispatch(actions.successfullyLoggedIn({ expiresAt, token })));
        });
    })
);

export const logoutUser = createAsyncThunk(`${sliceName}/logoutUser`, (_, { dispatch, getState }) => {
  if (Object.keys(getState().app.uploadsById).length) {
    return Promise.reject();
  }
  return GeneralApi.post(`${useradmApiUrl}/auth/logout`).finally(() => {
    cleanUp();
    clearAllRetryTimers(setSnackbar);
    return Promise.resolve(dispatch({ type: USER_LOGOUT }));
  });
});

export const switchUserOrganization = createAsyncThunk(`${sliceName}/switchUserOrganization`, (tenantId, { getState }) => {
  if (Object.keys(getState().app.uploadsById).length) {
    return Promise.reject();
  }
  return GeneralApi.get(`${useradmApiUrl}/users/tenants/${tenantId}/token`).then(({ data: token }) => {
    window.sessionStorage.setItem('tenantChanged', 'true');
    setSessionInfo({ ...getSessionInfo(), token });
    window.location.reload();
  });
});

export const passwordResetStart = createAsyncThunk(`${sliceName}/passwordResetStart`, (email, { dispatch }) =>
  GeneralApi.post(`${useradmApiUrl}/auth/password-reset/start`, { email }).catch(err =>
    commonErrorHandler(err, `The password reset request cannot be processed:`, dispatch, undefined, true)
  )
);

export const passwordResetComplete = createAsyncThunk(`${sliceName}/passwordResetComplete`, ({ secretHash, newPassword }, { dispatch }) =>
  GeneralApi.post(`${useradmApiUrl}/auth/password-reset/complete`, { secret_hash: secretHash, password: newPassword }).catch((err = {}) => {
    const { error, response = {} } = err;
    let errorMsg = '';
    if (response.status == 400) {
      errorMsg = 'the link you are using expired or the request is not valid, please try again.';
    } else {
      errorMsg = error;
    }
    dispatch(setSnackbar('The password reset request cannot be processed: ' + errorMsg));
    return Promise.reject(err);
  })
);

export const verifyEmailStart = createAsyncThunk(`${sliceName}/verifyEmailStart`, (_, { dispatch, getState }) =>
  GeneralApi.post(`${useradmApiUrl}/auth/verify-email/start`, { email: getCurrentUser(getState()).email })
    .catch(err => commonErrorHandler(err, 'An error occured starting the email verification process:', dispatch))
    .finally(() => Promise.resolve(dispatch(getUser(OWN_USER_ID))))
);

export const verifyEmailComplete = createAsyncThunk(`${sliceName}/verifyEmailComplete`, (secret_hash, { dispatch }) =>
  GeneralApi.post(`${useradmApiUrl}/auth/verify-email/complete`, { secret_hash })
    .catch(err => commonErrorHandler(err, 'An error occured completing the email verification process:', dispatch))
    .finally(() => Promise.resolve(dispatch(getUser(OWN_USER_ID))))
);

export const verify2FA = createAsyncThunk(`${sliceName}/verify2FA`, (tfaData, { dispatch }) =>
  UsersApi.putVerifyTFA(`${useradmApiUrl}/2faverify`, tfaData)
    .then(() => Promise.resolve(dispatch(getUser(OWN_USER_ID))))
    .catch(err =>
      commonErrorHandler(err, 'An error occured validating the verification code: failed to verify token, please try again.', dispatch, undefined, true)
    )
);

export const getUserList = createAsyncThunk(`${sliceName}/getUserList`, (_, { dispatch, getState }) =>
  GeneralApi.get(`${useradmApiUrl}/users`)
    .then(res => {
      const currentUsersById = getUsersById(getState());
      const users = res.data.reduce(
        (accu, item) => {
          accu[item.id] = {
            ...accu[item.id],
            ...item
          };
          return accu;
        },
        { ...currentUsersById }
      );
      return dispatch(actions.receivedUserList(users));
    })
    .catch(err => commonErrorHandler(err, `Users couldn't be loaded.`, dispatch, commonErrorFallback))
);

export const getUser = createAsyncThunk(`${sliceName}/getUser`, (id, { dispatch, rejectWithValue }) =>
  GeneralApi.get(`${useradmApiUrl}/users/${id}`)
    .then(({ data: user }) =>
      Promise.all([
        dispatch(actions.receivedUser(user)),
        dispatch(setHideAnnouncement({ shouldHide: false, userId: user.id })),
        dispatch(updateUserColumnSettings({ currentUserId: user.id })),
        user
      ])
    )
    .catch(e => rejectWithValue(e))
);

export const initializeSelf = createAsyncThunk(`${sliceName}/initializeSelf`, (_, { dispatch }) => dispatch(getUser(OWN_USER_ID)));

export const updateUserColumnSettings = createAsyncThunk(`${sliceName}/updateUserColumnSettings`, ({ columns, currentUserId }, { dispatch, getState }) => {
  const userId = currentUserId ?? getCurrentUser(getState()).id;
  const storageKey = `${userId}-column-widths`;
  let customColumns = [];
  if (!columns) {
    try {
      customColumns = JSON.parse(window.localStorage.getItem(storageKey)) || customColumns;
    } catch {
      // most likely the column info doesn't exist yet or is lost - continue
    }
  } else {
    customColumns = columns;
  }
  window.localStorage.setItem(storageKey, JSON.stringify(customColumns));
  return Promise.resolve(dispatch(actions.setCustomColumns(customColumns)));
});

const userActions = {
  add: {
    successMessage: 'The user was added successfully.',
    errorMessage: 'adding'
  },
  create: {
    successMessage: 'The user was created successfully.',
    errorMessage: 'creating'
  },
  edit: {
    successMessage: 'The user has been updated.',
    errorMessage: 'editing'
  },
  remove: {
    successMessage: 'The user was removed from the system.',
    errorMessage: 'removing'
  }
};

const userActionErrorHandler = (err, type, dispatch) => commonErrorHandler(err, `There was an error ${userActions[type].errorMessage} the user.`, dispatch);

export const createUser = createAsyncThunk(`${sliceName}/createUser`, ({ shouldResetPassword, ...userData }, { dispatch }) =>
  GeneralApi.post(`${useradmApiUrl}/users`, { ...userData, send_reset_password: shouldResetPassword })
    .then(() => Promise.all([dispatch(actions.createdUser(userData)), dispatch(getUserList()), dispatch(setSnackbar(userActions.create.successMessage))]))
    .catch(err => userActionErrorHandler(err, 'create', dispatch))
);

export const removeUser = createAsyncThunk(`${sliceName}/removeUser`, (userId, { dispatch }) =>
  GeneralApi.delete(`${useradmApiUrl}/users/${userId}`)
    .then(() => Promise.all([dispatch(actions.removedUser(userId)), dispatch(getUserList()), dispatch(setSnackbar(userActions.remove.successMessage))]))
    .catch(err => userActionErrorHandler(err, 'remove', dispatch))
);

export const editUser = createAsyncThunk(`${sliceName}/editUser`, ({ id, ...userData }, { dispatch, getState }) =>
  GeneralApi.put(`${useradmApiUrl}/users/${id}`, userData).then(() =>
    Promise.all([
      dispatch(actions.updatedUser({ ...userData, id: id === OWN_USER_ID ? getCurrentUser(getState()).id : id })),
      dispatch(setSnackbar(userActions.edit.successMessage))
    ])
  )
);

export const addUserToCurrentTenant = createAsyncThunk(`${sliceName}/addUserToTenant`, (userId, { dispatch, getState }) => {
  const { id } = getOrganization(getState());
  return GeneralApi.post(`${useradmApiUrl}/users/${userId}/assign`, { tenant_ids: [id] })
    .catch(err => commonErrorHandler(err, `There was an error adding the user to your organization:`, dispatch))
    .then(() => Promise.all([dispatch(setSnackbar(userActions.add.successMessage)), dispatch(getUserList())]));
});

export const enableUser2fa = createAsyncThunk(`${sliceName}/enableUser2fa`, (userId = OWN_USER_ID, { dispatch }) =>
  GeneralApi.post(`${useradmApiUrl}/users/${userId}/2fa/enable`)
    .catch(err => commonErrorHandler(err, `There was an error enabling Two Factor authentication for the user.`, dispatch))
    .then(() => Promise.resolve(dispatch(getUser(userId))))
);

export const disableUser2fa = createAsyncThunk(`${sliceName}/disableUser2fa`, (userId = OWN_USER_ID, { dispatch }) =>
  GeneralApi.post(`${useradmApiUrl}/users/${userId}/2fa/disable`)
    .catch(err => commonErrorHandler(err, `There was an error disabling Two Factor authentication for the user.`, dispatch))
    .then(() => Promise.all([dispatch(getUser(userId)), dispatch(actions.receivedQrCode(null))]))
);

/* RBAC related things follow:  */
const mapHttpPermission = permission =>
  Object.entries(uiPermissionsByArea).reduce(
    (accu, [area, definition]) => {
      const endpointMatches = definition.endpoints.filter(
        endpoint => endpoint.path.test(permission.value) && (endpoint.types.includes(permission.type) || permission.type === PermissionTypes.Any)
      );
      if (permission.value === PermissionTypes.Any || (permission.value.includes(apiRoot) && endpointMatches.length)) {
        const endpointUiPermission = endpointMatches.reduce((endpointAccu, endpoint) => [...endpointAccu, ...endpoint.uiPermissions], []);
        const collector = (endpointUiPermission || definition.uiPermissions)
          .reduce((permissionsAccu, uiPermission) => {
            if (permission.type === PermissionTypes.Any || (!endpointMatches.length && uiPermission.verbs.some(verb => verb === permission.type))) {
              permissionsAccu.push(uiPermission.value);
            }
            return permissionsAccu;
          }, [])
          .filter(duplicateFilter);
        if (Array.isArray(accu[area])) {
          accu[area] = [...accu[area], ...collector].filter(duplicateFilter);
        } else {
          accu[area] = mergePermissions(accu[area], { [scopedPermissionAreas[area].excessiveAccessSelector]: collector });
        }
      }
      return accu;
    },
    { ...emptyUiPermissions }
  );

const permissionActionTypes = {
  any: mapHttpPermission,
  CREATE_DEPLOYMENT: permission =>
    permission.type === PermissionTypes.DeviceGroup
      ? {
          deployments: [uiPermissionsById.deploy.value],
          groups: { [permission.value]: [uiPermissionsById.deploy.value] }
        }
      : {},
  http: mapHttpPermission,
  REMOTE_TERMINAL: permission =>
    permission.type === PermissionTypes.DeviceGroup
      ? {
          groups: { [permission.value]: [uiPermissionsById.connect.value] }
        }
      : {},
  VIEW_DEVICE: permission =>
    permission.type === PermissionTypes.DeviceGroup
      ? {
          groups: { [permission.value]: [uiPermissionsById.read.value] }
        }
      : {}
};

const combinePermissions = (existingPermissions, additionalPermissions = {}) =>
  Object.entries(additionalPermissions).reduce((accu, [name, permissions]) => {
    let maybeExistingPermissions = accu[name] || [];
    accu[name] = [...permissions, ...maybeExistingPermissions].filter(duplicateFilter);
    return accu;
  }, existingPermissions);

const tryParseCustomPermission = permission => {
  const uiPermissions = permissionActionTypes[permission.action](permission.object);
  const result = mergePermissions({ ...emptyUiPermissions }, uiPermissions);
  return { isCustom: true, permission, result };
};

const customPermissionHandler = (accu, permission) => {
  let processor = tryParseCustomPermission(permission);
  return {
    ...accu,
    isCustom: accu.isCustom || processor.isCustom,
    uiPermissions: mergePermissions(accu.uiPermissions, processor.result)
  };
};

const mapPermissionSet = (permissionSetName, names, scope, existingGroupsPermissions = {}) => {
  const permission = Object.values(uiPermissionsById).find(permission => permission.permissionSets[scope] === permissionSetName).value;
  const scopedPermissions = names.reduce((accu, name) => combinePermissions(accu, { [name]: [permission] }), existingGroupsPermissions);
  return Object.entries(scopedPermissions).reduce((accu, [key, permissions]) => ({ ...accu, [key]: deriveImpliedAreaPermissions(scope, permissions) }), {});
};

const isEmptyPermissionSet = permissionSet =>
  !Object.values(permissionSet).reduce((accu, permissions) => {
    if (Array.isArray(permissions)) {
      return accu || !!permissions.length;
    }
    return accu || !isEmpty(permissions);
  }, false);

const parseRolePermissions = ({ permission_sets_with_scope = [], permissions = [] }, permissionSets) => {
  const preliminaryResult = permission_sets_with_scope.reduce(
    (accu, permissionSet) => {
      let processor = permissionSets[permissionSet.name];
      if (!processor) {
        return accu;
      }
      const scope = Object.keys(scopedPermissionAreas).find(scope => uiPermissionsByArea[scope].scope === permissionSet.scope?.type);
      if (scope) {
        const result = mapPermissionSet(permissionSet.name, permissionSet.scope.value, scope, accu.uiPermissions[scope]);
        return { ...accu, uiPermissions: { ...accu.uiPermissions, [scope]: result } };
      } else if (isEmptyPermissionSet(processor.result)) {
        return processor.permissions.reduce(customPermissionHandler, accu);
      }
      return {
        ...accu,
        isCustom: accu.isCustom || processor.isCustom,
        uiPermissions: mergePermissions(accu.uiPermissions, processor.result)
      };
    },
    { isCustom: false, uiPermissions: { ...emptyUiPermissions, groups: {}, releases: {} } }
  );
  return permissions.reduce(customPermissionHandler, preliminaryResult);
};

export const normalizeRbacRoles = (roles, rolesById, permissionSets) =>
  roles.reduce(
    (accu, role) => {
      let normalizedPermissions;
      let isCustom = false;
      if (rolesById[role.name]) {
        normalizedPermissions = {
          ...rolesById[role.name].uiPermissions,
          groups: { ...rolesById[role.name].uiPermissions.groups },
          releases: { ...rolesById[role.name].uiPermissions.releases }
        };
      } else {
        const result = parseRolePermissions(role, permissionSets);
        normalizedPermissions = result.uiPermissions;
        isCustom = result.isCustom;
      }

      const roleState = accu[role.name] ?? { ...emptyRole };
      accu[role.name] = {
        ...roleState,
        ...role,
        description: roleState.description ? roleState.description : role.description,
        editable: !defaultRolesById[role.name] && !isCustom && (typeof roleState.editable !== 'undefined' ? roleState.editable : true),
        isCustom,
        name: roleState.name ? roleState.name : role.name,
        uiPermissions: normalizedPermissions
      };
      return accu;
    },
    { ...rolesById }
  );

export const getPermissionSets = createAsyncThunk(`${sliceName}/getPermissionSets`, (_, { dispatch, getState }) =>
  GeneralApi.get(`${useradmApiUrlv2}/permission_sets?per_page=500`)
    .then(({ data }) => {
      const permissionSets = data.reduce(
        (accu, permissionSet) => {
          const permissionSetState = accu[permissionSet.name] ?? {};
          let permissionSetObject = { ...permissionSetState, ...permissionSet };
          permissionSetObject.result = Object.values(uiPermissionsById).reduce(
            (accu, item) =>
              Object.entries(item.permissionSets).reduce((collector, [area, permissionSet]) => {
                if (scopedPermissionAreas[area]) {
                  return collector;
                }
                if (permissionSet === permissionSetObject.name) {
                  collector[area] = [...collector[area], item.value].filter(duplicateFilter);
                }
                return collector;
              }, accu),
            { ...emptyUiPermissions, ...(permissionSetObject.result ?? {}) }
          );
          const scopes = Object.values(scopedPermissionAreas).reduce((accu, { key, scopeType }) => {
            if (permissionSetObject.supported_scope_types?.includes(key) || permissionSetObject.supported_scope_types?.includes(scopeType)) {
              accu.push(key);
            }
            return accu;
          }, []);
          permissionSetObject = scopes.reduce((accu, scope) => {
            accu.result[scope] = mapPermissionSet(permissionSetObject.name, [scopedPermissionAreas[scope].excessiveAccessSelector], scope);
            return accu;
          }, permissionSetObject);
          accu[permissionSet.name] = permissionSetObject;
          return accu;
        },
        { ...getState().users.permissionSetsById }
      );
      return Promise.all([dispatch(actions.receivedPermissionSets(permissionSets)), permissionSets]);
    })
    .catch(() => console.log('Permission set retrieval failed - likely accessing a non-RBAC backend'))
);

export const getRoles = createAsyncThunk(`${sliceName}/getRoles`, (_, { dispatch, getState }) =>
  Promise.all([GeneralApi.get(`${useradmApiUrlv2}/roles?per_page=500`), dispatch(getPermissionSets())])
    .then(results => {
      if (!results) {
        return Promise.resolve();
      }
      const [{ data: roles }, { payload: permissionSetTasks }] = results;
      const rolesById = normalizeRbacRoles(roles, getRolesById(getState()), permissionSetTasks[permissionSetTasks.length - 1]);
      return Promise.resolve(dispatch(actions.receivedRoles(rolesById)));
    })
    .catch(() => console.log('Role retrieval failed - likely accessing a non-RBAC backend'))
);

const deriveImpliedAreaPermissions = (area, areaPermissions, skipPermissions = []) => {
  const highestAreaPermissionLevelSelected = areaPermissions.reduce(
    (highest, current) => (uiPermissionsById[current].permissionLevel > highest ? uiPermissionsById[current].permissionLevel : highest),
    1
  );
  return uiPermissionsByArea[area].uiPermissions.reduce((permissions, current) => {
    if ((current.permissionLevel < highestAreaPermissionLevelSelected || areaPermissions.includes(current.value)) && !skipPermissions.includes(current.value)) {
      permissions.push(current.value);
    }
    return permissions;
  }, []);
};

/**
 * transforms [{ group: "groupName",  uiPermissions: ["read", "manage", "connect"] }, ...] to
 * [{ name: "ReadDevices", scope: { type: "DeviceGroups", value: ["groupName", ...] } }, ...]
 */
const transformAreaRoleDataToScopedPermissionsSets = (area, areaPermissions, excessiveAccessSelector) => {
  const permissionSetObject = areaPermissions.reduce((accu, { item, uiPermissions }) => {
    // if permission area is release and item is release tag (not all releases) then exclude upload permission as it cannot be applied to tags
    const skipPermissions = scopedPermissionAreas.releases.key === area && item !== ALL_RELEASES ? [uiPermissionsById.upload.value] : [];
    const impliedPermissions = deriveImpliedAreaPermissions(area, uiPermissions, skipPermissions);
    accu = impliedPermissions.reduce((itemPermissionAccu, impliedPermission) => {
      const permissionSetState = itemPermissionAccu[uiPermissionsById[impliedPermission].permissionSets[area]] ?? {
        type: uiPermissionsByArea[area].scope,
        value: []
      };
      itemPermissionAccu[uiPermissionsById[impliedPermission].permissionSets[area]] = {
        ...permissionSetState,
        value: [...permissionSetState.value, item]
      };
      return itemPermissionAccu;
    }, accu);
    return accu;
  }, {});
  return Object.entries(permissionSetObject).map(([name, { value, ...scope }]) => {
    if (value.includes(excessiveAccessSelector)) {
      return { name };
    }
    return { name, scope: { ...scope, value: value.filter(duplicateFilter) } };
  });
};

const transformRoleDataToRole = (roleData, roleState = {}) => {
  const role = { ...roleState, ...roleData };
  const { description = '', name, uiPermissions = emptyUiPermissions } = role;
  const { maybeUiPermissions, remainderKeys } = Object.entries(emptyUiPermissions).reduce(
    (accu, [key, emptyPermissions]) => {
      if (!scopedPermissionAreas[key]) {
        accu.remainderKeys.push(key);
      } else if (uiPermissions[key]) {
        accu.maybeUiPermissions[key] = uiPermissions[key].reduce(itemUiPermissionsReducer, emptyPermissions);
      }
      return accu;
    },
    { maybeUiPermissions: {}, remainderKeys: [] }
  );
  const { permissionSetsWithScope, roleUiPermissions } = remainderKeys.reduce(
    (accu, area) => {
      const areaPermissions = role.uiPermissions[area];
      if (!Array.isArray(areaPermissions)) {
        return accu;
      }
      const impliedPermissions = deriveImpliedAreaPermissions(area, areaPermissions);
      accu.roleUiPermissions[area] = impliedPermissions;
      const mappedPermissions = impliedPermissions.map(uiPermission => ({ name: uiPermissionsById[uiPermission].permissionSets[area] }));
      accu.permissionSetsWithScope.push(...mappedPermissions);
      return accu;
    },
    { permissionSetsWithScope: [{ name: defaultPermissionSets.Basic.name }], roleUiPermissions: {} }
  );
  const scopedPermissionSets = Object.values(scopedPermissionAreas).reduce((accu, { key, excessiveAccessSelector }) => {
    if (!uiPermissions[key]) {
      return accu;
    }
    accu.push(...transformAreaRoleDataToScopedPermissionsSets(key, uiPermissions[key], excessiveAccessSelector));
    return accu;
  }, []);
  return {
    permissionSetsWithScope: [...permissionSetsWithScope, ...scopedPermissionSets],
    role: {
      ...emptyRole,
      name,
      description: description ? description : roleState.description,
      uiPermissions: {
        ...emptyUiPermissions,
        ...roleUiPermissions,
        ...maybeUiPermissions
      }
    }
  };
};

const roleActions = {
  create: {
    successMessage: 'The role was created successfully.',
    errorMessage: 'creating'
  },
  edit: {
    successMessage: 'The role has been updated.',
    errorMessage: 'editing'
  },
  remove: {
    successMessage: 'The role was deleted successfully.',
    errorMessage: 'removing'
  }
};

const roleActionErrorHandler = (err, type, dispatch) => commonErrorHandler(err, `There was an error ${roleActions[type].errorMessage} the role.`, dispatch);

export const createRole = createAsyncThunk(`${sliceName}/createRole`, (roleData, { dispatch }) => {
  const { permissionSetsWithScope, role } = transformRoleDataToRole(roleData);
  return GeneralApi.post(`${useradmApiUrlv2}/roles`, {
    name: role.name,
    description: role.description,
    permission_sets_with_scope: permissionSetsWithScope
  })
    .then(() => Promise.all([dispatch(actions.createdRole(role)), dispatch(getRoles()), dispatch(setSnackbar(roleActions.create.successMessage))]))
    .catch(err => roleActionErrorHandler(err, 'create', dispatch));
});

export const editRole = createAsyncThunk(`${sliceName}/editRole`, (roleData, { dispatch, getState }) => {
  const { permissionSetsWithScope, role } = transformRoleDataToRole(roleData, getRolesById(getState())[roleData.name]);
  return GeneralApi.put(`${useradmApiUrlv2}/roles/${role.name}`, {
    description: role.description,
    name: role.name,
    permission_sets_with_scope: permissionSetsWithScope
  })
    .then(() => Promise.all([dispatch(actions.createdRole(role)), dispatch(getRoles()), dispatch(setSnackbar(roleActions.edit.successMessage))]))
    .catch(err => roleActionErrorHandler(err, 'edit', dispatch));
});

export const removeRole = createAsyncThunk(`${sliceName}/removeRole`, (roleId, { dispatch }) =>
  GeneralApi.delete(`${useradmApiUrlv2}/roles/${roleId}`)
    .then(() => Promise.all([dispatch(actions.removedRole(roleId)), dispatch(getRoles()), dispatch(setSnackbar(roleActions.remove.successMessage))]))
    .catch(err => roleActionErrorHandler(err, 'remove', dispatch))
);

/*
  Global settings
*/
export const getGlobalSettings = createAsyncThunk(`${sliceName}/getGlobalSettings`, (_, { dispatch }) =>
  GeneralApi.get(`${useradmApiUrl}/settings`).then(({ data: settings, headers: { etag } }) => {
    window.sessionStorage.setItem(settingsKeys.initialized, true);
    return Promise.all([dispatch(actions.setGlobalSettings(settings)), dispatch(setOfflineThreshold()), etag]);
  })
);

export const saveGlobalSettings = createAsyncThunk(
  `${sliceName}/saveGlobalSettings`,
  ({ beOptimistic = false, notify = false, ...settings }, { dispatch, getState }) => {
    if (!window.sessionStorage.getItem(settingsKeys.initialized) && !beOptimistic) {
      return;
    }
    return dispatch(getGlobalSettings())
      .unwrap()
      .then(result => {
        let updatedSettings = { ...getState().users.globalSettings, ...settings };
        if (getCurrentUser(getState()).verified) {
          updatedSettings['2fa'] = twoFAStates.enabled;
        } else {
          delete updatedSettings['2fa'];
        }
        let tasks = [dispatch(actions.setGlobalSettings(updatedSettings))];
        const headers = result[result.length - 1] ? { 'If-Match': result[result.length - 1] } : {};
        return GeneralApi.post(`${useradmApiUrl}/settings`, updatedSettings, { headers })
          .then(() => {
            if (notify) {
              tasks.push(dispatch(setSnackbar('Settings saved successfully')));
            }
            return Promise.all(tasks);
          })
          .catch(err => {
            if (beOptimistic) {
              return Promise.all([tasks]);
            }
            console.log(err);
            return commonErrorHandler(err, `The settings couldn't be saved.`, dispatch);
          });
      });
  }
);

export const getUserSettings = createAsyncThunk(`${sliceName}/getUserSettings`, (_, { dispatch }) =>
  GeneralApi.get(`${useradmApiUrl}/settings/me`).then(({ data: settings, headers: { etag } }) => {
    window.sessionStorage.setItem(settingsKeys.initialized, true);
    return Promise.all([dispatch(actions.setUserSettings(settings)), etag]);
  })
);

export const saveUserSettings = createAsyncThunk(`${sliceName}/saveUserSettings`, (settings = { onboarding: {} }, { dispatch, getState }) => {
  if (!getCurrentUser(getState()).id) {
    return Promise.resolve();
  }
  return dispatch(getUserSettings())
    .unwrap()
    .then(result => {
      const userSettings = getUserSettingsSelector(getState());
      const onboardingState = getOnboardingState(getState());
      const tooltipState = getTooltipsState(getState());
      const updatedSettings = {
        ...userSettings,
        ...settings,
        onboarding: {
          ...onboardingState,
          ...settings.onboarding
        },
        tooltips: tooltipState
      };
      const headers = result[result.length - 1] ? { 'If-Match': result[result.length - 1] } : {};
      return Promise.all([
        Promise.resolve(dispatch(actions.setUserSettings(updatedSettings))),
        GeneralApi.post(`${useradmApiUrl}/settings/me`, updatedSettings, { headers })
      ]).catch(() => dispatch(actions.setUserSettings(userSettings)));
    });
});

export const get2FAQRCode = createAsyncThunk(`${sliceName}/get2FAQRCode`, (_, { dispatch }) =>
  GeneralApi.get(`${useradmApiUrl}/2faqr`).then(res => dispatch(actions.receivedQrCode(res.data.qr)))
);

export const setHideAnnouncement = createAsyncThunk(`${sliceName}/setHideAnnouncement`, ({ shouldHide, userId }, { dispatch, getState }) => {
  const currentUserId = userId || getCurrentUser(getState()).id;
  const hash = getState().app.hostedAnnouncement ? hashString(getState().app.hostedAnnouncement) : '';
  const announceCookie = cookies.get(`${currentUserId}${hash}`);
  if (shouldHide || (hash.length && typeof announceCookie !== 'undefined')) {
    cookies.set(`${currentUserId}${hash}`, true, { maxAge: 604800 });
    return Promise.resolve(dispatch(setAnnouncement()));
  }
  return Promise.resolve();
});

export const getTokens = createAsyncThunk(`${sliceName}/getTokens`, (_, { dispatch, getState }) =>
  GeneralApi.get(`${useradmApiUrl}/settings/tokens`).then(({ data: tokens }) => {
    const user = getCurrentUser(getState());
    const updatedUser = {
      ...user,
      tokens
    };
    return Promise.resolve(dispatch(actions.updatedUser(updatedUser)));
  })
);

const ONE_YEAR = 31536000;

export const generateToken = createAsyncThunk(`${sliceName}/generateToken`, ({ expiresIn = ONE_YEAR, name }, { dispatch }) =>
  GeneralApi.post(`${useradmApiUrl}/settings/tokens`, { name, expires_in: expiresIn })
    .then(({ data: token }) => Promise.all([dispatch(getTokens()), token]))
    .catch(err => commonErrorHandler(err, 'There was an error creating the token:', dispatch))
);

export const revokeToken = createAsyncThunk(`${sliceName}/revokeToken`, (token, { dispatch }) =>
  GeneralApi.delete(`${useradmApiUrl}/settings/tokens/${token.id}`).then(() => Promise.resolve(dispatch(getTokens())))
);

export const setTooltipReadState = createAsyncThunk(`${sliceName}/setTooltipReadState`, ({ persist, ...remainder }, { dispatch }) => {
  let tasks = [dispatch(actions.setTooltipState(remainder))];
  if (persist) {
    tasks.push(dispatch(saveUserSettings()));
  }
  return Promise.all(tasks);
});

export const setAllTooltipsReadState = createAsyncThunk(`${sliceName}/toggleHelptips`, (readState = READ_STATES.read, { dispatch }) => {
  const updatedTips = Object.keys(HELPTOOLTIPS).reduce((accu, id) => ({ ...accu, [id]: { readState } }), {});
  return Promise.resolve(dispatch(actions.setTooltipsState(updatedTips))).then(() => dispatch(saveUserSettings()));
});

export const submitFeedback = createAsyncThunk(`${sliceName}/submitFeedback`, ({ satisfaction, feedback, ...meta }, { dispatch }) =>
  GeneralApi.post(`${tenantadmApiUrlv2}/contact/support`, {
    subject: 'feedback submission',
    body: JSON.stringify({ feedback, satisfaction, meta })
  }).then(() => {
    const today = new Date();
    dispatch(saveUserSettings({ feedbackCollectedAt: today.toISOString().split('T')[0] }));
    setTimeout(() => dispatch(actions.setShowFeedbackDialog(false)), TIMEOUTS.threeSeconds);
  })
);
