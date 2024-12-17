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
import { useEffect, useRef } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { getDevicesById } from '@northern.tech/store/selectors';
import { getDeviceById } from '@northern.tech/store/thunks';
import { isUUID } from 'validator';

export const useDeploymentDevice = deploymentName => {
  const isLoading = useRef(false);
  const dispatch = useDispatch();
  const devicesById = useSelector(getDevicesById);
  const hasDeviceInfo = !!devicesById[deploymentName];

  useEffect(() => {
    if (isLoading.current) {
      return;
    }
    isLoading.current = true;
    if (isUUID(deploymentName) && !hasDeviceInfo) {
      dispatch(getDeviceById(deploymentName)).then(() => (isLoading.current = false));
    }
  }, [deploymentName, dispatch, hasDeviceInfo]);
};
