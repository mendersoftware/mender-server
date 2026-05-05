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
import { useState } from 'react';
import { useSelector } from 'react-redux';

import { Alert, Button, Chip, DialogActions, DialogContent, TextField, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { CopyTextToClipboard } from '@northern.tech/common-ui/CopyText';
import ExpandableAttribute from '@northern.tech/common-ui/ExpandableAttribute';
import { ToggleSetting } from '@northern.tech/common-ui/ToggleSetting';
import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import Form from '@northern.tech/common-ui/forms/Form';
import PasswordInput from '@northern.tech/common-ui/forms/PasswordInput';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import storeActions from '@northern.tech/store/actions';
import { DARK_MODE, LIGHT_MODE, OWN_USER_ID } from '@northern.tech/store/constants';
import { getCurrentSession, getCurrentUser, getFeatures, getIsDarkMode, getIsEnterprise, getUserSettings } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { editUser, passwordResetStart, saveUserSettings, verifyEmailStart } from '@northern.tech/store/thunks';
import { toggle } from '@northern.tech/utils/helpers';

import AccessTokenManagement from '../AccessTokenManagement';
import TwoFactorAuthSetup from './TwoFactorAuthSetup';
import { UserId, getUserSSOState } from './UserDefinition';
import { EmailVerificationConfirmation } from './twofactorauth-steps/EmailVerification';

const { setSnackbar } = storeActions;

const useStyles = makeStyles()(theme => ({
  formField: { width: 400, maxWidth: '100%' },
  jwt: { maxWidth: '70%' },
  widthLimit: { maxWidth: 777 },
  sessionTokenSection: { marginTop: theme.spacing(6) }
}));
export const notificationMap = {
  email: 'Email successfully verified.',
  enabled2fa: 'Two-factor authentication successfully enabled.'
};

export const SelfUserManagement = () => {
  const [editEmail, setEditEmail] = useState(false);
  const [editPass, setEditPass] = useState(false);
  const [confirmationShown, setConfirmationShown] = useState(false);

  const { classes } = useStyles();
  const dispatch = useAppDispatch();

  const { isHosted } = useSelector(getFeatures);
  const isEnterprise = useSelector(getIsEnterprise);
  const [confirmUnlink, setConfirmUnlink] = useState(false);
  const [resetNotice, setResetNotice] = useState(false);
  const canHave2FA = isEnterprise || isHosted;
  const currentUser = useSelector(getCurrentUser);
  const { isOAuth2, provider } = getUserSSOState(currentUser);
  const { email, id: userId } = currentUser;
  const hasTracking = useSelector(state => !!state.app.trackerCode);
  const { trackingConsentGiven: hasTrackingConsent } = useSelector(getUserSettings);
  const isDarkMode = useSelector(getIsDarkMode);
  const { token } = useSelector(getCurrentSession);
  const [showNotice, setShowNotice] = useState<string>('');

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

  const handleUnlinkConfirmed = () => {
    setConfirmUnlink(false);
    dispatch(passwordResetStart(email))
      .unwrap()
      .then(() => setResetNotice(true));
  };

  const handleEmail = () => setEditEmail(toggle);

  const toggleMode = () => {
    const newMode = isDarkMode ? LIGHT_MODE : DARK_MODE;
    dispatch(saveUserSettings({ mode: newMode }));
  };

  const startVerification = () => {
    dispatch(verifyEmailStart())
      .unwrap()
      .then(() => setConfirmationShown(true));
  };
  const handlePass = () => setEditPass(toggle);

  const needsVerification = currentUser.email && !currentUser.verified;
  return (
    <div className={`margin-top-small ${classes.widthLimit}`}>
      <Typography variant="h6" className="margin-top-small">
        My profile
      </Typography>
      {resetNotice && (
        <Alert className="margin-bottom-small" severity="warning" onClose={() => setResetNotice(false)}>
          We&rsquo;ve sent a reset password link. Please check your email.
        </Alert>
      )}

      {isOAuth2 && (
        <Alert severity="info" className="margin-top-small">
          Your Mender account is currently linked to your {provider.name} account. If you unlink these accounts, then your Mender account password will be reset
          and you will be able to change your email address or login method afterwards.
        </Alert>
      )}
      {confirmUnlink && (
        <BaseDialog open maxWidth="xs" title={`Unlink your ${provider.name} account`} onClose={() => setConfirmUnlink(false)}>
          <DialogContent>
            Please confirm that you would like to unlink your account. If you proceed, we’ll send an email to <b>{email}</b> with instructions for resetting
            your Mender account password.
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setConfirmUnlink(false)}>Cancel</Button>
            <Button variant="contained" color="error" onClick={handleUnlinkConfirmed}>
              Confirm
            </Button>
          </DialogActions>
        </BaseDialog>
      )}
      {needsVerification && (
        <Alert severity="warning" className="margin-bottom">
          Enhance your account security. We recommend you complete these essential steps:{' '}
          <ul className="margin-none padding-left">
            <li>Verify your email</li>
            {!isOAuth2 && <li>Enable two-factor authentication (2FA)</li>}
          </ul>
        </Alert>
      )}
      {!needsVerification && showNotice && (
        <Alert className="flexbox align-items-center margin-bottom-small" severity="success" onClose={() => setShowNotice('')}>
          <Typography>{showNotice}</Typography>
        </Alert>
      )}
      {confirmationShown && <EmailVerificationConfirmation onClose={() => setConfirmationShown(false)} email={email} />}
      <UserId className="margin-bottom-none margin-top-small" userId={userId} />
      {!editEmail && email ? (
        <>
          <div className="flexbox space-between margin-bottom-small">
            <div className="flexbox align-items-center">
              <TextField className={classes.formField} label="Email" key={email} disabled defaultValue={email} />
              <Chip
                size="small"
                label={needsVerification ? 'Not verified' : 'Verified'}
                variant="outlined"
                color={needsVerification ? 'warning' : 'success'}
                className="margin-left-x-small"
              />
            </div>
            {!isOAuth2 && (
              <Button color="primary" id="change_email" onClick={handleEmail}>
                Change email
              </Button>
            )}
          </div>
          {needsVerification && (
            <Button className="margin-top-x-small" variant="contained" color="primary" onClick={startVerification}>
              Verify
            </Button>
          )}
        </>
      ) : (
        <Form defaultValues={{ email }} onSubmit={editSubmit} handleCancel={handleEmail} submitLabel="Save" showButtons={editEmail}>
          <TextInput hint="Email" id="email" label="Email" validations="isLength:1,isEmail,trim" />
          <PasswordInput
            className="margin-top-x-small"
            id="current_password"
            label="Current password *"
            validations={`isLength:8:256,isNot:${email}`}
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
            <Form onSubmit={editSubmit} handleCancel={handlePass} submitLabel="Save" showButtons={editPass}>
              <PasswordInput
                className="margin-bottom-x-small"
                id="current_password"
                label="Current password *"
                validations={`isLength:8:256,isNot:${email}`}
                required
              />
              <PasswordInput id="password" label="Password *" validations={`isLength:8:256,isNot:${email}`} create generate required />
              <PasswordInput
                className="margin-top-x-small"
                id="password_confirmation"
                label="Confirm password *"
                validations={`isLength:8:256,isNot:${email}`}
                required
              />
            </Form>
          </>
        ))}
      {!isOAuth2 ? (
        canHave2FA && <TwoFactorAuthSetup setShowNotice={setShowNotice} needsVerification={needsVerification} />
      ) : (
        <Button color="neutral" variant="outlined" startIcon={provider.icon} onClick={() => setConfirmUnlink(true)}>
          Unlink from {provider.name}
        </Button>
      )}
      <ToggleSetting className="margin-top" title="Enable dark theme" onClick={toggleMode} value={isDarkMode} />
      <div className={`flexbox space-between ${classes.sessionTokenSection}`}>
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
        <div className="flexbox align-items-center">
          <CopyTextToClipboard token={token} />
        </div>
      </div>
      <AccessTokenManagement />
      {isEnterprise && hasTracking && (
        <ToggleSetting
          className="margin-top"
          description="Enable usage data and errors to be sent to help us improve our service."
          title="Help us improve Mender"
          onClick={() => dispatch(saveUserSettings({ trackingConsentGiven: !hasTrackingConsent }))}
          value={!!hasTrackingConsent}
        />
      )}
    </div>
  );
};

export default SelfUserManagement;
