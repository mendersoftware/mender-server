// Copyright 2017 Northern.tech AS
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
import { useSelector } from 'react-redux';

import { Button, Switch, TextField } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { CopyTextToClipboard } from '@northern.tech/common-ui/CopyText';
import ExpandableAttribute from '@northern.tech/common-ui/ExpandableAttribute';
import InfoText from '@northern.tech/common-ui/InfoText';
import Form from '@northern.tech/common-ui/forms/Form';
import PasswordInput from '@northern.tech/common-ui/forms/PasswordInput';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import storeActions from '@northern.tech/store/actions';
import { DARK_MODE, LIGHT_MODE, OWN_USER_ID } from '@northern.tech/store/constants';
import { getCurrentSession, getCurrentUser, getFeatures, getIsDarkMode, getIsEnterprise, getUserSettings } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { editUser, saveUserSettings } from '@northern.tech/store/thunks';
import { toggle } from '@northern.tech/utils/helpers';

import AccessTokenManagement from '../AccessTokenManagement';
import TwoFactorAuthSetup from './TwoFactorAuthSetup';
import { UserId, getUserSSOState } from './UserDefinition';

const { setSnackbar } = storeActions;

const useStyles = makeStyles()(() => ({
  formField: { width: 400, maxWidth: '100%' },
  infoText: { margin: 0, width: '75%' },
  jwt: { maxWidth: '70%' },
  oauthIcon: { fontSize: '36px', marginRight: 10 },
  widthLimit: { maxWidth: 750 }
}));

export const SelfUserManagement = () => {
  const [editEmail, setEditEmail] = useState(false);
  const [editPass, setEditPass] = useState(false);
  const { classes } = useStyles();
  const dispatch = useAppDispatch();

  const { isHosted } = useSelector(getFeatures);
  const isEnterprise = useSelector(getIsEnterprise);
  const canHave2FA = isEnterprise || isHosted;
  const currentUser = useSelector(getCurrentUser);
  const { isOAuth2, provider } = getUserSSOState(currentUser);
  const { email, id: userId } = currentUser;
  const hasTracking = useSelector(state => !!state.app.trackerCode);
  const { trackingConsentGiven: hasTrackingConsent } = useSelector(getUserSettings);
  const isDarkMode = useSelector(getIsDarkMode);
  const { token } = useSelector(getCurrentSession);

  const editSubmit = userData => {
    if (userData.password != userData.password_confirmation) {
      dispatch(setSnackbar(`The passwords don't match`));
    } else {
      dispatch(editUser({ ...userData, id: OWN_USER_ID }))
        .unwrap()
        .then(() => {
          setEditEmail(false);
          setEditPass(false);
        });
    }
  };

  const handleEmail = () => setEditEmail(toggle);

  const toggleMode = () => {
    const newMode = isDarkMode ? LIGHT_MODE : DARK_MODE;
    dispatch(saveUserSettings({ mode: newMode }));
  };

  const handlePass = () => setEditPass(toggle);

  return (
    <div className={`margin-top-small ${classes.widthLimit}`}>
      <h2 className="margin-top-small">My profile</h2>
      <UserId className="margin-bottom-none" userId={userId} />
      {!editEmail && email ? (
        <div className="flexbox space-between margin-bottom-small">
          <TextField className={classes.formField} label="Email" key={email} disabled defaultValue={email} />
          {!isOAuth2 && (
            <Button color="primary" id="change_email" onClick={handleEmail}>
              Change email
            </Button>
          )}
        </div>
      ) : (
        <Form defaultValues={{ email }} onSubmit={editSubmit} handleCancel={handleEmail} submitLabel="Save" showButtons={editEmail} buttonColor="secondary">
          <TextInput hint="Email" id="email" label="Email" validations="isLength:1,isEmail,trim" />
          <PasswordInput
            className="margin-top-x-small"
            id="current_password"
            label="Current password *"
            validations={`isLength:8,isNot:${email}`}
            required={true}
          />
        </Form>
      )}
      {!isOAuth2 &&
        (!editPass ? (
          <form className="flexbox space-between margin-top">
            <TextField className={classes.formField} label="Password" key="password-placeholder" disabled defaultValue="********" type="password" />
            <Button color="primary" id="change_password" onClick={handlePass}>
              Change password
            </Button>
          </form>
        ) : (
          <>
            <h3 className="margin-top">Change password</h3>
            <Form onSubmit={editSubmit} handleCancel={handlePass} submitLabel="Save" buttonColor="secondary" showButtons={editPass}>
              <PasswordInput
                className="margin-bottom-x-small"
                id="current_password"
                label="Current password *"
                validations={`isLength:8,isNot:${email}`}
                required
              />
              <PasswordInput id="password" label="Password *" validations={`isLength:8,isNot:${email}`} create generate required />
              <PasswordInput
                className="margin-top-x-small"
                id="password_confirmation"
                label="Confirm password *"
                validations={`isLength:8,isNot:${email}`}
                required
              />
            </Form>
          </>
        ))}
      <div className="clickable flexbox space-between margin-top" onClick={toggleMode}>
        <p className="help-content">Enable dark theme</p>
        <Switch checked={isDarkMode} />
      </div>
      {!isOAuth2 ? (
        canHave2FA && <TwoFactorAuthSetup />
      ) : (
        <div className="flexbox margin-top">
          <div className={classes.oauthIcon}>{provider.icon}</div>
          <div className="info">
            You are logging in using your <strong>{provider.name}</strong> account.
            <br />
            Please connect to {provider.name} to update your login settings.
          </div>
        </div>
      )}
      <div className="flexbox space-between margin-top-large">
        <div className={classes.jwt}>
          <div className="help-content">Session token</div>
          <ExpandableAttribute
            component="div"
            disableGutters
            dividerDisabled
            secondary={token}
            textClasses={{ secondary: 'inventory-text tenant-token-text' }}
          />
        </div>
        <div className="flexbox center-aligned">
          <CopyTextToClipboard token={token} />
        </div>
      </div>
      <AccessTokenManagement />
      {isEnterprise && hasTracking && (
        <div className="margin-top">
          <div className="clickable flexbox space-between" onClick={() => dispatch(saveUserSettings({ trackingConsentGiven: !hasTrackingConsent })).unwrap()}>
            <p className="help-content">Help us improve Mender</p>
            <Switch checked={!!hasTrackingConsent} />
          </div>
          <InfoText className={classes.infoText}>Enable usage data and errors to be sent to help us improve our service.</InfoText>
        </div>
      )}
    </div>
  );
};

export default SelfUserManagement;
