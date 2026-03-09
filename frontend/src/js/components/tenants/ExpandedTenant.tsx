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
import { useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Link } from 'react-router-dom';

import { Delete as DeleteIcon } from '@mui/icons-material';
import { Button, Checkbox, Divider, Drawer, FormControlLabel, Typography, formControlLabelClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { ConfirmModal } from '@northern.tech/common-ui/ConfirmModal';
import { DrawerTitle } from '@northern.tech/common-ui/DrawerTitle';
import EditableNameInput from '@northern.tech/common-ui/EditableNameInput';
import { TwoColumnData } from '@northern.tech/common-ui/TwoColumnData';
import Form from '@northern.tech/common-ui/forms/Form';
import actions from '@northern.tech/store/actions';
import { generateTenantPathById } from '@northern.tech/store/locationutils';
import { getSpLimits, getSsoConfig } from '@northern.tech/store/selectors';
import { AppDispatch } from '@northern.tech/store/store';
import { editTenant, removeTenant } from '@northern.tech/store/thunks';
import copy from 'copy-to-clipboard';

import { DeviceLimit } from '../header/DeviceNotifications';
import { DeviceLimitsInput } from './TenantCreateForm';
import { Tenant } from './types';

interface ExpandedTenantProps {
  onCloseClick: () => void;
  tenant: Tenant;
}

const useStyles = makeStyles()(theme => ({
  buttonWrapper: {
    '&.button-wrapper': {
      justifyContent: 'start',
      marginTop: 0
    }
  },
  formWrapper: { display: 'flex', flexDirection: 'column', gap: theme.spacing(4), maxWidth: 750, [`.${formControlLabelClasses.root}`]: { marginTop: 0 } },
  ssoLink: {
    marginLeft: `calc(1em + ${theme.spacing(1.5)})`, // 1em as the width of the checkbox + the padding around the checkbox
    marginTop: theme.spacing(-1)
  }
}));
const { setSnackbar } = actions;

export const ExpandedTenant = (props: ExpandedTenantProps) => {
  const { onCloseClick, tenant } = props;
  const spLimits = useSelector(getSpLimits);
  const { name, id, created_at, status, device_limits } = tenant;
  const [shouldDelete, setShouldDelete] = useState<boolean>(false);
  const [limitEdit, setLimitEdit] = useState<boolean>(false);
  const currentDeviceLimits = Object.fromEntries(Object.values(device_limits).map(limit => [limit.id, limit.limit]));
  const ssoConfig = useSelector(getSsoConfig);

  const { classes } = useStyles();
  const dispatch = useDispatch<AppDispatch>();

  const copyLinkToClipboard = () => {
    const location = window.origin + '/ui';
    copy(`${location}${generateTenantPathById(id)}`);
    dispatch(setSnackbar('Link copied to clipboard'));
  };

  const onNewLimitSubmit = async newLimits => {
    await dispatch(editTenant({ id, name, deviceLimits: newLimits }));
    setLimitEdit(false);
  };

  const onChangeName = async (newName: string) => {
    //TODO: endpoint is resetting some limits if you don't send the limits. Refactor if it is ever fixed
    await dispatch(editTenant({ id, name: newName, deviceLimits: currentDeviceLimits }));
  };
  const deleteTenant = () => dispatch(removeTenant({ id }));
  const twoColumnData = {
    name: <EditableNameInput id="tenant-name" isHovered name={name} placeholder={name} onSave={onChangeName} />,
    ID: id,
    'created at': created_at,
    status: status
  };

  return (
    <Drawer onClose={onCloseClick} open={true} PaperProps={{ style: { minWidth: '67vw' } }} anchor="right">
      <DrawerTitle
        title={`Tenant Information for ${name}`}
        onLinkCopy={copyLinkToClipboard}
        preCloser={
          <Button
            variant="outlined"
            className="margin-right-small"
            color="error"
            startIcon={<DeleteIcon className="red auth" />}
            onClick={() => setShouldDelete(true)}
          >
            Delete tenant
          </Button>
        }
        onClose={onCloseClick}
      />
      <Divider className="margin-bottom-large" />
      <div className={classes.formWrapper}>
        <TwoColumnData setSnackbar={(str: string) => dispatch(setSnackbar(str))} data={twoColumnData} />
        {!!ssoConfig && (
          <>
            <FormControlLabel
              control={<Checkbox color="primary" size="small" checked disabled />}
              label="Restrict to Service Provider’s Single Sign-On settings"
            />
            <Link className={classes.ssoLink} to="/settings/organization">
              View Single Sign-On settings
            </Link>
          </>
        )}
        <div>
          <Typography className="margin-bottom-x-small" variant="subtitle1">
            Device limits
          </Typography>
          {limitEdit ? (
            <Form
              initialValues={currentDeviceLimits}
              classes={classes}
              className={classes.formWrapper}
              handleCancel={() => setLimitEdit(false)}
              showButtons
              buttonColor="primary"
              onSubmit={onNewLimitSubmit}
              validationMode="onSubmit"
              submitLabel="Save changes"
              autocomplete="off"
            >
              <DeviceLimitsInput spLimits={spLimits} isEdit currentLimits={currentDeviceLimits} />
            </Form>
          ) : (
            <>
              {Object.values(device_limits)
                .filter(limit => !!spLimits[limit.id])
                .map(limit => (
                  <DeviceLimit
                    key={limit.id}
                    className="margin-bottom-x-small"
                    type={limit.id}
                    total={limit.current}
                    limit={limit.limit}
                    disabled={limit.current === limit.limit && limit.limit === 0}
                  />
                ))}
              <Button className="margin-top-x-small" variant="text" onClick={() => setLimitEdit(true)}>
                Manage device limits
              </Button>
            </>
          )}
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
