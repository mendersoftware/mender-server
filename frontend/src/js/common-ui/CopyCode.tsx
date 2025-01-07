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
import React, { useState } from 'react';
import CopyToClipboard from 'react-copy-to-clipboard';

import { FileCopy as CopyPasteIcon } from '@mui/icons-material';
import { Button, IconButton } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { TIMEOUTS } from '@northern.tech/store/constants';

const buttonStyle = { float: 'right', margin: '-20px 0 0 10px' };

const useStyles = makeStyles()(theme => ({
  code: {
    border: '1px solid',
    borderColor: theme.palette.background.lightgrey,
    backgroundColor: theme.palette.background.lightgrey
  }
}));

export const Code = ({ className = '', children, style = {} }) => {
  const { classes } = useStyles();
  return (
    <div className={`code ${classes.code} ${className}`} style={style}>
      {children}
    </div>
  );
};

export const CopyCode = ({ code, onCopy, withDescription }) => {
  const [copied, setCopied] = useState(false);

  const onCopied = (_text, result) => {
    setCopied(result);
    setTimeout(() => setCopied(false), TIMEOUTS.fiveSeconds);
    if (onCopy) {
      onCopy();
    }
  };

  return (
    <>
      <Code>
        <CopyToClipboard text={code} onCopy={onCopied}>
          {withDescription ? (
            <Button style={buttonStyle} startIcon={<CopyPasteIcon />}>
              Copy to clipboard
            </Button>
          ) : (
            <IconButton style={buttonStyle} size="large">
              <CopyPasteIcon />
            </IconButton>
          )}
        </CopyToClipboard>
        <span style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>{code}</span>
      </Code>
      <p>{copied && <span className="green fadeIn">Copied to clipboard.</span>}</p>
    </>
  );
};

export default CopyCode;
