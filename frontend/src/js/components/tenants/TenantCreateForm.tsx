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
import React, { Dispatch, SetStateAction, useEffect, useState } from 'react';
import { useFormContext } from 'react-hook-form';
import { useDispatch, useSelector } from 'react-redux';
import { Link } from 'react-router-dom';

import { Divider, Drawer } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { DrawerTitle } from '@northern.tech/common-ui/DrawerTitle';
import InfoHint from '@northern.tech/common-ui/InfoHint';
import Form from '@northern.tech/common-ui/forms/Form';
import FormCheckbox from '@northern.tech/common-ui/forms/FormCheckbox';
import PasswordInput from '@northern.tech/common-ui/forms/PasswordInput';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import { HELPTOOLTIPS, MenderHelpTooltip } from '@northern.tech/helptips/HelpTooltips';
import Api from '@northern.tech/store/api/general-api';
import { rolesByName, useradmApiUrlv1 } from '@northern.tech/store/constants';
import { getOrganization } from '@northern.tech/store/selectors';
import { AppDispatch } from '@northern.tech/store/store';
import { addTenant } from '@northern.tech/store/thunks';

import { PasswordLabel } from '../settings/user-management/UserForm';

interface TenantCreateFormProps {
  onCloseClick: () => void;
  open: boolean;
}

const useStyles = makeStyles()(theme => ({
  buttonWrapper: {
    '&.button-wrapper': {
      justifyContent: 'start'
    }
  },
  userInputContainer: { maxWidth: 500 },
  devLimitInput: { marginTop: 10, maxWidth: 150, minWidth: 130 },
  helpTooltip: {
    marginLeft: theme.spacing(9),
    alignSelf: 'flex-end'
  }
}));

interface UserInputsProps {
  adminExists: boolean;
  setAdminExists: Dispatch<SetStateAction<boolean>>;
}

const userExistsInfo =
  'This user already has a Mender account, and will be assigned as admin to the new tenant. If you want to create a brand new user, try a different email address.';
const newUserInfo = 'This will create a new user as admin of the new tenant.';

const UserInputs = (props: UserInputsProps) => {
  const { setAdminExists, adminExists } = props;
  const { classes } = useStyles();
  const [emailInfoText, setEmailInfoText] = useState<string>('');
  const checkEmailExists = async (email: string) => {
    const response = await Api.get(`${useradmApiUrlv1}/users/exists?email=${email}`);
    return response.data.exists;
  };

  const { watch, getFieldState } = useFormContext();

  const enteredEmail = watch('email');
  const isValidEmail = getFieldState('email');

  useEffect(() => {
    if (!(enteredEmail && isValidEmail)) {
      return;
    }
    const timeoutId = setTimeout(async () => {
      const exists = await checkEmailExists(enteredEmail);
      if (exists) {
        setAdminExists(true);
        setEmailInfoText(userExistsInfo);
      } else {
        setAdminExists(false);
        setEmailInfoText(newUserInfo);
      }
    }, 1000);
    return () => clearTimeout(timeoutId);
  }, [enteredEmail, isValidEmail, setAdminExists]);

  useEffect(() => {
    setAdminExists(false);
  }, [enteredEmail, setAdminExists]);

  return (
    <div className={classes.userInputContainer}>
      <div className="flexbox margin-bottom-small">
        <TextInput validations="isEmail,trim" required id="email" label="Admin user" />
        <div className={classes.helpTooltip}>
          <MenderHelpTooltip id={HELPTOOLTIPS.tenantAdmin.id} />
        </div>
      </div>
      {!adminExists && (
        <>
          <PasswordInput
            label={<PasswordLabel />}
            id="password"
            validations={`isLength:8,isNot:${enteredEmail}`}
            InputLabelProps={{ shrink: true }}
            edit={false}
            placeholder="Password"
            create
            generate
            className="margin-bottom-small"
          />
          <FormCheckbox className="margin-top-none" id="send_reset_password" label="Send an email to the user containing a link to reset the password" />
        </>
      )}
      {emailInfoText ? <InfoHint content={emailInfoText} /> : <div />}
    </div>
  );
};

const tenantAdminDefaults = { email: '', name: '', password: '', sso: false, binary_delta: false, device_limit: 0, send_reset_password: false };
export const TenantCreateForm = (props: TenantCreateFormProps) => {
  const { onCloseClick, open } = props;
  const { device_count: spDeviceUtilization, device_limit: spDeviceLimit } = useSelector(getOrganization);
  const dispatch = useDispatch<AppDispatch>();

  const { classes } = useStyles();
  const [adminExists, setAdminExists] = useState<boolean>(false);

  const quota = spDeviceLimit - spDeviceUtilization;
  const numericValidation = {
    min: { value: 1, message: `Device limit can't be less then 0` },
    max: { value: quota, message: `The device limit must be ${quota} or fewer` }
  };

  const submitNewTenant = async data => {
    const { email, password, device_limit, send_reset_password, ...remainder } = data;
    if (adminExists) {
      await dispatch(addTenant({ users: [{ role: rolesByName.admin, email }], device_limit: Number(device_limit), ...remainder }));
    } else {
      await dispatch(addTenant({ admin: { password, email, send_reset_password }, device_limit: Number(device_limit), ...remainder }));
    }
    onCloseClick();
  };

  return (
    <Drawer open={open} onClose={onCloseClick} anchor="right" PaperProps={{ style: { minWidth: '67vw' } }}>
      <DrawerTitle title="Add a tenant" onClose={onCloseClick} />
      <Divider className="margin-bottom" />
      <Form
        initialValues={tenantAdminDefaults}
        classes={classes}
        handleCancel={() => onCloseClick()}
        showButtons
        buttonColor="secondary"
        onSubmit={submitNewTenant}
        submitLabel="Create tenant"
        autocomplete="off"
      >
        <div className="flexbox column">
          <TextInput required validations="isLength:3,trim" id="name" hint="Name" label="Name" className="margin-bottom-large margin-top-large" />
          <UserInputs adminExists={adminExists} setAdminExists={setAdminExists} />
          <div className="flexbox margin-top-large margin-bottom-large">
            <TextInput
              required
              id="device_limit"
              hint="1000"
              type="number"
              label="Set device limit"
              className={classes.devLimitInput}
              numericValidations={numericValidation}
            />
            <div className={classes.helpTooltip}>
              <MenderHelpTooltip id={HELPTOOLTIPS.subTenantDeviceLimit.id} />
            </div>
          </div>
          <div className="flexbox">
            <FormCheckbox id="binary_delta" label="Enable Delta Artifact generation" />
            <div className={classes.helpTooltip}>
              <MenderHelpTooltip id={HELPTOOLTIPS.subTenantDeltaArtifactGeneration.id} />
            </div>
          </div>
          <div className="flexbox">
            <FormCheckbox id="sso" label="Restrict to Service Provider’s Single Sign-On settings" />
            <div className={classes.helpTooltip}>
              <MenderHelpTooltip id={HELPTOOLTIPS.subTenantSSO.id} />
            </div>
          </div>

          <div className="margin-top-x-small margin-bottom">
            <Link to="/settings/organization-and-billing">View Single Sign-On settings</Link>
          </div>
        </div>
      </Form>
    </Drawer>
  );
};
