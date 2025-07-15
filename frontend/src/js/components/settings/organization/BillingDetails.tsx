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

import { Alert, Button } from '@mui/material';

import { SupportLink } from '@northern.tech/common-ui/SupportLink';
import { getBillingProfile, getCard, getOrganization } from '@northern.tech/store/organizationSlice/selectors';

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
      <div>
        <b>{name}</b>
      </div>
      <div>{line1}</div>
      <div>
        {postal_code}, {city}
      </div>
      {country && <div>{displayNames.of(country) || ''}</div>}
      <div>{email}</div>
    </div>
  );
};

export const BillingDetails = props => {
  const { setChangeBilling, hideCard } = props;
  const card = useSelector(getCard);
  const organization = useSelector(getOrganization);
  const billing = useSelector(getBillingProfile);
  const { trial: isTrial } = organization;

  return (
    <>
      {billing && (
        <div>
          <div className="flexbox">
            {billing.address && <Address address={billing.address} email={billing.email} name={billing.name} />}
            {card && !hideCard && <CardDetails card={card} containerClass={billing.address ? 'margin-left-x-large' : ''} />}
          </div>
          <Button className="margin-top-x-small" onClick={() => setChangeBilling(true)} size="small">
            Edit
          </Button>
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
