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
import { useDispatch, useSelector } from 'react-redux';
import { Link } from 'react-router-dom';

import { Close as CloseIcon } from '@mui/icons-material';
import { Divider, Drawer, IconButton, TextField } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import Api from '@northern.tech/store/api/general-api';
import { useradmApiUrlv1 } from '@northern.tech/store/constants';
import { getOrganization } from '@northern.tech/store/selectors';
import { AppDispatch } from '@northern.tech/store/store';
import { addTenant } from '@northern.tech/store/thunks';
import validator from 'validator';

import Form from '../common/forms/form';
import FormCheckbox from '../common/forms/formcheckbox';
import PasswordInput from '../common/forms/passwordinput';
import TextInput from '../common/forms/textinput';
import InfoHint from '../common/info-hint';
import { HELPTOOLTIPS, MenderHelpTooltip } from '../helptips/helptooltips';
import { PasswordLabel } from '../settings/user-management/userform';

interface TenantCreateFormProps {
  onCloseClick: () => void;
  open: boolean;
}

const useStyles = makeStyles()(theme => ({
  tenantTitle: {
    fontSize: '17px',
    fontWeight: 700
  },
  buttonWrapper: {
    justifyContent: 'flex-start !important'
  },
  devLimitInput: { marginTop: 10, maxWidth: 150, minWidth: 130 },
  infoCard: {
    maxWidth: '500px'
  },
  helpTooltip: {
    marginLeft: theme.spacing(9),
    alignSelf: 'flex-end'
  }
}));
interface UserInputsProps {
  email: string;
  setEmail: Dispatch<SetStateAction<string>>;
}

const UserInputs = (props: UserInputsProps) => {
  const { email, setEmail } = props;
  const { classes } = useStyles();
  const [emailErrorText, setEmailErrorText] = useState<string>('');
  const [emailInfoText, setEmailInfoText] = useState<string>('');
  const checkEmailExists = async (email: string) => {
    const response = await Api.get(`${useradmApiUrlv1}/users/exists?email=${email}`);
    return response.data.exists;
  };

  useEffect(() => {
    if (validator.isEmail(email)) {
      setEmailErrorText('');
      const timeoutId = setTimeout(async () => {
        const exists = await checkEmailExists(email);
        if (exists) {
          setEmailInfoText(
            'This user already has a Mender account, and will be assigned as admin to the new tenant. If you want to create a brand new user, try a different email address.'
          );
        } else {
          setEmailInfoText('');
        }
      }, 1000);
      return () => clearTimeout(timeoutId);
    } else if (email) {
      setEmailErrorText('Please enter a valid email address');
      setEmailInfoText('');
    } else {
      setEmailInfoText('');
      setEmailErrorText('');
    }
  }, [email]);
  return (
    <>
      <div className="flexbox margin-bottom-small">
        <TextField
          style={{ width: 400 }}
          classes={{ root: 'required' }}
          onChange={e => setEmail(e.target.value)}
          required
          id="email"
          error={!!emailErrorText}
          helperText={emailErrorText}
          label="Admin user"
        />{' '}
        <div className={classes.helpTooltip}>
          <MenderHelpTooltip id={HELPTOOLTIPS.tenantAdmin.id} />
        </div>
      </div>
      {emailInfoText ? (
        <InfoHint className={classes.infoCard} content={emailInfoText} />
      ) : (
        <PasswordInput
          label={<PasswordLabel />}
          id="password"
          validations={`isLength:8,isNot:${email}`}
          InputLabelProps={{ shrink: true }}
          edit={false}
          placeholder="Password"
          create
          generate
          className="margin-bottom-small"
        />
      )}
    </>
  );
};

export const TenantCreateForm = (props: TenantCreateFormProps) => {
  const { onCloseClick, open } = props;
  const { device_count: spDeviceUtilization, device_limit: spDeviceLimit } = useSelector(getOrganization);
  const dispatch = useDispatch<AppDispatch>();

  const { classes } = useStyles();
  const [email, setEmail] = useState<string>('');

  const quota = spDeviceLimit - spDeviceUtilization;
  const numericValidation = {
    min: { value: 0, message: `Device limit can't be less then 0` },
    max: { value: quota, message: `Exceeds quota (${quota})` }
  };

  const submitNewTenant = async data => {
    const { name, password, sso, binary_delta, device_limit } = data;
    await dispatch(addTenant({ name, admin: { password, email }, sso, device_limit: Number(device_limit), binary_delta }));
    onCloseClick();
  };
  return (
    <Drawer open={open} onClose={onCloseClick} anchor="right" PaperProps={{ style: { minWidth: '67vw' } }}>
      <div className="flexbox center-aligned space-between">
        <div className="flexbox center-aligned">
          <h3 className={`${classes.tenantTitle}`}>Add a tenant</h3>
        </div>
        <div className="flexbox center-aligned">
          <IconButton onClick={onCloseClick} aria-label="close" size="large">
            <CloseIcon />
          </IconButton>
        </div>
      </div>
      <Divider className="margin-bottom" />
      <Form
        initialValues={{ name: '', password: '', sso: false, binary_delta: false, device_limit: 0 }}
        classes={classes}
        handleCancel
        showButtons
        buttonColor="secondary"
        onSubmit={submitNewTenant}
        submitLabel="Create tenant"
      >
        <div className="flexbox column">
          <TextInput required validations="isLength:3,trim" id="name" hint="Name" label="Tenant name" className="margin-bottom-large margin-top-large" />
          <UserInputs email={email} setEmail={setEmail} />
          <div className="flexbox">
            <TextInput
              required
              id="device_limit"
              hint="1000"
              label="Set device limit"
              className={`${classes.devLimitInput}`}
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
