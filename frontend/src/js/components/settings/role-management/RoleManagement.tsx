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
import React, { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';

// material ui
import { Add as AddIcon } from '@mui/icons-material';
import { Chip } from '@mui/material';

import DetailsIndicator from '@northern.tech/common-ui/DetailsIndicator';
import DetailsTable from '@northern.tech/common-ui/DetailsTable';
import { DOCSTIPS, DocsTooltip } from '@northern.tech/common-ui/DocsLink';
import EnterpriseNotification from '@northern.tech/common-ui/EnterpriseNotification';
import { InfoHintContainer } from '@northern.tech/common-ui/InfoHint';
import { Loader } from '@northern.tech/common-ui/Loader';
import { BENEFITS, UiRoleDefinition, emptyRole, settingsKeys } from '@northern.tech/store/constants';
import {
  getGroupsByIdWithoutUngrouped,
  getIsEnterprise,
  getOrganization,
  getReleaseTagsById,
  getRelevantRoles,
  getRolesInitialized
} from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { createRole, editRole, getDynamicGroups, getExistingReleaseTags, getGroups, getRoles, removeRole } from '@northern.tech/store/thunks';

import RoleDefinition from './RoleDefinition';

const columns = [
  { key: 'name', title: 'Role', render: ({ name }) => name },
  { key: 'description', title: 'Description', render: ({ description }) => description || '-' },
  { key: 'manage', title: 'Manage', render: DetailsIndicator }
];

export const RoleManagement = () => {
  const [adding, setAdding] = useState<boolean>(false);
  const [editing, setEditing] = useState<boolean>(false);
  const [role, setRole] = useState<UiRoleDefinition>({ ...emptyRole });
  const dispatch = useAppDispatch();
  const groups = useSelector(getGroupsByIdWithoutUngrouped);
  const releaseTags = useSelector(getReleaseTagsById);
  const isEnterprise = useSelector(getIsEnterprise);
  const { service_provider } = useSelector(getOrganization);
  const items = useSelector(getRelevantRoles);
  const isLikelyInitialized = window.sessionStorage.getItem(settingsKeys.initialized);
  const rolesInitialized = useSelector(getRolesInitialized);
  useEffect(() => {
    if (service_provider || !isLikelyInitialized) {
      return;
    }
    dispatch(getExistingReleaseTags());
  }, [dispatch, isLikelyInitialized, service_provider]);

  useEffect(() => {
    if (Object.keys(groups).length || service_provider || !isLikelyInitialized) {
      return;
    }
    dispatch(getDynamicGroups());
    dispatch(getGroups());
    dispatch(getRoles());
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dispatch, isLikelyInitialized, JSON.stringify(groups), service_provider]);

  const addRole = () => {
    setAdding(true);
    setEditing(false);
    setRole({ ...emptyRole });
  };

  const onEditRole = editedRole => {
    setAdding(false);
    setEditing(true);
    setRole(editedRole);
  };

  const onCancel = () => {
    setAdding(false);
    setEditing(false);
  };

  const onSubmit = submittedRole => {
    let action = editRole(submittedRole);
    if (adding) {
      action = createRole(submittedRole);
    }
    dispatch(action)
      .unwrap()
      .then(() => onCancel());
  };

  return (
    <div>
      <div className="flexbox center-aligned">
        <h2 style={{ marginLeft: 20 }}>Roles</h2>
        <InfoHintContainer>
          <EnterpriseNotification id={BENEFITS.rbac.id} />
          <DocsTooltip id={DOCSTIPS.rbac.id} />
        </InfoHintContainer>
      </div>
      {rolesInitialized ? <DetailsTable columns={columns} items={items} onItemClick={onEditRole} /> : <Loader show={true} />}
      <Chip color="primary" icon={<AddIcon />} label="Add a role" onClick={addRole} disabled={!isEnterprise} />
      <RoleDefinition
        adding={adding}
        editing={editing}
        isServiceProvider={!!service_provider}
        onCancel={onCancel}
        onSubmit={onSubmit}
        removeRole={name => dispatch(removeRole(name))}
        selectedRole={role}
        stateGroups={groups}
        stateReleaseTags={releaseTags}
      />
    </div>
  );
};

export default RoleManagement;
