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
import * as DeploymentConstants from '../constants/deploymentConstants';
import * as DeviceConstants from '../constants/deviceConstants';

export const initialState = {
  byId: {
    // [id]: { statistics: { status: {}, total_size }, devices: { [deploymentId]: { id, log } }, totalDeviceCount }
  },
  byStatus: {
    finished: { deploymentIds: [], total: 0 },
    inprogress: { deploymentIds: [], total: 0 },
    pending: { deploymentIds: [], total: 0 },
    scheduled: { deploymentIds: [], total: 0 }
  },
  config: {
    binaryDelta: {
      timeout: -1,
      duplicatesWindow: -1,
      compressionLevel: -1,
      disableChecksum: false,
      disableDecompression: false,
      inputWindow: -1,
      instructionBuffer: -1,
      sourceWindow: -1
    },
    binaryDeltaLimits: {
      timeout: { ...DeploymentConstants.limitDefault, default: 60, max: 3600, min: 60 },
      sourceWindow: DeploymentConstants.limitDefault,
      inputWindow: DeploymentConstants.limitDefault,
      duplicatesWindow: DeploymentConstants.limitDefault,
      instructionBuffer: DeploymentConstants.limitDefault
    }
  },
  deploymentDeviceLimit: 5000,
  selectedDeviceIds: [],
  selectionState: {
    finished: { ...DeviceConstants.DEVICE_LIST_DEFAULTS, endDate: undefined, search: '', selection: [], startDate: undefined, total: 0, type: '' },
    inprogress: { ...DeviceConstants.DEVICE_LIST_DEFAULTS, perPage: DeploymentConstants.DEFAULT_PENDING_INPROGRESS_COUNT, selection: [] },
    pending: { ...DeviceConstants.DEVICE_LIST_DEFAULTS, perPage: DeploymentConstants.DEFAULT_PENDING_INPROGRESS_COUNT, selection: [] },
    scheduled: { ...DeviceConstants.DEVICE_LIST_DEFAULTS, selection: [] },
    general: {
      state: DeploymentConstants.DEPLOYMENT_ROUTES.active.key,
      showCreationDialog: false,
      showReportDialog: false,
      reportType: null // DeploymentConstants.DEPLOYMENT_TYPES.configuration|DeploymentConstants.DEPLOYMENT_TYPES.software
    },
    selectedId: undefined
  }
};

const deploymentReducer = (state = initialState, action) => {
  switch (action.type) {
    case DeploymentConstants.CREATE_DEPLOYMENT:
      return {
        ...state,
        byId: {
          ...state.byId,
          [action.deploymentId]: {
            ...DeploymentConstants.deploymentPrototype,
            ...action.deployment
          }
        },
        byStatus: {
          ...state.byStatus,
          pending: {
            ...state.byStatus.pending,
            deploymentIds: [...state.byStatus.pending.deploymentIds, action.deploymentId],
            total: state.byStatus.pending.total + 1
          }
        },
        selectionState: {
          ...state.selectionState,
          pending: {
            ...state.selectionState.pending,
            selection: [action.deploymentId, ...state.selectionState.pending.selection],
            total: state.selectionState.pending.total + 1
          }
        }
      };
    case DeploymentConstants.REMOVE_DEPLOYMENT: {
      // eslint-disable-next-line no-unused-vars
      const { [action.deploymentId]: removedDeployment, ...remainder } = state.byId;
      return {
        ...state,
        byId: remainder
      };
    }
    case DeploymentConstants.RECEIVE_DEPLOYMENT:
    case DeploymentConstants.RECEIVE_DEPLOYMENT_DEVICE_LOG:
      return {
        ...state,
        byId: {
          ...state.byId,
          [action.deployment.id]: action.deployment
        }
      };
    case DeploymentConstants.RECEIVE_DEPLOYMENT_DEVICES:
      return {
        ...state,
        byId: {
          ...state.byId,
          [action.deploymentId]: {
            ...state.byId[action.deploymentId],
            devices: action.devices,
            totalDeviceCount: action.totalDeviceCount
          }
        },
        selectedDeviceIds: action.selectedDeviceIds
      };
    case DeploymentConstants.RECEIVE_DEPLOYMENTS:
      return {
        ...state,
        byId: {
          ...state.byId,
          ...action.deployments
        }
      };
    case DeploymentConstants.RECEIVE_INPROGRESS_DEPLOYMENTS:
    case DeploymentConstants.RECEIVE_PENDING_DEPLOYMENTS:
    case DeploymentConstants.RECEIVE_SCHEDULED_DEPLOYMENTS:
    case DeploymentConstants.RECEIVE_FINISHED_DEPLOYMENTS:
      return {
        ...state,
        byStatus: {
          ...state.byStatus,
          [action.status]: {
            ...state.byStatus[action.status],
            deploymentIds: action.deploymentIds,
            total: action.total
          }
        }
      };
    case DeploymentConstants.SELECT_INPROGRESS_DEPLOYMENTS:
    case DeploymentConstants.SELECT_PENDING_DEPLOYMENTS:
    case DeploymentConstants.SELECT_SCHEDULED_DEPLOYMENTS:
    case DeploymentConstants.SELECT_FINISHED_DEPLOYMENTS:
      return {
        ...state,
        selectionState: {
          ...state.selectionState,
          [action.status]: {
            ...state.selectionState[action.status],
            selection: action.deploymentIds,
            total: action.total
          }
        }
      };
    case DeploymentConstants.SET_DEPLOYMENTS_STATE:
      return {
        ...state,
        selectionState: action.state
      };
    case DeploymentConstants.SET_DEPLOYMENTS_CONFIG:
      return {
        ...state,
        config: action.config
      };
    default:
      return state;
  }
};

export default deploymentReducer;
