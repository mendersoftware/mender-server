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
import { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Link } from 'react-router-dom';

import { InfoOutlined as InfoIcon } from '@mui/icons-material';
import { Autocomplete, TextField, Typography } from '@mui/material';
import { createFilterOptions } from '@mui/material/useAutocomplete';

import CopyCode from '@northern.tech/common-ui/CopyCode';
import DocsLink from '@northern.tech/common-ui/DocsLink';
import { MenderTooltipClickable } from '@northern.tech/common-ui/helptips/MenderTooltip';
import { EXTERNAL_PROVIDER, onboardingSteps } from '@northern.tech/store/constants';
import {
  getCurrentSession,
  getFeatures,
  getFullVersionInformation,
  getHostAddress,
  getIsEnterprise,
  getIsPreview,
  getOnboardingState,
  getOrganization,
  getTenantCapabilities
} from '@northern.tech/store/selectors';
import { advanceOnboarding, setOnboardingApproach, setOnboardingDeviceType } from '@northern.tech/store/thunks';

import { getDebConfigurationCode } from '../../../utils/helpers';
import { HELPTOOLTIPS } from '../../helptips/HelpTooltips';
import { MenderHelpTooltip } from '../../helptips/MenderTooltip';

const filter = createFilterOptions();

const types = [
  { title: 'raspberrypi4', value: 'raspberrypi4' },
  { title: 'raspberrypi5', value: 'raspberrypi5' }
];

export const ConvertedImageNote = () => (
  <Typography variant="body1">
    We prepared an image, ready for Mender, for you to start with. You can find it in the{' '}
    <DocsLink path="get-started/preparation/prepare-a-raspberry-pi-device" title="Prepare a Raspberry Pi device" /> documentation, which also contains
    instructions for initial device setup. Once you&apos;re done flashing you can go ahead and proceed to the next step.
  </Typography>
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
    style={{ alignItems: 'end' }}
    title={
      <div style={{ maxWidth: 350 }}>
        {hasExternalIntegration ? (
          <p>
            Devices added here will be automatically integrated with the {EXTERNAL_PROVIDER[integrationProvider].title} you set in the <IntegrationsLink />.
          </p>
        ) : (
          <p>
            To connect your devices with {EXTERNAL_PROVIDER[integrationProvider].title}, go to <IntegrationsLink /> and set up the integration.
          </p>
        )}
      </div>
    }
  >
    <InfoIcon />
  </MenderTooltipClickable>
);

export const DeviceTypeSelectionStep = ({ hasConvertedImage, hasExternalIntegration, integrationProvider, onboardingState, onSelect, selection = '' }) => {
  const shouldShowOnboardingTip = !onboardingState.complete && onboardingState.showTips;
  return (
    <>
      <Typography variant="subtitle1" gutterBottom>
        Enter your device type
      </Typography>
      <Typography variant="body1" gutterBottom>
        Setting this attribute on the device ensures that the device will only receive updates for compatible software releases.
      </Typography>
      <div
        className="margin-top-small margin-bottom-small"
        style={{ display: 'grid', gridTemplateColumns: 'max-content max-content 150px', alignItems: 'center', gap: 16 }}
      >
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
        <ExternalProviderTip hasExternalIntegration={hasExternalIntegration} integrationProvider={integrationProvider} />
        {shouldShowOnboardingTip ? <MenderHelpTooltip id={HELPTOOLTIPS.deviceTypeTip.id} placement="bottom" /> : <div />}
      </div>
      {hasConvertedImage && <ConvertedImageNote />}
    </>
  );
};

export const InstallationStep = ({ advanceOnboarding, selection, ...remainingProps }) => {
  const codeToCopy = getDebConfigurationCode({ ...remainingProps, deviceType: selection });
  return (
    <>
      <Typography variant="subtitle1" gutterBottom>
        Log into your device and install the Mender client
      </Typography>
      <Typography className="margin-bottom-small" variant="body1">
        Copy & paste and run this command <b>on your device</b>:
      </Typography>
      <CopyCode code={codeToCopy} onCopy={() => advanceOnboarding(onboardingSteps.DASHBOARD_ONBOARDING_START)} withDescription={true} />
      <Typography variant="body1">
        This downloads the Mender client on the device, sets the configuration and starts the client. Once the client has started, your device will attempt to
        connect to the server. It will then appear in your Pending devices tab and you can continue.
      </Typography>
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
  const ipAddress = useSelector(getHostAddress);
  const isEnterprise = useSelector(getIsEnterprise);
  const { isHosted } = useSelector(getFeatures);
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
