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
import React from 'react';

import { FileCopyOutlined as CopyToClipboardIcon } from '@mui/icons-material';
import { Tooltip } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import TextOverflowMultiline from './TextOverflowMultiline';

const useStyles = makeStyles()(theme => ({
  copyable: {
    '& > svg': {
      cursor: 'pointer',
      fill: theme.palette.action.active,
      opacity: 0,
      transition: 'opacity 0.2s ease-in-out'
    },
    '&:hover > svg': {
      opacity: 1
    }
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
  const isComponent = React.isValidElement(children) || (Array.isArray(children) && children.some(child => React.isValidElement(child)));

  return (
    <div className={`flexbox align-items-center ${onCopy ? classes.copyable : ''}`}>
      <TextOverflowMultiline
        className={textClasses}
        component={isComponent ? 'div' : 'p'}
        lines={5}
        onClick={onCopy}
        style={{ textOverflow: 'ellipsis' }}
        title={title}
        variant="body2"
      >
        {children}
      </TextOverflowMultiline>
      {onCopy && (
        <Tooltip title="Copy to clipboard" placement="top">
          <CopyToClipboardIcon aria-label="Copy to clipboard" className="margin-left-x-small" color="action" fontSize="small" onClick={onCopy} />
        </Tooltip>
      )}
    </div>
  );
};
