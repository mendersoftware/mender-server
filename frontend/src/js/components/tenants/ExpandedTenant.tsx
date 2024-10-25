// Copyright 2024 Northern.tech AS
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
import React, { useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Link } from 'react-router-dom';

import { Close as CloseIcon, Link as LinkIcon } from '@mui/icons-material';
import { Button, Checkbox, Divider, Drawer, FormControlLabel, FormHelperText, IconButton, TextField } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import actions from '@northern.tech/store/actions';
import { getOrganization } from '@northern.tech/store/organizationSlice/selectors';
import { editTenantDeviceLimit, removeTenant } from '@northern.tech/store/organizationSlice/thunks';
import { AppDispatch } from '@northern.tech/store/store';
import copy from 'copy-to-clipboard';

import { generateTenantPathById } from '../../utils/locationutils';
import { ConfirmModal } from '../common/ConfirmModal/ConfirmModal';
import { TwoColumns } from '../common/configurationobject';
import { DeviceCount } from '../header/devicecount';
import { HELPTOOLTIPS, MenderHelpTooltip } from '../helptips/helptooltips';
import { Tenant } from './types';

interface ExpandedTenantProps {
  onCloseClick: () => void;
  tenant: Tenant;
}

const useStyles = makeStyles()(theme => ({
  devLimitLabel: {
    color: theme.palette.primary.main,
    fontSize: '11px',
    fontWeight: 400
  },
  tenantTitle: {
    fontSize: '17px',
    fontWeight: 700
  },
  devLimitInput: { marginTop: 10, maxWidth: 150, minWidth: 130 },
  tenantInitialAdminTooltip: {
    alignSelf: 'flex-end'
  },
  editLimitButton: {
    height: '42px',
    alignSelf: 'center'
  }
}));
const { setSnackbar } = actions;

export const ExpandedTenant = (props: ExpandedTenantProps) => {
  const { onCloseClick, tenant } = props;
  const { name, id, device_limit, device_count, binary_delta } = tenant;

  const [shouldDelete, setShouldDelete] = useState<boolean>(false);
  const [newLimitForm, setNewLimitForm] = useState<boolean>(false);
  const [newLimit, setNewLimit] = useState<number>(device_limit);
  const [limitErrorText, setLimitErrorText] = useState<string>('');

  const { device_count: spDeviceUtilization, device_limit: spDeviceLimit } = useSelector(getOrganization);
  const { classes } = useStyles();
  const dispatch = useDispatch<AppDispatch>();
  const copyLinkToClipboard = () => {
    const location = window.origin + '/ui';
    copy(`${location}${generateTenantPathById(id)}`);
    dispatch(setSnackbar('Link copied to clipboard'));
  };

  const onChangeLimit = ({ target: { validity, value } }) => {
    if (validity.valid) {
      setNewLimit(value);
      return setLimitErrorText('');
    }
    setLimitErrorText('Device limit');
  };
  const onNewLimitSubmit = async () => {
    await dispatch(editTenantDeviceLimit({ id, name, newLimit: Number(newLimit) }));
    setNewLimitForm(false);
  };
  const deleteTenant = () => {
    dispatch(removeTenant({ id }));
  };

  return (
    <Drawer onClose={onCloseClick} open={true} PaperProps={{ style: { minWidth: '67vw' } }} anchor="right">
      <div className="flexbox center-aligned space-between">
        <div className="flexbox center-aligned">
          <h3 className={classes.tenantTitle}>Tenant Information for {name}</h3>
          <IconButton onClick={copyLinkToClipboard} size="large">
            <LinkIcon />
          </IconButton>
        </div>
        <div className="flexbox center-aligned">
          <Button onClick={() => setShouldDelete(true)}>Delete tenant</Button>
          <IconButton onClick={onCloseClick} aria-label="close" size="large">
            <CloseIcon />
          </IconButton>
        </div>
      </div>
      <Divider className="margin-bottom" />
      <div className="margin-top">
        <div className="flexbox">
          <TwoColumns
            setSnackbar={(str: string) => dispatch(setSnackbar(str))}
            items={{
              name,
              ID: id,
              'Initial admin user': 'customer@email.com'
            }}
          />
          <div className={classes.tenantInitialAdminTooltip}>
            <MenderHelpTooltip id={HELPTOOLTIPS.tenantInitialAdmin.id} />
          </div>
        </div>
        <div className="flexbox column">
          <FormControlLabel
            control={<Checkbox color="primary" size="small" disabled={true} checked={binary_delta} />}
            label={'Enable Delta Artifact generation'}
          ></FormControlLabel>
          <FormControlLabel
            style={{ marginTop: 10 }}
            control={<Checkbox color="primary" size="small" checked={true} disabled={true} />}
            label={'Restrict to Service Provider’s Single Sign-On settings'}
          ></FormControlLabel>
          <div className="margin-top-x-small margin-bottom">
            <Link to="/settings/organization-and-billing">View Single Sign-On settings</Link>
          </div>
        </div>
        <div className="flexbox">
          <div>
            <DeviceCount current={device_count} max={device_limit} variant="detailed"></DeviceCount>
          </div>
          {!newLimitForm && (
            <Button className={`margin-left ${classes.editLimitButton}`} onClick={() => setNewLimitForm(true)}>
              Edit device limit
            </Button>
          )}
          <ConfirmModal
            header="Are you sure you want to delete this tenant?"
            description="All devices, users, artifacts and audit logs associated with the tenant will be removed."
            toType="delete"
            open={shouldDelete}
            close={() => setShouldDelete(false)}
            onConfirm={deleteTenant}
          />
          {newLimitForm && (
            <div className="margin-left">
              <div className={classes.devLimitLabel}>Set device limit</div>
              <div className="flexbox">
                <TextField
                  className={classes.devLimitInput}
                  type="number"
                  onChange={onChangeLimit}
                  inputProps={{ min: device_count, max: '10000' }}
                  error={!!limitErrorText}
                  value={newLimit}
                />
                <div className="margin-left">
                  <Button
                    onClick={() => {
                      setNewLimit(device_limit);
                      setNewLimitForm(false);
                    }}
                  >
                    Cancel
                  </Button>{' '}
                  <Button onClick={onNewLimitSubmit} color="primary" variant="contained">
                    Save
                  </Button>
                </div>
              </div>

              {!!limitErrorText && (
                <FormHelperText className="warning" component="div">
                  {limitErrorText}
                </FormHelperText>
              )}
              <FormHelperText className="info" component="div">
                Maximum limit: {spDeviceLimit - spDeviceUtilization} <br />
                {spDeviceUtilization} devices assigned of maximum {spDeviceLimit} across all tenants. <br />
                <a href="mailto:support@mender.io" target="_blank" rel="noopener noreferrer">
                  Contact support
                </a>{' '}
                to increase your total limit
              </FormHelperText>
            </div>
          )}
        </div>
      </div>
    </Drawer>
  );
};
