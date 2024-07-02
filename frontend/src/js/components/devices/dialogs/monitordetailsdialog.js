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
import React, { useState } from 'react';
import CopyToClipboard from 'react-copy-to-clipboard';

import {
  ArrowDropDown as ArrowDropDownIcon,
  ArrowDropUp as ArrowDropUpIcon,
  FileCopy as CopyPasteIcon,
  ReportProblemOutlined as WarningIcon
} from '@mui/icons-material';
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Button,
  Collapse,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  styled
} from '@mui/material';

import { TIMEOUTS } from '../../../constants/appConstants';
import { toggle } from '../../../helpers';

const CopyButton = ({ text, onCopy }) => (
  <CopyToClipboard text={text} onCopy={onCopy}>
    <IconButton size="small">
      <CopyPasteIcon fontSize="small" />
    </IconButton>
  </CopyToClipboard>
);

const LogLine = ({ beExplicit, line, prefix }) => {
  const [copied, setCopied] = useState(false);
  const [hovering, setHovering] = useState(false);

  const onCopied = () => {
    setCopied(true);
    setTimeout(() => setCopied(false), TIMEOUTS.threeSeconds);
  };

  const toggleHovering = () => setHovering(toggle);

  const { line_number, data } = line;

  return (
    <React.Fragment key={line_number}>
      <div className="log-line margin-right" onMouseLeave={toggleHovering} onMouseOver={setHovering}>
        {prefix ? prefix : <div />}
        {line_number !== undefined ? <code className={`align-right ${beExplicit ? 'red' : ''}`}>{line_number}</code> : <div />}
        <code className={beExplicit ? 'red' : ''}>{data}</code>
        {hovering && <CopyButton text={data} onCopy={onCopied} />}
      </div>
      <Collapse in={copied}>
        <div className="margin-left-small green fadeIn">Copied to clipboard.</div>
      </Collapse>
    </React.Fragment>
  );
};

const CustomAccordion = styled(Accordion)({
  root: {
    '&:before': {
      display: 'none'
    },
    '&$expanded': {
      backgroundColor: 'transparent',
      margin: 0
    }
  },
  expanded: {}
});

const LogSection = ({ section = 'previous', lines }) => {
  const [expanded, setExpanded] = useState(false);
  const onToggle = () => setExpanded(toggle);
  return (
    !!lines.length && (
      <CustomAccordion square expanded={expanded} onChange={onToggle}>
        <AccordionSummary style={{ paddingLeft: 0 }}>
          {expanded ? <ArrowDropUpIcon className="margin-right-small" /> : <ArrowDropDownIcon className="margin-right-small" />}
          <div>
            show {section} {lines.length} lines
          </div>
        </AccordionSummary>
        <AccordionDetails style={{ paddingLeft: 0, paddingRight: 0 }}>
          {lines.map(item => (
            <LogLine key={item.line_number} line={item} />
          ))}
        </AccordionDetails>
      </CustomAccordion>
    )
  );
};

const LogContent = ({ lines_before = [], lines_after = [], line_matching = '' }) => (
  <>
    <LogSection section="previous" lines={lines_before} />
    <LogLine beExplicit line={line_matching} prefix={<WarningIcon fontSize="small" />} />
    <LogSection section="next" lines={lines_after} />
  </>
);

const DescriptionContent = ({ description }) => <LogLine line={{ data: description }} />;

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
    details: { lines_before = [], lines_after = [], line_matching = '' }
  } = subject;

  const lines = [...lines_before, line_matching, ...lines_after].filter(i => i);

  const { component: Component, title } = lines.length ? detailTypes.log : detailTypes.description;
  return (
    <Dialog open={!!alert} maxWidth="md">
      <DialogTitle>{`${title} for ${name}`}</DialogTitle>
      <DialogContent style={{ minWidth: 600 }}>
        <Component {...subject.details} />
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Close</Button>
        {!!lines.length && (
          <Button variant="contained" color="primary" onClick={() => exportLog(name, lines)}>
            Export log
          </Button>
        )}
      </DialogActions>
    </Dialog>
  );
};

export default MonitorDetailsDialog;
