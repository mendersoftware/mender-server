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
import { useEffect, useRef, useState } from 'react';
import CopyToClipboard from 'react-copy-to-clipboard';

import { FileCopy as CopyPasteIcon } from '@mui/icons-material';
import { Button } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { TIMEOUTS, yes } from '@northern.tech/store/constants';

const useStyles = makeStyles()(() => ({
  copyNotification: { height: 15 }
}));

export const CopyTextToClipboard = ({ onCopy = yes, token }) => {
  const [copied, setCopied] = useState(false);
  const { classes } = useStyles();
  const timer = useRef<NodeJS.Timeout | undefined>();

  useEffect(() => {
    return () => clearTimeout(timer.current);
  }, []);

  const onCopied = () => {
    setCopied(true);
    onCopy();
    timer.current = setTimeout(() => setCopied(false), TIMEOUTS.fiveSeconds);
  };

  return (
    <div>
      <CopyToClipboard text={token} onCopy={onCopied}>
        <Button startIcon={<CopyPasteIcon />}>Copy to clipboard</Button>
      </CopyToClipboard>
      <p className={classes.copyNotification}>{copied && <span className="green fadeIn">Copied to clipboard.</span>}</p>
    </div>
  );
};
