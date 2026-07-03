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
import { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { Add as AddIcon } from '@mui/icons-material';
// material ui
import { Button, Typography } from '@mui/material';

import { ConfirmModal } from '@northern.tech/common-ui/ConfirmModal';
import storeActions from '@northern.tech/store/actions';
import {
  getCurrentUser,
  getFeatures,
  getIsEnterprise,
  getOrganization,
  getRelevantRoles,
  getUserCapabilities,
  getUsersList
} from '@northern.tech/store/selectors';
import { addUserToCurrentTenant, createUser, editUser, getUserList, removeUser } from '@northern.tech/store/thunks';

import { EmailVerificationWarning } from '../EmailVerificationWarning';
import { UserDefinition } from './UserDefinition';
import UserForm from './UserForm';
import UserList from './UserList';

const { setSnackbar } = storeActions;

const actions = {
  add: 'addUser',
  create: 'createUser',
  edit: 'editUser',
  remove: 'removeUser'
};

export const UserManagement = () => {
  const [showCreate, setShowCreate] = useState(false);
  const [removeDialog, setRemoveDialog] = useState(false);
  const [user, setUser] = useState({});
  const dispatch = useDispatch();

  const { canManageUsers } = useSelector(getUserCapabilities);
  const { isHosted } = useSelector(getFeatures);
  const isEnterprise = useSelector(getIsEnterprise);
  const currentUser = useSelector(getCurrentUser);
  const roles = useSelector(getRelevantRoles);
  const users = useSelector(getUsersList);
  const { trial: isTrial } = useSelector(getOrganization);
  const props = {
    canManageUsers,
    addUser: id => dispatch(addUserToCurrentTenant(id)),
    createUser: userData => dispatch(createUser(userData)),
    currentUser,
    editUser: (id, userData) => dispatch(editUser({ ...userData, id })),
    isEnterprise,
    isHosted,
    removeUser: id => dispatch(removeUser(id)),
    roles,
    users,
    isTrial
  };

  useEffect(() => {
    dispatch(getUserList());
  }, [dispatch]);

  const openEdit = user => {
    setUser(user);
    setRemoveDialog(false);
    dispatch(setSnackbar(''));
  };

  const openRemove = () => {
    dispatch(setSnackbar(''));
    setRemoveDialog(true);
  };

  const dialogDismiss = () => {
    setUser({});
    setShowCreate(false);
    setRemoveDialog(false);
  };

  const submit = async (userData, type, id) => {
    try {
      if (id) {
        await props[actions[type]](id, userData).unwrap();
      } else {
        await props[actions[type]](userData).unwrap();
      }
    } catch {
      // error already handled in thunk - leave open
    }
  };

  return (
    <div>
      <div className="flexbox space-between align-items-center margin-bottom-medium">
        <Typography variant="h6">Users</Typography>
        <Button color="primary" startIcon={<AddIcon />} onClick={setShowCreate} disabled={!currentUser.verified} variant="contained">
          Add new user
        </Button>
      </div>
      {!currentUser.verified && <EmailVerificationWarning action="add a new user" />}
      <UserList {...props} editUser={openEdit} />
      {showCreate && <UserForm {...props} closeDialog={dialogDismiss} submit={submit} />}
      <UserDefinition
        currentUser={currentUser}
        isEnterprise={isEnterprise}
        onRemove={openRemove}
        onCancel={dialogDismiss}
        onSubmit={submit}
        roles={roles}
        selectedUser={user}
      />
      <ConfirmModal
        header="Delete user?"
        description={
          <>
            Are you sure you want to delete the user with email <b>{user.email}</b>?
          </>
        }
        confirmButtonText="Delete user"
        open={removeDialog}
        close={dialogDismiss}
        onConfirm={() => submit(user, 'remove', user.id)}
      />
    </div>
  );
};

export default UserManagement;
