// Copyright 2026 Northern.tech AS
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

import { Chip, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

interface ContentSectionProps {
  children: ReactNode;
  className?: string;
  isAddOn?: boolean;
  postTitle?: ReactNode;
  title: string;
}

const useStyles = makeStyles()(theme => ({
  title: { display: 'grid', gridAutoFlow: 'column', gridAutoColumns: 'max-content', gap: theme.spacing(), alignItems: 'center' }
}));

export const ContentSection = ({ children, className = '', isAddOn = false, postTitle, title }: ContentSectionProps) => {
  const { classes } = useStyles();

  return (
    <div className={`margin-bottom margin-top-medium ${className}`}>
      <div className="flexbox space-between">
        <div className={`margin-bottom-small ${classes.title}`}>
          <Typography variant="subtitle1">{title}</Typography>
          {postTitle}
        </div>
        {isAddOn && <Chip label="Add-on" />}
      </div>
      {children}
    </div>
  );
};
