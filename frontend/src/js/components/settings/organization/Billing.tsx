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
import { Link } from 'react-router-dom';

// material ui
import { Error as ErrorIcon } from '@mui/icons-material';
import { Alert, Button, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { SupportLink } from '@northern.tech/common-ui/SupportLink';
import { ADDONS, PLANS } from '@northern.tech/store/constants';
import { getBillingProfile, getCard, getDeviceLimit, getIsEnterprise, getOrganization, getUserRoles } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { cancelRequest, getCurrentCard, getUserBilling } from '@northern.tech/store/thunks';
import { toggle } from '@northern.tech/utils/helpers';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime.js';
import pluralize from 'pluralize';

import { PlanExpanded } from '../PlanExpanded';
import CancelRequestDialog from '../dialogs/CancelRequest';
import { BillingDetails } from './BillingDetails';
import OrganizationSettingsItem from './OrganizationSettingsItem';

dayjs.extend(relativeTime);

const useStyles = makeStyles()(theme => ({
  fullWidthUpgrade: {
    '&.settings-item-main-content': {
      gridTemplateColumns: '1fr'
    }
  },
  upgradeSection: {
    backgroundColor: theme.palette.grey[100],
    borderRadius: theme.spacing(0.5),
    padding: theme.spacing(2),
    paddingTop: 0
  },
  wrapper: { gap: theme.spacing(2) }
}));

const AddOnDescriptor = ({ addOns = [], isTrial }: { addOns: string[]; isTrial: boolean }) => {
  if (!addOns.length) {
    return <>You currently don&apos;t have any add-ons</>;
  }
  return (
    <>
      You currently have the{' '}
      <b>
        {addOns.join(', ')} {pluralize('Add-on', addOns.length)}
      </b>
      {isTrial ? ' included in the trial plan' : ''}.
    </>
  );
};

export const PlanDescriptor = ({
  plan,
  isTrial,
  trialExpiration,
  deviceLimit
}: {
  deviceLimit: number;
  isTrial: boolean;
  plan: string;
  trialExpiration: string;
}) => {
  const deviceLimitNote = (
    <>
      Your device limit is <b>{deviceLimit}</b> {pluralize('device', deviceLimit)}
    </>
  );
  if (isTrial) {
    return (
      <>
        You&apos;re currently on the <b>Trial plan</b>, with your trial expiring in {dayjs().from(dayjs(trialExpiration), true)}.
        <br />
        {deviceLimitNote}
      </>
    );
  }
  return (
    <>
      You&apos;re currently on the <b>{plan} plan</b>.
      <br />
      {deviceLimitNote}
    </>
  );
};

export const DeviceLimitExpansionNotification = ({ isTrial }: { isTrial: boolean }) => (
  <div className="flexbox centered">
    <ErrorIcon className="muted margin-right-small" fontSize="small" />
    <div className="muted" style={{ marginRight: 4 }}>
      To increase your device limit,{' '}
    </div>
    {isTrial ? <Link to="/subscription">upgrade to a paid plan</Link> : <SupportLink variant="salesTeam" />}
    <div className="muted">.</div>
  </div>
);

export const CancelSubscriptionAlert = () => (
  <Alert className="margin-top-large" severity="error">
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

export const CancelSubscription = ({ handleCancelSubscription, isTrial }) => (
  <div className="margin-top-large flexbox column" style={{ gap: 8 }}>
    <Typography variant="h6" color="error">
      Delete account
    </Typography>
    <Typography variant="body2">Once you delete your account, it cannot be undone. Please be certain.</Typography>
    <div>
      <Button variant="outlined" onClick={handleCancelSubscription} color="error">
        Cancel {isTrial ? 'trial' : 'subscription'} and deactivate account
      </Button>
    </div>
  </div>
);

export const CardDetails = props => {
  const { card, containerClass } = props;
  return (
    <div className={containerClass || ''}>
      <div>Payment card ending: ****{card.last4}</div>
      <div>
        Expires {String(card.expiration.month).padStart(2, '0')}/{String(card.expiration.year).slice(-2)}
      </div>
    </div>
  );
};

const UpgradeNote = ({ isTrial }) => {
  const { classes } = useStyles();
  return (
    <div className={classes.upgradeSection}>
      <OrganizationSettingsItem
        classes={{ main: classes.fullWidthUpgrade }}
        title="Upgrade now"
        secondary={
          isTrial
            ? 'Upgrade to a paid plan to keep your access going, connect more devices, and get reliable support from our team.'
            : 'Upgrade to access more features, increase your device limit, and enhance your subscription with Add-ons.'
        }
      />
      <div className={`flexbox center-aligned margin-top-x-small ${classes.wrapper}`}>
        <Button component={Link} to="https://mender.io/pricing/plans" target="_blank" rel="noopener noreferrer" size="small">
          Compare all plans
        </Button>
        <Button color="primary" component={Link} to="/subscription" size="small" variant="contained">
          Upgrade
        </Button>
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
  const deviceLimit = useSelector(getDeviceLimit);
  const billing = useSelector(getBillingProfile);
  const { addons = [], plan: currentPlan = PLANS.os.id, trial: isTrial, trial_expiration } = organization;
  const dispatch = useAppDispatch();
  const { classes } = useStyles();

  const planName = PLANS[currentPlan].name;

  useEffect(() => {
    dispatch(getCurrentCard());
    dispatch(getUserBilling());
  }, [dispatch]);

  const enabledAddOns = addons.filter(({ enabled }) => enabled).map(({ name }) => ADDONS[name].title);

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
    <div style={{ maxWidth: 750 }}>
      <Typography variant="h6">Billing</Typography>
      <div className={`flexbox column ${classes.wrapper}`}>
        <OrganizationSettingsItem
          title="Current plan"
          secondary={<PlanDescriptor plan={planName} isTrial={isTrial} trialExpiration={trial_expiration} deviceLimit={deviceLimit} />}
        />
        <OrganizationSettingsItem title="Current Add-ons" secondary={<AddOnDescriptor addOns={enabledAddOns} isTrial={isTrial} />} />
        {!isEnterprise && <UpgradeNote isTrial={isTrial} />}
        <Typography className="margin-top-small" variant="subtitle1">
          Billing details
        </Typography>
        {isEnterprise ? (
          <Typography variant="body2">
            Enterprise plan payments are invoiced periodically to your organization. If you&apos;d like to make any changes to your plan, Add-ons, or billing
            details, please contact{' '}
            <a href="mailto:support@mender.io" target="_blank" rel="noopener noreferrer">
              support@mender.io
            </a>
            .
          </Typography>
        ) : (
          <BillingDetails setChangeBilling={setChangeBilling} />
        )}
      </div>
      {billing && changeBilling && <PlanExpanded isEdit onCloseClick={() => setChangeBilling(false)} currentBillingProfile={billing} card={card} />}
      {isAdmin && !cancelSubscriptionConfirmation && <CancelSubscription handleCancelSubscription={handleCancelSubscription} isTrial={isTrial} />}
      {cancelSubscriptionConfirmation && <CancelSubscriptionAlert />}
      {cancelSubscription && <CancelRequestDialog onCancel={() => setCancelSubscription(false)} onSubmit={cancelSubscriptionSubmit} />}
    </div>
  );
};

export default Billing;
