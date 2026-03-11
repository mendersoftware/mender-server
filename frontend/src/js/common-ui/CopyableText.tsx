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
import React, { ReactNode, useState } from 'react';

import { FileCopyOutlined as CopyToClipboardIcon } from '@mui/icons-material';
import { Tooltip, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

const useStyles = makeStyles()(theme => ({
  textContent: {
    display: '-webkit-box',
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    WebkitLineClamp: 5,
    WebkitBoxOrient: 'vertical',
    wordBreak: 'break-word'
  },
  copyIconOverride: {
    '&.copy-to-clipboard svg': {
      fill: theme.palette.action.active
    }
  },
  copyIconVisible: {
    opacity: 1
  }
}));

interface CopyableTextProps {
  children: ReactNode;
  onCopy?: () => void;
  textClasses?: string;
  title?: string;
}

export const CopyableText = ({ children, onCopy, textClasses = '', title }: CopyableTextProps) => {
  const { classes } = useStyles();
  const [isHovered, setIsHovered] = useState(false);
  const isComponent = React.isValidElement(children) || (Array.isArray(children) && children.some(child => React.isValidElement(child)));

  return (
    <div className={`flexbox align-items-center ${onCopy ? 'copy-to-clipboard' : ''} ${classes.copyIconOverride}`}>
      <Typography
        className={`${classes.textContent} ${textClasses}`}
        component={isComponent ? 'div' : 'p'}
        onClick={onCopy}
        onMouseEnter={() => setIsHovered(true)}
        onMouseLeave={() => setIsHovered(false)}
        title={title}
        variant="body2"
      >
        {children}
      </Typography>
      {onCopy && (
        <Tooltip title="Copy to clipboard" placement="top">
          <CopyToClipboardIcon
            aria-label="Copy to clipboard"
            className={`margin-left-x-small ${isHovered ? classes.copyIconVisible : ''}`}
            color="action"
            fontSize="small"
            onClick={onCopy}
          />
        </Tooltip>
      )}
    </div>
  );
};
