// Copyright 2025 Northern.tech AS
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
import { useSelector } from 'react-redux';

import { Alert, Button, Typography } from '@mui/material';

import { SupportLink } from '@northern.tech/common-ui/SupportLink';
import { getBillingProfile, getCard, getOrganization } from '@northern.tech/store/organizationSlice/selectors';
import { getCurrentUser } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { createBillingPortalSession } from '@northern.tech/store/thunks';

import { CardDetails } from './Billing';

const Address = props => {
  const {
    address: { city, country, line1, postal_code },
    name,
    email
  } = props;
  const displayNames = new Intl.DisplayNames('en', { type: 'region' });
  return (
    <div>
      <Typography color="textSecondary">{name}</Typography>
      <Typography>{line1}</Typography>
      <Typography>
        {postal_code}, {city}
      </Typography>
      {country && <Typography>{displayNames.of(country) || ''}</Typography>}
      <Typography>{email}</Typography>
    </div>
  );
};

export const BillingDetails = props => {
  const { setChangeBilling, hideCard, editDisabled } = props;
  const dispatch = useAppDispatch();
  const card = useSelector(getCard);
  const organization = useSelector(getOrganization);
  const billing = useSelector(getBillingProfile);
  const { id: currentUserId } = useSelector(getCurrentUser);
  const { trial: isTrial } = organization;

  // gated on 1) an authenticated session (currentUserId is only set once logged in) and
  // 2) an existing Stripe billing profile - opening a portal session makes no sense otherwise,
  // since Stripe has no customer record to manage yet.
  const canManageBilling = !!currentUserId && !!billing;

  const onManageBillingClick = () =>
    dispatch(createBillingPortalSession(undefined))
      .unwrap()
      .then(({ url }) => {
        window.location.assign(url);
      });

  return (
    <>
      {billing && (
        <div>
          <div className="flexbox">
            {billing.address && <Address address={billing.address} email={billing.email} name={billing.name} />}
            {card && !hideCard && <CardDetails card={card} containerClass={billing.address ? 'margin-left-x-large' : ''} />}
          </div>
          <div className="flexbox align-items-center" style={{ gap: 8 }}>
            <Button disabled={editDisabled} variant="outlined" className="margin-top-x-small" onClick={() => setChangeBilling(true)}>
              Edit
            </Button>
            {canManageBilling && (
              <Button variant="outlined" className="margin-top-x-small" onClick={onManageBillingClick}>
                Manage billing
              </Button>
            )}
          </div>
        </div>
      )}
      {!billing && !isTrial && (
        <Alert severity="warning">
          Your account is not set up for automatic billing. If you believe this is a mistake, please contact <SupportLink variant="email" />
        </Alert>
      )}
    </>
  );
};
