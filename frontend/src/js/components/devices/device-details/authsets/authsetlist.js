// Copyright 2018 Northern.tech AS
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
import React, { useState } from 'react';

import { accordionClasses, accordionDetailsClasses, accordionSummaryClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { canAccess } from '../../../../constants/appConstants';
import { DEVICE_STATES } from '../../../../constants/deviceConstants';
import { customSort } from '../../../../helpers';
import AuthsetListItem from './authsetlistitem';

const fourColumns = '0.5fr 1fr 2fr 2fr';
const useStyles = makeStyles()(theme => ({
  authsets: {
    [`.header, .${accordionClasses.root}`]: {
      borderBottom: `1px solid ${theme.palette.grey[600]}`
    },
    [`.columnHeader, .${accordionSummaryClasses.root}, .${accordionSummaryClasses.content}`]: {
      cursor: 'default'
    },
    [`.header, .body .${accordionSummaryClasses.content}`]: {
      display: 'grid',
      gridColumnGap: theme.spacing(2),
      gridTemplateColumns: '0.5fr 1fr 2fr 2fr 2fr'
    }
  },
  accordion: {
    backgroundColor: theme.palette.grey[50],
    '&:before': { display: 'none' },
    '&$expanded': { margin: 'auto' },
    [`.columns-4 .${accordionSummaryClasses.content}`]: {
      gridTemplateColumns: fourColumns
    },
    [`.${accordionDetailsClasses.root}`]: { flexDirection: 'row' }
  },
  divider: { marginTop: theme.spacing(), marginBottom: theme.spacing() },
  header: {
    padding: theme.spacing(2),
    '&.columns-4': { gridTemplateColumns: fourColumns }
  }
}));

export const defaultColumns = [
  { title: '', canAccess },
  { title: 'Status', canAccess },
  { title: 'Public key', canAccess },
  { title: 'Time of request', canAccess },
  { title: 'Actions', canAccess: ({ userCapabilities: { canManageDevices } }) => canManageDevices }
];

export const AuthsetList = ({ device, listRef, userCapabilities, ...remainingProps }) => {
  const [expandRow, setExpandRow] = useState();
  const { classes } = useStyles();
  const { auth_sets: authsets = [], status = DEVICE_STATES.accepted } = device;

  const availableColumns = defaultColumns.filter(column => column.canAccess({ userCapabilities }));

  let groupedAuthsets = authsets.reduce(
    // for each authset compare the device status and if it matches authset status, put it in correct list
    (accu, authset) => {
      if (authset.status === status) {
        accu.active.push(authset);
      } else if (authset.status === DEVICE_STATES.pending) {
        accu.pending.push(authset);
      } else {
        accu.inactive.push(authset);
      }
      return accu;
    },
    { active: [], inactive: [], pending: [] }
  );

  const orderedAuthsets = [
    ...groupedAuthsets.pending.sort(customSort(true, 'ts')),
    ...groupedAuthsets.active.sort(customSort(true, 'ts')),
    ...groupedAuthsets.inactive.sort(customSort(true, 'ts'))
  ];

  return (
    <div className={`authsets ${classes.authsets}`}>
      <div className={`header columns-${availableColumns.length} ${classes.header}`}>
        {availableColumns.map(({ title: headerName }, index) => (
          <div className="columnHeader" key={`columnHeader-${index}`}>
            {headerName}
          </div>
        ))}
      </div>
      <div className="body relative" ref={listRef}>
        {orderedAuthsets.map(authset => (
          <AuthsetListItem
            authset={authset}
            classes={classes}
            columns={availableColumns}
            device={device}
            isExpanded={expandRow === authset.id}
            key={`authset-${authset.id}`}
            onExpand={setExpandRow}
            userCapabilities={userCapabilities}
            {...remainingProps}
          />
        ))}
      </div>
    </div>
  );
};

export default AuthsetList;
