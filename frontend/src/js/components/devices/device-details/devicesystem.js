// Copyright 2022 Northern.tech AS
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
import { Link, useNavigate } from 'react-router-dom';

import { Button } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { setSnackbar } from '../../../actions/appActions';
import { getSystemDevices } from '../../../actions/deviceActions';
import { BENEFITS, SORTING_OPTIONS } from '../../../constants/appConstants';
import { DEVICE_LIST_DEFAULTS } from '../../../constants/deviceConstants';
import { getDemoDeviceAddress, toggle } from '../../../helpers';
import { getCurrentSession, getDevicesById, getIdAttribute, getIsPreview, getOrganization } from '../../../selectors';
import { TwoColumnData } from '../../common/configurationobject';
import DocsLink from '../../common/docslink';
import EnterpriseNotification from '../../common/enterpriseNotification';
import { getHeaders } from '../authorized-devices';
import { routes } from '../base-devices';
import Devicelist from '../devicelist';
import ConnectToGatewayDialog from '../dialogs/connecttogatewaydialog';
import DeviceDataCollapse from './devicedatacollapse';

const useStyles = makeStyles()(theme => ({ container: { maxWidth: 600, marginTop: theme.spacing(), marginBottom: theme.spacing() } }));

export const DeviceSystem = ({ columnSelection, device, onConnectToGatewayClick, openSettingsDialog }) => {
  const [columnHeaders, setColumnHeaders] = useState([]);
  const [headerKeys, setHeaderKeys] = useState([]);
  const [page, setPage] = useState(DEVICE_LIST_DEFAULTS.page);
  const [perPage, setPerPage] = useState(DEVICE_LIST_DEFAULTS.perPage);
  const { classes } = useStyles();
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const devicesById = useSelector(getDevicesById);
  const idAttribute = useSelector(getIdAttribute);

  const { systemDeviceIds = [], systemDeviceTotal = 0 } = device;
  const deviceIp = getDemoDeviceAddress([device]);

  const [sortOptions, setSortOptions] = useState([]);

  const onSortChange = attribute => {
    let changedOrder = SORTING_OPTIONS.asc;
    if (sortOptions.length && sortOptions[0].attribute == attribute.name) {
      changedOrder = sortOptions[0].order === SORTING_OPTIONS.desc ? SORTING_OPTIONS.asc : SORTING_OPTIONS.desc;
    }
    setSortOptions([{ attribute: attribute.name, scope: attribute.scope, order: changedOrder }]);
  };

  useEffect(() => {
    const columnHeaders = getHeaders(columnSelection, routes.allDevices.defaultHeaders, idAttribute, openSettingsDialog);
    setColumnHeaders(columnHeaders);
    setHeaderKeys(columnHeaders.map(({ attribute: { name, scope } }) => `${name}-${scope}`).join('-'));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [columnSelection, idAttribute.attribute, openSettingsDialog]);

  useEffect(() => {
    if (device.attributes) {
      dispatch(getSystemDevices(device.id, { page, perPage, sortOptions }));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dispatch, device.id, device.attributes?.mender_is_gateway, page, perPage, sortOptions]);

  const onDeviceClick = (device = {}) => navigate(`/devices/${device.status}?id=${device.id}&open=true&tab=identity`);

  return (
    <>
      <DeviceDataCollapse
        title={
          <div className="flexbox center-aligned">
            <h4>Mender Gateway</h4>
            <EnterpriseNotification className="margin-left-small" id={BENEFITS.gateway.id} />
          </div>
        }
      >
        <TwoColumnData config={{ 'Server IP': deviceIp }} compact setSnackbar={message => dispatch(setSnackbar(message))} />
      </DeviceDataCollapse>
      <DeviceDataCollapse className={classes.container} title="System for this gateway">
        {systemDeviceTotal ? (
          <Devicelist
            customColumnSizes={[]}
            columnHeaders={columnHeaders}
            devices={systemDeviceIds.map(id => devicesById[id])}
            deviceListState={{ page, perPage }}
            headerKeys={headerKeys}
            idAttribute={idAttribute}
            onChangeRowsPerPage={setPerPage}
            onExpandClick={onDeviceClick}
            onResizeColumns={false}
            onPageChange={setPage}
            onSelect={false}
            onSort={onSortChange}
            pageLoading={false}
            pageTotal={systemDeviceTotal}
          />
        ) : (
          <div className="dashboard-placeholder">
            <p>No devices have been connected to this gateway device yet.</p>
            <div>
              Visit the <DocsLink path="get-started/mender-gateway" title="full Mender Gateway documentation" /> to learn how to make the most of the gateway
              functionality.
            </div>
          </div>
        )}
      </DeviceDataCollapse>
      <div className="flexbox">
        <Button color="secondary" component={Link} to={`/deployments?deviceId=${device.id}&open=true`}>
          Create deployment for this system
        </Button>
        <Button onClick={onConnectToGatewayClick}>Connect devices</Button>
      </div>
    </>
  );
};

const DeviceSystemTab = ({ device, ...remainder }) => {
  const [open, setOpen] = useState(false);
  const isPreRelease = useSelector(getIsPreview);
  const { tenant_token: tenantToken } = useSelector(getOrganization);
  const { token } = useSelector(getCurrentSession);

  const gatewayIp = getDemoDeviceAddress([device]);
  const toggleDialog = () => setOpen(toggle);
  return (
    <>
      <DeviceSystem onConnectToGatewayClick={toggleDialog} {...{ device, ...remainder }} />
      {open && <ConnectToGatewayDialog gatewayIp={gatewayIp} isPreRelease={isPreRelease} onCancel={toggleDialog} tenantToken={tenantToken} token={token} />}
    </>
  );
};

export default DeviceSystemTab;
