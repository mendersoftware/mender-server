// Copyright 2019 Northern.tech AS
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
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useDispatch } from 'react-redux';
import { Link } from 'react-router-dom';

import { ErrorOutline as ErrorOutlineIcon } from '@mui/icons-material';
import { Autocomplete, TextField, Tooltip } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import AsyncAutocomplete from '@northern.tech/common-ui/AsyncAutocomplete';
import { getDeviceIdentityText } from '@northern.tech/common-ui/DeviceIdentity';
import InfoText from '@northern.tech/common-ui/InfoText';
import { HELPTOOLTIPS } from '@northern.tech/common-ui/helptips/HelpTooltips';
import { MenderHelpTooltip } from '@northern.tech/common-ui/helptips/MenderTooltip';
import { ALL_DEVICES, ATTRIBUTE_SCOPES, DEPLOYMENT_TYPES, DEVICE_FILTERING_OPTIONS, DEVICE_STATES } from '@northern.tech/store/constants';
import { formatDeviceSearch } from '@northern.tech/store/locationutils';
import { getReleases, getSystemDevices } from '@northern.tech/store/thunks';
import { stringToBoolean } from '@northern.tech/utils/helpers';
import { useWindowSize } from '@northern.tech/utils/resizehook';
import pluralize from 'pluralize';
import isUUID from 'validator/lib/isUUID';

const useStyles = makeStyles()(theme => ({
  infoStyle: {
    minWidth: 400,
    borderBottom: 'none'
  },
  selection: { minWidth: 'min-content', maxWidth: theme.spacing(50), minHeight: 96, marginBottom: theme.spacing(2) }
}));

const hardCodedStyle = {
  textField: {
    minWidth: 400
  }
};

export const getDevicesLink = ({ devices, filters = [], group, name }) => {
  let devicesLink = '/devices';
  if (filters.length) {
    return `${devicesLink}?${formatDeviceSearch({ pageState: {}, filters, selectedGroup: group })}`;
  }
  // older deployments won't have the filter set so we have to try to guess their targets based on other information
  if (devices.length && (!name || isUUID(name))) {
    devicesLink = `${devicesLink}?${devices.map(({ id }) => `id=${id}`).join('&')}`;
    if (devices.length === 1) {
      const { systemDeviceIds = [] } = devices[0];
      devicesLink = `${devicesLink}${systemDeviceIds.map(id => `&id=${id}`).join('')}`;
    }
  } else if (group) {
    devicesLink = `${devicesLink}?${formatDeviceSearch({ pageState: {}, filters, selectedGroup: group })}`;
  }
  return devicesLink;
};

const deploymentFiltersToTargetText = ({ devicesById, filter, idAttribute }) => {
  const { name, filters = [] } = filter;
  if (name) {
    return name;
  }
  if (
    filters.some(
      ({ operator, scope, value }) => scope === ATTRIBUTE_SCOPES.identity && value === DEVICE_STATES.accepted && operator === DEVICE_FILTERING_OPTIONS.$eq.key
    )
  ) {
    return ALL_DEVICES;
  }
  const groupFilter = filters.find(
    ({ operator, scope, key }) => scope === ATTRIBUTE_SCOPES.system && operator === DEVICE_FILTERING_OPTIONS.$eq.key && key === 'group'
  );
  if (groupFilter) {
    return groupFilter.value;
  }
  return filters
    .reduce((accu, { operator, scope, key, value }) => {
      if (!(key === 'id' && scope === ATTRIBUTE_SCOPES.identity)) {
        return accu;
      }
      if (operator === DEVICE_FILTERING_OPTIONS.$in.key) {
        const devices = value.map(deviceId => getDeviceIdentityText({ device: devicesById[deviceId], idAttribute }));
        return [...accu, ...devices];
      }
      accu.push(getDeviceIdentityText({ device: devicesById[value], idAttribute }));
      return accu;
    }, [])
    .join(', ');
};

export const getDeploymentTargetText = ({ deployment, devicesById, idAttribute }) => {
  const { devices = {}, filter = {}, group = '', name = '', type = DEPLOYMENT_TYPES.software } = deployment;
  const text = deploymentFiltersToTargetText({ devicesById, filter, idAttribute });
  if (text) {
    return text;
  }
  let deviceList = Array.isArray(devices) ? devices : Object.values(devices);
  if (isUUID(name) && devicesById[name]) {
    deviceList = [devicesById[name]];
  }
  if (type !== DEPLOYMENT_TYPES.configuration && (!deviceList.length || group || (deployment.name !== undefined && !isUUID(name)))) {
    return (group || name) ?? '';
  }
  return deviceList.map(device => getDeviceIdentityText({ device, idAttribute })).join(', ') || name;
};

export const ReleasesWarning = ({ lacksReleases }) => (
  <div className="flexbox center-aligned">
    <ErrorOutlineIcon fontSize="small" style={{ marginRight: 4, top: 4, color: 'rgb(171, 16, 0)' }} />
    <InfoText>
      There are no {lacksReleases ? 'compatible ' : ''}artifacts available.{lacksReleases ? <br /> : ' '}
      <Link to="/releases">Upload one to the repository</Link> to get started.
    </InfoText>
  </div>
);

export const Devices = ({
  deploymentObject,
  groupRef,
  groupNames,
  hasDevices,
  hasDynamicGroups,
  hasFullFiltering,
  hasPending,
  idAttribute,
  setDeploymentSettings
}) => {
  const { classes } = useStyles();
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const size = useWindowSize();
  const dispatch = useDispatch();

  const { deploymentDeviceCount = 0, devices = [], filter, group = null } = deploymentObject;
  // eslint-disable-next-line react-hooks/exhaustive-deps
  const device = useMemo(() => (devices.length === 1 ? devices[0] : {}), [devices.length]);

  useEffect(() => {
    const { attributes = {} } = device;
    const { mender_is_gateway } = attributes;
    if (!device.id || !stringToBoolean(mender_is_gateway)) {
      return;
    }
    dispatch(getSystemDevices({ id: device.id, perPage: 500 }));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [device.id, device.attributes?.mender_is_gateway, dispatch]);

  const deploymentSettingsUpdate = (e, value, reason) => {
    let update = { group: value };
    if (reason === 'clear') {
      update = { ...update, deploymentDeviceCount: 0, devices: [] };
    }
    setDeploymentSettings(update);
  };

  const { deviceText, devicesLink, targetDeviceCount, targetDevicesText } = useMemo(() => {
    const devicesLink = getDevicesLink({ devices, group, hasFullFiltering, filters: filter?.filters });
    let deviceText = getDeploymentTargetText({ deployment: deploymentObject, idAttribute });
    let targetDeviceCount = deploymentDeviceCount;
    let targetDevicesText = `${deploymentDeviceCount} ${pluralize('devices', deploymentDeviceCount)}`;
    if (device?.id) {
      const { attributes = {}, systemDeviceIds = [] } = device;
      const { mender_is_gateway } = attributes;
      deviceText = `${getDeviceIdentityText({ device, idAttribute })}${stringToBoolean(mender_is_gateway) ? ' (System)' : ''}`;
      // here we hope the number of systemDeviceIds doesn't exceed the queried 500 and add the gateway device
      targetDeviceCount = systemDeviceIds.length + 1;
    } else if (group) {
      deviceText = '';
      targetDevicesText = 'All devices';
      targetDeviceCount = 2;
      if (group !== ALL_DEVICES) {
        targetDevicesText = `${targetDevicesText} in this group`;
        targetDeviceCount = deploymentDeviceCount;
      }
    }
    return { deviceText, devicesLink, targetDeviceCount, targetDevicesText };
  }, [devices, filter, group, hasFullFiltering, deploymentObject, idAttribute, deploymentDeviceCount, device]);

  return (
    <>
      <h4 className="margin-top-none">Select a device group to target</h4>
      <div ref={groupRef} className={classes.selection}>
        {deviceText ? (
          <TextField value={deviceText} label={pluralize('device', devices.length)} disabled className={classes.infoStyle} />
        ) : (
          <div>
            <Autocomplete
              id="deployment-device-group-selection"
              autoSelect
              autoHighlight
              filterSelectedOptions
              handleHomeEndKeys
              disabled={!(hasDevices || hasDynamicGroups)}
              options={groupNames}
              onChange={deploymentSettingsUpdate}
              renderInput={params => (
                <TextField {...params} placeholder="Select a device group" InputProps={{ ...params.InputProps }} className={classes.textField} />
              )}
              value={group}
            />
            {!(hasDevices || hasDynamicGroups) && (
              <InfoText style={{ marginTop: '10px' }}>
                <ErrorOutlineIcon style={{ marginRight: '4px', fontSize: '18px', top: '4px', color: 'rgb(171, 16, 0)', position: 'relative' }} />
                There are no connected devices.{' '}
                {hasPending ? (
                  <span>
                    <Link to="/devices/pending">Accept pending devices</Link> to get started.
                  </span>
                ) : (
                  <span>
                    <Link to="/help/get-started">Read the help pages</Link> for help with connecting devices.
                  </span>
                )}
              </InfoText>
            )}
          </div>
        )}
        {!!targetDeviceCount && (
          <InfoText>
            {targetDevicesText} will be targeted. <Link to={devicesLink}>View the {pluralize('devices', targetDeviceCount)}</Link>
          </InfoText>
        )}
      </div>
    </>
  );
};

export const Software = ({ commonClasses, deploymentObject, releaseRef, releases, releasesById, setDeploymentSettings }) => {
  const [isLoadingReleases, setIsLoadingReleases] = useState(!releases.length);
  const dispatch = useDispatch();
  const { classes } = useStyles();
  const { devices = [], release: deploymentRelease = null, releaseSelectionLocked } = deploymentObject;
  const device = devices.length ? devices[0] : undefined;

  useEffect(() => {
    setIsLoadingReleases(!releases.length);
  }, [releases.length]);

  const releaseItems = useMemo(() => {
    let releaseItems = releases.map(rel => releasesById[rel]);
    if (device && device.attributes) {
      // If single device, don't show incompatible releases
      releaseItems = releaseItems.filter(rel => rel.device_types_compatible.some(type => device.attributes.device_type.includes(type)));
    }
    return releaseItems;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [device, releases]);

  const onReleaseSelectionChange = useCallback(
    release => {
      if (release !== deploymentObject.release) {
        setDeploymentSettings({ release });
      }
    },
    [deploymentObject.release, setDeploymentSettings]
  );

  const onReleaseInputChange = useCallback(
    inputValue => {
      setIsLoadingReleases(!releases.length);
      return dispatch(getReleases({ page: 1, perPage: 100, searchTerm: inputValue, searchOnly: true })).finally(() => setIsLoadingReleases(false));
    },
    [dispatch, releases.length]
  );

  const releaseDeviceTypes = (deploymentRelease && deploymentRelease.device_types_compatible) ?? [];
  const devicetypesInfo = (
    <Tooltip title={<p>{releaseDeviceTypes.join(', ')}</p>} placement="bottom">
      <span className="link">
        {releaseDeviceTypes.length} device {pluralize('types', releaseDeviceTypes.length)}
      </span>
    </Tooltip>
  );

  return (
    <>
      <h4 className="margin-top-none">Select a Release to deploy</h4>
      <div className={commonClasses.columns}>
        <div ref={releaseRef} className={classes.selection}>
          {releaseSelectionLocked ? (
            <TextField value={deploymentRelease?.name} label="Release" disabled className={classes.infoStyle} />
          ) : (
            <AsyncAutocomplete
              id="deployment-release-selection"
              initialValue={deploymentRelease?.name}
              labelAttribute="name"
              placeholder="Select a Release"
              selectionAttribute="name"
              options={releaseItems}
              onChange={onReleaseInputChange}
              onChangeSelection={onReleaseSelectionChange}
              isLoading={isLoadingReleases}
              styles={hardCodedStyle}
            />
          )}
          {!releaseItems.length ? (
            <ReleasesWarning lacksReleases />
          ) : (
            !!releaseDeviceTypes.length && <InfoText style={{ marginBottom: 0 }}>This Release is compatible with {devicetypesInfo}.</InfoText>
          )}
        </div>
        <div className="margin-left-small">
          <MenderHelpTooltip id={HELPTOOLTIPS.groupDeployment.id} />
        </div>
      </div>
    </>
  );
};
