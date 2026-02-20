// Copyright 2015 Northern.tech AS
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
import { useCallback, useEffect, useRef, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { Paper, Typography } from '@mui/material';

import Loader from '@northern.tech/common-ui/Loader';
import storeActions from '@northern.tech/store/actions';
import { DEPLOYMENT_STATES, onboardingSteps } from '@northern.tech/store/constants';
import {
  getDeploymentsByStatus as getDeploymentsByStatusSelector,
  getDeploymentsSelectionState,
  getDevicesById,
  getIdAttribute,
  getIsEnterprise,
  getMappedDeploymentSelection,
  getOnboardingState,
  getUserCapabilities
} from '@northern.tech/store/selectors';
import { getDeploymentsByStatus, setDeploymentsState } from '@northern.tech/store/thunks';
import { useWindowSize } from '@northern.tech/utils/resizehook';
import { clearAllRetryTimers, clearRetryTimer, setRetryTimer } from '@northern.tech/utils/retrytimer';

import { getOnboardingComponentFor } from '../../utils/onboardingManager';
import DeploymentsList from './DeploymentsList';
import { defaultRefreshDeploymentsLength as refreshDeploymentsLength } from './constants';

const { setSnackbar } = storeActions;

export const minimalRefreshDeploymentsLength = 2000;

export const Progress = ({ abort, createClick, ...remainder }) => {
  const { canConfigure, canDeploy } = useSelector(getUserCapabilities);
  const idAttribute = useSelector(getIdAttribute);
  const onboardingState = useSelector(getOnboardingState);
  const isEnterprise = useSelector(getIsEnterprise);
  const {
    finished: { total: pastDeploymentsCount },
    pending: { total: pendingCount },
    inprogress: { total: progressCount }
  } = useSelector(getDeploymentsByStatusSelector);
  const progress = useSelector(state => getMappedDeploymentSelection(state, DEPLOYMENT_STATES.inprogress));
  const pending = useSelector(state => getMappedDeploymentSelection(state, DEPLOYMENT_STATES.pending));
  const selectionState = useSelector(getDeploymentsSelectionState);
  const devices = useSelector(getDevicesById);
  const dispatch = useDispatch();
  const dispatchedSetSnackbar = useCallback((...args) => dispatch(setSnackbar(...args)), [dispatch]);

  const { page: progressPage, perPage: progressPerPage } = selectionState.inprogress;
  const { page: pendingPage, perPage: pendingPerPage } = selectionState.pending;

  const [doneLoading, setDoneLoading] = useState(!!(progressCount || pendingCount));
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const size = useWindowSize();

  const currentRefreshDeploymentLength = useRef(refreshDeploymentsLength);
  const inprogressRef = useRef<HTMLElement>();
  const dynamicTimer = useRef<HTMLElement>();

  // deploymentStatus = <inprogress|pending>
  const refreshDeployments = useCallback(
    deploymentStatus => {
      const { page, perPage } = selectionState[deploymentStatus];
      return dispatch(getDeploymentsByStatus({ status: deploymentStatus, page, perPage }))
        .then(({ payload }) => {
          clearRetryTimer(deploymentStatus, dispatchedSetSnackbar);
          const { total, deploymentIds } = payload[payload.length - 1];
          if (total && !deploymentIds.length) {
            return refreshDeployments(deploymentStatus);
          }
        })
        .catch(err => setRetryTimer(err, 'deployments', `Couldn't load deployments.`, refreshDeploymentsLength, dispatchedSetSnackbar))
        .finally(() => setDoneLoading(true));
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [dispatch, dispatchedSetSnackbar, pendingPage, pendingPerPage, progressPage, progressPerPage]
  );

  const setupDeploymentsRefresh = useCallback(
    (refreshLength = currentRefreshDeploymentLength.current) => {
      const tasks = [refreshDeployments(DEPLOYMENT_STATES.inprogress), refreshDeployments(DEPLOYMENT_STATES.pending)];
      if (!onboardingState.complete && !pastDeploymentsCount) {
        // retrieve past deployments outside of the regular refresh cycle to not change the selection state for past deployments
        dispatch(getDeploymentsByStatus({ status: DEPLOYMENT_STATES.finished, page: 1, perPage: 1, shouldSelect: false }));
      }
      return Promise.all(tasks)
        .then(() => {
          currentRefreshDeploymentLength.current = Math.min(refreshDeploymentsLength, refreshLength * 2);
          clearTimeout(dynamicTimer.current);
          dynamicTimer.current = setTimeout(setupDeploymentsRefresh, currentRefreshDeploymentLength.current);
        })
        .finally(() => setDoneLoading(true));
    },
    [dispatch, onboardingState.complete, pastDeploymentsCount, refreshDeployments]
  );

  useEffect(
    () => () => {
      clearTimeout(dynamicTimer.current);
    },
    []
  );

  useEffect(
    () => () => {
      clearAllRetryTimers(dispatchedSetSnackbar);
    },
    [dispatchedSetSnackbar]
  );

  useEffect(() => {
    clearTimeout(dynamicTimer.current);
    setupDeploymentsRefresh(minimalRefreshDeploymentsLength);
    return () => {
      clearTimeout(dynamicTimer.current);
    };
  }, [pendingCount, setupDeploymentsRefresh]);

  useEffect(() => {
    clearTimeout(dynamicTimer.current);
    setupDeploymentsRefresh();
    return () => {
      clearTimeout(dynamicTimer.current);
    };
  }, [progressPage, progressPerPage, pendingPage, pendingPerPage, setupDeploymentsRefresh]);

  const abortDeployment = id =>
    abort(id).then(() => Promise.all([refreshDeployments(DEPLOYMENT_STATES.inprogress), refreshDeployments(DEPLOYMENT_STATES.pending)]));

  const onChangePage = state => page => dispatch(setDeploymentsState({ [state]: { page } }));
  const onChangeRowsPerPage = state => perPage => dispatch(setDeploymentsState({ [state]: { page: 1, perPage } }));

  let onboardingComponent = null;
  if (!onboardingState.complete && inprogressRef.current) {
    const anchor = {
      left: inprogressRef.current.offsetLeft + (inprogressRef.current.offsetWidth / 100) * 90,
      top: inprogressRef.current.offsetTop + inprogressRef.current.offsetHeight
    };
    onboardingComponent = getOnboardingComponentFor(onboardingSteps.DEPLOYMENTS_INPROGRESS, onboardingState, { anchor });
  }
  const props = { ...remainder, canDeploy, canConfigure, devices, idAttribute, isEnterprise };
  return doneLoading ? (
    <div className="fadeIn">
      {!!progress.length && (
        <div className="margin-top margin-bottom-large">
          <Typography className="margin-bottom" variant="subtitle1">
            In progress now
          </Typography>
          <DeploymentsList
            {...props}
            abort={abortDeployment}
            count={progressCount}
            items={progress}
            page={progressPage}
            pageSize={progressPerPage}
            rootRef={inprogressRef}
            onChangeRowsPerPage={onChangeRowsPerPage(DEPLOYMENT_STATES.inprogress)}
            onChangePage={onChangePage(DEPLOYMENT_STATES.inprogress)}
            type={DEPLOYMENT_STATES.inprogress}
          />
        </div>
      )}
      {!!onboardingComponent && onboardingComponent}
      {!!pending.length && (
        <Paper variant="outlined" className="margin-top margin-bottom-large padding">
          <Typography variant="subtitle1">Pending</Typography>
          <DeploymentsList
            {...props}
            abort={abortDeployment}
            count={pendingCount}
            items={pending}
            page={pendingPage}
            pageSize={pendingPerPage}
            onChangeRowsPerPage={onChangeRowsPerPage(DEPLOYMENT_STATES.pending)}
            onChangePage={onChangePage(DEPLOYMENT_STATES.pending)}
            type={DEPLOYMENT_STATES.pending}
          />
        </Paper>
      )}
      {!(progressCount || pendingCount) && (
        <div className="dashboard-placeholder">
          <Typography>Pending and ongoing deployments will appear here.</Typography>
          {canDeploy && (
            <Typography>
              <a onClick={createClick}>Create a deployment</a> to get started
            </Typography>
          )}
        </div>
      )}
    </div>
  ) : (
    <Loader show={doneLoading} />
  );
};

export default Progress;
