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
import React, { useMemo } from 'react';

// material ui
import { Check as CheckIcon } from '@mui/icons-material';
import { Chip } from '@mui/material';

import DetailsIndicator from '@northern.tech/common-ui/DetailsIndicator';
import DetailsTable from '@northern.tech/common-ui/DetailsTable';
import Time, { RelativeTime } from '@northern.tech/common-ui/Time';
import { twoFAStates } from '@northern.tech/store/constants';

const columnData = [
  {
    key: 'email',
    disablePadding: false,
    title: 'Email',
    enterpriseOnly: false,
    render: user => (
      <>
        <span>{user.email}</span>
        {user.tfa_status === twoFAStates.enabled && (
          <Chip className="margin-left-small" icon={<CheckIcon titleAccess={`2FA ${twoFAStates.enabled}`} />} label="2FA" size="small" variant="outlined" />
        )}
      </>
    )
  },
  { key: 'created_ts', disablePadding: false, title: 'Date created', enterpriseOnly: false, render: ({ created_ts }) => <Time value={created_ts} /> },
  {
    key: 'updated_ts',
    disablePadding: false,
    title: 'Last updated',
    enterpriseOnly: false,
    render: ({ updated_ts }) => <RelativeTime updateTime={updated_ts} />
  },
  {
    key: 'roles',
    disablePadding: false,
    title: 'Role',
    enterpriseOnly: true,
    render: ({ roles: userRoles = [] }, { roles }) => userRoles.map(roleId => roles.find(({ value }) => roleId === value)?.name).join(', ')
  },
  {
    key: 'actions',
    disablePadding: false,
    title: 'Manage',
    enterpriseOnly: false,
    render: DetailsIndicator
  }
];

const UserList = ({ editUser, isEnterprise, roles, users }) => {
  const columns = useMemo(
    () =>
      columnData.reduce((accu, { enterpriseOnly, ...remainder }) => {
        if (enterpriseOnly && !isEnterprise) {
          return accu;
        }
        accu.push({ ...remainder, extras: { roles } });
        return accu;
      }, []),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [isEnterprise, JSON.stringify(roles)]
  );
  return <DetailsTable columns={columns} items={users} onItemClick={editUser} />;
};

export default UserList;
