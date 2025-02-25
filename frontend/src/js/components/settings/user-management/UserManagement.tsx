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
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { Add as AddIcon } from '@mui/icons-material';
// material ui
import { Button, Chip, DialogActions, DialogContent } from '@mui/material';

import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import storeActions from '@northern.tech/store/actions';
import { getCurrentUser, getFeatures, getIsEnterprise, getRelevantRoles, getUserCapabilities, getUsersList } from '@northern.tech/store/selectors';
import { addUserToCurrentTenant, createUser, editUser, getUserList, passwordResetStart, removeUser } from '@northern.tech/store/thunks';

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

const DeleteUserDialog = ({ dismiss, open, submit, user }) => (
  <BaseDialog title="Delete user?" open={open} onClose={dismiss}>
    <DialogContent style={{ overflow: 'hidden' }}>
      Are you sure you want to delete the user with email{' '}
      <b>
        <i>{user.email}</i>
      </b>
      ?
    </DialogContent>
    <DialogActions>
      <Button style={{ marginRight: 10 }} onClick={dismiss}>
        Cancel
      </Button>
      <Button variant="contained" color="primary" onClick={() => submit(user, 'remove', user.id)}>
        Delete user
      </Button>
    </DialogActions>
  </BaseDialog>
);

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
    users
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

  const submit = (userData, type, id, passwordResetEmail) => {
    if (userData) {
      let request = null;
      if (id) {
        request = props[actions[type]](id, userData);
      } else {
        request = props[actions[type]](userData);
      }
      return request.then(() => {
        if (passwordResetEmail) {
          dispatch(passwordResetStart(passwordResetEmail));
        }
        dialogDismiss();
      });
    }
    if (passwordResetEmail) {
      dispatch(passwordResetStart(passwordResetEmail));
    }
    return dialogDismiss();
  };

  return (
    <div>
      <div className="flexbox centered space-between" style={{ marginLeft: '20px' }}>
        <h2>Users</h2>
      </div>

      <UserList {...props} editUser={openEdit} />
      <Chip color="primary" icon={<AddIcon />} label="Add new user" onClick={setShowCreate} />
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
      <DeleteUserDialog dismiss={dialogDismiss} open={removeDialog} submit={submit} user={user} />
    </div>
  );
};

export default UserManagement;
