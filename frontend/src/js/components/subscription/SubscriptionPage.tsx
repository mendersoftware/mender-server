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
import { ChangeEvent, useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Controller, useFormContext, useWatch } from 'react-hook-form';
import { useSelector } from 'react-redux';
import { Link } from 'react-router-dom';

import { Alert, Button, Checkbox, FormControl, FormControlLabel, FormHelperText, Radio, RadioGroup, Typography, outlinedInputClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { SupportLink } from '@northern.tech/common-ui/SupportLink';
import { AddonSelect } from '@northern.tech/common-ui/forms/AddonSelect';
import Form from '@northern.tech/common-ui/forms/Form';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import { ADDONS, Addon, AddonId, AvailableAddon, AvailablePlans, PLANS, TIMEOUTS } from '@northern.tech/store/constants';
import { getAcceptedDevices, getAppInitDone, getCombinedLimit, getDeviceLimits, getOrganization, getStripeKey } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { getBillingPreview, getCurrentCard, getUserBilling, getUserSubscription, requestPlanChange } from '@northern.tech/store/thunks';
import type { DeviceTierLimits } from '@northern.tech/types/MenderTypes';
import { useDebounce } from '@northern.tech/utils/debouncehook';
import { Elements } from '@stripe/react-stripe-js';

import { HELPTOOLTIPS } from '../helptips/HelpTooltips';
import { MenderHelpTooltip } from '../helptips/MenderTooltip';
import { SubscriptionAddon } from './SubscriptionAddon';
import { SubscriptionDrawer } from './SubscriptionDrawer';
import { SubscriptionSummary } from './SubscriptionSummary';

let stripePromise = null;

const useStyles = makeStyles()(() => ({
  messageInput: {
    [`.${outlinedInputClasses.notchedOutline} > legend`]: {
      maxWidth: '100%'
    }
  }
}));
export const deviceTypes = {
  micro: { id: 'micro', stripeProductName: 'mender_micro', label: 'Micro devices', summaryLabel: 'Micro', tooltipId: 'mcuDevice', step: 100 },
  standard: { id: 'standard', stripeProductName: 'mender_standard', label: 'Standard devices', summaryLabel: 'Standard', tooltipId: 'standardDevice', step: 50 }
} as const;

type DeviceTypeId = keyof typeof deviceTypes;

export const deviceTypeIds = Object.keys(deviceTypes) as DeviceTypeId[];
const addonsEligibleDeviceTypes = ['standard'];

export type PreviewPrice = {
  [key in DeviceTypeId]: {
    price: number;
    quantity: number;
  };
} & {
  addons: { [key in AvailableAddon]: number };
  total: number;
};
const initialPreviewPrice: PreviewPrice = {
  addons: {},
  micro: { price: 0, quantity: 0 },
  standard: { price: 0, quantity: 0 },
  total: 0
};

const enterpriseDeviceCount = PLANS.enterprise.minimalDeviceCount;
const planOrder = Object.keys(PLANS);
const enterpriseRequestPlaceholder = 'Tell us a little about your requirements and device fleet size, so we can provide you with an accurate quote';
export type SelectedAddons = { [key in AvailableAddon]: boolean };

const contactReasons = {
  reduceLimit: {
    id: 'reduceLimit',
    alert: () => (
      <div>
        If you want to reduce your device limit, please contact <SupportLink variant="email" />.
      </div>
    )
  },
  overLimit: {
    id: 'overLimit',
    alert: (tier: DeviceTypeId) => (
      <div>
        For over {enterpriseDeviceCount[tier]} devices, please contact <SupportLink variant="email" /> for pricing.
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
  tier: DeviceTypeId;
}
const ContactReasonAlert = ({ reason, tier }: ContactReasonProps) => (
  <Alert severity="info" className="margin-bottom-x-small margin-top-x-small">
    {contactReasons[reason].alert(tier)}
  </Alert>
);
const DevicesConnectedAlert = ({ deviceCount, deviceLabel }) => (
  <Alert severity="error" className="margin-bottom-x-small margin-top-x-small">
    <Typography variant="body2" className="margin-bottom-x-small">
      You currently have {deviceCount} {deviceLabel} connected. To set {deviceLabel} devices to 0, you must first reject or decommission all {deviceLabel},
      otherwise you will need to subscribe to the minimum limit.
    </Typography>
    <Link to="/devices">View devices</Link>
  </Alert>
);

interface FormData {
  enterpriseMessage: string;
  limit: number;
  micro: number;
  selectedAddons: AddonId[];
  selectedPlan: string;
  standard: number;
}

interface SubscriptionFormProps {
  onShowUpgradeDrawer: () => void;
  onUpdateFormValues: (values: Partial<FormData>) => void;
  previewPrice: PreviewPrice;
  setOrder: (order: any) => void;
  setPreviewPrice: (price: PreviewPrice) => void;
  specialHandling: boolean;
}
type TiersEnabled = Record<DeviceTypeId, boolean>;
const areRequiredTiersEnabled = (deviceTierEnabled: TiersEnabled, currentLimits: DeviceTierLimits) =>
  deviceTypeIds.every(deviceTypeId => currentLimits[deviceTypeId] < 1 || deviceTierEnabled[deviceTypeId]);

const isLimitIncreased = (newLimits: number[], currentLimits: DeviceTierLimits, deviceTierEnabled: TiersEnabled) =>
  deviceTypeIds.some((deviceTypeId, index) => newLimits[index] > currentLimits[deviceTypeId] && deviceTierEnabled[deviceTypeId]);

const SubscriptionForm = ({ onShowUpgradeDrawer, onUpdateFormValues, previewPrice, setPreviewPrice, setOrder, specialHandling }: SubscriptionFormProps) => {
  const { setValue, watch, control, setFocus } = useFormContext<FormData>();
  const currentDeviceLimits = useSelector(getDeviceLimits);
  const totalDeviceLimit = useSelector(getCombinedLimit);
  const { counts: accepted } = useSelector(getAcceptedDevices);
  const deviceStatisticsLoaded = useSelector(getAppInitDone);
  const org = useSelector(getOrganization);
  const dispatch = useAppDispatch();

  const { addons: orgAddOns = [], plan: currentPlan = PLANS.os.id as AvailablePlans, trial: isTrial = true, id: orgId } = org;
  const isOrgLoaded = !!orgId;
  const plan = Object.values(PLANS).find(plan => plan.id === (isTrial ? PLANS.os.id : currentPlan)) || PLANS.os;
  const enabledAddons = useMemo(() => orgAddOns.filter(addon => addon.enabled), [orgAddOns]);
  const currentPlanId = plan.id;

  const selectedPlan = PLANS[watch('selectedPlan')] || PLANS.os;
  const selectedAddons = watch('selectedAddons');
  const limits = useWatch({ control, name: deviceTypeIds });
  const enterpriseMessage = watch('enterpriseMessage');
  const debouncedLimits = useDebounce(limits, TIMEOUTS.debounceDefault);

  const [contactReason, setContactReason] = useState<Record<DeviceTypeId, ContactReasonProps['reason'] | ''>>(
    deviceTypeIds.reduce((acc, curr) => ({ ...acc, [curr]: '' }), {} as Record<DeviceTypeId, ''>)
  );
  const [inputHelperText, setInputHelperText] = useState<Record<DeviceTypeId, string>>(
    deviceTypeIds.reduce((acc, curr) => ({ ...acc, [curr]: '' }), {} as Record<DeviceTypeId, string>)
  );
  const [isPreviewLoading, setIsPreviewLoading] = useState<boolean>(false);
  const [deviceTierEnabled, setDeviceTierEnabled] = useState<TiersEnabled>(deviceTypeIds.reduce((acc, curr) => ({ ...acc, [curr]: true }), {} as TiersEnabled));
  const initializedRef = useRef(false);
  const deviceLimitsInitialized = totalDeviceLimit > 0;
  const selectedAddonsLength = selectedAddons.length;
  const isNew =
    currentPlanId !== selectedPlan.id ||
    enabledAddons.length < selectedAddonsLength ||
    isLimitIncreased(debouncedLimits, currentDeviceLimits, deviceTierEnabled) ||
    isTrial;
  const couldGetPreview = isOrgLoaded && !specialHandling && selectedPlan.id !== PLANS.enterprise.id;

  // Enable tiers the user has devices or already bought
  useEffect(() => {
    if (!deviceStatisticsLoaded || initializedRef.current || !deviceLimitsInitialized) return;
    const newEnabled = deviceTypeIds.reduce<Record<DeviceTypeId, boolean>>(
      (acc, curr) => ({ ...acc, [curr]: accepted[curr] > 0 || (currentDeviceLimits[curr] > 0 && !isTrial) }),
      {} as Record<DeviceTypeId, boolean>
    );
    if (accepted.total === 0) {
      newEnabled.standard = true;
    }
    setDeviceTierEnabled(newEnabled);
    initializedRef.current = true;
  }, [accepted, isTrial, deviceLimitsInitialized, currentDeviceLimits, deviceStatisticsLoaded]);

  useEffect(() => {
    const [micro, standard] = limits;
    onUpdateFormValues({
      selectedPlan: selectedPlan.id,
      selectedAddons,
      micro,
      standard,
      enterpriseMessage
    });
  }, [selectedPlan.id, selectedAddons, limits, enterpriseMessage, onUpdateFormValues]);

  useEffect(() => {
    const eligibleAddonTierEnabled = deviceTierEnabled[addonsEligibleDeviceTypes[0]];
    const eligibleAddons = selectedAddons.filter(addonId => ADDONS[addonId].eligible.includes(selectedPlan.id));
    setValue('selectedAddons', eligibleAddonTierEnabled ? eligibleAddons : []);
    // Only depend on selectedPlan.id, not selectedAddons - we accept the risk of stale addons here to not run this repeatedly while still aligning addons w/ the selected plan
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedPlan.id, setValue, deviceTierEnabled]);

  useEffect(() => {
    if (specialHandling || !initializedRef.current) return;
    deviceTypeIds.forEach((deviceTypeId, index) => {
      const debouncedLimit = Number(debouncedLimits[index]);
      if (debouncedLimit >= enterpriseDeviceCount[deviceTypeId]) {
        setContactReasonByTier(deviceTypeId, contactReasons.overLimit.id);
        setInputTextByTier(deviceTypeId, `The maximum you can set is ${enterpriseDeviceCount[deviceTypeId]} devices.`);
      } else if (debouncedLimit < currentDeviceLimits[deviceTypeId]) {
        setContactReasonByTier(deviceTypeId, contactReasons.reduceLimit.id);
        setInputTextByTier(deviceTypeId, `Your current device limit is ${currentDeviceLimits[deviceTypeId]}.`);
      } else {
        setContactReasonByTier(deviceTypeId, '');
        setInputTextByTier(deviceTypeId, `The minimum limit for ${selectedPlan.name} is ${selectedPlan.minimalDeviceCount[deviceTypeId]}`);
      }
      if (debouncedLimit < selectedPlan.minimalDeviceCount[deviceTypeId]) {
        setValue(deviceTypeId, selectedPlan.minimalDeviceCount[deviceTypeId]);
      }
    });
  }, [currentDeviceLimits, debouncedLimits, selectedPlan.minimalDeviceCount, selectedPlan.name, setValue, specialHandling]);

  const setTierPreviewPrice = useCallback(
    preview => {
      const newPreviewPrice = {
        addons: preview.addons,
        total: preview.total
      };
      deviceTypeIds.forEach((deviceTypeId, index) => {
        if (deviceTierEnabled[deviceTypeId]) {
          newPreviewPrice[deviceTypeId] = {
            quantity: debouncedLimits[index] || 0,
            price: preview[deviceTypeId]
          };
        } else {
          newPreviewPrice[deviceTypeId] = {
            quantity: 0,
            price: 0
          };
        }
      });

      setPreviewPrice(newPreviewPrice);
    },
    [deviceTierEnabled, debouncedLimits, setPreviewPrice]
  );
  useEffect(() => {
    if (!couldGetPreview) {
      return;
    }

    const addons = selectedAddons.filter(addonId => ADDONS[addonId]?.eligible.includes(selectedPlan.id)).map(key => ({ name: key }));
    const products = [];
    for (const [index, deviceTypeId] of deviceTypeIds.entries()) {
      const debouncedLimit = Number(debouncedLimits[index]);
      const effectiveLimit = Math.min(Math.max(debouncedLimit, selectedPlan.minimalDeviceCount[deviceTypeId]), enterpriseDeviceCount[deviceTypeId]);
      if (!effectiveLimit || effectiveLimit % deviceTypes[deviceTypeId].step !== 0) {
        return;
      }
      if (effectiveLimit && deviceTierEnabled[deviceTypeId]) {
        const product = {
          name: deviceTypes[deviceTypeId].stripeProductName,
          quantity: effectiveLimit
        };
        product.addons = addonsEligibleDeviceTypes.includes(deviceTypeId) ? addons : [];
        products.push(product);
      }
    }
    setIsPreviewLoading(true);
    const order = {
      preview_mode: 'recurring',
      plan: selectedPlan.id,
      products
    };
    setOrder({ plan: order.plan, products: order.products });
    if (products.length === 0) {
      setTierPreviewPrice(initialPreviewPrice);
      setIsPreviewLoading(false);
    } else {
      dispatch(getBillingPreview(order))
        .unwrap()
        .then(setTierPreviewPrice)
        .finally(() => setIsPreviewLoading(false));
    }
  }, [
    couldGetPreview,
    debouncedLimits,
    deviceTierEnabled,
    dispatch,
    selectedAddons,
    selectedPlan.id,
    selectedPlan.minimalDeviceCount,
    setOrder,
    setPreviewPrice,
    setTierPreviewPrice
  ]);

  const handleDeviceLimitBlur = (event: ChangeEvent<HTMLInputElement>) => {
    const value = Number(event.target.value);
    const tier = event.target.id as DeviceTypeId;
    const { step } = deviceTypes[tier];
    const snappedValue = Math.ceil(value / step) * step;
    const effectiveLimit = Math.min(Math.max(snappedValue, selectedPlan.minimalDeviceCount[tier], currentDeviceLimits[tier]), enterpriseDeviceCount[tier]);
    if (value !== effectiveLimit) {
      setValue(tier, effectiveLimit);
    }
  };
  const setContactReasonByTier = (tier: DeviceTypeId, reason: ContactReasonProps['reason'] | '') => {
    setContactReason(curr => ({ ...curr, [tier]: reason }));
  };
  const setInputTextByTier = (tier: DeviceTypeId, text: string) => {
    setInputHelperText(curr => ({ ...curr, [tier]: text }));
  };
  const onToggleDeviceTier = (tier: DeviceTypeId) => {
    const isEnabling = !deviceTierEnabled[tier];
    setContactReasonByTier(tier, '');
    setDeviceTierEnabled(deviceTierEnabled => ({ ...deviceTierEnabled, [tier]: !deviceTierEnabled[tier] }));
    if (isEnabling) {
      // We need to wait for input to be enabled to focus on it
      setTimeout(() => setFocus(tier), 0);
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
    (!isTrial && !!enabledAddons.find(enabled => enabled.name === addon.id)) ||
    !addon.eligible.includes(selectedPlan.id) ||
    !deviceTierEnabled[addonsEligibleDeviceTypes[0]];

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
            {Object.values(deviceTypes).map(deviceType => (
              <div key={deviceType.id}>
                <FormControlLabel
                  control={<Checkbox checked={deviceTierEnabled[deviceType.id]} onChange={() => onToggleDeviceTier(deviceType.id)} />}
                  label={
                    <div className="flexbox">
                      <Typography>{deviceType.label}</Typography>
                      <MenderHelpTooltip id={HELPTOOLTIPS[deviceType.tooltipId].id} className="margin-left-small" />
                    </div>
                  }
                  disabled={(!isTrial && currentDeviceLimits[deviceType.id] > 0) || !deviceStatisticsLoaded}
                />
                <TextInput
                  id={deviceType.id}
                  label="Device limit"
                  disabled={!deviceTierEnabled[deviceType.id]}
                  type="number"
                  InputProps={{
                    inputProps: { min: Math.max(currentDeviceLimits[deviceType.id], selectedPlan.minimalDeviceCount[deviceType.id]), step: deviceType.step },
                    size: 'small',
                    onBlur: handleDeviceLimitBlur
                  }}
                  width="100%"
                />
                <FormHelperText className="margin-left-small">{inputHelperText[deviceType.id]}</FormHelperText>
                {contactReason[deviceType.id] && selectedPlan.id !== PLANS.enterprise.id && (
                  <ContactReasonAlert reason={contactReason[deviceType.id]} tier={deviceType.id} />
                )}
                {!deviceTierEnabled[deviceType.id] && accepted[deviceType.id] > 0 && (
                  <DevicesConnectedAlert deviceCount={accepted[deviceType.id]} deviceLabel={deviceType.label} />
                )}
              </div>
            ))}
            {Object.values(deviceTierEnabled).every(enabled => !enabled) && accepted.total === 0 && (
              <Alert severity="error">You must select at least one device limit for your new plan.</Alert>
            )}
          </>
        )}
        <Typography variant="subtitle1" className="margin-top">
          {selectedPlan.id === PLANS.enterprise.id || specialHandling ? 2 : 3}. Choose Add-ons
        </Typography>
        <Typography variant="caption">Add-ons are currently available for Standard devices only.</Typography>
        <div className="margin-top-x-small">
          {selectedPlan.id === PLANS.enterprise.id || specialHandling ? (
            <AddonSelect name="selectedAddons" />
          ) : (
            <Controller
              name="selectedAddons"
              render={({ field: { value = [], onChange } }) => (
                <>
                  {Object.values(ADDONS).map(addon => (
                    <SubscriptionAddon
                      selectedPlan={selectedPlan}
                      key={addon.id}
                      addon={addon}
                      disabled={isAddonDisabled(addon) && !specialHandling}
                      disabledDueToTier={!deviceTierEnabled[addonsEligibleDeviceTypes[0]]}
                      checked={value.includes(addon.id)}
                      onChange={(addonId, selected) => {
                        if (selected) {
                          return onChange([...value, addonId]);
                        }
                        onChange(value.filter(id => id !== addonId));
                      }}
                    />
                  ))}
                </>
              )}
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
  const { standard: currentDeviceLimit, micro: currentMicroDeviceLimit } = useSelector(getDeviceLimits);
  const stripeAPIKey = useSelector(getStripeKey);
  const org = useSelector(getOrganization);
  const dispatch = useAppDispatch();

  const { addons: orgAddOns = [], plan: currentPlan = PLANS.os.id as AvailablePlans, trial: isTrial = true } = org;
  const plan = Object.values(PLANS).find(plan => plan.id === (isTrial ? PLANS.os.id : currentPlan)) || PLANS.os;
  const enabledAddons = useMemo(() => orgAddOns.filter(addon => addon.enabled), [orgAddOns]);

  const [showUpgradeDrawer, setShowUpgradeDrawer] = useState(false);
  const [loadingFinished, setLoadingFinished] = useState(!stripeAPIKey);
  const [currentFormValues, setCurrentFormValues] = useState<Partial<FormData>>({});
  const [previewPrice, setPreviewPrice] = useState<PreviewPrice>(initialPreviewPrice);
  const [order, setOrder] = useState();
  const [specialHandling, setSpecialHandling] = useState(false);

  const initialValues: FormData = {
    selectedPlan: plan.id,
    selectedAddons: enabledAddons.filter(({ enabled }) => enabled && !isTrial).map(({ name }) => name),
    micro: Math.max(currentMicroDeviceLimit || 0, plan.minimalDeviceCount.micro),
    standard: Math.max(currentDeviceLimit || 0, plan.minimalDeviceCount.standard),
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
