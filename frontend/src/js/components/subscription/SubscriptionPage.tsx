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

import { Loader } from '@northern.tech/common-ui/Loader';
import { SupportLink } from '@northern.tech/common-ui/SupportLink';
import { AddonSelect } from '@northern.tech/common-ui/forms/AddonSelect';
import Form from '@northern.tech/common-ui/forms/Form';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import type { Addon as OrgAddon } from '@northern.tech/store/api/types';
import { Addon, AddonId, AvailableAddon, AvailablePlans, TIMEOUTS } from '@northern.tech/store/constants';
import { PricePreview, ProductConfig, ProductTier } from '@northern.tech/store/organizationSlice/types';
import {
  getAcceptedDevices,
  getAppInitDone,
  getCombinedLimit,
  getDeviceLimits,
  getFeatures,
  getOrganization,
  getProducts,
  getStripeKey
} from '@northern.tech/store/selectors';
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

//With hope in enum from backend types
type DeviceTypeId = string;

type DeviceTypeConfig = ProductTier & {
  label: string;
  stripeProductName: string;
  summaryLabel: string;
  tooltipId: string;
};

export type DeviceTypes = Record<DeviceTypeId, DeviceTypeConfig>;

export type PlanPreviewPriceItem = {
  addons: { [key in AvailableAddon]?: number };
  price: number;
  quantity: number;
};
export type PlanPreviewWithTotal = { items: Record<string, PlanPreviewPriceItem>; total: number };

const enterpriseRequestPlaceholder = 'Tell us a little about your requirements and device fleet size, so we can provide you with an accurate quote';

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
    alert: (limit: string) => (
      <div>
        For over {limit} devices, please contact <SupportLink variant="email" /> for pricing.
      </div>
    )
  }
} as const;

const addOnsToString = (addons: OrgAddon[] = []): string =>
  addons
    .filter(addon => addon.enabled)
    .map(({ name }) => name)
    .join(', ');

interface ContactReasonProps {
  limit: string;
  reason: keyof typeof contactReasons;
}
const ContactReasonAlert = ({ reason, limit }: ContactReasonProps) => (
  <Alert severity="info" className="margin-bottom-x-small margin-top-x-small">
    {contactReasons[reason].alert(limit)}
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
  [plan: string]: any; // number
  enterpriseMessage: string;
  selectedAddons: AddonId[];
  selectedPlan: string;
}

interface SubscriptionFormProps {
  deviceTypes: DeviceTypes;
  initialPreviewPrice: PlanPreviewWithTotal;
  onShowUpgradeDrawer: () => void;
  onUpdateFormValues: (values: Partial<FormData>) => void;
  previewPrice: PlanPreviewWithTotal;
  products: ProductConfig;
  setOrder: (order: any) => void;
  setPreviewPrice: (price: PlanPreviewWithTotal) => void;
  specialHandling: boolean;
}
type TiersEnabled = Record<DeviceTypeId, boolean>;

const SubscriptionForm = ({
  onShowUpgradeDrawer,
  onUpdateFormValues,
  previewPrice,
  setPreviewPrice,
  setOrder,
  specialHandling,
  deviceTypes,
  products,
  initialPreviewPrice
}: SubscriptionFormProps) => {
  const { setValue, watch, control, setFocus } = useFormContext<FormData>();
  const currentDeviceLimits = useSelector(getDeviceLimits);
  const totalDeviceLimit = useSelector(getCombinedLimit);
  const { hasMCUEnabled } = useSelector(getFeatures);
  const { counts: accepted } = useSelector(getAcceptedDevices);
  const deviceStatisticsLoaded = useSelector(getAppInitDone);
  const org = useSelector(getOrganization);
  const dispatch = useAppDispatch();
  const { plans: PLANS, addons: ADDONS } = products;
  const planOrder = Object.keys(PLANS);

  const { addons: orgAddOns = [], plan: currentPlan = PLANS.os.id as AvailablePlans, trial: isTrial = true, id: orgId } = org;
  const isOrgLoaded = !!orgId;
  const plan = Object.values(PLANS).find(plan => plan.id === (isTrial ? PLANS.os.id : currentPlan)) || PLANS.os;
  const enabledAddons = useMemo(() => orgAddOns.filter(addon => addon.enabled), [orgAddOns]);
  const currentPlanId = plan.id;

  const deviceTypeIds = useMemo(() => Object.keys(deviceTypes) as DeviceTypeId[], [deviceTypes]);
  const selectedPlan = PLANS[watch('selectedPlan')] || PLANS.os;
  const selectedAddons = watch('selectedAddons');
  const watchedLimits = useWatch({ control, name: deviceTypeIds });

  const limitValues = useMemo(() => {
    if (!watchedLimits) return {};
    return Object.fromEntries(deviceTypeIds.map((id, i) => [id, watchedLimits[i]]));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(watchedLimits)]);

  const enterpriseMessage = watch('enterpriseMessage');
  const debouncedLimits = useDebounce(limitValues, TIMEOUTS.debounceDefault);

  const [contactReason, setContactReason] = useState<Record<DeviceTypeId, ContactReasonProps['reason'] | ''>>(
    deviceTypeIds.reduce((acc, curr) => ({ ...acc, [curr]: '' }), {} as Record<DeviceTypeId, ''>)
  );
  const [inputHelperText, setInputHelperText] = useState<Record<DeviceTypeId, string>>(
    deviceTypeIds.reduce((acc, curr) => ({ ...acc, [curr]: '' }), {} as Record<DeviceTypeId, string>)
  );
  const [isPreviewLoading, setIsPreviewLoading] = useState<boolean>(false);
  const [deviceTierEnabled, setDeviceTierEnabled] = useState<TiersEnabled>(deviceTypeIds.reduce((acc, curr) => ({ ...acc, [curr]: true }), {} as TiersEnabled));
  const initializedRef = useRef(false);

  const eligibleAddonTiers = useMemo(() => {
    const allAddons = Object.values(deviceTypes)
      .filter(tier => deviceTierEnabled[tier.id])
      .flatMap(tier => tier.addonsByPlan[selectedPlan.id] || []);

    return new Set(allAddons);
  }, [deviceTypes, deviceTierEnabled, selectedPlan.id]);

  const deviceLimitsInitialized = totalDeviceLimit > 0;
  const selectedAddonsLength = selectedAddons.length;
  const areRequiredTiersEnabled = (deviceTierEnabled: TiersEnabled, currentLimits: DeviceTierLimits) =>
    deviceTypeIds.every(deviceTypeId => currentLimits[deviceTypeId] < 1 || deviceTierEnabled[deviceTypeId]);

  const isLimitIncreased = (newLimits: Record<string, number>, currentLimits: DeviceTierLimits, deviceTierEnabled: TiersEnabled) =>
    deviceTypeIds.some(deviceTypeId => newLimits[deviceTypeId] > currentLimits[deviceTypeId] && deviceTierEnabled[deviceTypeId]);

  const isNew =
    currentPlanId !== selectedPlan.id ||
    enabledAddons.length < selectedAddonsLength ||
    isLimitIncreased(debouncedLimits, currentDeviceLimits, deviceTierEnabled) ||
    isTrial;
  const couldGetPreview = isOrgLoaded && !specialHandling && selectedPlan.id !== PLANS.enterprise.id && initializedRef.current;

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
  }, [accepted, isTrial, deviceLimitsInitialized, currentDeviceLimits, deviceStatisticsLoaded, deviceTypeIds, hasMCUEnabled, deviceTypes]);

  useEffect(() => {
    onUpdateFormValues({
      selectedPlan: selectedPlan.id,
      selectedAddons,
      enterpriseMessage,
      ...limitValues
    });
  }, [selectedPlan.id, selectedAddons, limitValues, enterpriseMessage, onUpdateFormValues]);

  useEffect(() => {
    const eligibleSelectedAddons = selectedAddons.filter(addonId => eligibleAddonTiers.has(addonId));
    if (eligibleSelectedAddons.length < selectedAddons.length) {
      setValue('selectedAddons', eligibleSelectedAddons);
    }
    // Only depend on selectedPlan.id, not selectedAddons - we accept the risk of stale addons here to not run this repeatedly while still aligning addons w/ the selected plan
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedPlan.id, setValue, deviceTierEnabled]);

  useEffect(() => {
    if (specialHandling || !initializedRef.current || selectedPlan.id === 'enterprise') return;
    Object.values(deviceTypes).forEach(({ id: deviceTypeId, limitConstrains }) => {
      const debouncedLimit = Number(debouncedLimits[deviceTypeId]);
      if (debouncedLimit >= limitConstrains[selectedPlan.id].max) {
        setContactReasonByTier(deviceTypeId, contactReasons.overLimit.id);
        setInputTextByTier(deviceTypeId, `The maximum you can set is ${limitConstrains[selectedPlan.id].max} devices.`);
      } else if (debouncedLimit < currentDeviceLimits[deviceTypeId]) {
        setContactReasonByTier(deviceTypeId, contactReasons.reduceLimit.id);
        setInputTextByTier(deviceTypeId, `Your current device limit is ${currentDeviceLimits[deviceTypeId]}.`);
      } else {
        setContactReasonByTier(deviceTypeId, '');
        setInputTextByTier(deviceTypeId, `The minimum limit for ${selectedPlan.name} is ${limitConstrains[selectedPlan.id].min}`);
      }
      if (debouncedLimit < limitConstrains[selectedPlan.id].min) {
        setValue(deviceTypeId, limitConstrains[selectedPlan.id].min);
      }
    });
  }, [currentDeviceLimits, debouncedLimits, deviceTypes, selectedPlan, setValue, specialHandling]);

  const setTierPreviewPrice = useCallback(
    (preview: PricePreview) => {
      const newPreviewPrice = {
        total: preview.total,
        items: {}
      };
      deviceTypeIds.forEach(deviceTypeId => {
        if (deviceTierEnabled[deviceTypeId] && preview.items[deviceTypeId]) {
          newPreviewPrice.items[deviceTypeId] = {
            quantity: debouncedLimits[deviceTypeId] || 0,
            addons: preview.items[deviceTypeId].addons,
            price: preview.items[deviceTypeId].amount
          };
        } else {
          newPreviewPrice.items[deviceTypeId] = {
            quantity: 0,
            addons: {},
            price: 0
          };
        }
      });

      setPreviewPrice(newPreviewPrice);
    },
    [deviceTypeIds, setPreviewPrice, deviceTierEnabled, debouncedLimits]
  );
  useEffect(() => {
    if (!couldGetPreview) {
      return;
    }

    const addons = selectedAddons.map(key => ({ name: key }));
    const products = [];
    for (const [deviceTypeId, deviceType] of Object.entries(deviceTypes)) {
      const debouncedLimit = Number(debouncedLimits[deviceTypeId]);
      const effectiveLimit = Math.min(
        Math.max(debouncedLimit, deviceType.limitConstrains[selectedPlan.id].min),
        deviceType.limitConstrains[selectedPlan.id].max
      );
      if (!effectiveLimit || effectiveLimit % deviceType.limitConstrains[selectedPlan.id].div !== 0) {
        return;
      }
      if (effectiveLimit && deviceTierEnabled[deviceTypeId]) {
        const product = {
          name: deviceType.stripeProductName,
          quantity: effectiveLimit
        };
        product.addons = addons.filter(({ name }) => deviceType.addonsByPlan[selectedPlan.id]?.includes(name));
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
      setPreviewPrice(initialPreviewPrice);
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
    deviceTypeIds,
    dispatch,
    selectedAddons,
    selectedPlan.id,
    setOrder,
    setPreviewPrice,
    setTierPreviewPrice,
    deviceTypes,
    initialPreviewPrice
  ]);

  const handleDeviceLimitBlur = (event: ChangeEvent<HTMLInputElement>) => {
    const value = Number(event.target.value);
    const tier = event.target.id as DeviceTypeId;
    const { div: step } = deviceTypes[tier].limitConstrains[selectedPlan.id];
    const snappedValue = Math.ceil(value / step) * step;
    const effectiveLimit = Math.min(
      Math.max(snappedValue, selectedPlan.tierLimitsConstrains[tier].min, currentDeviceLimits[tier]),
      selectedPlan.tierLimitsConstrains[tier].max
    );
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
      .then(() => setValue('enterpriseMessage', ''));

  const isAddonDisabled = (addon: Addon) => (!isTrial && !!enabledAddons.find(enabled => enabled.name === addon.id)) || !eligibleAddonTiers.has(addon.id);

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
            {Object.values(deviceTypes).map(
              deviceType =>
                deviceTierEnabled[deviceType.id] !== undefined && (
                  <div key={deviceType.id}>
                    <FormControlLabel
                      control={
                        <Checkbox
                          disabled={(!isTrial && currentDeviceLimits[deviceType.id] > 0) || !deviceStatisticsLoaded}
                          checked={deviceTierEnabled[deviceType.id]}
                          onChange={() => onToggleDeviceTier(deviceType.id)}
                        />
                      }
                      label={
                        <div className="flexbox">
                          <Typography color="textPrimary" className="capitalized-start">
                            {deviceType.label}
                          </Typography>
                          <MenderHelpTooltip id={HELPTOOLTIPS[deviceType.tooltipId].id} className="margin-left-small" />
                        </div>
                      }
                    />
                    <TextInput
                      id={deviceType.id}
                      label="Device limit"
                      disabled={!deviceTierEnabled[deviceType.id]}
                      type="number"
                      InputProps={{
                        inputProps: {
                          min: Math.max(currentDeviceLimits[deviceType.id], deviceType.limitConstrains[selectedPlan.id].min),
                          step: deviceType.limitConstrains[selectedPlan.id].div
                        },
                        size: 'small',
                        onBlur: handleDeviceLimitBlur
                      }}
                      width="100%"
                    />
                    <FormHelperText className="margin-left-small">{inputHelperText[deviceType.id]}</FormHelperText>
                    {contactReason[deviceType.id] && selectedPlan.id !== PLANS.enterprise.id && (
                      <ContactReasonAlert reason={contactReason[deviceType.id]} limit={deviceType.limitConstrains[selectedPlan.id].max} />
                    )}
                    {!deviceTierEnabled[deviceType.id] && accepted[deviceType.id] > 0 && (
                      <DevicesConnectedAlert deviceCount={accepted[deviceType.id]} deviceLabel={deviceType.label} />
                    )}
                  </div>
                )
            )}
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
                      disabledDueToTier={!eligibleAddonTiers.has(addon.id)}
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
              deviceTypes={deviceTypes}
              title="Your subscription:"
              isEnabled={isNew && areRequiredTiersEnabled(deviceTierEnabled, accepted)}
              previewPrice={previewPrice}
              onAction={onShowUpgradeDrawer}
            />
          </div>
        )}
      </div>
    </div>
  );
};

export const SubscriptionPageContent = () => {
  const currentDeviceTierLimits = useSelector(getDeviceLimits);
  const stripeAPIKey = useSelector(getStripeKey);
  const org = useSelector(getOrganization);
  const products = useSelector(getProducts);
  const dispatch = useAppDispatch();
  const { plans: PLANS = {}, tiers: backendTiers = [] } = products || {};
  const deviceTypes = useMemo(
    () =>
      Object.fromEntries(
        backendTiers.map(tier => [tier.id, { ...tier, label: `${tier.title} devices`, summaryLabel: tier.title, tooltipId: `${tier.id}Device` }])
      ),
    [backendTiers]
  );

  const initialPreviewPrice: PlanPreviewWithTotal = useMemo(
    () => ({
      items: Object.fromEntries(backendTiers.map(tier => [tier.id, { addons: {}, price: 0, quantity: 0 }])),
      total: 0
    }),
    [backendTiers]
  );

  const { addons: orgAddOns = [], plan: currentPlan = PLANS.os.id as AvailablePlans, trial: isTrial = true } = org;
  const plan = Object.values(PLANS).find(plan => plan.id === (isTrial ? PLANS.os.id : currentPlan)) || PLANS.os;
  const enabledAddons = useMemo(() => orgAddOns.filter(addon => addon.enabled), [orgAddOns]);

  const [showUpgradeDrawer, setShowUpgradeDrawer] = useState(false);
  const [loadingFinished, setLoadingFinished] = useState(!stripeAPIKey);
  const [currentFormValues, setCurrentFormValues] = useState<Partial<FormData>>({});
  const [previewPrice, setPreviewPrice] = useState<PlanPreviewWithTotal>(initialPreviewPrice);
  const [order, setOrder] = useState();
  const [specialHandling, setSpecialHandling] = useState(false);

  const tierInitialLimits = Object.fromEntries(
    Object.values(backendTiers).map(({ id, limitConstrains }) => [id, Math.max(limitConstrains[PLANS.os.id].min, currentDeviceTierLimits[id])])
  );
  const defaultValues = {
    selectedPlan: PLANS.os.id,
    selectedAddons: [],
    enterpriseMessage: '',
    ...tierInitialLimits
  };
  const initialValues: FormData = {
    selectedPlan: plan.id,
    selectedAddons: enabledAddons.filter(({ enabled }) => enabled && !isTrial).map(({ name }) => name),
    enterpriseMessage: '',
    ...tierInitialLimits
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
          deviceTypes={deviceTypes}
          products={products}
          onShowUpgradeDrawer={setShowUpgradeDrawer}
          initialPreviewPrice={initialPreviewPrice}
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
            deviceTypes={deviceTypes}
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
export const SubscriptionPage = () => {
  const products = useSelector(getProducts);

  const isLoaded = products && typeof products === 'object';

  if (!isLoaded) {
    return (
      <div className="flexbox centered">
        <Loader show />;
      </div>
    );
  }

  return <SubscriptionPageContent />;
};
