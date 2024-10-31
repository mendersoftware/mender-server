// Copyright 2024 Northern.tech AS
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
import { useState } from 'react';
import { useSelector } from 'react-redux';

import { Add as AddIcon } from '@mui/icons-material';
import { Chip } from '@mui/material';

import { getTenantsList } from '@northern.tech/store/organizationSlice/selectors';

import { TenantCreateForm } from './TenantCreateForm';
import { TenantList } from './TenantList';

interface TenantsEmptyStateProps {
  openModal: () => void;
}
const TenantsEmptyState = (props: TenantsEmptyStateProps) => {
  const { openModal } = props;
  return (
    <div className="dashboard-placeholder">
      <p>You are not currently managing any tenants. </p>
      <p>
        <a onClick={openModal}>Add a tenant</a> to get started.
      </p>
    </div>
  );
};
export const TenantPage = () => {
  const [showCreate, setShowCreate] = useState<boolean>(false);
  const { tenants } = useSelector(getTenantsList);
  return (
    <div>
      <h2>Tenants</h2>
      {tenants.length ? <TenantList /> : <TenantsEmptyState openModal={() => setShowCreate(true)} />}
      <Chip color="primary" icon={<AddIcon />} label="Add tenant" onClick={() => setShowCreate(true)} />
      <TenantCreateForm open={showCreate} onCloseClick={() => setShowCreate(false)} />
    </div>
  );
};
