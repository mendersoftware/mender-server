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
import React from 'react';

import { Divider, ListItem, ListItemText } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

const defaultItemProps = { alignItems: 'flex-start' };
export const maxWidth = 500;

const useStyles = makeStyles()(theme => ({
  divider: { marginBottom: theme.spacing(), marginLeft: theme.spacing(-2), width: maxWidth },
  text: { display: 'grid', width: '100%', marginBottom: theme.spacing() },
  secondary: { display: 'grid', width: '100%', marginBottom: 0 }
}));

const OrganizationSettingsItem = ({ title, content: { action, description }, secondary, sideBarContent, notification }) => {
  const { classes } = useStyles();

  const secondaryContent = secondary ? (
    secondary
  ) : (
    <>
      <div>{description}</div>
      {action.internal ? (
        <a className="align-right" onClick={action.action}>
          {action.title}
        </a>
      ) : (
        <a className="align-right" href={action.target} target="_blank" rel="noopener noreferrer">
          {action.title}
        </a>
      )}
    </>
  );
  return (
    <li className="margin-top-small org-settings-item">
      <ListItem {...defaultItemProps} component="div" classes={{ root: 'flexbox column' }}>
        <ListItemText
          className={secondary ? classes.secondary : classes.text}
          classes={{ secondary: secondary ? '' : 'two-columns' }}
          primary={title}
          secondaryTypographyProps={{ component: 'div' }}
          secondary={<>{secondaryContent}</>}
        />
        <Divider className={classes.divider} />
        {notification}
      </ListItem>
      {sideBarContent}
    </li>
  );
};

export default OrganizationSettingsItem;
