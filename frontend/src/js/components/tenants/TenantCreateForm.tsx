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
import React, { Dispatch, SetStateAction, useCallback, useEffect, useState } from 'react';
import { useFormContext } from 'react-hook-form';
import { useSelector } from 'react-redux';
import { Link } from 'react-router-dom';

import { ErrorOutline as ErrorOutlineIcon } from '@mui/icons-material';
import { Alert, Divider, Drawer, formControlLabelClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { DrawerTitle } from '@northern.tech/common-ui/DrawerTitle';
import InfoHint from '@northern.tech/common-ui/InfoHint';
import Form from '@northern.tech/common-ui/forms/Form';
import FormCheckbox from '@northern.tech/common-ui/forms/FormCheckbox';
import PasswordInput from '@northern.tech/common-ui/forms/PasswordInput';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import { HELPTOOLTIPS } from '@northern.tech/common-ui/helptips/HelpTooltips';
import { MenderHelpTooltip } from '@northern.tech/common-ui/helptips/MenderTooltip';
import Api from '@northern.tech/store/api/general-api';
import { TIMEOUTS, rolesByName, useradmApiUrlv1 } from '@northern.tech/store/constants';
import { getOrganization, getSsoConfig } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { addTenant, getSsoConfigs } from '@northern.tech/store/thunks';
import { useDebounce } from '@northern.tech/utils/debouncehook';

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
  formWrapper: {
    display: 'flex',
    flexDirection: 'column',
    gap: theme.spacing(2),
    [`.${formControlLabelClasses.root}`]: { marginTop: 0 },
    '.required .relative': { marginLeft: theme.spacing(10) }
  },
  devLimitInput: { marginTop: 10, maxWidth: 150, minWidth: 130 }
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
  const [emailInfoText, setEmailInfoText] = useState<string>('');
  const checkEmailExists = async (email: string) => {
    const response = await Api.get(`${useradmApiUrlv1}/users/exists?email=${encodeURIComponent(email)}`);
    return response.data.exists;
  };

  const { watch, getFieldState, setValue } = useFormContext();

  const enteredEmail = watch('email');
  const debouncedEmail = useDebounce(enteredEmail, TIMEOUTS.debounceDefault);

  useEffect(() => {
    const { invalid: isInvalidEmail } = getFieldState('email');
    if (!debouncedEmail || isInvalidEmail) {
      return;
    }
    const existingEmailCheck = async () => {
      const exists = await checkEmailExists(debouncedEmail);
      if (exists) {
        setAdminExists(true);
        setEmailInfoText(userExistsInfo);
        setValue('password', '');
      } else {
        setAdminExists(false);
        setEmailInfoText(newUserInfo);
      }
    };
    existingEmailCheck();
  }, [debouncedEmail, getFieldState, setAdminExists, setValue]);

  return (
    <>
      <div className="flexbox center-aligned">
        <TextInput validations="isEmail,trim" required id="email" label="Admin user" />
        <MenderHelpTooltip className="required" id={HELPTOOLTIPS.tenantAdmin.id} />
      </div>
      {!adminExists && (
        <>
          <PasswordInput
            className="margin-bottom-small"
            label={<PasswordLabel />}
            id="password"
            InputLabelProps={{ shrink: true }}
            validations={`isLength:8,isNot:${enteredEmail}`}
            placeholder="Password"
            create
            generate
          />
          <FormCheckbox id="send_reset_password" label="Send an email to the user containing a link to reset the password" />
        </>
      )}
      {emailInfoText ? <InfoHint content={emailInfoText} /> : <div />}
    </>
  );
};

const tenantAdminDefaults = { email: '', name: '', password: '', sso: false, binary_delta: false, device_limit: undefined, send_reset_password: false };
export const TenantCreateForm = (props: TenantCreateFormProps) => {
  const { onCloseClick, open } = props;
  const { device_count: spDeviceUtilization = 0, device_limit: spDeviceLimit = 0 } = useSelector(getOrganization);
  const ssoConfig = useSelector(getSsoConfig);
  const dispatch = useAppDispatch();

  const { classes } = useStyles();
  const [adminExists, setAdminExists] = useState<boolean>(false);
  const [hasError, setHasError] = useState<boolean>(false);

  const quota = spDeviceLimit - spDeviceUtilization || 0;
  const numericValidation = {
    min: { value: 1, message: 'The limit must be 1 or more' },
    max: { value: quota, message: `The device limit must be ${quota} or fewer` }
  };

  useEffect(() => {
    dispatch(getSsoConfigs());
  }, [dispatch]);

  const submitNewTenant = useCallback(
    async data => {
      const { email, password, device_limit, send_reset_password, ...remainder } = data;
      let selectionState = { device_limit: Number(device_limit), ...remainder };
      if (adminExists) {
        selectionState = { users: [{ role: rolesByName.admin, email }], ...selectionState };
      } else {
        selectionState = { admin: { password, email, send_reset_password }, ...selectionState };
      }
      try {
        await dispatch(addTenant(selectionState)).unwrap(); // only awaiting the thunk resolution to not get rejected
        onCloseClick();
      } catch {
        setHasError(true);
      }
    },
    [adminExists, dispatch, onCloseClick]
  );

  return (
    <Drawer open={open} onClose={onCloseClick} anchor="right" PaperProps={{ style: { minWidth: '67vw' } }}>
      <DrawerTitle title="Add a tenant" onClose={onCloseClick} />
      <Divider className="margin-bottom-large" />
      <Form
        initialValues={tenantAdminDefaults}
        classes={classes}
        className={classes.formWrapper}
        handleCancel={() => onCloseClick()}
        showButtons
        buttonColor="secondary"
        onSubmit={submitNewTenant}
        submitLabel="Create tenant"
        autocomplete="off"
      >
        {hasError && (
          <Alert icon={<ErrorOutlineIcon />} severity="error">
            There was an error while creating the tenant. Please try again, or contact support.
          </Alert>
        )}
        <TextInput required validations="isLength:3,trim" id="name" hint="Name" label="Name" />
        <UserInputs adminExists={adminExists} setAdminExists={setAdminExists} />
        <div className="flexbox center-aligned">
          <TextInput
            required
            id="device_limit"
            hint={`${quota}`}
            type="number"
            label="Set device limit"
            className={classes.devLimitInput}
            InputProps={{ inputProps: { min: 1, max: quota } }}
            numericValidations={numericValidation}
          />
          <MenderHelpTooltip className="required" id={HELPTOOLTIPS.subTenantDeviceLimit.id} />
        </div>
        <div className="flexbox center-aligned">
          <FormCheckbox id="binary_delta" label="Enable Delta Artifact generation" />
          <MenderHelpTooltip id={HELPTOOLTIPS.subTenantDeltaArtifactGeneration.id} />
        </div>
        {!!ssoConfig && (
          <>
            <div className="flexbox center-aligned">
              <FormCheckbox id="sso" label="Restrict to Service Provider’s Single Sign-On settings" />
              <MenderHelpTooltip className="flexbox center-aligned" id={HELPTOOLTIPS.subTenantSSO.id} />
            </div>
            <div className="margin-top-x-small margin-bottom">
              <Link to="/settings/organization-and-billing">View Single Sign-On settings</Link>
            </div>
          </>
        )}
      </Form>
    </Drawer>
  );
};
