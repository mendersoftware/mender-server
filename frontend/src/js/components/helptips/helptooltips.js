// Copyright 2017 Northern.tech AS
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
import React, { useCallback } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import ConfigurationObject from '@northern.tech/common-ui/configurationobject';
import DocsLink from '@northern.tech/common-ui/docslink';
import { HelpTooltip } from '@northern.tech/common-ui/mendertooltip';
import storeActions from '@northern.tech/store/actions';
import { READ_STATES, yes } from '@northern.tech/store/constants';
import { getDeviceById, getFeatures, getTooltipsState } from '@northern.tech/store/selectors';
import { setAllTooltipsReadState, setTooltipReadState } from '@northern.tech/store/thunks';

const { setSnackbar } = storeActions;

const AuthExplainButton = () => (
  <>
    <h3>Device authorization status</h3>
    <p>
      Each device sends an authentication request containing its identity attributes and its current public key. You can accept, reject or dismiss these
      requests to determine the authorization status of the device.
    </p>
    <p>
      In cases such as key rotation, each device may have more than one identity/key combination listed. See the documentation for more on{' '}
      <DocsLink path="overview/device-authentication" title="Device authentication" />.
    </p>
  </>
);

const AuthButton = () => (
  <>
    <h3>Authorize devices</h3>
    <p>
      Expand this section to view the authentication options for this device. You can decide whether to accept it, reject it, or just dismiss this device for
      now.
    </p>
    <p>
      See the documentation for more on <DocsLink path="overview/device-authentication" title="Device authentication" />.
    </p>
  </>
);

const AddGroup = () => (
  <>
    <h3>Device groups</h3>
    <p>
      It is possible to create groups of devices. Once you have created a group and added one or more devices to it, you can deploy an update to that specific
      group only.
    </p>
  </>
);

const ExpandArtifact = () => (
  <>
    <h3>Device type compatibility</h3>
    <p>
      Mender Artifacts have <b>Device types compatible</b> as part of their metadata. All devices report which device type they are, as part of their inventory
      information. During a deployment, Mender makes sure that a device will only download and install an Artifact it is compatible with.
    </p>
    <p>You can click on each Artifact in the Release to expand the row and view more information about it.</p>
    <p>
      For more information on how to specify the device type compatibility and other artifact metadata,{' '}
      <DocsLink path="artifact-creation/create-an-artifact" title="see the documentation" />.
    </p>
  </>
);

const DeviceSupportTip = () => (
  <p>
    The steps in the guide should work on most operating systems in the Debian family (e.g. Debian, Ubuntu, Raspberry Pi OS) and devices based on ARMv6 or newer
    (e.g. Raspberry Pi 2/3/4, Beaglebone). Visit <DocsLink path="overview/device-support" title="our documentation" /> for more information about device
    support.
  </p>
);

const ConfigureTimezoneTip = () => (
  <>
    To see the effects of applying a configuration to your device you can set one of the below values to modify the timezone of your device. While all values
    from <i>timedatectl list-timezones</i> will work, to easily see the impact of the changed value you can use one of the following values:
    <ul>
      <li>Europe/Oslo</li>
      <li>America/Los_Angeles</li>
      <li>Asia/Tokyo</li>
    </ul>
    Once the configuration has been applied you can see the effect by opening the Remote Terminal to the device and executing the <i>date</i> command.
  </>
);

const ConfigureRaspberryLedTip = () => {
  const dispatch = useDispatch();
  return (
    <>
      To see the effects of applying a configuration to your device you can set one of the below values to modify the behaviour of your Raspberry Pi green
      status LED
      <ConfigurationObject
        className="margin-top-small margin-bottom-small"
        config={{
          mmc0: 'The default, which blinks the led on storage activity',
          on: 'Turn on the light permanently',
          off: 'Turn off the light permanently',
          heartbeat: 'Enable heartbeat blinking'
        }}
        compact
        setSnackbar={(...args) => dispatch(setSnackbar(...args))}
      />
      There are other possible values, but we won&apos;t advertise them here. See
      <a href="http://www.d3noob.org/2020/07/controlling-activity-led-on-raspberry-pi.html" target="_blank" rel="noopener noreferrer">
        this blog post
      </a>{' '}
      or{' '}
      <a href="https://www.raspberrypi.org/forums/viewtopic.php?t=273194#p1658930" target="_blank" rel="noopener noreferrer">
        in the Raspberry Pi forums
      </a>{' '}
      for more information.
    </>
  );
};

const ConfigureAddOnTip = () => (
  <p>
    Mender deploys the configuration attributes using the same mechanisms as software updates. The configuration is stored as a JSON file at
    <code>/var/lib/mender-configure/device-config.json</code> on the device and then all the scripts in{' '}
    <code>/usr/lib/mender-configure/apply-device-config.d/</code> are executed to apply the configuration attributes. To add a new configuration attribute, you
    simply need to input it in the UI and add a script to that directory that applies it accordingly.
  </p>
);

const NameTagTip = () => (
  <>
    The <i>Name</i> tag will be available as a device indentifier too.
  </>
);

const NameFilterTip = () => <>Filtering by name is limited to devices with a previously defined name.</>;

const DeviceTypeTip = () => (
  <>
    <p>
      If you don&apos;t see your exact device on the list, choose <i>Generic ARMv6 or newer</i> to continue the tutorial for now.
    </p>
    <p>
      (Note: if your device is <i>not</i> based on ARMv6 or newer, the tutorial won&apos;t work - instead, go back and use the virtual device)
    </p>
  </>
);

const TwoFactorNote = ({ className }) => (
  <div className={className}>
    Two Factor Authentication is enabled for your account. If you haven&apos;t set up a 3rd party authentication app with a verification code, please contact an
    administrator.
  </div>
);

const AuditlogExplanation = () => <>The audit log shows the history of changes made to your Devices, Artifacts, Deployments, and user management.</>;

const DashboardWidget = () => (
  <>Add dashboard widgets to visualize the software distribution or geographical location of all your devices, or a group of your devices.</>
);

const ScheduleDeployment = () => (
  <>
    This time is relative to the server only – each device&apos;s time zone will not be taken into account. Devices across different time zones will receive the
    update at the same time.
  </>
);

const GroupDeployment = () => (
  <>The deployment will skip any devices in the group that are already on the target Release version, or that have an incompatible device type.</>
);

const ForceDeployment = () => (
  <>
    <h3>Force update</h3>
    <p>This will make the Mender client install the update even if the selected release is already installed.</p>
  </>
);

const ArtifactUpload = () => <>Upload an Artifact to an existing or new Release</>;

const PhasedPausedDeployments = () => (
  <>
    This feature is not available on <b>phased deployments</b>. If you&apos;d like to set pause states between update steps, go back and adjust the rollout
    schedule to a <b>single phase</b>.
  </>
);

const ResetHistory = () => <>Greyed out items will not be considered during deployment roll out</>;

const MenderArtifactUpload = () => (
  <>
    If there is no Release matching this Artifact’s name, a new Release will be created for this Artifact.
    <br />
    <br />
    If there is already a Release matching this Artifact’s name, the Artifact will be grouped in that Release.
  </>
);

const SingleFileUpload = () => <>This will generate a single file application update Artifact, which requires some additional metadata to be entered.</>;

const GlobalSettings = () => <>These settings apply to all users, so changes made here may affect other users&apos; experience.</>;

const Webhooks = () => (
  <>Use webhooks to send data about device lifecycle events to third-party systems. Currently you can only have one integration set up at a time.</>
);

const WebhookEvents = () => (
  <>
    You can select which type(s) of events the webhook will receive. Device authentication includes when devices are provisioned, decommissioned, or
    authentication status changes.
  </>
);

const WebhookSecret = () => (
  <>
    The secret is used for signing the requests sent to your webhook, to verify their authenticity. It is optional, but highly recommended for security. The
    secret must be a hexidecimal string (including only characters from A-F and 0-9).
  </>
);

const TenantToken = () => (
  <>
    <h3>Organization token</h3>
    This token is unique for your organization and ensures that only devices that you own are able to connect to your account.
  </>
);

const SsoMetadata = () => <>Submit the metadata document from your Identity Provider</>;

const ReleaseName = () => (
  <>
    If a Release with this name already exists, this new Artifact may be grouped into a Release with other Artifacts of the same name - so long as they are
    compatible with different device types
  </>
);

const TenantAdmin = () => (
  <>
    Set the user who will have the admin role when this tenant is created. This can be a user who already has a Mender account, or a brand new user. The tenant
    admin will be able to change which user(s) have this role once they begin using the tenant.
  </>
);

const SubTenantDeviceLimit = () => (
  <>
    Set the maximum number of accepted devices this tenant can have connected to the server at any time. You can adjust this later. Each tenants’ amount of
    accepted devices will count towards your total device limit.
  </>
);

const SubTenantDeltaArtifactGeneration = () => <>TBD: is this feature still alive?</>;
const SubTenantSSO = () => (
  <>
    The created tenant will inherit the same Single Sign-On configuration as this Service Provider tenant. The created tenant’s admin user will not be able to
    change the SSO settings later.
  </>
);

const AttributeLimit = () => {
  const { isHosted } = useSelector(getFeatures);
  return isHosted ? (
    <>
      Expand to see the list of attributes currently in use. Please{' '}
      <a href="mailto:support@mender.io" target="_blank" rel="noopener noreferrer">
        contact our team
      </a>{' '}
      if your use case requires a different set of attributes.
    </>
  ) : (
    <>Expand to see the list of attributes currently in use.</>
  );
};
const TenantInitialAdmin = () => {
  return (
    <>
      The user that was assigned as admin for this tenant when it was created. There is a chance the admin user will have changed or this user no longer exists
      in this tenant.
    </>
  );
};

export const HELPTOOLTIPS = {
  addGroup: { id: 'addGroup', Component: AddGroup },
  artifactUpload: { id: 'artifactUpload', Component: ArtifactUpload },
  attributeLimit: { id: 'attributeLimit', Component: AttributeLimit },
  auditlogExplanation: { id: 'auditlogExplanation', Component: AuditlogExplanation },
  authButton: { id: 'authButton', Component: AuthButton },
  authExplainButton: { id: 'authExplainButton', Component: AuthExplainButton },
  configureAddOnTip: { id: 'configureAddOnTip', Component: ConfigureAddOnTip },
  configureRaspberryLedTip: {
    id: 'configureRaspberryLedTip',
    Component: ConfigureRaspberryLedTip,
    isRelevant: ({ device = {} }) => {
      const { attributes = {} } = device;
      return ['raspberry', 'rpi'].some(type => attributes.device_type?.some(deviceType => deviceType.startsWith(type)));
    }
  },
  configureTimezoneTip: {
    id: 'configureTimezoneTip',
    Component: ConfigureTimezoneTip,
    isRelevant: ({ device = {} }) => {
      const { attributes = {} } = device;
      return ['raspberry', 'rpi', 'qemux86-64'].some(type => attributes.device_type?.some(deviceType => deviceType.startsWith(type)));
    }
  },
  dashboardWidget: { id: 'dashboardWidget', Component: DashboardWidget },
  deviceSupportTip: { id: 'deviceSupportTip', Component: DeviceSupportTip },
  deviceTypeTip: { id: 'deviceTypeTip', Component: DeviceTypeTip },
  expandArtifact: { id: 'expandArtifact', Component: ExpandArtifact },
  forceDeployment: { id: 'forceDeployment', Component: ForceDeployment },
  globalSettings: { id: 'globalSettings', Component: GlobalSettings },
  groupDeployment: { id: 'groupDeployment', Component: GroupDeployment },
  menderArtifactUpload: { id: 'menderArtifactUpload', Component: MenderArtifactUpload },
  nameFilterTip: { id: 'nameFilterTip', Component: NameFilterTip },
  nameTagTip: { id: 'nameTagTip', Component: NameTagTip },
  phasedPausedDeployments: { id: 'phasedPausedDeployments', Component: PhasedPausedDeployments },
  releaseName: { id: 'releaseName', Component: ReleaseName },
  resetHistory: { id: 'resetHistory', Component: ResetHistory },
  ssoMetadata: { id: 'ssoMetadata', Component: SsoMetadata },
  scheduleDeployment: { id: 'scheduleDeployment', Component: ScheduleDeployment },
  singleFileUpload: { id: 'singleFileUpload', Component: SingleFileUpload },
  subTenantDeltaArtifactGeneration: { id: 'subTenantDeltaArtifactGeneration', Component: SubTenantDeltaArtifactGeneration },
  subTenantDeviceLimit: { id: 'subTenantDeviceLimit', Component: SubTenantDeviceLimit },
  subTenantSSO: { id: 'subTenantSSO', Component: SubTenantSSO },
  tenantAdmin: { id: 'tenantAdmin', Component: TenantAdmin },
  tenantInitialAdmin: { id: 'tenantInitialAdmin', Component: TenantInitialAdmin },
  tenantToken: { id: 'tenantToken', Component: TenantToken },
  twoFactorNote: { id: 'twoFactorNote', SpecialComponent: TwoFactorNote },
  webhookEvents: { id: 'webhookEvents', Component: WebhookEvents },
  webhooks: { id: 'webhooks', Component: Webhooks },
  webhookSecret: { id: 'webhookSecret', Component: WebhookSecret }
};

export const MenderHelpTooltip = props => {
  const { id, contentProps = {} } = props;
  const tooltipsById = useSelector(getTooltipsState);
  const dispatch = useDispatch();
  const device = useSelector(state => getDeviceById(state, contentProps.deviceId));
  const { readState = READ_STATES.unread } = tooltipsById[id] || {};
  const { Component, SpecialComponent, isRelevant = yes } = HELPTOOLTIPS[id];

  const onSetTooltipReadState = useCallback((...args) => dispatch(setTooltipReadState(...args)), [dispatch]);
  const onSetAllTooltipsReadState = state => dispatch(setAllTooltipsReadState(state));

  return (
    <HelpTooltip
      setAllTooltipsReadState={onSetAllTooltipsReadState}
      setTooltipReadState={onSetTooltipReadState}
      device={device}
      tooltip={{ Component, SpecialComponent, isRelevant, readState }}
      {...props}
    />
  );
};
