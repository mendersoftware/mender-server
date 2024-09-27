// Copyright 2023 Northern.tech AS
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
// @ts-nocheck
import { createSelector } from '@reduxjs/toolkit';

import { DEPLOYMENT_STATES } from './constants';

export const getDeploymentsById = state => state.deployments.byId;
export const getDeploymentsByStatus = state => state.deployments.byStatus;
export const getSelectedDeploymentDeviceIds = state => state.deployments.selectedDeviceIds;
export const getDeploymentsSelectionState = state => state.deployments.selectionState;

export const getMappedDeploymentSelection = createSelector(
  [getDeploymentsSelectionState, (_, deploymentsState) => deploymentsState, getDeploymentsById],
  (selectionState, deploymentsState, deploymentsById) => {
    const { selection = [] } = selectionState[deploymentsState] ?? {};
    return selection.reduce((accu, id) => {
      if (deploymentsById[id]) {
        accu.push(deploymentsById[id]);
      }
      return accu;
    }, []);
  }
);

const relevantDeploymentStates = [DEPLOYMENT_STATES.pending, DEPLOYMENT_STATES.inprogress, DEPLOYMENT_STATES.finished];
export const DEPLOYMENT_CUTOFF = 3;
export const getRecentDeployments = createSelector([getDeploymentsById, getDeploymentsByStatus], (deploymentsById, deploymentsByStatus) =>
  Object.entries(deploymentsByStatus).reduce(
    (accu, [state, byStatus]) => {
      if (!relevantDeploymentStates.includes(state) || !byStatus.deploymentIds.length) {
        return accu;
      }
      accu[state] = byStatus.deploymentIds
        .reduce((accu, id) => {
          if (deploymentsById[id]) {
            accu.push(deploymentsById[id]);
          }
          return accu;
        }, [])
        .slice(0, DEPLOYMENT_CUTOFF);
      accu.total += byStatus.total;
      return accu;
    },
    { total: 0 }
  )
);
