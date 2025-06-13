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
import type { ReactNode } from 'react';

import { Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

type OrganizationSettingsItemClasses = {
  base: string;
  content: string;
  main: string;
};

interface OrganizationSettingsItemProps {
  classes?: OrganizationSettingsItemClasses;
  description?: string;
  notification?: ReactNode;
  secondary: string | ReactNode;
  sideBarContent?: ReactNode;
  title: string;
}

export const maxWidth = 500;

const useStyles = makeStyles()(({ spacing }) => ({
  base: { gap: spacing(1) },
  content: {
    '> *': { maxWidth }
  },
  mainContent: {
    alignItems: 'baseline',
    display: 'grid',
    gridTemplateColumns: '500px 1fr',
    gridColumnGap: spacing(2),
    maxWidth: 'initial'
  }
}));

const defaultClasses: OrganizationSettingsItemClasses = { base: '', content: '', main: '' };

const OrganizationSettingsItem = ({ classes = defaultClasses, description, notification, secondary, sideBarContent, title }: OrganizationSettingsItemProps) => {
  const { classes: localClasses } = useStyles();
  return (
    <div className={`flexbox column settings-item-base ${localClasses.base} margin-top-small ${classes.base ?? ''}`}>
      <div className={`flexbox column settings-item-content ${localClasses.base} ${localClasses.content} ${classes.content ?? ''}`}>
        <Typography variant="subtitle1">{title}</Typography>
        {description && <Typography variant="body2">{description}</Typography>}
        <div className={`settings-item-main-content ${localClasses.mainContent} ${classes.main ?? ''}`}>
          <Typography variant="body2" component="div">
            {secondary}
          </Typography>
          {sideBarContent}
        </div>
      </div>
      {notification}
    </div>
  );
};

export default OrganizationSettingsItem;
