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
import { Button } from '@mui/material';

import { ContentSection } from '@northern.tech/common-ui/ContentSection';
import DeviceIdentityDisplay from '@northern.tech/common-ui/DeviceIdentity';
import { Link } from '@northern.tech/common-ui/Link';
import { TwoColumnData } from '@northern.tech/common-ui/TwoColumnData';
import { AUDIT_LOGS_TYPES, BEGINNING_OF_TIME, rootfsImageVersion } from '@northern.tech/store/constants';
import { formatAuditlogs } from '@northern.tech/store/locationutils';

const deviceAuditlogType = AUDIT_LOGS_TYPES.find(type => type.value === 'device');

export const DeviceDetails = ({ device, idAttribute, onClose }) => {
  const { attributes, id: deviceId } = device;
  const { name, device_type: deviceTypes, artifact_name } = attributes || {};
  const { attribute } = idAttribute;
  const usesId = attribute === 'id' || attribute === 'Device ID';
  const nameContainer = name ? { Name: name } : {};
  const deviceDetails = {
    ...nameContainer,
    [usesId ? 'Device ID' : attribute]: (
      <Link className="flexbox align-items-center" to={`/devices?id=${deviceId}`}>
        <DeviceIdentityDisplay device={device} isEditable={false} />
      </Link>
    ),
    'Device type': deviceTypes,
    'Operating system version': device[rootfsImageVersion] || artifact_name || '-'
  };

  return (
    <ContentSection title="Device details">
      <TwoColumnData data={deviceDetails} />
      <div className="margin-top-small">
        <Button
          color="secondary"
          component={Link}
          to={`/auditlog?${formatAuditlogs({ pageState: { type: deviceAuditlogType, detail: deviceId, startDate: BEGINNING_OF_TIME } }, {})}`}
          onClick={onClose}
        >
          List all log entries for this device
        </Button>
      </div>
    </ContentSection>
  );
};

export default DeviceDetails;
