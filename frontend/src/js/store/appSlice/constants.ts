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
import { BarChart as BarChartIcon, PieChartOutline as PieChartIcon } from '@mui/icons-material';

import FlagCN from '../../../assets/img/flag-cn.svg';
import FlagEU from '../../../assets/img/flag-eu.svg';
import FlagUS from '../../../assets/img/flag-us.svg';

const startingDeviceCount = {
  os: 'for first 50 devices',
  professional: 'for first 250 devices'
};

export const apiRoot = '/api/management';
export const apiUrl = {
  v1: `${apiRoot}/v1`,
  v2: `${apiRoot}/v2`
};

export const headerNames = {
  link: 'link',
  location: 'location',
  total: 'x-total-count'
};

export const chartTypes = {
  bar: { key: 'bar', Icon: BarChartIcon },
  pie: { key: 'pie', Icon: PieChartIcon }
};
export const emptyChartSelection = { software: '', group: '', chartType: chartTypes.bar.key, attribute: 'artifact_name' };
export const defaultReportType = 'distribution';
export const defaultReports = [{ ...emptyChartSelection, group: null, attribute: 'artifact_name', type: defaultReportType }];

export const BEGINNING_OF_TIME = '2016-01-01T00:00:00.000Z';

export const locations = {
  cn: { key: 'cn', title: 'China', location: 'cn.hosted.mender.io', icon: FlagCN },
  eu: { key: 'eu', title: 'EU', location: 'eu.hosted.mender.io', icon: FlagEU },
  us: { key: 'us', title: 'US', location: 'hosted.mender.io', icon: FlagUS }
};
export type AvailablePlans = 'os' | 'professional' | 'enterprise';

export type Plan = {
  description: string;
  deviceCount: string;
  features: string[];
  id: AvailablePlans;
  minimalDeviceCount: number;
  name: string;
  offer?: boolean;
  offerprice?: string;
  price: string;
  price2?: string;
};

export type AvailableAddon = 'configure' | 'troubleshoot' | 'monitor';

export const PLANS: { [key in AvailablePlans]: Plan } = {
  os: {
    id: 'os',
    name: 'Basic',
    minimalDeviceCount: 50,
    offer: true,
    price: '$34 / month',
    deviceCount: startingDeviceCount.os,
    description:
      'The core features of Mender. To continue using Enterprise Trial features—like Delta updates, scheduled deployments, phased rollouts, device filtering, dynamic groups, RBAC, audit logs, and more—please upgrade to a higher plan.',
    features: ['Access to core features of Mender', 'Basic support']
  },
  professional: {
    id: 'professional',
    name: 'Professional',
    minimalDeviceCount: 250,
    offer: true,
    price: '$291 / month',
    deviceCount: startingDeviceCount.professional,
    description: 'Everything in Basic, plus enhanced update management and automation features.',
    features: ['Advanced OTA features', 'Higher priority support']
  },
  enterprise: {
    id: 'enterprise',
    name: 'Enterprise',
    minimalDeviceCount: 1000,
    price: 'Custom pricing',
    deviceCount: 'Unlimited devices',
    description:
      'Every advanced feature of Mender, tailored for complex and large-scale deployments. Not available as a monthly subscription — ask us for a quote.',
    features: ['All Mender features', 'Advanced security features', 'SLA-backed support']
  }
} as const;
export type Addon = {
  description: string;
  eligible: AvailablePlans[];
  id: AvailableAddon;
  needs: string[];
  title: string;
} & {
  [key in Exclude<AvailablePlans, 'enterprise'>]: {
    deviceCount: string;
    price: string;
  };
};

export type AddonId = keyof typeof ADDONS;

// the needs names need to be aligned with the name of the features in the appReducer, as they will be checked in the addonselection
export const ADDONS: { [key in AvailableAddon]: Addon } = {
  configure: {
    id: 'configure',
    title: 'Configure',
    description: 'Seamlessly configure applications and devices remotely – configure each device to its environment.',
    needs: ['hasDeviceConfig'],
    os: {
      price: '$11/month',
      deviceCount: startingDeviceCount.os
    },
    professional: {
      price: '$70/month',
      deviceCount: startingDeviceCount.professional
    },
    eligible: ['os', 'professional', 'enterprise']
  },
  monitor: {
    id: 'monitor',
    title: 'Monitor',
    description: 'Secure, remote access to your devices – quickly diagnose and fix issues in real time.',
    needs: ['hasMonitor'],
    os: {
      price: '-',
      deviceCount: '-'
    },
    professional: {
      price: '$93/month',
      deviceCount: startingDeviceCount.professional
    },
    eligible: ['professional', 'enterprise']
  },
  troubleshoot: {
    id: 'troubleshoot',
    title: 'Troubleshoot',
    description: 'Detect and analyze health issues of devices, services and applications. Set up alerts so you can act quickly.',
    needs: ['hasDeviceConnect'],
    os: {
      price: '$28/month',
      deviceCount: startingDeviceCount.os
    },
    professional: {
      price: '$79/month',
      deviceCount: startingDeviceCount.professional
    },
    eligible: ['os', 'professional', 'enterprise']
  }
};

export const BENEFITS = {
  auditlog: { id: 'auditlog', benefit: 'trace change across your devices and access troubleshooting session replay', requiredPlan: PLANS.professional.id },
  dashboard: { id: 'dashboard', benefit: 'actionable insights into the devices you are updating with Mender', requiredPlan: PLANS.enterprise.id },
  deltaGeneration: {
    id: 'deltaGeneration',
    benefit: 'automatic delta artifacts generation to minimize data transfer and improve the update delivery',
    requiredPlan: PLANS.enterprise.id
  },
  deviceConfiguration: { id: 'deviceConfiguration', benefit: 'device configuration features', requiredAddon: ADDONS.configure.id },
  deviceMonitor: { id: 'deviceMonitor', benefit: 'device monitoring features', requiredAddon: ADDONS.monitor.id, requiredPlan: PLANS.professional.id },
  deviceTroubleshoot: { id: 'deviceTroubleshoot', benefit: 'device troubleshooting features', requiredAddon: ADDONS.troubleshoot.id },
  dynamicGroups: { id: 'dynamicGroups', benefit: 'create dynamic groups to ease device management', requiredPlan: PLANS.enterprise.id },
  fullFiltering: { id: 'fullFiltering', benefit: 'filtering by multiple attributes to improve the device overview', requiredPlan: PLANS.professional.id },
  gateway: { id: 'gateway', benefit: 'see devices connected to your gateway device for easy access', requiredPlans: PLANS.professional.id },
  pausedDeployments: {
    id: 'pausedDeployments',
    benefit: 'granular control about update rollout to allow synchronization across your fleet',
    requiredPlan: PLANS.enterprise.id
  },
  phasedDeployments: { id: 'phasedDeployments', benefit: 'choose to roll out deployments in multiple phases', requiredPlan: PLANS.enterprise.id },
  rbac: { id: 'rbac', benefit: 'granular role based access control', requiredPlan: PLANS.enterprise.id },
  retryDeployments: { id: 'retryDeployments', benefit: 'optional retries for failed rollout attempts', requiredPlan: PLANS.professional.id },
  scheduledDeployments: {
    id: 'scheduledDeployments',
    benefit: 'scheduled deployments to steer the distribution of your updates.',
    requiredPlan: PLANS.professional.id
  },
  webhookEvents: {
    id: 'webhookEvents',
    benefit: 'receive inventory events and select which type(s) of events the webhook will receive',
    requiredPlan: PLANS.professional.id
  },

  default: { id: 'default', benefit: 'gain access to this feature', requiredPlan: PLANS.enterprise.id }
};

export const yes = () => true;
export const canAccess = yes;

export const DARK_MODE = 'dark';
export const LIGHT_MODE = 'light';

export const APPLICATION_JSON_CONTENT_TYPE = 'application/json';
export const APPLICATION_JWT_CONTENT_TYPE = 'application/jwt';
