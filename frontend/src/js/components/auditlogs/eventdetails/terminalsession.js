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
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { useTheme } from '@mui/material/styles';

import Loader from '@northern.tech/common-ui/loader';
import Time from '@northern.tech/common-ui/time';
import { getAuditlogDevice, getCurrentSession, getIdAttribute, getUserCapabilities } from '@northern.tech/store/selectors';
import { getDeviceById, getSessionDetails } from '@northern.tech/store/thunks';
import dayjs from 'dayjs';
import durationDayJs from 'dayjs/plugin/duration';

import DeviceDetails, { DetailInformation } from './devicedetails';
import TerminalPlayer from './terminalplayer';

dayjs.extend(durationDayJs);

export const TerminalSession = ({ item, onClose }) => {
  const theme = useTheme();
  const [sessionDetails, setSessionDetails] = useState();
  const dispatch = useDispatch();
  const { action, actor, meta, object = {}, time } = item;
  const { canReadDevices } = useSelector(getUserCapabilities);
  const device = useSelector(getAuditlogDevice);
  const idAttribute = useSelector(getIdAttribute);
  const { token } = useSelector(getCurrentSession);

  useEffect(() => {
    if (canReadDevices) {
      dispatch(getDeviceById(object.id));
    }
    dispatch(
      getSessionDetails({
        sessionId: meta.session_id[0],
        deviceId: object.id,
        userId: actor.id,
        startDate: action.startsWith('open') ? time : undefined,
        endDate: action.startsWith('close') ? time : undefined
      })
    )
      .unwrap()
      .then(setSessionDetails);
  }, [action, actor.id, canReadDevices, dispatch, meta.session_id, object.id, time]);

  if (!sessionDetails || (canReadDevices && !device)) {
    return <Loader show={true} />;
  }

  const sessionMeta = {
    'Session ID': item.meta.session_id[0],
    'Start time': <Time value={sessionDetails.start} />,
    'End time': <Time value={sessionDetails.end} />,
    'Duration': dayjs.duration(dayjs(sessionDetails.end).diff(sessionDetails.start)).format('HH:mm:ss:SSS'),
    User: item.actor.email
  };

  return (
    <div className="flexbox" style={{ flexWrap: 'wrap' }}>
      <TerminalPlayer className="flexbox column margin-top" item={item} sessionInitialized={!!sessionDetails} token={token} />
      <div className="flexbox column" style={{ margin: theme.spacing(3), minWidth: 'min-content' }}>
        {canReadDevices && <DeviceDetails device={device} idAttribute={idAttribute} onClose={onClose} />}
        <DetailInformation title="session" details={sessionMeta} />
      </div>
    </div>
  );
};

export default TerminalSession;
