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

// material ui
import { Button, Dialog, DialogActions, DialogContent, DialogTitle, FormControlLabel, FormHelperText, Radio, RadioGroup } from '@mui/material';

import FileUpload from '../../common/forms/fileupload';

export const ConfigImportDialog = ({ onCancel, onSubmit }) => {
  const [config, setConfig] = useState(null);
  const [errortext, setErrortext] = useState();
  const [importType, setImportType] = useState('default');
  const [isSubmitDisabled, setIsSubmitDisabled] = useState(false);

  const onHandleSubmit = () => {
    onSubmit({ importType, config });
  };

  const onFileChange = file => {
    let importedConfig;
    try {
      importedConfig = JSON.parse(file);
    } catch {
      if (file) {
        return setErrortext('Could not parse the selected file, please ensure the validity of the JSON structure.');
      }
    }
    setErrortext();
    setConfig(importedConfig);
    setIsSubmitDisabled(!file);
  };

  return (
    <Dialog open={true}>
      <DialogTitle>Import configuration</DialogTitle>
      <DialogContent className="margin-small" style={{ overflow: 'hidden' }}>
        <RadioGroup
          value={importType}
          onChange={e => {
            setErrortext();
            setIsSubmitDisabled(e.target.value === 'file' && !config);
            setImportType(e.target.value);
          }}
        >
          <FormControlLabel
            control={<Radio size="small" value="default" />}
            name="default"
            label="Import the current default configuration"
            style={{ marginTop: 0 }}
          />
          <div className="margin margin-left-small">or</div>
          <FormControlLabel
            control={<Radio size="small" value="file" />}
            name="file"
            label={
              <FileUpload
                placeholder={
                  <>
                    Drag & drop or <a>select a JSON file</a> to import configuration data
                  </>
                }
                onFileChange={onFileChange}
              />
            }
            style={{ marginTop: 0 }}
          />
          {errortext && <FormHelperText>{errortext}</FormHelperText>}
        </RadioGroup>
      </DialogContent>
      <DialogActions>
        <Button onClick={onCancel}>Cancel</Button>
        <Button variant="contained" disabled={isSubmitDisabled} onClick={onHandleSubmit} color="secondary" style={{ marginLeft: 10 }}>
          Import
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default ConfigImportDialog;
