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

// material ui
import { TextField } from '@mui/material';

import historyImage from '../../../assets/img/history.png';
import { setSnackbar } from '../../actions/appActions';
import { getDeploymentsByStatus, setDeploymentsState } from '../../actions/deploymentActions';
import { advanceOnboarding } from '../../actions/onboardingActions';
import { BEGINNING_OF_TIME, SORTING_OPTIONS } from '../../constants/appConstants';
import { DEPLOYMENT_STATES, DEPLOYMENT_TYPES } from '../../constants/deploymentConstants';
import { onboardingSteps } from '../../constants/onboardingConstants';
import { getISOStringBoundaries } from '../../helpers';
import {
  getDeploymentsSelectionState,
  getDevicesById,
  getGroupNames,
  getIdAttribute,
  getMappedDeploymentSelection,
  getOnboardingState,
  getUserCapabilities
} from '../../selectors';
import { getOnboardingComponentFor } from '../../utils/onboardingmanager';
import useWindowSize from '../../utils/resizehook';
import { clearAllRetryTimers, clearRetryTimer, setRetryTimer } from '../../utils/retrytimer';
import { ControlledAutoComplete } from '../common/forms/autocomplete';
import Filters from '../common/forms/filters';
import TimeframePicker from '../common/forms/timeframe-picker';
import { DeploymentSize, DeploymentStatus } from './deploymentitem';
import { defaultRefreshDeploymentsLength as refreshDeploymentsLength } from './deployments';
import DeploymentsList, { defaultHeaders } from './deploymentslist';

const headers = [
  ...defaultHeaders.slice(0, defaultHeaders.length - 1),
  { title: 'Status', renderer: DeploymentStatus },
  { title: 'Data downloaded', renderer: DeploymentSize }
];

const type = DEPLOYMENT_STATES.finished;

export const Past = props => {
  const { createClick, isShowingDetails } = props;
  // eslint-disable-next-line no-unused-vars
  const size = useWindowSize();
  const [tonight] = useState(getISOStringBoundaries(new Date()).end);
  const [loading, setLoading] = useState(false);
  const deploymentsRef = useRef();
  const timer = useRef();

  const dispatch = useDispatch();
  const dispatchedSetSnackbar = useCallback((...args) => dispatch(setSnackbar(...args)), [dispatch]);

  const { finished: pastSelectionState } = useSelector(getDeploymentsSelectionState);
  const past = useSelector(state => getMappedDeploymentSelection(state, type));
  const { canConfigure, canDeploy } = useSelector(getUserCapabilities);
  const idAttribute = useSelector(getIdAttribute);
  const onboardingState = useSelector(getOnboardingState);
  const devices = useSelector(getDevicesById);
  const groupNames = useSelector(getGroupNames);

  const { endDate, page, perPage, search: deviceGroup, startDate, total: count, type: deploymentType } = pastSelectionState;

  /*
  / refresh only finished deployments
  /
  */
  const refreshPast = useCallback(
    (
      currentPage = page,
      currentPerPage = perPage,
      currentStartDate = startDate,
      currentEndDate = endDate,
      currentDeviceGroup = deviceGroup,
      currentType = deploymentType
    ) => {
      const roundedStartDate = Math.round(Date.parse(currentStartDate) / 1000);
      const roundedEndDate = Math.round(Date.parse(currentEndDate) / 1000);
      setLoading(true);
      return dispatch(getDeploymentsByStatus(type, currentPage, currentPerPage, roundedStartDate, roundedEndDate, currentDeviceGroup, currentType))
        .then(deploymentsAction => {
          setLoading(false);
          clearRetryTimer(type, dispatchedSetSnackbar);
          const { total, deploymentIds } = deploymentsAction[deploymentsAction.length - 1];
          if (total && !deploymentIds.length) {
            return refreshPast(currentPage, currentPerPage, currentStartDate, currentEndDate, currentDeviceGroup);
          }
        })
        .catch(err => setRetryTimer(err, 'deployments', `Couldn't load deployments.`, refreshDeploymentsLength, dispatchedSetSnackbar));
    },
    [deploymentType, deviceGroup, dispatch, dispatchedSetSnackbar, endDate, page, perPage, startDate]
  );

  useEffect(() => {
    const roundedStartDate = Math.round(Date.parse(startDate || BEGINNING_OF_TIME) / 1000);
    const roundedEndDate = Math.round(Date.parse(endDate) / 1000);
    setLoading(true);
    dispatch(getDeploymentsByStatus(type, page, perPage, roundedStartDate, roundedEndDate, deviceGroup, deploymentType, true, SORTING_OPTIONS.desc))
      .then(deploymentsAction => {
        const deploymentsList = deploymentsAction ? Object.values(deploymentsAction[0].deployments) : [];
        if (deploymentsList.length) {
          let newStartDate = new Date(deploymentsList[deploymentsList.length - 1].created);
          const { start } = getISOStringBoundaries(newStartDate);
          dispatch(setDeploymentsState({ [DEPLOYMENT_STATES.finished]: { startDate: startDate || start } }));
        }
      })
      .finally(() => setLoading(false));
    return () => {
      clearAllRetryTimers(dispatchedSetSnackbar);
    };
  }, [deploymentType, deviceGroup, dispatch, dispatchedSetSnackbar, endDate, page, perPage, startDate]);

  useEffect(() => {
    clearInterval(timer.current);
    timer.current = setInterval(refreshPast, refreshDeploymentsLength);
    // refreshPast();
    return () => {
      clearInterval(timer.current);
    };
  }, [page, perPage, startDate, endDate, deviceGroup, deploymentType, refreshPast]);

  useEffect(() => {
    if (!past.length || onboardingState.complete) {
      return;
    }
    const pastDeploymentsFailed = past.reduce(
      (accu, item) =>
        item.status === 'failed' ||
        (item.statistics?.status &&
          item.statistics.status.noartifact + item.statistics.status.failure + item.statistics.status['already-installed'] + item.statistics.status.aborted >
            0) ||
        accu,
      false
    );
    let onboardingStep = onboardingSteps.DEPLOYMENTS_PAST;
    if (pastDeploymentsFailed) {
      onboardingStep = onboardingSteps.DEPLOYMENTS_PAST_COMPLETED_FAILURE;
    }
    dispatch(advanceOnboarding(onboardingStep));
  }, [dispatch, onboardingState.complete, past]);

  let onboardingComponent = null;
  if (deploymentsRef.current) {
    const detailsButtons = deploymentsRef.current.getElementsByClassName('MuiButton-contained');
    const left = detailsButtons.length
      ? deploymentsRef.current.offsetLeft + detailsButtons[0].offsetLeft + detailsButtons[0].offsetWidth / 2 + 15
      : deploymentsRef.current.offsetWidth;
    let anchor = { left: deploymentsRef.current.offsetWidth / 2, top: deploymentsRef.current.offsetTop };
    onboardingComponent = getOnboardingComponentFor(onboardingSteps.DEPLOYMENTS_PAST_COMPLETED, onboardingState, {
      anchor,
      setSnackbar: dispatchedSetSnackbar
    });
    onboardingComponent = getOnboardingComponentFor(
      onboardingSteps.DEPLOYMENTS_PAST_COMPLETED_FAILURE,
      onboardingState,
      { anchor: { left, top: detailsButtons[0].parentElement.offsetTop + detailsButtons[0].parentElement.offsetHeight } },
      onboardingComponent
    );
  }

  const onFiltersChange = useCallback(
    ({ endDate, group, startDate, type }) =>
      dispatch(setDeploymentsState({ [DEPLOYMENT_STATES.finished]: { page: 1, search: group, type, startDate, endDate } })),
    [dispatch]
  );

  const autoCompleteProps = { autoHighlight: true, autoSelect: true, filterSelectedOptions: true, freeSolo: true, handleHomeEndKeys: true };
  return (
    <div className="fadeIn margin-left margin-top-large">
      <Filters
        initialValues={{ startDate, endDate, group: deviceGroup, type: deploymentType }}
        defaultValues={{ startDate: '', endDate: tonight, group: '', type: '' }}
        filters={[
          {
            key: 'group',
            title: 'Device group',
            Component: ControlledAutoComplete,
            componentProps: {
              ...autoCompleteProps,
              options: groupNames,
              renderInput: params => <TextField {...params} label="Target devices" placeholder="Select a group" InputProps={{ ...params.InputProps }} />
            }
          },
          {
            key: 'type',
            title: 'Contains Artifact type',
            Component: ControlledAutoComplete,
            componentProps: {
              ...autoCompleteProps,
              options: Object.keys(DEPLOYMENT_TYPES),
              renderInput: params => <TextField {...params} label="Deployment type" placeholder="Select a type" InputProps={{ ...params.InputProps }} />
            }
          },
          {
            key: 'timeframe',
            title: 'Start time',
            Component: TimeframePicker,
            componentProps: {
              tonight
            }
          }
        ]}
        onChange={onFiltersChange}
      />
      <div className="deploy-table-contain">
        {/* TODO: fix status retrieval for past deployments to decide what to show here - */}
        {!loading && !!past.length && !!onboardingComponent && !isShowingDetails && onboardingComponent}
        {!!past.length && (
          <DeploymentsList
            {...props}
            canConfigure={canConfigure}
            canDeploy={canDeploy}
            componentClass="margin-left-small"
            count={count}
            devices={devices}
            headers={headers}
            idAttribute={idAttribute}
            items={past}
            loading={loading}
            onChangePage={page => dispatch(setDeploymentsState({ [DEPLOYMENT_STATES.finished]: { page } }))}
            onChangeRowsPerPage={perPage => dispatch(setDeploymentsState({ [DEPLOYMENT_STATES.finished]: { page: 1, perPage } }))}
            page={page}
            pageSize={perPage}
            rootRef={deploymentsRef}
            showPagination
            type={type}
          />
        )}
        {!(loading || past.length) && (
          <div className="dashboard-placeholder">
            <p>No finished deployments were found.</p>
            <p>
              Try adjusting the filters, or <a onClick={createClick}>Create a new deployment</a> to get started
            </p>
            <img src={historyImage} alt="Past" />
          </div>
        )}
      </div>
    </div>
  );
};

export default Past;
