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

import { Button, Checkbox, Divider, Drawer, FormControl, FormControlLabel, FormHelperText, TextField, formControlLabelClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { TwoColumns } from '@northern.tech/common-ui/ConfigurationObject';
import { ConfirmModal } from '@northern.tech/common-ui/ConfirmModal';
import { DrawerTitle } from '@northern.tech/common-ui/DrawerTitle';
import actions from '@northern.tech/store/actions';
import { getOrganization, getSsoConfig } from '@northern.tech/store/organizationSlice/selectors';
import { editTenantDeviceLimit, removeTenant } from '@northern.tech/store/organizationSlice/thunks';
import { AppDispatch } from '@northern.tech/store/store';
import { generateTenantPathById } from '@northern.tech/utils/locationutils';
import copy from 'copy-to-clipboard';

import { DeviceCount } from '../header/devicecount';
import { Tenant } from './types';

interface ExpandedTenantProps {
  onCloseClick: () => void;
  tenant: Tenant;
}

const useStyles = makeStyles()(theme => ({
  devLimitInput: { minWidth: 150 },
  formWrapper: { display: 'flex', flexDirection: 'column', gap: theme.spacing(2), maxWidth: 750, [`.${formControlLabelClasses.root}`]: { marginTop: 0 } },
  ssoLink: {
    marginLeft: `calc(1em + ${theme.spacing(1.5)})`, // 1em as the width of the checkbox + the padding around the checkbox
    marginTop: theme.spacing(-1)
  }
}));
const { setSnackbar } = actions;

export const ExpandedTenant = (props: ExpandedTenantProps) => {
  const { onCloseClick, tenant } = props;
  const { name, id, device_limit, device_count, binary_delta } = tenant;

  const [shouldDelete, setShouldDelete] = useState<boolean>(false);
  const [newLimitForm, setNewLimitForm] = useState<boolean>(false);
  const [newLimit, setNewLimit] = useState<number>(device_limit);
  const [hasLimitError, setHasLimitError] = useState<boolean>(false);

  const { device_count: spDeviceUtilization, device_limit: spDeviceLimit } = useSelector(getOrganization);
  const ssoConfig = useSelector(getSsoConfig);

  const currentLimit = spDeviceLimit - spDeviceUtilization + device_limit;
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
      return setHasLimitError(false);
    }
    setHasLimitError(true);
  };

  const onNewLimitSubmit = async () => {
    await dispatch(editTenantDeviceLimit({ id, name, newLimit: Number(newLimit) }));
    setNewLimitForm(false);
  };

  const deleteTenant = () => dispatch(removeTenant({ id }));

  return (
    <Drawer onClose={onCloseClick} open={true} PaperProps={{ style: { minWidth: '67vw' } }} anchor="right">
      <DrawerTitle
        title={`Tenant Information for ${name}`}
        onLinkCopy={copyLinkToClipboard}
        preCloser={<Button onClick={() => setShouldDelete(true)}>Delete tenant</Button>}
        onClose={onCloseClick}
      />
      <Divider className="margin-bottom-large" />
      <div className={classes.formWrapper}>
        <TwoColumns className="align-self-start" setSnackbar={(str: string) => dispatch(setSnackbar(str))} items={{ name, ID: id }} />
        <FormControlLabel control={<Checkbox color="primary" size="small" disabled checked={binary_delta} />} label="Enable Delta Artifact generation" />
        {!!ssoConfig && (
          <>
            <FormControlLabel
              control={<Checkbox color="primary" size="small" checked disabled />}
              label="Restrict to Service Provider’s Single Sign-On settings"
            />
            <Link className={classes.ssoLink} to="/settings/organization-and-billing">
              View Single Sign-On settings
            </Link>
          </>
        )}
        <div className={`flexbox ${newLimitForm ? '' : 'center-aligned'} margin-top-small`}>
          <DeviceCount current={device_count} max={device_limit} variant="detailed" />
          <div className="margin-left">
            {newLimitForm ? (
              <FormControl className={classes.formWrapper}>
                <div className="flexbox center-aligned">
                  <TextField
                    className={classes.devLimitInput}
                    label="Set device limit"
                    type="number"
                    onChange={onChangeLimit}
                    slotProps={{ htmlInput: { min: device_count, max: currentLimit, 'data-testid': 'dev-limit-input' } }}
                    error={hasLimitError}
                    value={newLimit}
                  />
                  <Button
                    className="margin-left"
                    onClick={() => {
                      setNewLimit(device_limit);
                      setNewLimitForm(false);
                    }}
                  >
                    Cancel
                  </Button>
                  <Button className="margin-left-x-small" onClick={onNewLimitSubmit} color="primary" variant="contained">
                    Save
                  </Button>
                </div>
                <FormHelperText className={`${hasLimitError ? 'warning' : 'info'} margin-top-none`}>Maximum limit: {currentLimit}</FormHelperText>
                <FormHelperText className="info margin-top-none">
                  {spDeviceUtilization} devices assigned of maximum {spDeviceLimit} across all tenants.
                  <br />
                  <a href="mailto:support@mender.io" target="_blank" rel="noopener noreferrer">
                    Contact support
                  </a>{' '}
                  to increase your total limit
                </FormHelperText>
              </FormControl>
            ) : (
              <Button onClick={() => setNewLimitForm(true)}>Edit device limit</Button>
            )}
          </div>
        </div>
        <ConfirmModal
          header="Are you sure you want to delete this tenant?"
          description="All devices, users, artifacts and audit logs associated with the tenant will be removed."
          toType="delete"
          open={shouldDelete}
          close={() => setShouldDelete(false)}
          onConfirm={() => {
            deleteTenant();
            setShouldDelete(false);
            onCloseClick();
          }}
        />
      </div>
    </Drawer>
  );
};
