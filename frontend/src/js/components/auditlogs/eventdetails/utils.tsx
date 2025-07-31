// Copyright 2025 Northern.tech AS
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
import { ReactNode, useEffect, useState } from 'react';
import { useSelector } from 'react-redux';

import Time from '@northern.tech/common-ui/Time';
import type { AuditLog, Device, Object } from '@northern.tech/store/api/types/MenderTypes';
import type { IdAttribute } from '@northern.tech/store/constants';
import { getAuditlogDevice, getIdAttribute, getUserCapabilities } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { getDeviceById, getSessionDetails } from '@northern.tech/store/thunks';
import dayjs from 'dayjs';
import duration from 'dayjs/plugin/duration';

dayjs.extend(duration);

export interface SessionDetailsEventProps {
  item: AuditLog;
  onClose: () => void;
}

interface SessionDetails {
  end: string;
  start: string;
}

interface SessionMeta {
  Duration: string;
  'End time': ReactNode;
  'Session ID': string;
  'Start time': ReactNode;
  User?: string;
}

interface UseSessionDetailsReturn {
  canReadDevices: boolean;
  device?: Device;
  idAttribute: IdAttribute | string;
  isLoading: boolean;
  sessionDetails?: SessionDetails;
  sessionMeta: SessionMeta | Record<string, never>;
}

export const useSessionDetails = (auditLogItem: AuditLog): UseSessionDetailsReturn => {
  const [sessionDetails, setSessionDetails] = useState<SessionDetails | undefined>();
  const dispatch = useAppDispatch();
  const { action, actor, meta, object = {} as Object, time } = auditLogItem;
  const { canReadDevices } = useSelector(getUserCapabilities);
  const idAttribute = useSelector(getIdAttribute) as string;
  const device = useSelector(getAuditlogDevice) as Device | undefined;

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
      .catch(e => {
        console.error('failed to retrieve session details for auditlog event', e);
        setSessionDetails({ end: time, start: time });
      })
      .then(setSessionDetails);
  }, [action, actor.id, canReadDevices, dispatch, meta.session_id, object.id, time]);

  const isLoading = !sessionDetails || (canReadDevices && !device);

  const sessionMeta: SessionMeta | Record<string, never> = sessionDetails
    ? {
        'Session ID': auditLogItem.meta.session_id[0],
        'Start time': <Time value={sessionDetails.start} />,
        'End time': <Time value={sessionDetails.end} />,
        Duration: dayjs.duration(dayjs(sessionDetails.end).diff(sessionDetails.start)).format('HH:mm:ss:SSS'),
        User: auditLogItem.actor.email
      }
    : {};

  return {
    sessionDetails,
    device,
    idAttribute,
    canReadDevices,
    isLoading,
    sessionMeta
  };
};
