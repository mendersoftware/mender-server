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
import React, { useState } from 'react';

import { Button } from '@mui/material';

import { CardElement, useElements, useStripe } from '@stripe/react-stripe-js';

import stripeImage from '../../../assets/img/powered_by_stripe.png';
import InfoText from '../common/infotext';
import Loader from '../common/loader';

const CardSection = ({ isSignUp, onCancel, onComplete, onSubmit, setSnackbar }) => {
  const stripe = useStripe();
  const elements = useElements();
  const [errors, setErrors] = useState(false);
  const [loading, setLoading] = useState(false);
  const [empty, setEmpty] = useState(true);

  const handleSubmit = async event => {
    event.preventDefault();
    setLoading(true);
    return onSubmit()
      .then(confirmCard)
      .finally(() => setLoading(false));
  };

  const confirmCard = async secret => {
    // Use elements.getElement to get a reference to the mounted Element.
    const cardElement = elements.getElement(CardElement);

    // Use your card Element with other Stripe.js APIs
    try {
      const result = await stripe.confirmCardSetup(secret, {
        payment_method: {
          card: cardElement
        }
      });

      if (result.error) {
        setSnackbar(`Error while confirming card: ${result.error.message}`);
        onCancel();
      } else {
        setSnackbar(`Card confirmed. Updating your account...`);
        onComplete();
      }
    } catch (err) {
      setSnackbar(`Something went wrong while submitting the form. Please contact support.`);
      onCancel();
    }
  };

  const stripeElementChange = event => {
    setEmpty(false);
    if (event.complete) {
      // enable payment button
      setErrors(false);
    } else if (event.error) {
      // show validation to customer
      setErrors(true);
    }
  };

  return (
    <form className="margin-top-small" onSubmit={handleSubmit} onReset={onCancel}>
      <CardElement className="warning" onChange={event => stripeElementChange(event)} />
      {!!errors && <p className="warning">There is an error in the form. Please check that your details are correct</p>}

      <div id="poweredByStripe">
        <div>All standard credit card fees apply (e.g. foreign transaction fee, if your card has one)</div>
        <img src={stripeImage} />
      </div>

      {isSignUp && <InfoText>Billing will be scheduled monthly, starting from today. You can cancel at any time.</InfoText>}

      <div className="flexbox center-aligned margin-top-small" style={{ justifyContent: 'flex-end' }}>
        {!isSignUp && (
          <Button type="reset" disabled={loading} style={{ marginRight: 15 }}>
            Cancel
          </Button>
        )}
        <Button variant="contained" color="secondary" type="submit" disabled={errors || loading || empty}>
          {isSignUp ? 'Sign up' : 'Save'}
        </Button>
      </div>
      <Loader show={loading} />
    </form>
  );
};

export default CardSection;
