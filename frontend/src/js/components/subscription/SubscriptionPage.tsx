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

import { Alert, Button, FormControl, FormControlLabel, FormHelperText, Radio, RadioGroup, TextField, Typography } from '@mui/material';

import { ADDONS, Addon, AvailableAddon, AvailablePlans, PLANS, Plan } from '@northern.tech/store/appSlice/constants';
import { TIMEOUTS } from '@northern.tech/store/commonConstants';
import { getDeviceLimit } from '@northern.tech/store/devicesSlice/selectors';
import { getOrganization } from '@northern.tech/store/organizationSlice/selectors';
import { requestPlanChange } from '@northern.tech/store/organizationSlice/thunks';
import { useAppDispatch } from '@northern.tech/store/store';
import { useDebounce } from '@northern.tech/utils/debouncehook';

import { addOnsToString } from '../settings/Upgrade';
import { SubscriptionAddon } from './SubscriptionAddon';
import { SubscriptionSummary } from './SubscriptionSummary';

const DIVISIBILITY_STEP = 50;
const enterpriseDeviceCount = PLANS.enterprise.minimalDeviceCount;
const enterpriseRequestPlaceholder = 'Tell us a little about your requirements and device fleet size, so we can provide you with an accurate quote';
type SelectedAddons = { [key in AvailableAddon]: boolean };
export const SubscriptionPage = () => {
  const [selectedPlan, setSelectedPlan] = useState<Plan>(PLANS.os);
  const [selectedAddons, setSelectedAddons] = useState<SelectedAddons>({ configure: false, monitor: false, troubleshoot: false });
  const [contactReason, setContactReason] = useState<string>('');
  const [inputHelperText, setInputHelperText] = useState<string>(`The minimum limit for ${selectedPlan.name} is ${selectedPlan.minimalDeviceCount}`);
  const [limit, setLimit] = useState<number>(selectedPlan.minimalDeviceCount);
  const [enterpriseMessage, setEnterpriseMessage] = useState('');

  const dispatch = useAppDispatch();
  const currentDeviceLimit = useSelector(getDeviceLimit);
  const org = useSelector(getOrganization);
  const { addons: orgAddOns = [], plan: currentPlan = PLANS.os.id as AvailablePlans, trial: isTrial = true } = org;
  const plan = Object.values(PLANS).find(plan => plan.id === (isTrial ? PLANS.os.id : currentPlan));
  const enabledAddons = useMemo(() => orgAddOns.filter(addon => addon.enabled), [orgAddOns]);
  const currentPlanId = plan ? plan.id : null;
  const isProfessionalDisabled = !isTrial && currentPlanId === 'enterprise';
  const isBasicDisabled = (!isTrial && currentPlanId === 'professional') || isProfessionalDisabled;
  const debouncedLimit = useDebounce(limit, TIMEOUTS.debounceDefault);

  const onChangeLimit = ({ target: { value } }) => {
    setLimit(value);
  };

  useEffect(() => {
    if (plan) {
      setSelectedPlan(plan);
    }
  }, [plan]);

  useEffect(() => {
    const newSelectedAddons: Record<AvailableAddon, boolean> = {} as Record<AvailableAddon, boolean>;
    enabledAddons.forEach(addon => {
      newSelectedAddons[addon.name] = addon.enabled && !isTrial;
    });
    setSelectedAddons(newSelectedAddons);
  }, [enabledAddons]);

  useEffect(() => {
    if (debouncedLimit >= enterpriseDeviceCount) {
      setContactReason(`For over ${enterpriseDeviceCount} devices, please contact for pricing.`);
      setLimit(enterpriseDeviceCount);
      setInputHelperText(`The maximum you can set is ${enterpriseDeviceCount} devices.`);
    } else if (debouncedLimit < currentDeviceLimit) {
      setLimit(currentDeviceLimit);
      setContactReason(`If you want to reduce your device limit, please contact support@mender.io.`);
      setInputHelperText(`Your current device limit is ${currentDeviceLimit}.`);
    }
    const snappedValue = Math.ceil(debouncedLimit / DIVISIBILITY_STEP) * DIVISIBILITY_STEP;
    if (snappedValue !== limit) {
      setLimit(snappedValue);
    }
  }, [debouncedLimit]);

  useEffect(() => {
    setInputHelperText(`The minimum limit for ${selectedPlan.name} is ${selectedPlan.minimalDeviceCount}`);
  }, [selectedPlan]);

  const onChangePlan = planId => {
    //we need to reset unavailable addons from selection
    const unavailableAddons = Object.keys(selectedAddons).filter(addonId => !ADDONS[addonId].eligible.includes(planId));
    const newAddons = { ...selectedAddons, ...unavailableAddons.reduce((acc, addon) => ({ ...acc, [addon]: false }), {}) };
    setSelectedAddons(newAddons);
    setSelectedPlan(PLANS[planId]);
    setContactReason('');
    setLimit(PLANS[planId].minimalDeviceCount);
  };
  const handleBlur = (event: ChangeEvent<HTMLInputElement>) => {
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
  const isNew = currentPlanId !== selectedPlan.id || enabledAddons.length < selectedAddonsLength || debouncedLimit > currentDeviceLimit;
  return (
    <div>
      <Typography variant="h4" className="margin-bottom-large">
        Upgrade your subscription
      </Typography>
      <Typography variant="body2">Current plan: {isTrial ? ' Free trial' : PLANS[currentPlan].name}</Typography>
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
              <FormControlLabel disabled={isBasicDisabled} value="os" control={<Radio />} label="Basic" />
              <FormControlLabel disabled={isProfessionalDisabled} value="professional" control={<Radio />} label="Professional" />
              <FormControlLabel value="enterprise" control={<Radio />} label="Enterprise" />
            </RadioGroup>
          </FormControl>
          <Typography variant="body2" sx={{ minHeight: '56px' }}>
            {selectedPlan.description}
          </Typography>
          {selectedPlan.id !== 'enterprise' && (
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
                    onBlur={handleBlur}
                    slotProps={{ htmlInput: { min: selectedPlan.minimalDeviceCount, step: DIVISIBILITY_STEP } }}
                    value={limit}
                    fullWidth
                  />
                </div>
                <FormHelperText className="info margin-top-none">{inputHelperText}</FormHelperText>
              </FormControl>
              {contactReason && (
                <Alert severity="info" className="margin-bottom-x-small">
                  {contactReason}
                </Alert>
              )}
            </>
          )}
          <Typography variant="subtitle1" className="margin-top">
            {selectedPlan.id === 'enterprise' ? 2 : 3}. Choose Add-ons
          </Typography>
          <div className="margin-top-x-small">
            {Object.values(ADDONS)
              .sort((a, b) => a.id.localeCompare(b.id))
              .map(addon => (
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
          {enabledAddons.length > 0 && <Typography>To remove active Add-ons from your plan, please contact us</Typography>}
          {selectedPlan.id === 'enterprise' && (
            <>
              <Typography variant="subtitle1" className="margin-top">
                3. Request a quote
              </Typography>
              <FormControl fullWidth className="margin-top-none">
                <FormHelperText>Your message</FormHelperText>
                <TextField
                  fullWidth
                  multiline
                  placeholder={enterpriseRequestPlaceholder}
                  value={enterpriseMessage}
                  onChange={e => setEnterpriseMessage(e.target.value)}
                />
              </FormControl>
              <Button
                sx={{ textTransform: 'none' }}
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
          {selectedPlan.id !== 'enterprise' && (
            <SubscriptionSummary plan={selectedPlan} addons={selectedAddons} deviceLimit={limit} title="Your subscription:" isNew={isNew} />
          )}
        </div>
      </div>
    </div>
  );
};
