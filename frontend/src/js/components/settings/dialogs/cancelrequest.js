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
import React, { useEffect, useState } from 'react';

import { Button, Dialog, DialogActions, DialogContent, DialogTitle, FormControl, FormControlLabel, Input, Radio, RadioGroup } from '@mui/material';

const defaultCancelSubscriptionReasons = [
  'Just learning about Mender',
  'Too expensive',
  'Lack of features',
  'Mender did not work as expected',
  'Chose a different solution',
  'My project is delayed',
  'My project is cancelled'
];

export const CancelRequestDialog = ({ onCancel, onSubmit }) => {
  const [confirm, setConfirm] = useState(false);
  const [cancelSubscriptionReason, setCancelSubscriptionReason] = useState('');
  const [cancelSubscriptionReasonOther, setCancelSubscriptionReasonOther] = useState('');
  const [cancelSubscriptionSuggestions, setCancelSubscriptionSuggestions] = useState('');
  const [cancelSubscriptionReasons] = useState(defaultCancelSubscriptionReasons.sort(() => Math.random() - 0.5).concat(['Other']));

  useEffect(() => {
    if (cancelSubscriptionReason !== 'Other') {
      setCancelSubscriptionReasonOther('');
    }
  }, [cancelSubscriptionReason]);

  return (
    <Dialog open={true}>
      <DialogTitle>{!confirm ? 'Cancel subscription and deactivate account?' : 'Confirm deactivation'}</DialogTitle>
      {!confirm ? (
        <DialogContent>
          <p className="margin-top-small">
            We&#39;ll be sorry to see you go. If you&#39;re sure you&#39;re ready to deactivate your account please let us know the reason you&#39;re canceling,
            and we&#39;ll start the process.
          </p>
          <p>Please select the reason for your cancellation to help us improve our service:</p>
          <FormControl component="fieldset" style={{ marginTop: '0' }}>
            <RadioGroup name="cancellation-selection" onChange={e => setCancelSubscriptionReason(e.target.value)}>
              {cancelSubscriptionReasons.map((item, index) => (
                <FormControlLabel value={item} control={<Radio />} label={item} key={index} style={{ marginTop: '0px' }} />
              ))}
            </RadioGroup>
            <Input
              id="reason_other"
              name="reason_other"
              value={cancelSubscriptionReasonOther}
              disabled={cancelSubscriptionReason !== 'Other'}
              onChange={e => setCancelSubscriptionReasonOther(e.target.value)}
              placeholder="Fill in reason"
              style={{ marginLeft: '30px' }}
            />
          </FormControl>
          <p className="margin-top">Which key areas should we improve?</p>
          <Input
            id="suggestions"
            name="suggestions"
            className="margin-bottom"
            value={cancelSubscriptionSuggestions}
            onChange={e => setCancelSubscriptionSuggestions(e.target.value)}
            placeholder="Fill in suggestions"
            style={{ width: '100%' }}
          />
        </DialogContent>
      ) : (
        <DialogContent>
          <p>
            You can now click <strong>confirm deactivation</strong> to cancel your subscription and deactivate your Mender account.
          </p>
          <p>Meanwhile, you can continue using your account until the end of your current billing cycle.</p>
          <p>Thank you for using Mender!</p>
        </DialogContent>
      )}
      <DialogActions>
        <Button onClick={onCancel}>Stay subscribed</Button>
        <div style={{ flexGrow: 1 }} />
        {!confirm ? (
          <Button
            variant="contained"
            color="secondary"
            onClick={() => setConfirm(true)}
            disabled={cancelSubscriptionReason == '' || (cancelSubscriptionReason == 'Other' && cancelSubscriptionReasonOther == '')}
          >
            Continue to deactivate
          </Button>
        ) : (
          <Button
            variant="contained"
            color="secondary"
            onClick={() => onSubmit((cancelSubscriptionReasonOther || cancelSubscriptionReason) + '\n' + cancelSubscriptionSuggestions)}
          >
            Confirm deactivation
          </Button>
        )}
      </DialogActions>
    </Dialog>
  );
};

export default CancelRequestDialog;
