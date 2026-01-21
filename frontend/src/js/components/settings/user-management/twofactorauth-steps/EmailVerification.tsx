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
import { useState } from 'react';

import { Button, DialogActions, DialogContent, Typography } from '@mui/material';

import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';

export const EmailVerificationConfirmation = ({ onClose, email }) => (
  <BaseDialog open maxWidth="sm" title="Email verification" onClose={onClose}>
    <DialogContent>
      <Typography className="margin-bottom-small">
        We sent a verification link to you at <b>{email}</b>
      </Typography>
      <Typography>Check your email for instructions to continue. If it&#39;s not there, take a quick look in your spam folder.</Typography>
    </DialogContent>
    <DialogActions>
      <Button variant="text" onClick={onClose}>
        Close
      </Button>
    </DialogActions>
  </BaseDialog>
);
export const EmailVerification = ({ verifyEmailStart, email }) => {
  const [confirmationShown, setConfirmationShown] = useState(false);

  const startVerification = () => {
    verifyEmailStart().then(() => setConfirmationShown(true));
  };

  return (
    <div className="margin-top-x-small">
      Please verify your email address first, to enable Two Factor Authentication.
      {confirmationShown && <EmailVerificationConfirmation onClose={() => setConfirmationShown(false)} email={email} />}
      <div className="flexbox center-aligned">
        <Button className="margin-top-x-small" variant="contained" color="primary" onClick={startVerification}>
          Verify
        </Button>
      </div>
    </div>
  );
};

export default EmailVerification;
