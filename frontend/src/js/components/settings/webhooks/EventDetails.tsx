// Copyright 2024 Northern.tech AS
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
import { ReactNode } from 'react';

import { ArrowCircleLeftOutlined as ArrowLeftIcon } from '@mui/icons-material';
import { Divider, IconButton } from '@mui/material';

import Editor, { loader } from '@monaco-editor/react';
import { TwoColumnData } from '@northern.tech/common-ui/ConfigurationObject';
import { CopyTextToClipboard } from '@northern.tech/common-ui/CopyText';
import { ClassesOverrides } from '@northern.tech/common-ui/List';
import Loader from '@northern.tech/common-ui/Loader';
import { Event } from '@northern.tech/store/api/types';
import { Webhook } from '@northern.tech/store/constants';

import { WebhookColumns } from './Management';

loader.config({ paths: { vs: '/ui/vs' } });
const editorProps = {
  height: 600,
  loading: <Loader show />,
  language: 'json',
  options: {
    autoClosingOvertype: 'auto',
    codeLens: false,
    contextmenu: false,
    enableSplitViewResizing: false,
    formatOnPaste: true,
    lightbulb: { enabled: false },
    lineNumbers: 'off',
    minimap: { enabled: false },
    quickSuggestions: false,
    readOnly: true,
    renderOverviewRuler: false,
    scrollBeyondLastLine: false,
    wordWrap: 'on'
  }
};

interface SetSnackbarProps {
  action: () => void;
  autoHideDuration: number;
  children: ReactNode;
  message: string;
  onClick: () => void;
  onClose: () => void;
}

interface WebhookEventDetailsProps extends ClassesOverrides {
  columns: WebhookColumns;
  entry?: Event | undefined;
  onClickBack: () => void;
  setSnackbar: (args: string | SetSnackbarProps) => void;
  webhook: Webhook;
}

const WebhookEventDetails = ({ classes, columns, entry = {}, onClickBack, setSnackbar, webhook }: WebhookEventDetailsProps) => {
  const { data = {} } = entry;

  const content = columns.slice(0, columns.length - 1).reduce((accu, column) => ({ ...accu, [column.title]: column.render(entry, { webhook, classes }) }), {});

  return (
    <>
      <div className="clickable" onClick={onClickBack}>
        <IconButton>
          <ArrowLeftIcon />
        </IconButton>
        Back to webhook
      </div>
      <Divider className={classes.divider} />
      <h4>Event details</h4>
      <TwoColumnData className="margin-top margin-bottom" config={content} setSnackbar={setSnackbar} />
      <h4>Payload</h4>
      {data && <Editor {...editorProps} className="editor modified" value={JSON.stringify(data, null, '\t')} />}
      <Divider className={classes.divider} />
      <CopyTextToClipboard token={data} />
    </>
  );
};

export default WebhookEventDetails;
