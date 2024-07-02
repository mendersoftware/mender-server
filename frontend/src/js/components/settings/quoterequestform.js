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

import { Button, FormControl, FormHelperText, TextField } from '@mui/material';

import { ADDONS, PLANS } from '../../constants/appConstants';

const quoteRequest = {
  default: {
    title: 'Request a change to your plan',
    note: `If you have any questions regarding plan pricing or device limits, leave us a message and we'll respond as soon as we can.`
  },
  trial: {
    title: '2. Request a quote from our team',
    note: `If you have any notes about your requirements or device fleet size, leave a message and we'll respond as soon as we can`
  },
  enterprise: {
    title: 'Request a quote from our team',
    note: `If you have any notes about your requirements or device fleet size, leave a message and we'll respond as soon as we can`
  }
};

export const QuoteRequestForm = ({ addOns, currentPlan, isTrial, onSendMessage, updatedPlan }) => {
  const [message, setMessage] = useState('');

  const isEnterpriseUpgrade = updatedPlan === 'enterprise';
  const isUpgrade = Object.keys(PLANS).indexOf(updatedPlan) > Object.keys(PLANS).indexOf(currentPlan);
  let { note, title } = isEnterpriseUpgrade ? quoteRequest.enterprise : quoteRequest.default;
  title = isTrial ? `2. ${title}` : title;
  return (
    <div className="flexbox column margin-bottom-large">
      <h3 className="margin-top-large">{title}</h3>
      {isEnterpriseUpgrade ? 'You are requesting a quote for the following subscription:' : 'You are requesting the following changes:'}
      <div>
        <p>
          Plan: {isUpgrade ? 'Upgrade to ' : ''}
          <b>{PLANS[updatedPlan].name}</b>
          {!isTrial && !!addOns.length && (
            <span>
              <br />
              Add-ons: {addOns.map(addon => ADDONS[addon.name].title).join(', ')}
            </span>
          )}
        </p>
      </div>
      <FormControl style={{ marginBottom: 30, marginTop: 0 }}>
        <FormHelperText>Your message</FormHelperText>
        <TextField fullWidth multiline placeholder={note} value={message} onChange={e => setMessage(e.target.value)} />
      </FormControl>
      <p>
        {isTrial || isEnterpriseUpgrade
          ? `After we receive your request, we'll be in touch to discuss your needs and provide a quote.`
          : `We'll send you a confirmation of any changes to your subscription.`}
      </p>
      <Button
        className="margin-top margin-bottom"
        color="secondary"
        disabled={!(message || (!isTrial && addOns.length))}
        onClick={() => onSendMessage(message)}
        style={{ alignSelf: 'flex-start' }}
        variant="contained"
      >
        Submit request
      </Button>
    </div>
  );
};

export default QuoteRequestForm;
