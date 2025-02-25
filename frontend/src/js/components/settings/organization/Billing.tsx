// Copyright 2024 Northern.tech AS
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
import { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { Link, useNavigate } from 'react-router-dom';

// material ui
import { Error as ErrorIcon, OpenInNew as OpenInNewIcon } from '@mui/icons-material';
import { Button, List } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import Alert from '@northern.tech/common-ui/Alert';
import InfoText from '@northern.tech/common-ui/InfoText';
import { ADDONS, PLANS } from '@northern.tech/store/constants';
import { getBillingProfile, getCard, getIsEnterprise, getOrganization, getUserRoles } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { cancelRequest, getCurrentCard } from '@northern.tech/store/thunks';
import { toggle } from '@northern.tech/utils/helpers';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';

import { PlanExpanded } from '../PlanExpanded';
import CancelRequestDialog from '../dialogs/CancelRequest';
import OrganizationSettingsItem, { maxWidth } from './OrganizationSettingsItem';

const useStyles = makeStyles()(theme => ({
  wrapper: {
    marginTop: theme.spacing(4),
    padding: theme.spacing(2),
    paddingBottom: theme.spacing(6),
    '&>h5': { marginTop: 0, marginBottom: 0 }
  },
  billingSection: {
    backgroundColor: theme.palette.background.lightgrey,
    padding: theme.spacing(2)
  }
}));

dayjs.extend(relativeTime);

export const TrialExpirationNote = ({ trial_expiration }) => (
  <div className="flexbox centered muted">
    <ErrorIcon fontSize="small" />
    <span className="margin-left-small">
      Your trial expires in {dayjs().from(dayjs(trial_expiration), true)}. <Link to="/settings/upgrade">Upgrade to a paid plan</Link>
    </span>
  </div>
);

export const DeviceLimitExpansionNotification = ({ isTrial }) => (
  <div className="flexbox centered">
    <ErrorIcon className="muted margin-right-small" fontSize="small" />
    <div className="muted" style={{ marginRight: 4 }}>
      To increase your device limit,{' '}
    </div>
    {isTrial ? (
      <Link to="/settings/upgrade">upgrade to a paid plan</Link>
    ) : (
      <a href="mailto:support@mender.io" target="_blank" rel="noopener noreferrer">
        contact our sales team
      </a>
    )}
    <div className="muted">.</div>
  </div>
);

export const CancelSubscriptionAlert = () => (
  <Alert className="margin-top-large" severity="error" style={{ maxWidth }}>
    <p>We&#39;ve started the process to cancel your plan and deactivate your account.</p>
    <p>
      We&#39;ll send you an email confirming your deactivation. If you have any question at all, contact us at our{' '}
      <strong>
        <a href="https://support.northern.tech" target="_blank" rel="noopener noreferrer">
          support portal
        </a>
        .
      </strong>
    </p>
  </Alert>
);

export const CancelSubscriptionButton = ({ handleCancelSubscription, isTrial }) => (
  <p className="margin-left-small margin-right-small" style={{ maxWidth }}>
    <a href="" onClick={handleCancelSubscription}>
      {isTrial ? 'End trial' : 'Cancel subscription'} and deactivate account
    </a>
  </p>
);
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
export const CardDetails = props => {
  const { card, containerClass } = props;
  return (
    <div className={containerClass || ''}>
      <div>Payment card ending: **** {card.last4}</div>
      <div>
        Expires {String(card.expiration.month).padStart(2, '0')}/{String(card.expiration.year).slice(-2)}
      </div>
    </div>
  );
};

export const Billing = () => {
  const [cancelSubscription, setCancelSubscription] = useState(false);
  const [changeBilling, setChangeBilling] = useState<boolean>(false);
  const [cancelSubscriptionConfirmation, setCancelSubscriptionConfirmation] = useState(false);
  const { isAdmin } = useSelector(getUserRoles);
  const isEnterprise = useSelector(getIsEnterprise);
  const organization = useSelector(getOrganization);
  const card = useSelector(getCard);
  const billing = useSelector(getBillingProfile);
  const { plan: currentPlan = PLANS.os.id } = organization;
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const { classes } = useStyles();

  const planName = PLANS[currentPlan].name;

  useEffect(() => {
    dispatch(getCurrentCard());
  }, [dispatch]);

  const enabledAddOns =
    organization.addons?.reduce((accu: string[], addon) => {
      if (addon.enabled) {
        const { title } = ADDONS[addon.name];
        let addonPrice = '';
        if (!organization.trial && !isEnterprise) {
          const planAddon = ADDONS[addon.name][currentPlan] ? ADDONS[addon.name][currentPlan] : ADDONS[addon.name].os;
          addonPrice = ` - ${planAddon.price}`;
        }
        accu.push(`${title}${addonPrice}`);
      }
      return accu;
    }, []) || [];

  const cancelSubscriptionSubmit = async reason =>
    dispatch(cancelRequest(reason)).then(() => {
      setCancelSubscription(false);
      setCancelSubscriptionConfirmation(true);
    });

  const handleCancelSubscription = e => {
    if (e !== undefined) {
      e.preventDefault();
    }
    setCancelSubscription(toggle);
  };

  return (
    <div className={classes.wrapper}>
      <h5>Billing</h5>
      <List>
        <OrganizationSettingsItem
          title="Current plan"
          content={{
            action: { title: 'Compare product plans', internal: false, target: 'https://mender.io/plans/pricing' },
            description: organization.trial ? 'Trial' : planName
          }}
          notification={organization.trial ? <TrialExpirationNote trial_expiration={organization.trial_expiration} /> : null}
        />
        <OrganizationSettingsItem
          title="Current add-ons"
          content={{
            action: { title: 'Purchase an add-on', internal: true, action: () => navigate('/settings/upgrade') },
            description: enabledAddOns.length ? enabledAddOns.join(', ') : `You currently don't have any add-ons`
          }}
          notification={organization.trial && <TrialExpirationNote trial_expiration={organization.trial_expiration} />}
          sideBarContent={
            <div className="margin-left-small margin-bottom">
              {/* eslint-disable-next-line react/jsx-no-target-blank */}
              <a className="flexbox center-aligned" href="https://mender.io/plans/pricing" target="_blank" rel="noopener">
                <div style={{ maxWidth: 200 }}>Compare plans and add-ons at mender.io</div>
                <OpenInNewIcon fontSize="small" />
              </a>
            </div>
          }
        />
        {billing && changeBilling && <PlanExpanded isEdit onCloseClick={() => setChangeBilling(false)} currentBillingProfile={billing} card={card} />}
        <div className={classes.billingSection}>
          <div className="flexbox center-aligned">
            <div className="padding-right-x-small">
              <b>Billing details</b>
            </div>
            {!isEnterprise && billing && (
              <Button className="margin-left" onClick={() => setChangeBilling(true)}>
                Edit
              </Button>
            )}
          </div>
          {isEnterprise ? (
            <InfoText>
              Enterprise plan payments are invoiced periodically to your organization. If you have any questions about your billing, <br /> please contact{' '}
              <a href="mailto:support@mender.io" target="_blank" rel="noopener noreferrer">
                support@mender.io
              </a>
            </InfoText>
          ) : billing ? (
            <div className="flexbox">
              {billing.address && <Address address={billing.address} email={billing.email} name={billing.name} />}
              {card && <CardDetails card={card} containerClass={billing.address ? 'margin-left-x-large' : ''} />}
            </div>
          ) : (
            <InfoText>
              Your account is not set up for automatic billing. If you believe this is a mistake, please contact{' '}
              <a href="mailto:support@mender.io" target="_blank" rel="noopener noreferrer">
                support@mender.io
              </a>
            </InfoText>
          )}
        </div>
      </List>
      {cancelSubscriptionConfirmation && <CancelSubscriptionAlert />}
      {isAdmin && !cancelSubscriptionConfirmation && (
        <CancelSubscriptionButton handleCancelSubscription={handleCancelSubscription} isTrial={organization.trial} />
      )}
      {cancelSubscription && <CancelRequestDialog onCancel={() => setCancelSubscription(false)} onSubmit={cancelSubscriptionSubmit} />}
    </div>
  );
};

export default Billing;
