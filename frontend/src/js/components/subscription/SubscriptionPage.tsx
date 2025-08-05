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
import { ChangeEvent, useEffect, useMemo, useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';

import { Alert, Button, FormControl, FormControlLabel, FormHelperText, Radio, RadioGroup, TextField, Typography } from '@mui/material';

import { SupportLink } from '@northern.tech/common-ui/SupportLink';
import { ADDONS, Addon, AvailableAddon, AvailablePlans, PLANS, Plan } from '@northern.tech/store/appSlice/constants';
import { getStripeKey } from '@northern.tech/store/appSlice/selectors';
import { TIMEOUTS } from '@northern.tech/store/commonConstants';
import { getDeviceLimit } from '@northern.tech/store/devicesSlice/selectors';
import { getHasCurrentPricing, getOrganization } from '@northern.tech/store/organizationSlice/selectors';
import { getBillingPreview, getCurrentCard, getUserBilling, getUserSubscription, requestPlanChange } from '@northern.tech/store/organizationSlice/thunks';
import { useAppDispatch } from '@northern.tech/store/store';
import { useDebounce } from '@northern.tech/utils/debouncehook';
import { Elements } from '@stripe/react-stripe-js';

import { addOnsToString } from '../settings/Upgrade';
import { SubscriptionAddon } from './SubscriptionAddon';
import { SubscriptionDrawer } from './SubscriptionDrawer';
import { SubscriptionSummary } from './SubscriptionSummary';

let stripePromise = null;
export type PreviewPrice = { addons: { [key in AvailableAddon]: number }; plan: number; total: number };

const DIVISIBILITY_STEP = 50;
const enterpriseDeviceCount = PLANS.enterprise.minimalDeviceCount;
const planOrder = Object.keys(PLANS);
const enterpriseRequestPlaceholder = 'Tell us a little about your requirements and device fleet size, so we can provide you with an accurate quote';
type SelectedAddons = { [key in AvailableAddon]: boolean };

const contactReasons = {
  reduceLimit: {
    id: 'reduceLimit',
    alert: (
      <div>
        For over {enterpriseDeviceCount} devices, please contact <SupportLink variant="email" /> for pricing.
      </div>
    )
  },
  overLimit: {
    id: 'overLimit',
    alert: (
      <div>
        If you want to reduce your device limit, please contact <SupportLink variant="email" />.
      </div>
    )
  }
} as const;
interface ContactReasonProps {
  reason: keyof typeof contactReasons;
}
const ContactReasonAlert = ({ reason }: ContactReasonProps) => (
  <Alert severity="info" className="margin-bottom-x-small margin-top-x-small">
    {contactReasons[reason].alert}
  </Alert>
);

export const SubscriptionPage = () => {
  const [selectedPlan, setSelectedPlan] = useState<Plan>(PLANS.os);
  const [selectedAddons, setSelectedAddons] = useState<SelectedAddons>({ configure: false, monitor: false, troubleshoot: false });
  const [previewPrice, setPreviewPrice] = useState<PreviewPrice>();
  const [order, setOrder] = useState();
  const [contactReason, setContactReason] = useState<ContactReasonProps['reason'] | ''>('');
  const [inputHelperText, setInputHelperText] = useState<string>(`The minimum limit for ${selectedPlan.name} is ${selectedPlan.minimalDeviceCount}`);
  const [limit, setLimit] = useState<number>(selectedPlan.minimalDeviceCount);
  const [enterpriseMessage, setEnterpriseMessage] = useState('');
  const stripeAPIKey = useSelector(getStripeKey);
  const [isPreviewLoading, setIsPreviewLoading] = useState(false);
  const [showUpgradeDrawer, setShowUpgradeDrawer] = useState(false);
  const [loadingFinished, setLoadingFinished] = useState(!stripeAPIKey);

  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const currentDeviceLimit = useSelector(getDeviceLimit);
  const hasCurrentPricing = useSelector(getHasCurrentPricing);
  const org = useSelector(getOrganization);
  const { addons: orgAddOns = [], plan: currentPlan = PLANS.os.id as AvailablePlans, trial: isTrial = true, id: orgId } = org;
  const isOrgLoaded = !!orgId;
  const plan = Object.values(PLANS).find(plan => plan.id === (isTrial ? PLANS.os.id : currentPlan)) || PLANS.os;
  const enabledAddons = useMemo(() => orgAddOns.filter(addon => addon.enabled), [orgAddOns]);
  const currentPlanId = plan.id;
  const debouncedLimit = useDebounce(limit, TIMEOUTS.debounceDefault);

  // redirect customers paying old price to billing page to see the alert
  useEffect(() => {
    if (org.id && !isTrial && !hasCurrentPricing) {
      navigate('/settings/billing');
    }
  }, [isTrial, navigate, hasCurrentPricing, org]);

  //Fetch Billing profile & subscription
  useEffect(() => {
    if (isTrial) {
      return;
    }
    dispatch(getCurrentCard());
    dispatch(getUserBilling());
    dispatch(getUserSubscription());
  }, [isTrial, dispatch]);

  //Loading stripe Component
  useEffect(() => {
    // Make sure to call `loadStripe` outside of a component’s render to avoid recreating
    // the `Stripe` object on every render - but don't initialize twice.
    if (!stripePromise) {
      import(/* webpackChunkName: "stripe" */ '@stripe/stripe-js').then(({ loadStripe }) => {
        if (stripeAPIKey) {
          stripePromise = loadStripe(stripeAPIKey).finally(() => setLoadingFinished(true));
        }
      });
    } else {
      const notStripePromise = new Promise(resolve => setTimeout(resolve, TIMEOUTS.debounceDefault));
      Promise.race([stripePromise, notStripePromise]).then(result => setLoadingFinished(result !== notStripePromise));
    }
  }, [stripeAPIKey]);

  useEffect(() => {
    if (plan) {
      setSelectedPlan(plan);
    }
    setLimit(updatedLimit => (currentDeviceLimit && updatedLimit < currentDeviceLimit ? currentDeviceLimit : updatedLimit));
  }, [plan, currentDeviceLimit]);

  useEffect(() => {
    const newSelectedAddons = enabledAddons.reduce(
      (acc, addon) => {
        acc[addon.name] = addon.enabled && !isTrial;
        return acc;
      },
      {} as Record<AvailableAddon, boolean>
    );
    setSelectedAddons(newSelectedAddons);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [enabledAddons]);

  useEffect(() => {
    if (debouncedLimit >= enterpriseDeviceCount) {
      setContactReason(contactReasons.overLimit.id);
      setLimit(enterpriseDeviceCount);
      setInputHelperText(`The maximum you can set is ${enterpriseDeviceCount} devices.`);
    } else if (debouncedLimit < selectedPlan.minimalDeviceCount) {
      setLimit(selectedPlan.minimalDeviceCount);
      setInputHelperText(`The minimum limit for ${selectedPlan.name} is ${selectedPlan.minimalDeviceCount}`);
    } else if (debouncedLimit < currentDeviceLimit) {
      setLimit(currentDeviceLimit);
      setContactReason(contactReasons.reduceLimit.id);
      setInputHelperText(`Your current device limit is ${currentDeviceLimit}.`);
    } else {
      const snappedValue = Math.ceil(debouncedLimit / DIVISIBILITY_STEP) * DIVISIBILITY_STEP;
      if (snappedValue !== limit) {
        setLimit(snappedValue);
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [debouncedLimit]);

  useEffect(() => {
    setInputHelperText(`The minimum limit for ${selectedPlan.name} is ${selectedPlan.minimalDeviceCount}`);
  }, [selectedPlan]);

  useEffect(() => {
    if (!isOrgLoaded || selectedPlan.id === PLANS.enterprise.id || limit % DIVISIBILITY_STEP !== 0) {
      return;
    }
    const addons = Object.entries(selectedAddons)
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      .filter(([_, value]) => value)
      .map(([key]) => ({ name: key }));
    setIsPreviewLoading(true);
    const order = {
      preview_mode: 'recurring',
      plan: selectedPlan.id,
      products: [{ name: 'mender_standard', quantity: limit, addons }]
    };
    setOrder({ plan: order.plan, products: order.products });

    dispatch(getBillingPreview(order))
      .unwrap()
      .then(previewPrice => setPreviewPrice(previewPrice))
      .finally(() => setIsPreviewLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dispatch, selectedPlan, debouncedLimit, JSON.stringify(selectedAddons), isOrgLoaded]);

  const onChangePlan = planId => {
    //we need to reset unavailable addons from selection
    const unavailableAddons = Object.keys(selectedAddons).filter(addonId => !ADDONS[addonId].eligible.includes(planId));
    const newAddons = { ...selectedAddons, ...unavailableAddons.reduce((acc, addon) => ({ ...acc, [addon]: false }), {}) };
    setSelectedAddons(newAddons);
    setSelectedPlan(PLANS[planId]);
    setContactReason('');
    if (limit < PLANS[planId].minimalDeviceCount) {
      setLimit(PLANS[planId].minimalDeviceCount);
    }
  };

  const onChangeLimit = ({ target: { value } }) => {
    setContactReason('');
    setLimit(Number(value));
  };

  const handleDeviceLimitBlur = (event: ChangeEvent<HTMLInputElement>) => {
    const value = Number(event.target.value);
    const snappedValue = Math.ceil(value / DIVISIBILITY_STEP) * DIVISIBILITY_STEP;
    setLimit(snappedValue);
  };

  const onSendRequest = async (message = '', requestedAddons = '') => {
    await dispatch(
      requestPlanChange({
        tenantId: org.id,
        content: {
          current_plan: PLANS[currentPlan || PLANS.os.id].name,
          requested_plan: selectedPlan.name,
          current_addons: addOnsToString(org.addons) || '-',
          requested_addons: requestedAddons || addOnsToString(org.addons) || '-',
          user_message: message
        }
      })
    );
  };

  const onEnterpriseRequest = ({ message }: { message: string }) => {
    const requestedAddons = Object.entries(selectedAddons)
      .filter(([, selected]) => selected)
      .map(([key]) => key);
    setEnterpriseMessage('');
    onSendRequest(message, requestedAddons.join(', '));
  };

  const onSelectAddon = (addon: AvailableAddon, selected: boolean) => {
    setSelectedAddons({ ...selectedAddons, [addon]: selected });
  };
  const isAddonDisabled = (addon: Addon) =>
    (!isTrial && !!enabledAddons.find(enabled => enabled.name === addon.id)) || !addon.eligible.includes(selectedPlan.id);
  const selectedAddonsLength = Object.values(selectedAddons).reduce((acc, curr) => acc + Number(curr), 0);
  const isNew = currentPlanId !== selectedPlan.id || enabledAddons.length < selectedAddonsLength || debouncedLimit > currentDeviceLimit || isTrial;
  return (
    <div>
      <Typography variant="h4" className="margin-bottom-large">
        Upgrade your subscription
      </Typography>
      <Typography className="margin-bottom-small" variant="body2">
        Current plan: {isTrial ? ' Free trial' : PLANS[currentPlan].name}
      </Typography>
      <Typography variant="body1">
        Upgrade your plan or purchase an Add-on package to connect more devices, access more features and advanced support. <br />
        See the full details of plans and features at{' '}
        <a href="https://mender.io/plans/pricing" target="_blank" rel="noopener noreferrer">
          mender.io/plans/pricing
        </a>
      </Typography>
      <div className="flexbox">
        <div style={{ maxWidth: '550px' }}>
          <Typography className="margin-top" variant="subtitle1">
            1. Choose a plan
          </Typography>
          <FormControl component="fieldset">
            <RadioGroup
              row
              aria-labelledby="plan-selection"
              name="plan-selection-radio-group"
              value={selectedPlan ? selectedPlan.id : null}
              onChange={(_, value) => onChangePlan(value)}
            >
              {Object.values(PLANS).map((plan, index) => (
                <FormControlLabel
                  key={plan.id}
                  disabled={!isTrial && planOrder.indexOf(currentPlan) > index}
                  value={plan.id}
                  control={<Radio />}
                  label={plan.name}
                />
              ))}
            </RadioGroup>
          </FormControl>
          <Typography variant="body2" style={{ minHeight: '56px' }}>
            {selectedPlan.description}
          </Typography>
          {selectedPlan.id !== PLANS.enterprise.id && (
            <>
              <Typography variant="subtitle1" className="margin-top">
                2. Set a device limit
              </Typography>
              <FormControl fullWidth>
                <div className="flexbox center-aligned margin-top-x-small">
                  <TextField
                    label="Number of devices"
                    size="small"
                    type="number"
                    onChange={onChangeLimit}
                    onBlur={handleDeviceLimitBlur}
                    slotProps={{ htmlInput: { min: selectedPlan.minimalDeviceCount, step: DIVISIBILITY_STEP } }}
                    value={limit}
                    fullWidth
                  />
                </div>
                <FormHelperText className="info margin-top-none">{inputHelperText}</FormHelperText>
              </FormControl>
              {contactReason && <ContactReasonAlert reason={contactReason} />}
            </>
          )}
          <Typography variant="subtitle1" className="margin-top">
            {selectedPlan.id === PLANS.enterprise.id ? 2 : 3}. Choose Add-ons
          </Typography>
          <div className="margin-top-x-small">
            {Object.values(ADDONS).map(addon => (
              <SubscriptionAddon
                selectedPlan={selectedPlan}
                key={addon.id}
                addon={addon}
                disabled={isAddonDisabled(addon)}
                checked={selectedAddons[addon.id]}
                onChange={onSelectAddon}
              />
            ))}
          </div>
          {enabledAddons.length > 0 && !isTrial && (
            <Typography variant="body2" className="margin-bottom">
              To remove active Add-ons from your plan, please contact <SupportLink variant="email" />
            </Typography>
          )}
          {selectedPlan.id === PLANS.enterprise.id && (
            <>
              <Typography variant="subtitle1" className="margin-top">
                3. Request a quote
              </Typography>
              <FormControl fullWidth className="margin-top-small">
                <TextField
                  slotProps={{
                    inputLabel: {
                      shrink: true
                    }
                  }}
                  label="Your message"
                  name="enterprise-request-message"
                  fullWidth
                  multiline
                  placeholder={enterpriseRequestPlaceholder}
                  value={enterpriseMessage}
                  onChange={e => setEnterpriseMessage(e.target.value)}
                />
              </FormControl>
              <Button
                className="margin-top"
                color="secondary"
                disabled={!enterpriseMessage}
                onClick={() => onEnterpriseRequest({ message: enterpriseMessage })}
                variant="contained"
              >
                Submit request
              </Button>
            </>
          )}
        </div>
        <div>
          {selectedPlan.id !== PLANS.enterprise.id && previewPrice && (
            <div className="margin-top margin-left-x-large">
              <SubscriptionSummary
                isPreviewLoading={isPreviewLoading}
                plan={selectedPlan}
                addons={selectedAddons}
                deviceLimit={limit}
                title="Your subscription:"
                isNew={isNew}
                previewPrice={previewPrice}
                onAction={() => setShowUpgradeDrawer(true)}
              />
            </div>
          )}
        </div>
      </div>
      {loadingFinished && showUpgradeDrawer && (
        <Elements stripe={stripePromise}>
          <SubscriptionDrawer
            order={order}
            isTrial={isTrial}
            previewPrice={previewPrice}
            organization={org}
            plan={selectedPlan}
            addons={selectedAddons}
            onClose={() => setShowUpgradeDrawer(false)}
            currentPlanId={currentPlanId}
          />
        </Elements>
      )}
    </div>
  );
};
