// Copyright 2020 Northern.tech AS
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
import { ReactNode, useEffect, useRef, useState } from 'react';
import { useDispatch } from 'react-redux';
import { Link } from 'react-router-dom';

import { buttonClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { SupportLink } from '@northern.tech/common-ui/SupportLink';
import Form from '@northern.tech/common-ui/forms/Form';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import { passwordResetStart } from '@northern.tech/store/thunks';

import LoginLogo from '../../../assets/img/loginlogo.svg';
import VeryMuch from '../../../assets/img/verymuch.svg';
import { LocationWarning } from './Login';

const useStyles = makeStyles()(theme => ({
  loginBox: { marginTop: 190, maxWidth: 400, '&#login-box': { alignSelf: 'start' } },
  logo: { maxWidth: 215 },
  buttonWrapper: { [`.${buttonClasses.root}`]: { width: '100%' } },
  requiredReset: { '.required:after': { content: 'none' } },
  ntBrandingLeft: { bottom: `${theme.mixins.toolbar.minHeight}px`, left: '1vw', zIndex: 0 },
  ntBrandingRight: { right: '2vw', top: '-3vh', transform: 'rotate(90deg)', zIndex: 0 }
}));

export const PasswordScreenContainer = ({ children, hasReturn = true, title }: { children: ReactNode; hasReturn?: boolean; title: string }) => {
  const { classes } = useStyles();
  return (
    <>
      <LocationWarning />
      <div className={`flexbox column content ${classes.loginBox}`} id="login-box">
        <a href="https://mender.io/" target="_blank" rel="noopener noreferrer">
          <LoginLogo alt="mender-logo" className={`flexbox margin-bottom-large ${classes.logo}`} />
        </a>
        <h1>{title}</h1>
        {children}
        {hasReturn && (
          <div className="margin-top-large flexbox centered">
            <Link to="/login">Return to login page</Link>
          </div>
        )}
      </div>
      <VeryMuch className={`absolute ${classes.ntBrandingLeft}`} />
      <VeryMuch className={`absolute ${classes.ntBrandingRight}`} />
    </>
  );
};

interface PasswordResetState {
  email: string;
}

const PasswordForgotRequest = ({ onSubmit }: { onSubmit: (formValues: PasswordResetState) => Promise<void> }) => {
  const { classes } = useStyles();
  const inputRef = useRef<HTMLInputElement | undefined>(undefined);

  useEffect(() => {
    if (inputRef.current) {
      inputRef.current.focus();
    }
  }, []);

  return (
    <>
      <p className="margin-bottom">Enter the email address associated with your account, and we&apos;ll send you a reset link.</p>
      <Form
        buttonColor="primary"
        classes={{ buttonWrapper: classes.buttonWrapper }}
        className={classes.requiredReset}
        defaultValues={{ email: '' }}
        onSubmit={onSubmit}
        showButtons={true}
        submitLabel="Send password reset link"
      >
        <TextInput
          className="full-width"
          controlRef={inputRef}
          hint="Your email"
          id="email"
          label="Your email"
          required
          validations="isLength:1,isEmail,trim"
        />
      </Form>
    </>
  );
};

const PasswordResetInfo = ({ email }: { email: string }) => (
  <>
    <p>
      If there is a Mender account with email address <b>{email}</b>, you&apos;ll receive an email with a link and instructions to reset your password.
    </p>
    <Link className="margin-top-small margin-bottom-small" to="/login">
      Return to login page
    </Link>
    <p>
      If you still haven&apos;t received the email, check your spam folder or <SupportLink variant="support" />.
    </p>
  </>
);

export const Password = () => {
  const [confirm, setConfirm] = useState<boolean>(false);
  const [email, setEmail] = useState<string>('');

  const dispatch = useDispatch();

  const handleSubmit = (formData: PasswordResetState) =>
    dispatch(passwordResetStart(formData.email))
      .unwrap()
      .then(() => {
        setEmail(formData.email);
        setConfirm(true);
      });

  return (
    <PasswordScreenContainer title={confirm ? 'Reset your password' : 'Forgot password?'} hasReturn={!confirm}>
      {confirm ? <PasswordResetInfo email={email} /> : <PasswordForgotRequest onSubmit={handleSubmit} />}
    </PasswordScreenContainer>
  );
};

export default Password;
