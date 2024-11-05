// Copyright 2021 Northern.tech AS
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
import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { useTheme } from '@mui/material/styles';

import Loader from '@northern.tech/common-ui/loader';
import { getAuditlogDevice, getIdAttribute, getUserCapabilities } from '@northern.tech/store/selectors';
import { getDeviceById } from '@northern.tech/store/thunks';

import DeviceDetails, { DetailInformation } from './devicedetails';

export const FileTransfer = ({ item, onClose }) => {
  const dispatch = useDispatch();
  const {
    actor,
    meta: { path = [] },
    object = {}
  } = item;
  const device = useSelector(getAuditlogDevice);
  const { canReadDevices } = useSelector(getUserCapabilities);
  const idAttribute = useSelector(getIdAttribute);
  const theme = useTheme();

  useEffect(() => {
    if (canReadDevices) {
      dispatch(getDeviceById(object.id));
    }
  }, [canReadDevices, dispatch, object.id]);

  if (canReadDevices && !device) {
    return <Loader show={true} />;
  }

  const sessionMeta = {
    Path: path.join(','),
    User: actor.email
  };

  return (
    <div className="flexbox column" style={{ margin: theme.spacing(3), minWidth: 'min-content' }}>
      {canReadDevices && <DeviceDetails device={device} idAttribute={idAttribute} onClose={onClose} />}
      <DetailInformation title="file transfer" details={sessionMeta} />
    </div>
  );
};

export default FileTransfer;
