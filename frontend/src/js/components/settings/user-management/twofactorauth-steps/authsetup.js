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
import React, { useEffect, useRef, useState } from 'react';

import { CheckCircle as CheckCircleIcon } from '@mui/icons-material';
import { Button } from '@mui/material';

import { twoFAStates } from '../../../../constants/userConstants';
import Form from '../../../common/forms/form';
import TextInput from '../../../common/forms/textinput';
import Loader from '../../../common/loader';

export const AuthSetup = ({ currentUser, handle2FAState, has2FA, onClose, qrImage, verify2FA }) => {
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

  const validate2faSetup = formData => {
    setValidating2fa(true);
    formData.email = currentUser.email;
    return verify2FA(formData)
      .then(() => setValidated2fa(true))
      .catch(() => setValidated2fa(false))
      .finally(() => setValidating2fa(false));
  };

  return (
    <div className="margin-top">
      Setup:
      <div className="flexbox margin-top">
        <ol className="spaced-list margin-right-large" style={{ paddingInlineStart: 20 }}>
          <li className="margin-top-none">
            To use Two Factor Authentication, first download a third party authentication app such as{' '}
            <a href="https://authy.com/download/" target="_blank" rel="noopener noreferrer">
              Authy
            </a>{' '}
            or{' '}
            <a href="https://support.google.com/accounts/answer/1066447" target="_blank" rel="noopener noreferrer">
              Google Authenticator
            </a>
            .
          </li>
          <li>Scan the QR code on the right with the authenticator app you just downloaded on your device.</li>
          <li>
            <div>
              Type in the generated code in the input field below and click Verify.
              {validated2fa ? (
                <div className="flexbox space-between centered margin-top margin-right margin-bottom" style={{ justifyContent: 'flex-end' }}>
                  <CheckCircleIcon className="green" />
                  <h3 className="green margin-left-small" style={{ textTransform: 'uppercase' }}>
                    Verified
                  </h3>
                </div>
              ) : (
                <>
                  <Form showButtons={!validating2fa} buttonColor="primary" onSubmit={validate2faSetup} submitLabel="Verify">
                    <TextInput hint="Verification code" label="Verification code" id="token2fa" validations="isLength:6,isNumeric" required={true} />
                  </Form>
                  {validating2fa && (
                    <div className="flexbox" style={{ alignItems: 'flex-end', justifyContent: 'flex-end', height: 'min-content' }}>
                      <Loader show={true} />
                      <Button variant="contained" color="primary" disabled={true} style={{ marginLeft: 30 }}>
                        Verifying...
                      </Button>
                    </div>
                  )}
                </>
              )}
            </div>
          </li>
          <li>Then each time you log in, you will be asked for a verification code which you can retrieve from the authentication app on your device.</li>
        </ol>
        {!qrImage ? <Loader show={!qrImage} /> : <img src={`data:image/png;base64,${qrImage}`} style={{ maxHeight: '20vh' }} />}
      </div>
      <div className="flexbox" style={{ justifyContent: 'flex-end' }}>
        <Button onClick={() => handle2FAState(twoFAStates.disabled)} style={{ marginRight: 10 }}>
          Cancel
        </Button>
        <Button variant="contained" color="secondary" disabled={!validated2fa} onClick={onClose}>
          Save
        </Button>
      </div>
    </div>
  );
};

export default AuthSetup;
