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
import { DEVICE_LIST_DEFAULTS, SORTING_OPTIONS, apiUrl } from '@northern.tech/store/constants';

const alreadyInstalled = 'already-installed';

export const deploymentsApiUrl = `${apiUrl.v1}/deployments`;
export const deploymentsApiUrlV2 = `${apiUrl.v2}/deployments`;

export const deploymentSubstates = {
  aborted: 'aborted',
  alreadyInstalled,
  decommissioned: 'decommissioned',
  downloading: 'downloading',
  failure: 'failure',
  installing: 'installing',
  noartifact: 'noartifact',
  pause_before_committing: 'pause_before_committing',
  pause_before_installing: 'pause_before_installing',
  pause_before_rebooting: 'pause_before_rebooting',
  pending: 'pending',
  rebooting: 'rebooting',
  success: 'success'
};

export const deploymentStatesToSubstates = {
  failures: [deploymentSubstates.failure, deploymentSubstates.aborted, deploymentSubstates.decommissioned],
  inprogress: [deploymentSubstates.downloading, deploymentSubstates.installing, deploymentSubstates.rebooting],
  paused: [deploymentSubstates.pause_before_installing, deploymentSubstates.pause_before_rebooting, deploymentSubstates.pause_before_committing],
  pending: [deploymentSubstates.pending],
  successes: [deploymentSubstates.success, deploymentSubstates.alreadyInstalled, deploymentSubstates.noartifact]
};

export const deploymentStatesToSubstatesWithSkipped = {
  ...deploymentStatesToSubstates,
  failures: [deploymentSubstates.failure],
  skipped: [deploymentSubstates.aborted, deploymentSubstates.noartifact, deploymentSubstates.alreadyInstalled, deploymentSubstates.decommissioned],
  successes: [deploymentSubstates.success]
};
export const installationSubstatesMap = {
  download: {
    title: 'download',
    done: 'downloaded',
    successIndicators: [deploymentSubstates.installing, deploymentSubstates.rebooting, ...deploymentStatesToSubstates.paused, deploymentSubstates.success],
    failureIndicators: deploymentStatesToSubstates.failures,
    pauseConfigurationIndicator: 'ArtifactInstall_Enter'
  },
  install: {
    title: 'install',
    done: 'installed',
    successIndicators: [
      deploymentSubstates.rebooting,
      deploymentSubstates.pause_before_rebooting,
      deploymentSubstates.pause_before_committing,
      deploymentSubstates.success
    ],
    failureIndicators: deploymentStatesToSubstates.failures,
    pauseConfigurationIndicator: 'ArtifactReboot_Enter'
  },
  reboot: {
    title: 'reboot',
    done: 'rebooted',
    successIndicators: [deploymentSubstates.pause_before_committing, deploymentSubstates.success],
    failureIndicators: deploymentStatesToSubstates.failures,
    pauseConfigurationIndicator: 'ArtifactCommit_Enter'
  },
  commit: {
    title: 'commit',
    done: 'committed',
    successIndicators: deploymentStatesToSubstates.successes,
    failureIndicators: deploymentStatesToSubstates.failures,
    pauseConfigurationIndicator: undefined
  }
};

export const DEPLOYMENT_STATES = {
  finished: 'finished',
  inprogress: 'inprogress',
  pending: 'pending',
  scheduled: 'scheduled'
};

export const listDefaultsByState = {
  [DEPLOYMENT_STATES.inprogress]: { page: 1, perPage: 10 },
  [DEPLOYMENT_STATES.pending]: { page: 1, perPage: 10 },
  [DEPLOYMENT_STATES.scheduled]: { ...DEVICE_LIST_DEFAULTS },
  [DEPLOYMENT_STATES.finished]: { ...DEVICE_LIST_DEFAULTS },
  sort: { direction: SORTING_OPTIONS.desc }
};

export const DEFAULT_PENDING_INPROGRESS_COUNT = 10;
export const DEPLOYMENT_ROUTES = {
  active: {
    key: 'active',
    route: '/deployments/active',
    states: [DEPLOYMENT_STATES.pending, DEPLOYMENT_STATES.inprogress],
    title: 'Active'
  },
  finished: {
    key: 'finished',
    route: '/deployments/finished',
    states: [DEPLOYMENT_STATES.finished],
    title: 'Finished'
  },
  scheduled: {
    key: 'scheduled',
    route: '/deployments/scheduled',
    states: [DEPLOYMENT_STATES.scheduled],
    title: 'Scheduled'
  }
};
export const DEPLOYMENT_TYPES = {
  software: 'software',
  configuration: 'configuration'
};
export const defaultStats = {
  [deploymentSubstates.aborted]: 0,
  [deploymentSubstates.alreadyInstalled]: 0,
  [deploymentSubstates.decommissioned]: 0,
  [deploymentSubstates.downloading]: 0,
  [deploymentSubstates.failure]: 0,
  [deploymentSubstates.installing]: 0,
  [deploymentSubstates.noartifact]: 0,
  [deploymentSubstates.pause_before_committing]: 0,
  [deploymentSubstates.pause_before_installing]: 0,
  [deploymentSubstates.pause_before_rebooting]: 0,
  [deploymentSubstates.pending]: 0,
  [deploymentSubstates.rebooting]: 0,
  [deploymentSubstates.success]: 0
};
export const deploymentDisplayStates = {
  finished: 'Finished',
  scheduled: 'Scheduled',
  skipped: 'Skipped',
  paused: 'Paused',
  pending: 'Pending',
  inprogress: 'In Progress',
  success: 'Success',
  successes: 'Success',
  failure: 'Fail',
  failures: 'Fail'
};

export const deploymentPrototype = {
  devices: {},
  name: undefined,
  statistics: { status: {} }
};

export const pauseMap = {
  [deploymentSubstates.pause_before_installing]: {
    title: installationSubstatesMap.download.done,
    followUp: installationSubstatesMap.download.pauseConfigurationIndicator
  },
  [deploymentSubstates.pause_before_rebooting]: {
    title: installationSubstatesMap.install.done,
    followUp: installationSubstatesMap.install.pauseConfigurationIndicator
  },
  [deploymentSubstates.pause_before_committing]: {
    title: installationSubstatesMap.reboot.done,
    followUp: installationSubstatesMap.reboot.pauseConfigurationIndicator
  }
};
export const limitDefault = { min: 5, max: 100, default: 10 };
