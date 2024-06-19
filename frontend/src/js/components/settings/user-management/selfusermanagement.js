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
import { useDispatch, useSelector } from 'react-redux';

import { Button, Switch, TextField } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { setSnackbar } from '../../../actions/appActions';
import { editUser, saveUserSettings } from '../../../actions/userActions';
import { DARK_MODE, LIGHT_MODE } from '../../../constants/appConstants';
import * as UserConstants from '../../../constants/userConstants';
import { isDarkMode, toggle } from '../../../helpers';
import { getCurrentSession, getCurrentUser, getFeatures, getIsEnterprise, getUserSettings } from '../../../selectors';
import ExpandableAttribute from '../../common/expandable-attribute';
import Form from '../../common/forms/form';
import PasswordInput from '../../common/forms/passwordinput';
import TextInput from '../../common/forms/textinput';
import InfoText from '../../common/infotext';
import AccessTokenManagement from '../accesstokenmanagement';
import { CopyTextToClipboard } from '../organization/organization';
import TwoFactorAuthSetup from './twofactorauthsetup';
import { UserId, getUserSSOState } from './userdefinition';

const useStyles = makeStyles()(() => ({
  formField: { width: 400, maxWidth: '100%' },
  changeButton: { margin: '30px 0 0 15px' },
  infoText: { margin: 0, width: '75%' },
  jwt: { maxWidth: '70%' },
  oauthIcon: { fontSize: '36px', marginRight: 10 },
  userId: { '&.full-width': { maxWidth: '100%' } },
  widthLimit: { maxWidth: 750 }
}));

export const SelfUserManagement = () => {
  const [editEmail, setEditEmail] = useState(false);
  const [editPass, setEditPass] = useState(false);
  const { classes } = useStyles();
  const dispatch = useDispatch();

  const { isHosted } = useSelector(getFeatures);
  const isEnterprise = useSelector(getIsEnterprise);
  const canHave2FA = isEnterprise || isHosted;
  const currentUser = useSelector(getCurrentUser);
  const { isOAuth2, provider } = getUserSSOState(currentUser);
  const { email, id: userId } = currentUser;
  const hasTracking = useSelector(state => !!state.app.trackerCode);
  const { trackingConsentGiven: hasTrackingConsent, mode } = useSelector(getUserSettings);
  const { token } = useSelector(getCurrentSession);

  const editSubmit = userData => {
    if (userData.password != userData.password_confirmation) {
      dispatch(setSnackbar(`The passwords don't match`));
    } else {
      dispatch(editUser(UserConstants.OWN_USER_ID, userData)).then(() => {
        setEditEmail(false);
        setEditPass(false);
      });
    }
  };

  const handleEmail = () => setEditEmail(toggle);

  const toggleMode = () => {
    const newMode = isDarkMode(mode) ? LIGHT_MODE : DARK_MODE;
    dispatch(saveUserSettings({ mode: newMode }));
  };

  const handlePass = () => setEditPass(toggle);

  return (
    <div className={`margin-top-small ${classes.widthLimit}`}>
      <h2 className="margin-top-small">My profile</h2>
      <UserId className={`full-width ${classes.userId}`} userId={userId} />
      {!editEmail && email ? (
        <div className="flexbox space-between margin-bottom-small">
          <TextField className={classes.formField} label="Email" key={email} InputLabelProps={{ shrink: !!email }} disabled defaultValue={email} />
          {!isOAuth2 && (
            <Button className={`inline-block ${classes.changeButton}`} color="primary" id="change_email" onClick={handleEmail}>
              Change email
            </Button>
          )}
        </div>
      ) : (
        <Form defaultValues={{ email }} onSubmit={editSubmit} handleCancel={handleEmail} submitLabel="Save" showButtons={editEmail} buttonColor="secondary">
          <TextInput disabled={false} hint="Email" id="email" InputLabelProps={{ shrink: !!email }} label="Email" validations="isLength:1,isEmail,trim" />
          <PasswordInput id="current_password" label="Current password *" validations={`isLength:8,isNot:${email}`} required={true} />
        </Form>
      )}
      {!isOAuth2 &&
        (!editPass ? (
          <form className="flexbox space-between">
            <TextField className={classes.formField} label="Password" key="password-placeholder" disabled defaultValue="********" type="password" />
            <Button className={classes.changeButton} color="primary" id="change_password" onClick={handlePass}>
              Change password
            </Button>
          </form>
        ) : (
          <>
            <h3 className="margin-top margin-bottom-none">Change password</h3>
            <Form onSubmit={editSubmit} handleCancel={handlePass} submitLabel="Save" buttonColor="secondary" showButtons={editPass}>
              <PasswordInput id="current_password" label="Current password *" validations={`isLength:8,isNot:${email}`} required />
              <PasswordInput className="edit-pass" id="password" label="Password *" validations={`isLength:8,isNot:${email}`} create generate required />
              <PasswordInput id="password_confirmation" label="Confirm password *" validations={`isLength:8,isNot:${email}`} required />
            </Form>
          </>
        ))}
      <div className="clickable flexbox space-between margin-top" onClick={toggleMode}>
        <p className="help-content">Enable dark theme</p>
        <Switch checked={isDarkMode(mode)} />
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
          <div className="clickable flexbox space-between" onClick={() => dispatch(saveUserSettings({ trackingConsentGiven: !hasTrackingConsent }))}>
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
