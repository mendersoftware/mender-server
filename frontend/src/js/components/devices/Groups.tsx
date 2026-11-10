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
import { useState } from 'react';

import { Delete, Info as InfoIcon, InfoOutlined as InfoOutlinedIcon, MoreHoriz as MoreHorizIcon } from '@mui/icons-material';
import {
  Button,
  IconButton,
  List,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  ListSubheader,
  Menu,
  MenuItem,
  Tooltip,
  Typography,
  listItemButtonClasses,
  listItemSecondaryActionClasses
} from '@mui/material';
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
      background: theme.palette.background.default,
      position: 'absolute'
    },
    '.group-border': {
      background: theme.palette.divider,
      width: '100%',
      height: 1
    }
  },
  list: {
    [`.${listItemButtonClasses.root}.Mui-selected`]: {
      backgroundColor: theme.palette.action.selected
    }
  },
  groupItem: {
    [`.${listItemSecondaryActionClasses.root}`]: {
      visibility: 'hidden'
    },
    [`&:hover .${listItemSecondaryActionClasses.root}, &:focus-within .${listItemSecondaryActionClasses.root}`]: {
      visibility: 'visible'
    }
  },
  disabledInfo: {
    cursor: 'pointer',
    pointerEvents: 'auto'
  },
  groupItemActive: {
    [`.${listItemSecondaryActionClasses.root}`]: {
      visibility: 'visible'
    }
  }
}));

export const GroupsSubheader = ({ heading }) => {
  const { classes } = useStyles();
  return (
    <ListSubheader
      classes={{ root: 'heading-lined' }}
      className={`flexbox align-items-center relative ${classes.header}`}
      disableGutters
      disableSticky
      key="static-groups-sub"
    >
      <span>{heading}</span>
      <div className="group-border" />
    </ListSubheader>
  );
};

export const GroupItem = ({ canRemove, changeGroup, groupname, selectedGroup, name, onRemove }) => {
  const { classes } = useStyles();
  const [anchorEl, setAnchorEl] = useState(null);
  const onRemoveClick = () => {
    setAnchorEl(null);
    onRemove(name);
  };

  return (
    <ListItem
      className={`${classes.groupItem} ${anchorEl ? classes.groupItemActive : ''}`}
      disablePadding
      secondaryAction={
        !!onRemove && (
          <>
            <IconButton aria-label="group actions" edge="end" size="small" onClick={e => setAnchorEl(e.currentTarget)}>
              <MoreHorizIcon />
            </IconButton>
            <Menu
              anchorEl={anchorEl}
              anchorOrigin={{ horizontal: 'left', vertical: 'bottom' }}
              onClose={() => setAnchorEl(null)}
              open={Boolean(anchorEl)}
              transformOrigin={{ horizontal: 'left', vertical: 'top' }}
            >
              <MenuItem className={canRemove ? 'red' : ''} disabled={!canRemove} onClick={onRemoveClick}>
                <ListItemIcon>
                  <Delete className={canRemove ? 'red' : ''} />
                </ListItemIcon>
                Delete group
                {!canRemove && (
                  <Tooltip arrow title="View only. Contact admin to edit">
                    <InfoOutlinedIcon className={`margin-left-small ${classes.disabledInfo}`} fontSize="small" onClick={e => e.stopPropagation()} />
                  </Tooltip>
                )}
              </MenuItem>
            </Menu>
          </>
        )
      }
    >
      <ListItemButton classes={{ root: 'grouplist' }} selected={name === selectedGroup || groupname === selectedGroup} onClick={() => changeGroup(name)}>
        <ListItemText className="margin-left" primary={decodeURIComponent(name)} />
      </ListItemButton>
    </ListItem>
  );
};

export const Groups = ({ acceptedCount, canRemoveGroups, changeGroup, className, groups, onRemoveGroup, openGroupDialog, selectedGroup }) => {
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
          <GroupItem
            canRemove={canRemoveGroups}
            changeGroup={changeGroup}
            groupname={name}
            key={name + index}
            name={groupId}
            onRemove={onRemoveGroup}
            selectedGroup={selectedGroup}
          />
        ))}
        {!!staticGroups.length && <GroupsSubheader heading="Static" />}
        {staticGroups.map(({ groupId, name }, index) => (
          <GroupItem
            canRemove={canRemoveGroups}
            changeGroup={changeGroup}
            groupname={name}
            key={name + index}
            name={groupId}
            onRemove={onRemoveGroup}
            selectedGroup={selectedGroup}
          />
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
