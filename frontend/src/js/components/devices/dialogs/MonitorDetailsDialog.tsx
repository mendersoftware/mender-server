// Copyright 2021 Northern.tech AS
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
import { useEffect, useRef } from 'react';
import CopyToClipboard from 'react-copy-to-clipboard';

import { Button, DialogActions, DialogContent, Typography } from '@mui/material';
import { alpha } from '@mui/material/styles';
import { makeStyles } from 'tss-react/mui';

import { Code } from '@northern.tech/common-ui/CopyCode';
import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';

const useStyles = makeStyles()(theme => ({
  container: {
    position: 'relative',
    maxHeight: 400,
    overflowY: 'auto',
    border: `1px solid ${theme.palette.divider}`,
    borderRadius: theme.shape.borderRadius,
    '& code': { padding: theme.spacing(0.5, 1.5) }
  },
  row: {
    display: 'flex',
    borderInlineStart: '3px solid transparent'
  },
  matchRow: {
    backgroundColor: alpha(theme.palette.error.main, 0.1),
    borderInlineStart: `3px solid ${theme.palette.error.main}`,
    '& code': { color: theme.palette.error.main, fontWeight: 500 }
  },
  gutter: {
    flexShrink: 0,
    minWidth: 52,
    textAlign: 'right',
    userSelect: 'none',
    color: theme.palette.text.disabled,
    borderInlineEnd: `1px solid ${theme.palette.divider}`
  },
  content: {
    color: theme.palette.text.secondary,
    whiteSpace: 'pre-wrap',
    wordBreak: 'break-all'
  }
}));

const LogRow = ({ line, isMatch, ref }) => {
  const { classes } = useStyles();
  const { line_number, data } = line;
  return (
    <div ref={ref} className={`${classes.row} ${isMatch ? classes.matchRow : ''}`}>
      {line_number !== undefined && (
        <Typography component="code" variant="code2" className={classes.gutter}>
          {line_number}
        </Typography>
      )}
      <Typography component="code" variant="code2" className={classes.content}>
        {data}
      </Typography>
    </div>
  );
};

const LogContent = ({ lines_before = [], lines_after = [], line_matching = {} }) => {
  const { classes } = useStyles();
  const containerRef = useRef(null);
  const matchRef = useRef(null);
  const rows = [...lines_before, line_matching, ...lines_after];

  //scroll matching line into view
  useEffect(() => {
    const container = containerRef.current;
    const match = matchRef.current;
    if (container && match) {
      container.scrollTop = match.offsetTop - (container.clientHeight - match.clientHeight) / 2;
    }
  }, []);

  return (
    <div ref={containerRef} className={classes.container}>
      {rows.map(line => {
        const isMatch = line.line_number === line_matching.line_number;
        return <LogRow key={line.line_number} ref={isMatch ? matchRef : undefined} line={line} isMatch={isMatch} />;
      })}
    </div>
  );
};

const DescriptionContent = ({ description }) => <Code>{description}</Code>;

const detailTypes = {
  log: {
    component: LogContent,
    title: 'Log excerpt'
  },
  description: {
    component: DescriptionContent,
    title: 'Details'
  }
};

const exportLog = (name, lines) => {
  const max = lines.reduce((accu, item) => Math.max(accu, item.line_number), 0);
  const length = `${max}`.length;
  const logData = lines
    .reduce((accu, item) => {
      const paddedLineNumber = `${item.line_number}`.padStart(length, '0');
      accu.push(`${paddedLineNumber}   ${item.data}`);
      return accu;
    }, [])
    .join('\n');
  const uriContent = `data:application/octet-stream,${encodeURIComponent(logData)}`;
  window.open(uriContent, `Mender-Monitor-${name.replace(/ /g, '_')}.log`);
};

export const MonitorDetailsDialog = ({ alert, onClose }) => {
  const { name, subject = { details: {} } } = alert ?? {};
  const {
    details: { description = '', lines_before = [], lines_after = [], line_matching = '' }
  } = subject;

  const lines = [...lines_before, line_matching, ...lines_after].filter(i => i);

  const { component: Component, title } = lines.length ? detailTypes.log : detailTypes.description;
  return (
    <BaseDialog open={!!alert} title={`${title} for ${name}`} maxWidth="md" onClose={onClose}>
      <DialogContent style={{ minWidth: 600 }}>
        <Component {...subject.details} />
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Close</Button>
        {!lines.length && !!description && (
          <CopyToClipboard text={description}>
            <Button variant="contained">Copy to clipboard</Button>
          </CopyToClipboard>
        )}
        {!!lines.length && (
          <Button variant="contained" onClick={() => exportLog(name, lines)}>
            Export log
          </Button>
        )}
      </DialogActions>
    </BaseDialog>
  );
};

export default MonitorDetailsDialog;
