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
import { Controller, useFormContext } from 'react-hook-form';
import { useSelector } from 'react-redux';

import { Alert, Button, FormControl, FormControlLabel, FormHelperText, Radio, RadioGroup, Typography, outlinedInputClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { SupportLink } from '@northern.tech/common-ui/SupportLink';
import { AddonSelect } from '@northern.tech/common-ui/forms/AddonSelect';
import Form from '@northern.tech/common-ui/forms/Form';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import { ADDONS, Addon, AddonId, AvailableAddon, AvailablePlans, PLANS, TIMEOUTS } from '@northern.tech/store/constants';
import { getDeviceLimit, getOrganization, getStripeKey } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { getBillingPreview, getCurrentCard, getUserBilling, getUserSubscription, requestPlanChange } from '@northern.tech/store/thunks';
import { useDebounce } from '@northern.tech/utils/debouncehook';
import { Elements } from '@stripe/react-stripe-js';

import { SubscriptionAddon } from './SubscriptionAddon';
import { SubscriptionDrawer } from './SubscriptionDrawer';
import { SubscriptionSummary } from './SubscriptionSummary';

let stripePromise = null;
export type PreviewPrice = { addons: { [key in AvailableAddon]: number }; plan: number; total: number };

const useStyles = makeStyles()(() => ({
  messageInput: {
    [`.${outlinedInputClasses.notchedOutline} > legend`]: {
      maxWidth: '100%'
    }
  }
}));

const DIVISIBILITY_STEP = 50;
const enterpriseDeviceCount = PLANS.enterprise.minimalDeviceCount;
const planOrder = Object.keys(PLANS);
const enterpriseRequestPlaceholder = 'Tell us a little about your requirements and device fleet size, so we can provide you with an accurate quote';
export type SelectedAddons = { [key in AvailableAddon]: boolean };

const contactReasons = {
  reduceLimit: {
    id: 'reduceLimit',
    alert: (
      <div>
        If you want to reduce your device limit, please contact <SupportLink variant="email" />.
      </div>
    )
  },
  overLimit: {
    id: 'overLimit',
    alert: (
      <div>
        For over {enterpriseDeviceCount} devices, please contact <SupportLink variant="email" /> for pricing.
      </div>
    )
  }
} as const;

const addOnsToString = (addons: Addon[] = []) =>
  addons
    .reduce((accu: string[], item) => {
      if (item.enabled) {
        accu.push(item.name);
      }
      return accu;
    }, [])
    .join(', ');

interface ContactReasonProps {
  reason: keyof typeof contactReasons;
}
const ContactReasonAlert = ({ reason }: ContactReasonProps) => (
  <Alert severity="info" className="margin-bottom-x-small margin-top-x-small">
    {contactReasons[reason].alert}
  </Alert>
);

interface FormData {
  enterpriseMessage: string;
  limit: number;
  selectedAddons: AddonId[];
  selectedPlan: string;
}

interface SubscriptionFormProps {
  onShowUpgradeDrawer: () => void;
  onUpdateFormValues: (values: Partial<FormData>) => void;
  previewPrice: PreviewPrice;
  setOrder: (order: any) => void;
  setPreviewPrice: (price: PreviewPrice) => void;
  specialHandling: boolean;
}

const SubscriptionForm = ({ onShowUpgradeDrawer, onUpdateFormValues, previewPrice, setPreviewPrice, setOrder, specialHandling }: SubscriptionFormProps) => {
  const { setValue, watch } = useFormContext<FormData>();
  const currentDeviceLimit = useSelector(getDeviceLimit);
  const org = useSelector(getOrganization);
  const dispatch = useAppDispatch();

  const { addons: orgAddOns = [], plan: currentPlan = PLANS.os.id as AvailablePlans, trial: isTrial = true, id: orgId } = org;
  const isOrgLoaded = !!orgId;
  const plan = Object.values(PLANS).find(plan => plan.id === (isTrial ? PLANS.os.id : currentPlan)) || PLANS.os;
  const enabledAddons = useMemo(() => orgAddOns.filter(addon => addon.enabled), [orgAddOns]);
  const currentPlanId = plan.id;

  const selectedPlan = PLANS[watch('selectedPlan')] || PLANS.os;
  const selectedAddons = watch('selectedAddons');
  const limit = Number(watch('limit'));
  const enterpriseMessage = watch('enterpriseMessage');
  const debouncedLimit = useDebounce(limit, TIMEOUTS.debounceDefault);

  const [contactReason, setContactReason] = useState<ContactReasonProps['reason'] | ''>('');
  const [inputHelperText, setInputHelperText] = useState<string>('');
  const [isPreviewLoading, setIsPreviewLoading] = useState(false);

  const selectedAddonsLength = selectedAddons.length;
  const isNew = currentPlanId !== selectedPlan.id || enabledAddons.length < selectedAddonsLength || debouncedLimit > currentDeviceLimit || isTrial;
  const couldGetPreview = isOrgLoaded && !specialHandling && selectedPlan.id !== PLANS.enterprise.id;

  useEffect(() => {
    onUpdateFormValues({
      selectedPlan: selectedPlan.id,
      selectedAddons,
      limit,
      enterpriseMessage
    });
  }, [selectedPlan.id, selectedAddons, limit, enterpriseMessage, onUpdateFormValues]);

  useEffect(() => {
    const eligibleAddons = selectedAddons.filter(addonId => ADDONS[addonId].eligible.includes(selectedPlan.id));
    setValue('selectedAddons', eligibleAddons);
    // Only depend on selectedPlan.id, not selectedAddons - we accept the risk of stale addons here to not run this repeatedly while still aligning addons w/ the selected plan
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedPlan.id, setValue]);

  useEffect(() => {
    if (specialHandling) return;

    if (debouncedLimit >= enterpriseDeviceCount) {
      setContactReason(contactReasons.overLimit.id);
      setInputHelperText(`The maximum you can set is ${enterpriseDeviceCount} devices.`);
    } else if (debouncedLimit < currentDeviceLimit) {
      setContactReason(contactReasons.reduceLimit.id);
      setInputHelperText(`Your current device limit is ${currentDeviceLimit}.`);
    } else {
      setContactReason('');
      setInputHelperText(`The minimum limit for ${selectedPlan.name} is ${selectedPlan.minimalDeviceCount}`);
    }
    if (debouncedLimit < selectedPlan.minimalDeviceCount) {
      setValue('limit', selectedPlan.minimalDeviceCount);
    }
  }, [currentDeviceLimit, debouncedLimit, selectedPlan.minimalDeviceCount, selectedPlan.name, setValue, specialHandling]);

  useEffect(() => {
    if (!couldGetPreview) {
      return;
    }
    const effectiveLimit = Math.min(Math.max(debouncedLimit, selectedPlan.minimalDeviceCount), enterpriseDeviceCount);
    if (!effectiveLimit || effectiveLimit % DIVISIBILITY_STEP !== 0) {
      return;
    }

    const addons = selectedAddons.filter(addonId => ADDONS[addonId]?.eligible.includes(selectedPlan.id)).map(key => ({ name: key }));
    setIsPreviewLoading(true);
    const order = {
      preview_mode: 'recurring',
      plan: selectedPlan.id,
      products: [{ name: 'mender_standard', quantity: effectiveLimit, addons }]
    };
    setOrder({ plan: order.plan, products: order.products });

    dispatch(getBillingPreview(order))
      .unwrap()
      .then(setPreviewPrice)
      .finally(() => setIsPreviewLoading(false));
  }, [couldGetPreview, debouncedLimit, dispatch, selectedPlan.id, selectedPlan.minimalDeviceCount, selectedAddons, setOrder, setPreviewPrice]);

  const handleDeviceLimitBlur = (event: ChangeEvent<HTMLInputElement>) => {
    const value = Number(event.target.value);
    const snappedValue = Math.ceil(value / DIVISIBILITY_STEP) * DIVISIBILITY_STEP;
    const effectiveLimit = Math.min(Math.max(snappedValue, selectedPlan.minimalDeviceCount), enterpriseDeviceCount);
    if (value !== effectiveLimit) {
      setValue('limit', effectiveLimit);
    }
  };

  const onEnterpriseRequest = ({ message }: { message: string }) =>
    dispatch(
      requestPlanChange({
        tenantId: org.id,
        content: {
          current_plan: PLANS[currentPlan || PLANS.os.id].name,
          requested_plan: selectedPlan.name,
          current_addons: addOnsToString(org.addons) || '-',
          requested_addons: selectedAddons.join(', ') || addOnsToString(org.addons) || '-',
          user_message: message
        }
      })
    )
      .unwrap()
      .then(() => setValue('enterpriseMessage', defaultValues.enterpriseMessage));

  const isAddonDisabled = (addon: Addon) =>
    (!isTrial && !!enabledAddons.find(enabled => enabled.name === addon.id)) || !addon.eligible.includes(selectedPlan.id);

  const { classes } = useStyles();

  return (
    <div className="flexbox">
      <div style={{ maxWidth: '550px' }}>
        <Typography className="margin-top" variant="subtitle1">
          1. Choose a plan
        </Typography>
        <Controller
          name="selectedPlan"
          render={({ field: { value, onChange } }) => (
            <FormControl component="fieldset">
              <RadioGroup
                row
                aria-labelledby="plan-selection"
                name="plan-selection-radio-group"
                value={value || ''}
                onChange={(_, newValue) => onChange(newValue)}
              >
                {Object.values(PLANS).map((plan, index) => (
                  <FormControlLabel
                    key={plan.id}
                    disabled={!isTrial && planOrder.indexOf(currentPlan) > index && !specialHandling}
                    value={plan.id}
                    control={<Radio />}
                    label={plan.name}
                  />
                ))}
              </RadioGroup>
            </FormControl>
          )}
        />
        <Typography variant="body2" style={{ minHeight: '56px' }}>
          {selectedPlan.description}
        </Typography>
        {selectedPlan.id !== PLANS.enterprise.id && !specialHandling && (
          <>
            <Typography variant="subtitle1" className="margin-top margin-bottom-x-small">
              2. Set a device limit
            </Typography>
            <TextInput
              id="limit"
              label="Number of devices"
              type="number"
              InputProps={{
                inputProps: { min: Math.max(currentDeviceLimit, selectedPlan.minimalDeviceCount), step: DIVISIBILITY_STEP },
                size: 'small',
                onBlur: handleDeviceLimitBlur
              }}
              width="100%"
            />
            <FormHelperText className="margin-left-small">{inputHelperText}</FormHelperText>
          </>
        )}
        {contactReason && selectedPlan.id !== PLANS.enterprise.id && <ContactReasonAlert reason={contactReason} />}
        <Typography variant="subtitle1" className="margin-top">
          {selectedPlan.id === PLANS.enterprise.id || specialHandling ? 2 : 3}. Choose Add-ons
        </Typography>
        <div className="margin-top-x-small">
          {selectedPlan.id === PLANS.enterprise.id || specialHandling ? (
            <AddonSelect name="selectedAddons" />
          ) : (
            <Controller
              name="selectedAddons"
              render={({ field: { value = [], onChange } }) =>
                Object.values(ADDONS).map(addon => (
                  <SubscriptionAddon
                    selectedPlan={selectedPlan}
                    key={addon.id}
                    addon={addon}
                    disabled={isAddonDisabled(addon) && !specialHandling}
                    checked={value.includes(addon.id)}
                    onChange={(addonId, selected) => {
                      if (selected) {
                        return onChange([...value, addonId]);
                      }
                      onChange(value.filter(id => id !== addonId));
                    }}
                  />
                ))
              }
            />
          )}
        </div>
        {enabledAddons.length > 0 && !isTrial && !specialHandling && selectedPlan.id !== PLANS.enterprise.id && (
          <Typography variant="body2" className="margin-bottom">
            To remove active Add-ons from your plan, please contact <SupportLink variant="email" />
          </Typography>
        )}
        {(selectedPlan.id === PLANS.enterprise.id || specialHandling) && (
          <>
            <Typography variant="subtitle1" className="margin-top">
              3. Request a quote
            </Typography>
            <div className="margin-top-small">
              <TextInput
                id="enterpriseMessage"
                label="Your message"
                InputLabelProps={{ shrink: true }}
                InputProps={{ className: classes.messageInput, multiline: true, placeholder: enterpriseRequestPlaceholder }}
                width="100%"
              />
            </div>
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
        {selectedPlan.id !== PLANS.enterprise.id && previewPrice && !specialHandling && (
          <div className="margin-top margin-left-x-large">
            <SubscriptionSummary
              isPreviewLoading={isPreviewLoading}
              plan={selectedPlan}
              addons={selectedAddons}
              deviceLimit={specialHandling ? limit : Math.min(Math.max(debouncedLimit, selectedPlan.minimalDeviceCount), enterpriseDeviceCount)}
              title="Your subscription:"
              isNew={isNew}
              previewPrice={previewPrice}
              onAction={onShowUpgradeDrawer}
            />
          </div>
        )}
      </div>
    </div>
  );
};

const defaultValues = {
  selectedPlan: PLANS.os.id,
  selectedAddons: [],
  limit: 50,
  enterpriseMessage: ''
};

export const SubscriptionPage = () => {
  const currentDeviceLimit = useSelector(getDeviceLimit);
  const stripeAPIKey = useSelector(getStripeKey);
  const org = useSelector(getOrganization);
  const dispatch = useAppDispatch();

  const { addons: orgAddOns = [], plan: currentPlan = PLANS.os.id as AvailablePlans, trial: isTrial = true } = org;
  const plan = Object.values(PLANS).find(plan => plan.id === (isTrial ? PLANS.os.id : currentPlan)) || PLANS.os;
  const enabledAddons = useMemo(() => orgAddOns.filter(addon => addon.enabled), [orgAddOns]);

  const [showUpgradeDrawer, setShowUpgradeDrawer] = useState(false);
  const [loadingFinished, setLoadingFinished] = useState(!stripeAPIKey);
  const [currentFormValues, setCurrentFormValues] = useState<Partial<FormData>>({});
  const [previewPrice, setPreviewPrice] = useState<PreviewPrice>();
  const [order, setOrder] = useState();
  const [specialHandling, setSpecialHandling] = useState(false);

  const initialValues: FormData = {
    selectedPlan: plan.id,
    selectedAddons: enabledAddons.filter(({ enabled }) => enabled && !isTrial).map(({ name }) => name),
    limit: Math.max(currentDeviceLimit || 0, plan.minimalDeviceCount),
    enterpriseMessage: ''
  };

  //Loading stripe Component
  useEffect(() => {
    // Make sure to call `loadStripe` outside of a component's render to avoid recreating
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

  //Fetch Billing profile & subscription
  useEffect(() => {
    dispatch(getUserBilling());
    if (isTrial) {
      return;
    }
    dispatch(getCurrentCard());
    //We need to handle special enterprise-like agreements
    dispatch(getUserSubscription())
      .unwrap()
      .catch(error => {
        if (!isTrial && error.message && error.message.includes('404')) {
          setSpecialHandling(true);
        }
      });
  }, [isTrial, dispatch]);

  // Form submission is handled by individual components within the form
  const onSubmit = (data: FormData) => console.log(JSON.stringify(data));

  return (
    <div className="padding-bottom-x-large">
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

      <Form initialValues={initialValues} defaultValues={defaultValues} onSubmit={onSubmit} autocomplete="off">
        <SubscriptionForm
          onShowUpgradeDrawer={setShowUpgradeDrawer}
          onUpdateFormValues={setCurrentFormValues}
          setOrder={setOrder}
          setPreviewPrice={setPreviewPrice}
          previewPrice={previewPrice}
          specialHandling={specialHandling}
        />
      </Form>

      {loadingFinished && showUpgradeDrawer && (
        <Elements stripe={stripePromise}>
          <SubscriptionDrawer
            order={order}
            isTrial={isTrial}
            previewPrice={previewPrice}
            organization={org}
            plan={PLANS[currentFormValues.selectedPlan || plan.id]}
            addons={currentFormValues.selectedAddons || initialValues.selectedAddons}
            onClose={() => setShowUpgradeDrawer(false)}
            currentPlanId={plan.id}
          />
        </Elements>
      )}
    </div>
  );
};
