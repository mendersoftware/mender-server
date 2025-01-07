// Copyright 2022 Northern.tech AS
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
import { NavLink } from 'react-router-dom';

// material ui
import { List, ListItem, ListItemIcon, ListItemText, ListSubheader } from '@mui/material';
import { listItemTextClasses } from '@mui/material/ListItemText';
import { makeStyles } from 'tss-react/mui';

import { isDarkMode } from '@northern.tech/store/utils';

const useStyles = makeStyles()(theme => ({
  list: {
    [`&.navLink .${listItemTextClasses.root}`]: {
      color: isDarkMode(theme.palette.mode) ? theme.palette.text.primary : theme.palette.grey[900]
    },
    [`&.navLink.active .${listItemTextClasses.root}`]: {
      color: isDarkMode(theme.palette.mode) ? theme.palette.grey[900] : theme.palette.text.primary
    },
    '&.active': {
      background: theme.palette.grey[400]
    }
  },
  listItem: {
    [`.${listItemTextClasses.primary}`]: {
      fontSize: '0.8rem',
      '&:hover': {
        color: theme.palette.primary.main
      }
    }
  }
}));

export const LeftNav = ({ sections }) => {
  const { classes } = useStyles();
  return (
    <List className="leftFixed">
      {sections.map(({ itemClass = '', items = [], title = '' }, index) => (
        <React.Fragment key={`${itemClass}-${index}`}>
          <ListSubheader disableSticky={true}>{title}</ListSubheader>
          {items.map(({ exact, path, icon = null, style = {}, title = '', url }) => {
            const props = url
              ? { component: 'a', exact: `${exact}`, href: url, rel: 'noopener', target: '_blank', to: url }
              : { component: NavLink, end: exact, to: path };
            return (
              <ListItem className={`navLink ${itemClass} ${classes.list}`} key={path} style={style} {...props}>
                <ListItemText className={classes.listItem} primary={title} url={url} />
                {!!icon && <ListItemIcon>{icon}</ListItemIcon>}
              </ListItem>
            );
          })}
        </React.Fragment>
      ))}
    </List>
  );
};

export default LeftNav;
