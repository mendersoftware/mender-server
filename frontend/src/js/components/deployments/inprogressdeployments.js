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
import React, { useCallback, useEffect, useRef, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { Refresh as RefreshIcon } from '@mui/icons-material';
import { makeStyles } from 'tss-react/mui';

import { setSnackbar } from '../../actions/appActions';
import { getDeploymentsByStatus, setDeploymentsState } from '../../actions/deploymentActions';
import { DEPLOYMENT_STATES } from '../../constants/deploymentConstants';
import { onboardingSteps } from '../../constants/onboardingConstants';
import {
  getDeploymentsByStatus as getDeploymentsByStatusSelector,
  getDeploymentsSelectionState,
  getDevicesById,
  getIdAttribute,
  getIsEnterprise,
  getMappedDeploymentSelection,
  getOnboardingState,
  getUserCapabilities
} from '../../selectors';
import { getOnboardingComponentFor } from '../../utils/onboardingmanager';
import useWindowSize from '../../utils/resizehook';
import { clearAllRetryTimers, clearRetryTimer, setRetryTimer } from '../../utils/retrytimer';
import LinedHeader from '../common/lined-header';
import Loader from '../common/loader';
import { defaultRefreshDeploymentsLength as refreshDeploymentsLength } from './deployments';
import DeploymentsList from './deploymentslist';

export const minimalRefreshDeploymentsLength = 2000;

const useStyles = makeStyles()(theme => ({
  deploymentsPending: {
    borderColor: 'rgba(0, 0, 0, 0.06)',
    backgroundColor: theme.palette.background.light,
    color: theme.palette.text.primary,
    ['.dashboard-header span']: {
      backgroundColor: theme.palette.background.light,
      color: theme.palette.text.primary
    },
    ['.MuiButtonBase-root']: {
      color: theme.palette.text.primary
    }
  }
}));

export const Progress = ({ abort, createClick, ...remainder }) => {
  const { canConfigure, canDeploy } = useSelector(getUserCapabilities);
  const { attribute: idAttribute } = useSelector(getIdAttribute);
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
  // eslint-disable-next-line no-unused-vars
  const size = useWindowSize();

  const currentRefreshDeploymentLength = useRef(refreshDeploymentsLength);
  const inprogressRef = useRef();
  const dynamicTimer = useRef();

  const { classes } = useStyles();

  // deploymentStatus = <inprogress|pending>
  const refreshDeployments = useCallback(
    deploymentStatus => {
      const { page, perPage } = selectionState[deploymentStatus];
      return dispatch(getDeploymentsByStatus(deploymentStatus, page, perPage))
        .then(deploymentsAction => {
          clearRetryTimer(deploymentStatus, dispatchedSetSnackbar);
          const { total, deploymentIds } = deploymentsAction[deploymentsAction.length - 1];
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
      let tasks = [refreshDeployments(DEPLOYMENT_STATES.inprogress), refreshDeployments(DEPLOYMENT_STATES.pending)];
      if (!onboardingState.complete && !pastDeploymentsCount) {
        // retrieve past deployments outside of the regular refresh cycle to not change the selection state for past deployments
        dispatch(getDeploymentsByStatus(DEPLOYMENT_STATES.finished, 1, 1, undefined, undefined, undefined, undefined, false));
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

  useEffect(() => {
    return () => {
      clearTimeout(dynamicTimer.current);
    };
  }, []);

  useEffect(() => {
    return () => {
      clearAllRetryTimers(dispatchedSetSnackbar);
    };
  }, [dispatchedSetSnackbar]);

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
        <div className="margin-left">
          <LinedHeader className="margin-top-large  margin-right" heading="In progress now" />
          <DeploymentsList
            {...props}
            abort={abortDeployment}
            count={progressCount}
            items={progress}
            listClass="margin-right-small"
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
        <div className={`deployments-pending margin-top margin-bottom-large ${classes.deploymentsPending}`}>
          <LinedHeader className="margin-small margin-top" heading="Pending" />
          <DeploymentsList
            {...props}
            abort={abortDeployment}
            componentClass="margin-left-small"
            count={pendingCount}
            items={pending}
            page={pendingPage}
            pageSize={pendingPerPage}
            onChangeRowsPerPage={onChangeRowsPerPage(DEPLOYMENT_STATES)}
            onChangePage={onChangePage(DEPLOYMENT_STATES.pending)}
            type={DEPLOYMENT_STATES.pending}
          />
        </div>
      )}
      {!(progressCount || pendingCount) && (
        <div className="dashboard-placeholder">
          <p>Pending and ongoing deployments will appear here. </p>
          {canDeploy && (
            <p>
              <a onClick={createClick}>Create a deployment</a> to get started
            </p>
          )}
          <RefreshIcon className="flip-horizontal" style={{ fill: '#e3e3e3', width: 111, height: 111 }} />
        </div>
      )}
    </div>
  ) : (
    <Loader show={doneLoading} />
  );
};

export default Progress;
