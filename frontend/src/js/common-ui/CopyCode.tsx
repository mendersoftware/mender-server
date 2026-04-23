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
import type { CSSProperties, ReactNode } from 'react';
import { useState } from 'react';
import CopyToClipboard from 'react-copy-to-clipboard';

import { ContentCopy as CopyPasteIcon } from '@mui/icons-material';
import { Button, IconButton, Typography, type TypographyProps } from '@mui/material';
import { alpha } from '@mui/material/styles';
import { makeStyles } from 'tss-react/mui';

import { TIMEOUTS } from '@northern.tech/store/constants';

const sizeMaxHeights = {
  small: 200,
  medium: 368
};

const useStyles = makeStyles()(theme => {
  const isDarkMode = theme.palette.mode === 'dark';
  return {
    button: { float: 'right', marginRight: theme.spacing(-1), marginTop: theme.spacing(-1) },
    code: {
      backgroundColor: alpha(theme.palette.info.contrastText, isDarkMode ? 0.08 : 0.04),
      fontFamily: 'monospace',
      color: alpha(theme.palette.text.primary, isDarkMode ? 0.87 : 1),
      borderRadius: theme.spacing(0.5),
      padding: theme.spacing(2),
      overflowY: 'auto',
      position: 'relative',
      whiteSpace: 'pre-line',
      '.copyable-content': {
        whiteSpace: 'pre-wrap',
        wordBreak: 'break-word'
      }
    },
    inlineCode: {
      backgroundColor: theme.palette.action.hover,
      border: `1px solid ${theme.palette.divider}`,
      fontFamily: 'monospace',
      borderRadius: theme.spacing(0.5),
      padding: '2px 6px',
      fontSize: '0.875em'
    }
  };
});

type CodeSize = 'small' | 'medium' | 'full';

type CodeVariant = 'code1' | 'code2';

interface CodeProps {
  children: ReactNode;
  className?: string;
  size?: CodeSize;
  style?: CSSProperties;
}

export const Code = ({ className = '', children, size = 'full', style = {} }: CodeProps) => {
  const { classes } = useStyles();
  const maxHeight = sizeMaxHeights[size];
  return (
    <div className={`${classes.code} ${className}`} style={{ ...style, ...(maxHeight ? { maxHeight } : {}) }}>
      {children}
    </div>
  );
};

export const InlineCode = ({ children, variant = 'code1', ...props }: TypographyProps) => {
  const { classes } = useStyles();
  const { className, ...restProps } = props;
  return (
    <Typography component="code" variant={variant} className={`${className ? className : ''} ${classes.inlineCode}`} {...restProps}>
      {children}
    </Typography>
  );
};

interface CopyCodeProps {
  code: string;
  onCopy?: () => void;
  size?: CodeSize;
  variant?: CodeVariant;
  withDescription?: boolean;
}

export const CopyCode = ({ code, onCopy, size = 'full', variant = 'code1', withDescription }: CopyCodeProps) => {
  const [copied, setCopied] = useState(false);
  const { classes } = useStyles();

  const onCopied = (_text: string, result: boolean) => {
    setCopied(result);
    setTimeout(() => setCopied(false), TIMEOUTS.fiveSeconds);
    if (onCopy) {
      onCopy();
    }
  };

  return (
    <>
      <Code size={size}>
        <CopyToClipboard text={code} onCopy={onCopied}>
          {withDescription ? (
            <Button color="inherit" size="large" variant="text" className={classes.button} startIcon={<CopyPasteIcon />} title="Copy to clipboard">
              Copy to clipboard
            </Button>
          ) : (
            <IconButton className={classes.button} size="large" title="Copy to clipboard">
              <CopyPasteIcon />
            </IconButton>
          )}
        </CopyToClipboard>
        <Typography component="code" variant={variant} className="copyable-content">
          {code}
        </Typography>
      </Code>
      <p>{copied && <span className="green fadeIn">Copied to clipboard.</span>}</p>
    </>
  );
};

export default CopyCode;
