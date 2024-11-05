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
import React from 'react';
import { Link } from 'react-router-dom';

import { Launch as LaunchIcon } from '@mui/icons-material';
import { makeStyles } from 'tss-react/mui';

import { TwoColumns } from '@northern.tech/common-ui/configurationobject';
import DeviceIdentityDisplay from '@northern.tech/common-ui/deviceidentity';
import { AUDIT_LOGS_TYPES, BEGINNING_OF_TIME, rootfsImageVersion } from '@northern.tech/store/constants';
import { formatAuditlogs } from '@northern.tech/utils/locationutils';

const useStyles = makeStyles()(theme => ({
  eventDetails: { gridTemplateColumns: 'minmax(max-content, 150px) max-content', rowGap: theme.spacing(2.5) },
  deviceLink: { color: theme.palette.text.secondary, fontWeight: 'initial' }
}));

export const DetailInformation = ({ title, details }) => {
  const { classes } = useStyles();
  return (
    <div key={`${title}-details`} className="flexbox column margin-top-small">
      <b className="margin-bottom-small capitalized-start">{title} details</b>
      <TwoColumns className={classes.eventDetails} items={details} />
    </div>
  );
};

const deviceAuditlogType = AUDIT_LOGS_TYPES.find(type => type.value === 'device');

export const DeviceDetails = ({ device, idAttribute, onClose }) => {
  const { classes } = useStyles();
  const { attributes, id: deviceId } = device;
  const { name, device_type: deviceTypes, artifact_name } = attributes || {};
  const { attribute } = idAttribute;
  const usesId = attribute === 'id' || attribute === 'Device ID';
  const nameContainer = name ? { Name: name } : {};
  const deviceDetails = {
    ...nameContainer,
    [usesId ? 'Device ID' : attribute]: (
      <Link className={`flexbox center-aligned ${classes.deviceLink}`} to={`/devices?id=${deviceId}`}>
        <DeviceIdentityDisplay device={device} isEditable={false} />
        <LaunchIcon className="margin-left-small link-color" fontSize="small" />
      </Link>
    ),
    'Device type': deviceTypes,
    'Operating system version': device[rootfsImageVersion] || artifact_name || '-',
    ' ': (
      <Link
        to={`/auditlog?${formatAuditlogs({ pageState: { type: deviceAuditlogType, detail: deviceId, startDate: BEGINNING_OF_TIME } }, {})}`}
        onClick={onClose}
      >
        List all log entries for this device
      </Link>
    )
  };

  return <DetailInformation title="device" details={deviceDetails} />;
};

export default DeviceDetails;
