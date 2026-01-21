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
import { useEffect, useRef, useState } from 'react';

import { Button, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import Loader from '@northern.tech/common-ui/Loader';
import Form from '@northern.tech/common-ui/forms/Form';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import { twoFAStates } from '@northern.tech/store/constants';

import { notificationMap } from '../SelfUserManagement';

const useStyles = makeStyles()(theme => ({
  orderedList: {
    marginBottom: 0,
    marginTop: theme.spacing(),
    maxWidth: '550px'
  },
  qrImg: { maxWidth: '188px' },
  codeInput: { maxWidth: '220px', marginTop: theme.spacing(1) },
  buttonWrapper: { justifyContent: 'flex-end', flexDirection: 'row-reverse', marginTop: theme.spacing(2) },
  cancelButton: { marginLeft: theme.spacing() }
}));

export const AuthSetup = ({ currentUser, handle2FAState, has2FA, onClose, qrImage, verify2FA, setShowNotice }) => {
  const current2FA = useRef(has2FA);
  const [validated2fa, setValidated2fa] = useState(false);
  const [validating2fa, setValidating2fa] = useState(false);

  useEffect(() => {
    current2FA.current = has2FA;
  }, [has2FA]);

  useEffect(() => {
    const onUnload = e => {
      if (!e || (validated2fa && current2FA.current) || !qrImage) {
        return;
      }
      e.returnValue = '2fa setup incomplete';
      return e.returnValue;
    };

    window.addEventListener('beforeunload', onUnload);
    return () => {
      if (!current2FA.current && qrImage) {
        handle2FAState(twoFAStates.disabled);
      }
      window.removeEventListener('beforeunload', onUnload);
    };
  }, [handle2FAState, qrImage, validated2fa]);
  const { classes } = useStyles();

  const validate2faSetup = formData => {
    setValidating2fa(true);
    formData.email = currentUser.email;
    return verify2FA(formData)
      .unwrap()
      .then(() => {
        setValidated2fa(true);
        setShowNotice(notificationMap.enabled2fa);
        onClose();
      })
      .catch(() => setValidated2fa(false))
      .finally(() => setValidating2fa(false));
  };

  return (
    <div className="margin-top-x-small">
      Setup:
      <div className="flexbox">
        <ol className={classes.orderedList}>
          <li className="margin-top-none margin-bottom-small">
            <Typography>Download a third party authentication app such as Authy or Google authenticator.</Typography>
          </li>
          <li className="margin-bottom-small">
            <Typography>On your authenticator application, tap the “+” icon and then select “Scan QR code” to scan the QR code below.</Typography>
          </li>
          {!qrImage ? <Loader show={!qrImage} /> : <img className={classes.qrImg} src={`data:image/png;base64,${qrImage}`} />}
          <li>
            <div>
              Once the code has been scanned, type in the generated code from the app to verify.
              <Form
                classes={classes}
                showButtons={!validating2fa}
                buttonColor="primary"
                onSubmit={validate2faSetup}
                handleCancel={onClose}
                submitLabel="Verify and Save"
              >
                <TextInput className={classes.codeInput} hint="Verification code" label="Verification code" id="token2fa" validations="isLength:6,isNumeric" />
              </Form>
              {validating2fa && (
                <div className="flexbox">
                  <Loader show={true} />
                  <Button variant="contained" color="primary" disabled={true}>
                    Verifying...
                  </Button>
                </div>
              )}
            </div>
          </li>
        </ol>
      </div>
    </div>
  );
};

export default AuthSetup;
