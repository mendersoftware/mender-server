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
import type { ReactNode } from 'react';
import { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';

import Loader from '@northern.tech/common-ui/Loader';
import Time from '@northern.tech/common-ui/Time';
import type { AuditLog, Device, Object } from '@northern.tech/store/api/types';
import { getAuditlogDevice, getIdAttribute, getUserCapabilities } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { getDeviceById, getSessionDetails } from '@northern.tech/store/thunks';
import dayjs from 'dayjs';
import duration from 'dayjs/plugin/duration.js';

import DeviceDetails, { DetailInformation } from './DeviceDetails';

dayjs.extend(duration);

export const stringifyEvent = (item: unknown, space?: number) => {
  try {
    return JSON.stringify(item, null, space);
  } catch (error) {
    return `error parsing the logged event:\n${error}`;
  }
};

export const parseConfigChange = (change: string) => {
  try {
    return JSON.parse(change);
  } catch (error) {
    return { error: `An error occurred processing the changed config:\n${error}` };
  }
};

export interface EventDetailsProps {
  item: AuditLog;
  onClose: () => void;
}

export const useAuditlogDevice = () => {
  const dispatch = useAppDispatch();
  const { canReadDevices } = useSelector(getUserCapabilities);
  const device = useSelector(getAuditlogDevice) as Device;
  const idAttribute = useSelector(getIdAttribute);
  const [isLoading, setIsLoading] = useState(canReadDevices && !device.attributes?.device_type?.length);
  const { id: deviceId } = device;

  useEffect(() => {
    if (canReadDevices) {
      dispatch(getDeviceById(deviceId))
        .unwrap()
        .then(() => setIsLoading(false));
    }
  }, [canReadDevices, deviceId, dispatch]);

  return { canReadDevices, device, idAttribute, isLoading };
};

export const DeviceDetailsSection = ({ onClose }: Pick<EventDetailsProps, 'onClose'>) => {
  const { canReadDevices, device, idAttribute, isLoading } = useAuditlogDevice();

  if (!canReadDevices) {
    return null;
  }
  return isLoading ? <Loader show={true} /> : <DeviceDetails device={device} idAttribute={idAttribute} onClose={onClose} />;
};

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
  isLoading: boolean;
  sessionDetails?: SessionDetails;
  sessionMeta: SessionMeta | Record<string, never>;
}

export const useSessionDetails = (auditLogItem: AuditLog): UseSessionDetailsReturn => {
  const [sessionDetails, setSessionDetails] = useState<SessionDetails | undefined>();
  const dispatch = useAppDispatch();
  const { action, actor, meta, object = {} as Object, time } = auditLogItem;

  useEffect(() => {
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
      .then(setSessionDetails)
      .catch(e => {
        console.error('failed to retrieve session details for auditlog event', e);
        setSessionDetails({ end: time, start: time });
      });
  }, [action, actor.id, dispatch, meta.session_id, object.id, time]);

  const sessionMeta: SessionMeta | Record<string, never> = sessionDetails
    ? {
        'Session ID': auditLogItem.meta.session_id[0],
        'Start time': <Time value={sessionDetails.start} />,
        'End time': <Time value={sessionDetails.end} />,
        Duration: dayjs.duration(dayjs(sessionDetails.end).diff(sessionDetails.start)).format('HH:mm:ss:SSS'),
        User: auditLogItem.actor.email
      }
    : {};

  return { sessionDetails, isLoading: !sessionDetails, sessionMeta };
};

interface SessionInfoProps extends Pick<EventDetailsProps, 'onClose'>, Pick<UseSessionDetailsReturn, 'sessionMeta'> {
  title: string;
}

export const SessionInfo = ({ sessionMeta, onClose, title }: SessionInfoProps) => (
  <div className="flexbox column">
    <DeviceDetailsSection onClose={onClose} />
    <DetailInformation title={title} details={sessionMeta} />
  </div>
);
