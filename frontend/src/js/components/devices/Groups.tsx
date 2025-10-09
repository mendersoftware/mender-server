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
// material ui
import { Info as InfoIcon } from '@mui/icons-material';
import { Button, List, ListItemButton, ListItemText, ListSubheader, Typography, listItemButtonClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { ALL_DEVICES } from '@northern.tech/store/constants';

import { HELPTOOLTIPS } from '../helptips/HelpTooltips';
import { MenderHelpTooltip } from '../helptips/MenderTooltip';

const useStyles = makeStyles()(theme => ({
  header: {
    height: theme.spacing(6),
    '&.heading-lined span': {
      padding: 0,
      minWidth: 70,
      background: theme.palette.background.default
    },
    '.group-border': {
      background: theme.palette.divider
    }
  },
  list: {
    [`.${listItemButtonClasses.root}.Mui-selected`]: {
      backgroundColor: theme.palette.action.selected
    }
  }
}));

export const GroupsSubheader = ({ heading }) => {
  const { classes } = useStyles();
  return (
    <ListSubheader classes={{ root: 'heading-lined' }} className={classes.header} disableGutters disableSticky key="static-groups-sub">
      <span>{heading}</span>
      <div className="group-border" />
    </ListSubheader>
  );
};

export const GroupItem = ({ changeGroup, groupname, selectedGroup, name }) => (
  <ListItemButton classes={{ root: 'grouplist' }} selected={name === selectedGroup || groupname === selectedGroup} onClick={() => changeGroup(name)}>
    <ListItemText className="margin-left" primary={decodeURIComponent(name)} />
  </ListItemButton>
);

export const Groups = ({ acceptedCount, changeGroup, className, groups, openGroupDialog, selectedGroup }) => {
  const { classes } = useStyles();
  const { dynamic: dynamicGroups, static: staticGroups, ungrouped } = groups;
  return (
    <div className={className}>
      <div className="flexbox margin-bottom-small margin-top-small">
        <Typography variant="subtitle1">Groups</Typography>
        {!!acceptedCount && <MenderHelpTooltip id={HELPTOOLTIPS.addGroup.id} className="margin-left-small" />}
      </div>
      <List className={classes.list}>
        <ListItemButton classes={{ root: 'grouplist' }} key="All" selected={!selectedGroup} onClick={() => changeGroup()}>
          <ListItemText primary={ALL_DEVICES} />
        </ListItemButton>
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
      </List>
      <Button className="margin-top" startIcon={<InfoIcon />} onClick={openGroupDialog} color="inherit">
        Create a group
      </Button>
    </div>
  );
};

export default Groups;
