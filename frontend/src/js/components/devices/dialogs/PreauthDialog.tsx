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
import { useCallback, useState } from 'react';

// material ui
import { Alert, Button, DialogActions, DialogContent, Typography } from '@mui/material';

import { ContentSection } from '@northern.tech/common-ui/ContentSection';
import { Link } from '@northern.tech/common-ui/Link';
import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import FileUpload from '@northern.tech/common-ui/forms/FileUpload';
import KeyValueEditor from '@northern.tech/common-ui/forms/KeyValueEditor';
import { useAppDispatch } from '@northern.tech/store/store';
import { preauthDevice } from '@northern.tech/store/thunks';
import { isEmpty } from '@northern.tech/utils/helpers';

const publicKeyEnvelope = /^-----BEGIN PUBLIC KEY-----([A-Za-z0-9+/=\s]+)-----END PUBLIC KEY-----$/;

const generalFileHint = 'Please upload a PEM encoded public key, starting with "-----BEGIN PUBLIC KEY-----".';
const invalidKeyError = `This file does not contain a public key. ${generalFileHint}`;
const emptyKeyError = `This file is empty. ${generalFileHint}`;

export const DeviceLimitWarning = ({ acceptedDevices, deviceLimit }) => (
  <Alert severity="error">
    You have reached your limit of authorized devices: {acceptedDevices} of {deviceLimit}
  </Alert>
);

export const PreauthDialog = ({ acceptedDevices, deviceLimit, limitMaxed, onCancel, onSubmit }) => {
  const [jsonIdentity, setJsonIdentity] = useState<Record<string, string>>({});
  const [keyError, setKeyError] = useState<string>('');
  const [publicKey, setPublicKey] = useState<string>('');
  const [submitError, setSubmitError] = useState<string>('');
  const dispatch = useAppDispatch();

  const convertIdentityToJSON = useCallback(jsonIdentity => {
    setSubmitError('');
    setJsonIdentity(jsonIdentity);
  }, []);

  const onKeyChange = (content?: string) => {
    setSubmitError('');
    setPublicKey('');
    if (content === undefined) {
      setKeyError('');
      return;
    }
    if (!content) {
      setKeyError(emptyKeyError);
      return;
    }
    if (!publicKeyEnvelope.test(content)) {
      setKeyError(invalidKeyError);
      return;
    }
    setKeyError('');
    setPublicKey(content);
  };

  const onHandleSubmit = async shouldClose => {
    setSubmitError('');
    const authset = {
      pubkey: publicKey,
      identity_data: jsonIdentity
    };
    try {
      await dispatch(preauthDevice(authset)).unwrap();
      onSubmit(shouldClose);
    } catch (error) {
      setSubmitError(typeof error === 'string' ? error : (error as Error)?.message);
    }
  };

  const isSubmitDisabled = !publicKey || isEmpty(jsonIdentity) || !!limitMaxed;
  return (
    <BaseDialog open title="Preauthorize devices" onClose={onCancel}>
      <DialogContent>
        <Typography>Add a device&apos;s authentication set to authorize it automatically as soon as it connects.</Typography>
        <ContentSection title="Public key">
          <FileUpload
            isValid={!!publicKey}
            placeholder={
              <>
                Drag and drop or <Link>browse</Link> to upload a file
              </>
            }
            onFileChange={onKeyChange}
          />
          {!!keyError && (
            <Alert className="margin-top-small" severity="error">
              {keyError}
            </Alert>
          )}
        </ContentSection>
        <ContentSection title="Identity data">
          <KeyValueEditor onInputChange={convertIdentityToJSON} />
        </ContentSection>
        {!!limitMaxed && <DeviceLimitWarning acceptedDevices={acceptedDevices} deviceLimit={deviceLimit} />}
        {!!submitError && <Alert severity="error">The device could not be added: {submitError}</Alert>}
      </DialogContent>
      <DialogActions>
        <Button onClick={onCancel}>Cancel</Button>
        <Button variant="outlined" disabled={isSubmitDisabled} onClick={() => onHandleSubmit(false)}>
          Save and add another
        </Button>
        <Button variant="contained" disabled={isSubmitDisabled} onClick={() => onHandleSubmit(true)}>
          Save
        </Button>
      </DialogActions>
    </BaseDialog>
  );
};

export default PreauthDialog;
