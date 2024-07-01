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
import React, { useCallback, useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { Collapse, Switch } from '@mui/material';

import { setSnackbar } from '../../../actions/appActions';
import { disableUser2fa, enableUser2fa, get2FAQRCode, verify2FA, verifyEmailComplete, verifyEmailStart } from '../../../actions/userActions';
import { twoFAStates } from '../../../constants/userConstants';
import { getCurrentUser, getHas2FA } from '../../../selectors';
import InfoText from '../../common/infotext';
import AuthSetup from './twofactorauth-steps/authsetup';
import EmailVerification from './twofactorauth-steps/emailverification';

export const TwoFactorAuthSetup = () => {
  const activationCode = useSelector(state => state.users.activationCode);
  const currentUser = useSelector(getCurrentUser);
  const has2FA = useSelector(getHas2FA);
  const qrImage = useSelector(state => state.users.qrCode);
  const [qrExpanded, setQrExpanded] = useState(false);
  const [is2FAEnabled, setIs2FAEnabled] = useState(has2FA);
  const [showEmailVerification, setShowEmailVerification] = useState(false);
  const dispatch = useDispatch();

  useEffect(() => {
    if ((currentUser.verified || currentUser.email?.endsWith('@example.com')) && is2FAEnabled && !has2FA) {
      setShowEmailVerification(false);
      setQrExpanded(true);
    }
  }, [currentUser.email, currentUser.verified, is2FAEnabled, has2FA]);

  useEffect(() => {
    if (activationCode) {
      setIs2FAEnabled(true);
      dispatch(verifyEmailComplete(activationCode))
        .catch(() => {
          setShowEmailVerification(true);
          setQrExpanded(false);
        })
        // we have to explicitly call this, to not send the returned promise as user to activate 2fa for
        .then(() => dispatch(enableUser2fa()))
        .then(() => dispatch(get2FAQRCode()));
    }
  }, [activationCode, dispatch]);

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
      is2FAEnabled ? dispatch(disableUser2fa()) : handle2FAState(twoFAStates.unverified);
      setQrExpanded(!is2FAEnabled);
      setIs2FAEnabled(!is2FAEnabled);
    }
  }, [currentUser.email, currentUser.verified, dispatch, handle2FAState, has2FA, is2FAEnabled, showEmailVerification]);

  return (
    <div className="margin-top">
      <div className="clickable flexbox space-between" onClick={onToggle2FAClick}>
        <p className="help-content">Enable Two Factor authentication</p>
        <Switch checked={is2FAEnabled} />
      </div>
      <InfoText style={{ width: '75%', margin: 0 }}>
        Two Factor Authentication adds a second layer of protection to your account by asking for an additional verification code each time you log in.
      </InfoText>
      {showEmailVerification && (
        <EmailVerification
          activationCode={activationCode}
          verifyEmailComplete={data => dispatch(verifyEmailComplete(data))}
          verifyEmailStart={() => dispatch(verifyEmailStart())}
        />
      )}
      <Collapse in={qrExpanded} timeout="auto" unmountOnExit>
        <AuthSetup
          currentUser={currentUser}
          handle2FAState={handle2FAState}
          has2FA={has2FA}
          qrImage={qrImage}
          verify2FA={data => dispatch(verify2FA(data))}
          onClose={() => setQrExpanded(false)}
        />
      </Collapse>
    </div>
  );
};

export default TwoFactorAuthSetup;
