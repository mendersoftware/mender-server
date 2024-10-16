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
import { ArrowCircleLeftOutlined as ArrowLeftIcon } from '@mui/icons-material';
import { Divider, IconButton } from '@mui/material';

import Editor, { loader } from '@monaco-editor/react';

import { TwoColumnData } from '../../common/configurationobject';
import { CopyTextToClipboard } from '../../common/copytext';
import Loader from '../../common/loader';

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

const WebhookEventDetails = ({
  classes,
  columns,
  entry = {},
  onClickBack,
  setSnackbar,
  webhook
}: {
  classes: any;
  columns: any;
  entry?: {} | undefined;
  onClickBack: any;
  setSnackbar: any;
  webhook: any;
}) => {
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
      <TwoColumnData className={classes.twoColumnsMultiple} config={content} setSnackbar={setSnackbar} />
      <h4>Payload</h4>
      {data && <Editor {...editorProps} className="editor modified" value={JSON.stringify(data, null, '\t')} />}
      <Divider className={classes.divider} />
      <CopyTextToClipboard token={data} />
    </>
  );
};

export default WebhookEventDetails;
