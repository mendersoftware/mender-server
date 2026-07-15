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
import { useCallback, useEffect, useState } from 'react';
import { useFormState } from 'react-hook-form';
import { useSelector } from 'react-redux';

import { Alert, Button, Chip, DialogActions, DialogContent, TextField, Typography, formControlClasses, textFieldClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { ConfirmModal } from '@northern.tech/common-ui/ConfirmModal';
import { CopyTextToClipboard } from '@northern.tech/common-ui/CopyText';
import ExpandableAttribute from '@northern.tech/common-ui/ExpandableAttribute';
import { SettingsItem, ToggleSettingsItem } from '@northern.tech/common-ui/SettingsItem';
import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import Form from '@northern.tech/common-ui/forms/Form';
import PasswordInput from '@northern.tech/common-ui/forms/PasswordInput';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import { DARK_MODE, LIGHT_MODE, OWN_USER_ID } from '@northern.tech/store/constants';
import { getCurrentSession, getCurrentUser, getFeatures, getIsDarkMode, getIsEnterprise, getUserSettings } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import {
  cancelEmailChange,
  editUser,
  getPendingEmailChange,
  initiateEmailChange,
  passwordResetStart,
  saveUserSettings,
  verifyEmailStart
} from '@northern.tech/store/thunks';
import '@northern.tech/types/MenderTypes';
import type { PendingEmailChange } from '@northern.tech/types/MenderTypes';
import { toggle } from '@northern.tech/utils/helpers';

import AccessTokenManagement from '../AccessTokenManagement';
import { SETTINGS_CONTENT_MAX_WIDTH, SETTINGS_INPUT_WIDTH } from '../constants';
import TwoFactorAuthSetup from './TwoFactorAuthSetup';
import { UserId, getUserSSOState } from './UserDefinition';
import { EmailVerificationConfirmation } from './twofactorauth-steps/EmailVerification';

const useStyles = makeStyles()(theme => ({
  formField: { width: SETTINGS_INPUT_WIDTH, maxWidth: '100%' },
  oauthIcon: { fontSize: '36px', marginRight: 10 },
  alert: { width: SETTINGS_INPUT_WIDTH },
  widthLimit: {
    maxWidth: SETTINGS_CONTENT_MAX_WIDTH,
    [`.${textFieldClasses.root},.${formControlClasses.root}`]: { width: SETTINGS_INPUT_WIDTH },
    '.required:after': { content: 'none' }
  },
  buttonReset: { '.button-wrapper': { justifyContent: 'start', marginTop: theme.spacing(1.5) } },
  requiredReset: { '.required:after': { content: 'none' } },
  columnWidths: {
    '&.settings-item-main-content': {
      gridTemplateColumns: `${SETTINGS_INPUT_WIDTH}px 1fr`
    }
  }
}));

const PasswordConfirmButton = () => {
  const { isValid } = useFormState();
  return (
    <Button variant="contained" type="submit" disabled={!isValid}>
      Confirm
    </Button>
  );
};

export const notificationMap = {
  email: 'Email successfully verified.',
  enabled2fa: 'Two-factor authentication successfully enabled.'
};

export const SelfUserManagement = () => {
  const [editEmail, setEditEmail] = useState(false);
  const [editPass, setEditPass] = useState(false);
  const [pendingEmail, setPendingEmail] = useState('');
  const [pendingChange, setPendingChange] = useState<PendingEmailChange | null>(null);
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
  const { hasMultitenancy } = useSelector(getFeatures);
  const isOsInstallation = !isEnterprise && !hasMultitenancy;

  const editSubmit = userData => {
    dispatch(editUser({ ...userData, id: OWN_USER_ID }))
      .unwrap()
      .then(() => {
        setEditEmail(false);
        setEditPass(false);
      })
      .catch(() => {});
  };
  const onEmailSubmit = ({ email }) => setPendingEmail(email);

  const fetchPendingEmailChange = useCallback(
    () =>
      dispatch(getPendingEmailChange())
        .unwrap()
        .then(change => setPendingChange(change)),
    [dispatch]
  );
  const onCancelEmailChange = () => {
    dispatch(cancelEmailChange()).then(() => setPendingChange(null));
  };
  const onConfirmEmailChange = ({ current_password }) => {
    //Open source user use basic user change functionality
    if (isOsInstallation) {
      dispatch(editUser({ email: pendingEmail, current_password, id: OWN_USER_ID }))
        .unwrap()
        .then(() => {
          setEditEmail(false);
        });
    } else {
      dispatch(initiateEmailChange({ email: pendingEmail, current_password }))
        .unwrap()
        .then(() => {
          fetchPendingEmailChange();
          setPendingEmail('');
          setEditEmail(false);
        })
        .catch(() => {});
    }
  };

  const handleUnlinkConfirmed = () => {
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

  useEffect(() => {
    if (!isOsInstallation) {
      fetchPendingEmailChange();
    }
  }, [fetchPendingEmailChange, isOsInstallation]);

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
        <Alert severity="info" className="margin-top-small margin-bottom-small">
          Your Mender account is currently linked to your {provider.name} account. If you unlink these accounts, your Mender account password will be reset, and
          you will then be able to change your email address or login method.
        </Alert>
      )}
      {confirmUnlink && (
        <ConfirmModal
          open
          header={`Unlink your ${provider.name} account`}
          description={
            <>
              Please confirm that you would like to unlink your account. If you proceed, we’ll send an email to <b>{email}</b> with instructions for resetting
              your Mender account password.
            </>
          }
          onConfirm={handleUnlinkConfirmed}
          close={() => setConfirmUnlink(false)}
        />
      )}
      {needsVerification && !editEmail && !pendingChange && (
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
      {pendingChange && (
        <Alert className={`${classes.alert} margin-top-small`} severity="warning">
          There is a pending email change to {pendingChange.new_email} Click the verification link in the email to complete changes.
        </Alert>
      )}
      <UserId className="margin-top-small profile-settings" userId={userId} />
      {email && (
        <>
          {editEmail && (
            <Alert severity="info" className={`${classes.alert} margin-bottom-small`}>
              For security reasons, changes to your email address require verification. The update won’t be finalized until the new address is confirmed.
            </Alert>
          )}
          <div className="flexbox align-items-center">
            <TextField className={classes.formField} label={editEmail ? 'Current email address' : 'Your email'} key={email} disabled defaultValue={email} />
            <Chip
              size="small"
              label={needsVerification ? 'Not verified' : 'Verified'}
              variant="outlined"
              color={needsVerification ? 'warning' : 'success'}
              className="margin-left-small"
            />
          </div>
          <div className="flexbox column" style={{ alignItems: 'flex-start' }}>
            {needsVerification && !editEmail && (
              <Button component="div" className="margin-top-x-small" variant="contained" color="primary" onClick={startVerification}>
                Verify
              </Button>
            )}

            {!isOAuth2 && !editEmail && !pendingChange && (
              <Button className="margin-top-x-small" color="primary" id="change_email" onClick={handleEmail}>
                Change email address
              </Button>
            )}
          </div>
          {pendingChange && (
            <Button className="margin-top-small" onClick={() => onCancelEmailChange()} color="error">
              Cancel change request
            </Button>
          )}

          {editEmail && (
            <Form
              className={`${classes.buttonReset} margin-top-medium`}
              defaultValues={{ email }}
              onSubmit={onEmailSubmit}
              handleCancel={handleEmail}
              submitLabel="Save changes"
              showButtons={editEmail}
            >
              <TextInput
                hint="Email"
                id="email"
                label="New email address"
                helperText="Ensure to enter your new email address correctly."
                validations="isLength:1,isEmail,trim"
                width={null}
              />
            </Form>
          )}
          {!!pendingEmail && (
            <BaseDialog open title="Confirm email change" onClose={() => setPendingEmail('')}>
              <Form className={classes.requiredReset} onSubmit={onConfirmEmailChange} showButtons={false}>
                <DialogContent>
                  <Typography className="margin-bottom-small">Please enter your password to continue.</Typography>
                  <PasswordInput id="current_password" label="Password" required />
                </DialogContent>
                <DialogActions>
                  <Button variant="outlined" color="info" onClick={() => setPendingEmail('')}>
                    Cancel
                  </Button>
                  <PasswordConfirmButton />
                </DialogActions>
              </Form>
            </BaseDialog>
          )}
        </>
      )}
      {!isOAuth2 && (
        <SettingsItem
          title="Password"
          secondary={
            !editPass ? (
              <div className="flexbox column">
                <TextField className={classes.formField} label="Password" disabled defaultValue="********" type="password" />
                <Button className="margin-top-x-small align-self-start" color="primary" onClick={handlePass}>
                  Change password
                </Button>
              </div>
            ) : (
              <>
                <Form
                  classes={{ buttonWrapper: 'justify-content-start margin-top-x-small' }}
                  onSubmit={editSubmit}
                  handleCancel={handlePass}
                  submitLabel="Save changes"
                  showButtons={editPass}
                  validationMode="onSubmit"
                >
                  <PasswordInput className="margin-bottom-medium" id="current_password" label="Current password *" required width={null} />
                  <PasswordInput id="password" label="New password *" validations={`isLength:8:256,isNot:${email}`} create required width={null} />
                  <PasswordInput
                    className="margin-top-x-small"
                    id="password_confirmation"
                    label="Confirm new password *"
                    validations={`isLength:8:256,isNot:${email}`}
                    required
                    width={null}
                  />
                </Form>
              </>
            )
          }
        />
      )}
      <ToggleSettingsItem title="Enable dark theme" onClick={toggleMode} checked={isDarkMode} />
      {!isOAuth2 ? (
        canHave2FA && <TwoFactorAuthSetup setShowNotice={setShowNotice} needsVerification={needsVerification} />
      ) : (
        <div className="margin-top-x-small">
          <Button color="neutral" variant="outlined" startIcon={provider.icon} onClick={() => setConfirmUnlink(true)}>
            Unlink from {provider.name}
          </Button>
        </div>
      )}
      <SettingsItem
        classes={{ main: classes.columnWidths }}
        title="Session token"
        secondary={
          <ExpandableAttribute
            component="div"
            disableGutters
            dividerDisabled
            secondary={token}
            textClasses={{ secondary: 'inventory-text tenant-token-text' }}
          />
        }
        sideBarContent={<CopyTextToClipboard token={token} />}
      />
      <AccessTokenManagement />
      {isEnterprise && hasTracking && (
        <ToggleSettingsItem
          description="Enable usage data and errors to be sent to help us improve our service."
          title="Help us improve Mender"
          onClick={() => dispatch(saveUserSettings({ trackingConsentGiven: !hasTrackingConsent }))}
          checked={!!hasTrackingConsent}
        />
      )}
    </div>
  );
};

export default SelfUserManagement;
