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
import React, { useEffect, useState } from 'react';

import { Button } from '@mui/material';

import { TIMEOUTS } from '../../../../constants/appConstants';
import Form from '../../../common/forms/form';
import TextInput from '../../../common/forms/textinput';
import Loader from '../../../common/loader';

export const EmailVerification = ({ activationCode, verifyEmailComplete, verifyEmailStart }) => {
  const [isVerifying, setIsVerifying] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    setIsVerifying(Boolean(activationCode));
  }, [activationCode]);

  const startVerification = () => {
    setIsLoading(true);
    setTimeout(() => {
      setIsVerifying(true);
      setIsLoading(false);
    }, TIMEOUTS.threeSeconds);
    verifyEmailStart().catch(() => setIsVerifying(false));
  };

  const completeVerification = formData => {
    verifyEmailComplete(formData.emailVerification);
  };

  return (
    <div className="margin-top">
      Please verify your email address first, to enable Two Factor Authentication.
      <br />
      Once you click the button below we will send you an email with a confirmation link in it and a confirmation field will appear below. Click on the link to
      complete the verification. If the link does not work, you can also enter the confirmation code from the link in a confirmation field that will appear
      below once you clicked the button.
      {!isVerifying ? (
        <div className="flexbox center-aligned">
          <Button variant="contained" disabled={isLoading} color="primary" onClick={startVerification} style={{ marginTop: 20, marginRight: 30 }}>
            Verify your email address
          </Button>
          <Loader show={isLoading} />
        </div>
      ) : (
        <Form showButtons={isVerifying} buttonColor="primary" onSubmit={completeVerification} submitLabel="Verify">
          <TextInput hint="Verification code" label="Verification code" id="emailVerification" required={true} value={activationCode} />
        </Form>
      )}
    </div>
  );
};

export default EmailVerification;
