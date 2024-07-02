// Copyright 2015 Northern.tech AS
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
import React from 'react';

// material ui
import { InfoOutlined as InfoIcon } from '@mui/icons-material';
import { List, ListItem, ListItemIcon, ListItemText, ListSubheader } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { ALL_DEVICES } from '../../constants/deviceConstants';
import { HELPTOOLTIPS, MenderHelpTooltip } from '../helptips/helptooltips';

const useStyles = makeStyles()(theme => ({
  header: {
    color: theme.palette.grey[800],
    height: theme.spacing(6)
  },
  groupBorder: {
    background: theme.palette.grey[50]
  },
  groupHeading: {
    background: theme.palette.background.default
  }
}));

export const GroupsSubheader = ({ heading }) => {
  const { classes } = useStyles();
  return (
    <ListSubheader classes={{ root: 'heading-lined' }} className={classes.header} disableGutters disableSticky key="static-groups-sub">
      <span className={classes.groupHeading}>{heading}</span>
      <div className={classes.groupBorder}></div>
    </ListSubheader>
  );
};

export const GroupItem = ({ changeGroup, groupname, selectedGroup, name }) => (
  <ListItem classes={{ root: 'grouplist' }} button selected={name === selectedGroup || groupname === selectedGroup} onClick={() => changeGroup(name)}>
    <ListItemText primary={decodeURIComponent(name)} />
  </ListItem>
);

export const Groups = ({ acceptedCount, changeGroup, className, groups, openGroupDialog, selectedGroup }) => {
  const { dynamic: dynamicGroups, static: staticGroups, ungrouped } = groups;
  return (
    <div className={className}>
      <div className="flexbox margin-bottom-small margin-top-small">
        <div className="muted">Groups</div>
        {!!(acceptedCount && staticGroups.length + dynamicGroups.length <= 1) && <MenderHelpTooltip id={HELPTOOLTIPS.addGroup.id} className="margin-left" />}
      </div>
      <List>
        <ListItem classes={{ root: 'grouplist' }} button key="All" selected={!selectedGroup} onClick={() => changeGroup()}>
          <ListItemText primary={ALL_DEVICES} />
        </ListItem>
        {!!dynamicGroups.length && <GroupsSubheader heading="Dynamic" />}
        {dynamicGroups.map(({ groupId, name }, index) => (
          <GroupItem changeGroup={changeGroup} groupname={name} key={name + index} name={groupId} selectedGroup={selectedGroup} />
        ))}
        {!!staticGroups.length && <GroupsSubheader heading="Static" />}
        {staticGroups.map(({ groupId, name }, index) => (
          <GroupItem changeGroup={changeGroup} groupname={name} key={name + index} name={groupId} selectedGroup={selectedGroup} />
        ))}
        {!!staticGroups.length &&
          ungrouped.map(({ groupId, name }, index) => (
            <GroupItem changeGroup={changeGroup} groupname={name} key={name + index} name={groupId} selectedGroup={selectedGroup} />
          ))}
        <ListItem button classes={{ root: 'grouplist' }} style={{ marginTop: 30 }} onClick={openGroupDialog}>
          <ListItemIcon>
            <InfoIcon />
          </ListItemIcon>
          <ListItemText primary="Create a group" />
        </ListItem>
      </List>
    </div>
  );
};

export default Groups;
