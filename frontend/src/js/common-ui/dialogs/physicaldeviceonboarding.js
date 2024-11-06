// Copyright 2019 Northern.tech AS
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
import { Link } from 'react-router-dom';

import { InfoOutlined as InfoIcon } from '@mui/icons-material';
import { Autocomplete, TextField } from '@mui/material';
import { createFilterOptions } from '@mui/material/useAutocomplete';

import { HELPTOOLTIPS, MenderHelpTooltip } from '@northern.tech/helptips/helptooltips';
import { EXTERNAL_PROVIDER, onboardingSteps } from '@northern.tech/store/constants';
import {
  getCurrentSession,
  getFeatures,
  getFullVersionInformation,
  getIsEnterprise,
  getIsPreview,
  getOnboardingState,
  getOrganization,
  getTenantCapabilities
} from '@northern.tech/store/selectors';
import { advanceOnboarding, setOnboardingApproach, setOnboardingDeviceType } from '@northern.tech/store/thunks';
import { getDebConfigurationCode, versionCompare } from '@northern.tech/utils/helpers';

import CopyCode from '../copy-code';
import DocsLink from '../docslink';
import { MenderTooltipClickable } from '../mendertooltip';

const filter = createFilterOptions();

const types = [
  { title: 'Raspberry Pi 3', value: 'raspberrypi3' },
  { title: 'Raspberry Pi 4', value: 'raspberrypi4' }
];

export const ConvertedImageNote = () => (
  <p>
    We prepared an image, ready for Mender, for you to start with. You can find it in the{' '}
    <DocsLink path="get-started/preparation/prepare-a-raspberry-pi-device" title="Prepare a Raspberry Pi device" /> documentation, which also contains
    instructions for initial device setup. Once you&apos;re done flashing you can go ahead and proceed to the next step.
  </p>
);

const IntegrationsLink = () => (
  <Link to="/settings/integrations" target="_blank">
    Integration settings
  </Link>
);

export const ExternalProviderTip = ({ hasExternalIntegration, integrationProvider }) => (
  <MenderTooltipClickable
    className="clickable flexbox muted"
    placement="bottom"
    style={{ alignItems: 'end', marginBottom: 3 }}
    title={
      <div style={{ maxWidth: 350 }}>
        {hasExternalIntegration ? (
          <p>
            Devices added here will be automatically integrated with the <i>{EXTERNAL_PROVIDER[integrationProvider].title}</i> you set in the{' '}
            <IntegrationsLink />.
          </p>
        ) : (
          <p>
            To connect your devices with <i>{EXTERNAL_PROVIDER[integrationProvider].title}</i>, go to <IntegrationsLink /> and set up the integration.
          </p>
        )}
      </div>
    }
  >
    <InfoIcon />
  </MenderTooltipClickable>
);

export const DeviceTypeSelectionStep = ({
  hasConvertedImage,
  hasExternalIntegration,
  integrationProvider,
  onboardingState,
  onSelect,
  selection = '',
  version
}) => {
  const shouldShowOnboardingTip = !onboardingState.complete && onboardingState.showTips;
  const hasExternalIntegrationSupport = versionCompare(version, '3.2') > -1;
  return (
    <>
      <h4>Enter your device type</h4>
      <p>Setting this attribute on the device ensures that the device will only receive updates for compatible software releases.</p>

      <div style={{ display: 'grid', gridTemplateColumns: 'max-content 50px 150px', placeItems: 'end center' }}>
        <Autocomplete
          id="device-type-selection"
          autoSelect
          autoHighlight
          filterSelectedOptions
          freeSolo
          getOptionLabel={option => {
            // Value selected with enter, right from the input
            if (typeof option === 'string') {
              return option;
            }
            if (option.key === 'custom' && option.value === selection) {
              return option.value;
            }
            return option.title;
          }}
          handleHomeEndKeys
          includeInputInList
          filterOptions={(options, params) => {
            const filtered = filter(options, params);
            if (filtered.length !== 1 && params.inputValue !== '') {
              filtered.push({
                value: params.inputValue,
                key: 'custom',
                title: `Use "${params.inputValue}"`
              });
            }
            return filtered;
          }}
          options={types}
          onChange={onSelect}
          renderInput={params => (
            <TextField {...params} label="Device type" placeholder="Choose a device type" InputProps={{ ...params.InputProps }} style={{ marginTop: 0 }} />
          )}
          style={{ maxWidth: 300 }}
          value={selection}
        />
        {hasExternalIntegrationSupport && <ExternalProviderTip hasExternalIntegration={hasExternalIntegration} integrationProvider={integrationProvider} />}
        {shouldShowOnboardingTip ? <MenderHelpTooltip id={HELPTOOLTIPS.deviceTypeTip.id} placement="bottom" /> : <div />}
      </div>
      {hasConvertedImage && <ConvertedImageNote />}
    </>
  );
};

export const InstallationStep = ({ advanceOnboarding, selection, onboardingState, ...remainingProps }) => {
  const codeToCopy = getDebConfigurationCode({ ...remainingProps, deviceType: selection, isOnboarding: !onboardingState.complete });
  return (
    <>
      <h4>Log into your device and install the Mender client</h4>
      <p>
        Copy & paste and run this command <b>on your device</b>:
      </p>
      <CopyCode code={codeToCopy} onCopy={() => advanceOnboarding(onboardingSteps.DASHBOARD_ONBOARDING_START)} withDescription={true} />
      <p>This downloads the Mender client on the device, sets the configuration and starts the client.</p>
      <p>
        Once the client has started, your device will attempt to connect to the server. It will then appear in your Pending devices tab and you can continue.
      </p>
    </>
  );
};

const steps = {
  1: DeviceTypeSelectionStep,
  2: InstallationStep
};

const integrationProvider = EXTERNAL_PROVIDER['iot-hub'].provider;

export const PhysicalDeviceOnboarding = ({ progress }) => {
  const [selection, setSelection] = useState('');
  const hasExternalIntegration = useSelector(state => {
    const { credentials = {} } = state.organization.externalDeviceIntegrations.find(integration => integration.provider === integrationProvider) ?? {};
    const { [EXTERNAL_PROVIDER['iot-hub'].credentialsAttribute]: azureConnectionString = '' } = credentials;
    return !!azureConnectionString;
  });
  const ipAddress = useSelector(state => state.app.hostAddress);
  const isEnterprise = useSelector(getIsEnterprise);
  const { isDemoMode, isHosted } = useSelector(getFeatures);
  const isPreRelease = useSelector(getIsPreview);
  const onboardingState = useSelector(getOnboardingState);
  const { tenant_token: tenantToken } = useSelector(getOrganization);
  const { Integration: version } = useSelector(getFullVersionInformation);
  const { token } = useSelector(getCurrentSession);
  const { hasMonitor } = useSelector(getTenantCapabilities);
  const dispatch = useDispatch();

  useEffect(() => {
    dispatch(setOnboardingApproach('physical'));
  }, [dispatch]);

  const onSelect = (e, deviceType, reason) => {
    if (reason === 'selectOption') {
      dispatch(setOnboardingDeviceType(deviceType.value));
      setSelection(deviceType.value);
    } else if (reason === 'clear') {
      dispatch(setOnboardingDeviceType(''));
      setSelection('');
    }
  };

  const hasConvertedImage = !!selection && selection.length && (selection.startsWith('raspberrypi3') || selection.startsWith('raspberrypi4'));

  const ComponentToShow = steps[progress];
  return (
    <ComponentToShow
      advanceOnboarding={step => dispatch(advanceOnboarding(step))}
      hasConvertedImage={hasConvertedImage}
      hasExternalIntegration={hasExternalIntegration}
      hasMonitor={hasMonitor}
      integrationProvider={integrationProvider}
      ipAddress={ipAddress}
      isEnterprise={isEnterprise}
      isHosted={isHosted}
      isDemoMode={isDemoMode}
      isPreRelease={isPreRelease}
      onboardingState={onboardingState}
      onSelect={onSelect}
      selection={selection}
      tenantToken={tenantToken}
      token={token}
      version={version}
    />
  );
};

export default PhysicalDeviceOnboarding;
