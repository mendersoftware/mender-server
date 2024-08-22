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
import React, { useEffect, useMemo, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

// material ui
import { Add as AddIcon, ArrowRightAlt as ArrowRightAltIcon } from '@mui/icons-material';
import { Chip } from '@mui/material';

import { getDynamicGroups, getGroups } from '../../actions/deviceActions';
import { getExistingReleaseTags } from '../../actions/releaseActions.js';
import { createRole, editRole, getRoles, removeRole } from '../../actions/userActions';
import { BENEFITS } from '../../constants/appConstants';
import { emptyRole, rolesById } from '../../constants/userConstants';
import { getGroupsByIdWithoutUngrouped, getIsEnterprise, getReleaseTagsById, getRolesList } from '../../selectors';
import DetailsTable from '../common/detailstable';
import { DocsTooltip } from '../common/docslink';
import EnterpriseNotification from '../common/enterpriseNotification';
import { InfoHintContainer } from '../common/info-hint';
import RoleDefinition from './roledefinition';

const columns = [
  { key: 'name', title: 'Role', render: ({ name }) => name },
  { key: 'description', title: 'Description', render: ({ description }) => description || '-' },
  {
    key: 'manage',
    title: 'Manage',
    render: () => (
      <div className="bold flexbox center-aligned link-color margin-right-small uppercased" style={{ whiteSpace: 'nowrap' }}>
        view details <ArrowRightAltIcon />
      </div>
    )
  }
];

export const RoleManagement = () => {
  const [adding, setAdding] = useState(false);
  const [editing, setEditing] = useState(false);
  const [role, setRole] = useState({ ...emptyRole });
  const dispatch = useDispatch();
  const groups = useSelector(getGroupsByIdWithoutUngrouped);
  const releaseTags = useSelector(getReleaseTagsById);
  const roles = useSelector(getRolesList);
  const isEnterprise = useSelector(getIsEnterprise);

  useEffect(() => {
    dispatch(getExistingReleaseTags());
  }, [dispatch]);

  useEffect(() => {
    if (Object.keys(groups).length) {
      return;
    }
    dispatch(getDynamicGroups());
    dispatch(getGroups());
    dispatch(getRoles());
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dispatch, JSON.stringify(groups)]);

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
    if (adding) {
      dispatch(createRole(submittedRole));
    } else {
      dispatch(editRole(submittedRole));
    }
    onCancel();
  };

  const items = useMemo(
    () =>
      Object.keys(rolesById)
        .reverse()
        .reduce((accu, key) => {
          const index = accu.findIndex(({ id }) => id === key);
          accu = [accu[index], ...accu.filter((item, itemIndex) => index !== itemIndex)];
          return accu;
        }, roles),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [JSON.stringify(roles)]
  );

  return (
    <div>
      <div className="flexbox center-aligned">
        <h2 style={{ marginLeft: 20 }}>Roles</h2>
        <InfoHintContainer>
          <EnterpriseNotification id={BENEFITS.rbac.id} />
          <DocsTooltip />
        </InfoHintContainer>
      </div>
      <DetailsTable columns={columns} items={items} onItemClick={onEditRole} />
      <Chip color="primary" icon={<AddIcon />} label="Add a role" onClick={addRole} disabled={!isEnterprise} />
      <RoleDefinition
        adding={adding}
        editing={editing}
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
