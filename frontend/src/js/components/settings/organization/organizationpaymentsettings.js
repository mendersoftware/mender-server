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
import { useDispatch, useSelector } from 'react-redux';

// material ui
import { Error as ErrorIcon } from '@mui/icons-material';

import { setSnackbar } from '../../../actions/appActions';
import { confirmCardUpdate, getCurrentCard, startCardUpdate } from '../../../actions/organizationActions';
import CardSection from '../cardsection';
import OrganizationSettingsItem from './organizationsettingsitem';

export const OrganizationPaymentSettings = () => {
  const [isUpdatingPaymentDetails, setIsUpdatingPaymentDetails] = useState(false);
  const card = useSelector(state => state.organization.card);
  const hasUnpaid = useSelector(state => state.organization.billing);
  const dispatch = useDispatch();

  useEffect(() => {
    dispatch(getCurrentCard());
  }, [dispatch]);

  const onCardConfirm = async () => {
    await dispatch(confirmCardUpdate());
    dispatch(getCurrentCard());
    setIsUpdatingPaymentDetails(false);
  };

  const { last4, expiration, brand } = card;
  return (
    <OrganizationSettingsItem
      title="Payment card"
      content={{
        action: { title: `${last4 ? 'Update' : 'Enter'} payment card`, internal: true, action: () => setIsUpdatingPaymentDetails(true) },
        description: last4 ? (
          <div>
            <div>
              {brand} ending in {last4}
            </div>
            <div>
              Expires {`0${expiration.month}`.slice(-2)}/{`${expiration.year}`.slice(-2)}
            </div>
          </div>
        ) : (
          <div>
            The introduction of the PSD2 regulation in Europe requires re-entering your card details before we can allow the modification of payment details or
            access to the payment history.
          </div>
        )
      }}
      secondary={
        isUpdatingPaymentDetails && (
          <CardSection
            onCancel={() => setIsUpdatingPaymentDetails(false)}
            onComplete={onCardConfirm}
            onSubmit={() => dispatch(startCardUpdate())}
            setSnackbar={text => dispatch(setSnackbar(text))}
          />
        )
      }
      notification={
        hasUnpaid && (
          <div className="red flexbox centered">
            <ErrorIcon fontSize="small" />
            <span className="margin-left-small">You have an unpaid invoice. Please check your payment card details</span>
          </div>
        )
      }
    />
  );
};

export default OrganizationPaymentSettings;
