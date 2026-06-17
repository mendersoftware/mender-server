// Copyright 2019 Northern.tech AS
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
import { useSelector } from 'react-redux';

import { Alert, Button, Chip, Collapse, Switch, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { ConfirmModal } from '@northern.tech/common-ui/ConfirmModal';
import { SettingsItem } from '@northern.tech/common-ui/SettingsItem';
import storeActions from '@northern.tech/store/actions';
import { twoFAStates } from '@northern.tech/store/constants';
import { getCurrentUser, getHas2FA } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { disableUser2fa, enableUser2fa, get2FAQRCode, verify2FA, verifyEmailStart } from '@northern.tech/store/thunks';

import { SETTINGS_INPUT_WIDTH } from '../constants';
import AuthSetup from './twofactorauth-steps/AuthSetup';
import EmailVerification from './twofactorauth-steps/EmailVerification';

const { setSnackbar } = storeActions;

const useStyles = makeStyles()(() => ({
  columnWidths: {
    '&.settings-item-content': {
      width: SETTINGS_INPUT_WIDTH
    }
  }
}));

export const TwoFactorAuthSetup = ({ needsVerification, setShowNotice }) => {
  const currentUser = useSelector(getCurrentUser);
  const has2FA = useSelector(getHas2FA);
  const qrImage = useSelector(state => state.users.qrCode);
  const [confirmDisable, setConfirmDisable] = useState(false);
  const [qrExpanded, setQrExpanded] = useState(false);
  const [is2FAEnabled, setIs2FAEnabled] = useState(has2FA);
  const [showEmailVerification, setShowEmailVerification] = useState(false);
  const dispatch = useAppDispatch();
  const { classes } = useStyles();

  useEffect(() => {
    if ((currentUser.verified || currentUser.email?.endsWith('@example.com')) && is2FAEnabled && !has2FA) {
      setShowEmailVerification(false);
      setQrExpanded(true);
    }
  }, [currentUser.email, currentUser.verified, is2FAEnabled, has2FA]);

  useEffect(() => {
    if (has2FA) {
      setIs2FAEnabled(has2FA);
    }
  }, [has2FA]);

  const handle2FAState = useCallback(
    state => {
      setIs2FAEnabled(state !== twoFAStates.disabled);
      setQrExpanded(state === twoFAStates.unverified);
      let request;
      if (state === twoFAStates.disabled) {
        request = dispatch(disableUser2fa());
      } else {
        request = dispatch(enableUser2fa());
      }
      request.then(() => {
        if (state === twoFAStates.unverified) {
          dispatch(get2FAQRCode());
        } else if (state === twoFAStates.enabled) {
          setSnackbar('Two Factor authentication set up successfully.');
        }
      });
    },
    [dispatch]
  );

  const onToggle2FAClick = useCallback(() => {
    if (!(currentUser.verified || currentUser.email?.endsWith('@example.com'))) {
      setShowEmailVerification(!showEmailVerification);
      setIs2FAEnabled(!showEmailVerification);
      return;
    }
    if (has2FA) {
      handle2FAState(twoFAStates.disabled);
    } else {
      if (is2FAEnabled) {
        dispatch(disableUser2fa());
      } else {
        handle2FAState(twoFAStates.unverified);
      }
      setQrExpanded(!is2FAEnabled);
      setIs2FAEnabled(!is2FAEnabled);
    }
  }, [currentUser.email, currentUser.verified, dispatch, handle2FAState, has2FA, is2FAEnabled, showEmailVerification]);

  return (
    <>
      <SettingsItem
        classes={{ content: classes.columnWidths }}
        title={
          <div className="flexbox align-items-center">
            <Typography
              className={`capitalized-start margin-right-x-small ${has2FA ? '' : 'clickable'}`}
              onClick={has2FA ? undefined : onToggle2FAClick}
              variant="subtitle1"
            >
              {has2FA ? '' : 'Enable '}Two Factor authentication
            </Typography>
            {has2FA ? (
              <Chip size="small" label="Enabled" variant="outlined" color="success" className="margin-left-x-small" />
            ) : (
              <Switch className="margin-left-small" checked={!!has2FA || qrExpanded} onChange={onToggle2FAClick} />
            )}
          </div>
        }
        description="Two-factor authentication adds a second layer of protection to your account by asking for an additional verification code each time you log in."
        secondary={
          <>
            {showEmailVerification ? (
              <EmailVerification email={currentUser.email} verifyEmailStart={() => dispatch(verifyEmailStart()).unwrap()} />
            ) : (
              <>
                {!qrExpanded && has2FA && (
                  <Button variant="outlined" color="error" onClick={() => setConfirmDisable(true)}>
                    Disable 2FA
                  </Button>
                )}
              </>
            )}
            <Collapse in={qrExpanded} timeout="auto" unmountOnExit>
              <AuthSetup
                setShowNotice={setShowNotice}
                currentUser={currentUser}
                handle2FAState={handle2FAState}
                has2FA={has2FA}
                qrImage={qrImage}
                verify2FA={data => dispatch(verify2FA(data))}
                onClose={() => setQrExpanded(false)}
              />
            </Collapse>
            {!has2FA && !qrExpanded && !needsVerification && (
              <Alert severity="warning">Two-factor authentication is not enabled yet. Enable it now to prevent unauthorized access.</Alert>
            )}
          </>
        }
      />
      <ConfirmModal
        open={confirmDisable}
        description="Are you sure you want to turn off 2FA? This will make your account less secure."
        header="Disable two-factor authentication"
        close={() => setConfirmDisable(false)}
        confirmButtonText="Disable 2FA"
        onConfirm={onToggle2FAClick}
      />
    </>
  );
};

export default TwoFactorAuthSetup;
