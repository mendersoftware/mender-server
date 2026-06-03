// Copyright 2026 Northern.tech AS
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
import { useEffect, useState } from 'react';
import { FormProvider, useForm } from 'react-hook-form';
import { Link as RouterLink } from 'react-router';

import { Alert, Button, Paper, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import PasswordInput from '@northern.tech/common-ui/forms/PasswordInput';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import { useAppDispatch } from '@northern.tech/store/store';
import { confirmOAuthLink } from '@northern.tech/store/thunks';
import Cookies from 'universal-cookie';

import DotsGradient from '../../../assets/img/dots-gradient.svg';
import LoginLogo from '../../../assets/img/loginlogo.svg';
import DotsWhite from '../../../assets/img/verymuch.svg';
import { OAuth2Providers } from './OAuth2Providers';

const cookies = new Cookies();

const useStyles = makeStyles()(theme => ({
  loginBox: { maxWidth: 525, justifySelf: 'center' },
  ntBrandingLeft: {
    top: 0,
    left: 0,
    bottom: 0,
    width: '25%',
    maxWidth: 365,
    overflow: 'hidden',
    zIndex: -2,
    '&::before': {
      content: '""',
      position: 'absolute',
      inset: 0,
      background: '#284D68',
      borderTopRightRadius: 120,
      transform: 'skewY(3deg)',
      transformOrigin: '0 0'
    }
  },
  ntBrandingRight: {
    right: 0,
    bottom: 0,
    width: 88,
    height: 289,
    overflow: 'hidden',
    zIndex: -2,
    '&::before': {
      content: '""',
      display: 'block',
      width: 638,
      height: 704,
      background: 'linear-gradient(to top, #970F57, #02AFCF)',
      borderTopLeftRadius: 120
    }
  },
  dotsOverlayTop: {
    top: -10,
    right: -60,
    zIndex: 1,
    pointerEvents: 'none'
  },
  dotsOverlayBottom: {
    bottom: -24,
    left: -40,
    zIndex: 1,
    pointerEvents: 'none',
    '& path': { stroke: theme.palette.primary.light }
  },
  wideAlert: {
    width: 600,
    maxWidth: 'none'
  }
}));

type Step = 'confirm' | 'verify';

type ConfirmStepProps = {
  onContinue: () => void;
  provider: string;
};

const ConfirmStep = ({ provider, onContinue }: ConfirmStepProps) => (
  <>
    <Typography variant="h4">Link {provider} account?</Typography>
    <Typography className="margin-top-small">
      By linking your {provider} account, we will switch your Mender account&#39;s authentication method from email and password to {provider} OAuth. Moving
      forward, your credentials, personal information, and two-factor authentication will be managed through your {provider} account.
    </Typography>
    <div className="margin-top-medium flexbox">
      <Button variant="outlined" color="secondary" component={RouterLink} to="/login">
        Cancel
      </Button>
      <Button className="margin-left-small" variant="contained" color="secondary" onClick={onContinue}>
        Link my accounts
      </Button>
    </div>
  </>
);

type VerifyStepProps = {
  classes: { wideAlert: string };
  email: string;
  onCancel: () => void;
  provider: string;
};

interface VerifyFormState {
  email: string;
  password: string;
  token2fa: string;
}

const VerifyStep = ({ classes, provider, email, onCancel }: VerifyStepProps) => {
  const [show2fa, setShow2fa] = useState<boolean>(false);
  const [errorMessage, setErrorMessage] = useState<string>('');
  const dispatch = useAppDispatch();
  const methods = useForm<VerifyFormState>({ mode: 'onSubmit', defaultValues: { email, password: '', token2fa: '' } });
  const { handleSubmit } = methods;

  const onSubmit = ({ email, password, token2fa }: VerifyFormState) => {
    setErrorMessage('');
    dispatch(confirmOAuthLink({ email, password, stayLoggedIn: true, ...(token2fa ? { token2fa } : {}) }))
      .unwrap()
      .catch(err => {
        if (err?.error?.includes('2fa')) {
          return setShow2fa(true);
        }
        setErrorMessage(err?.error || 'Something went wrong. Please try again.');
      });
  };

  return (
    <FormProvider {...methods}>
      <form noValidate onSubmit={handleSubmit(onSubmit)}>
        <Typography variant="h4">{show2fa ? 'Two-factor authentication' : 'Verify your identity'}</Typography>
        {errorMessage && (
          <Alert severity="error" className={`margin-top-small ${classes.wideAlert}`}>
            {errorMessage}
          </Alert>
        )}
        {show2fa ? (
          <>
            <Typography className="margin-top-small">
              To finish linking your account, please enter the code from your authenticator app for your Mender account.
            </Typography>
            <TextInput className="margin-top-small" hint="Verification code" id="token2fa" label="Verification code" validations="isLength:6,isNumeric" />
          </>
        ) : (
          <>
            <Alert severity="warning" className={`margin-top-small ${classes.wideAlert}`}>
              Verify your identity by entering your password for your Mender account below.
            </Alert>
            <Typography className="margin-top-small">This is the {provider} email address we matched to your Mender account.</Typography>
            <TextInput className="margin-top-small" disabled hint="Your email" id="email" label="Your email" />
            <PasswordInput className="margin-top-small" id="password" label="Password" required />
          </>
        )}
        <div className="margin-top-medium flexbox">
          <Button variant="outlined" color="secondary" onClick={onCancel}>
            Cancel
          </Button>
          <Button className="margin-left-small" variant="contained" color="secondary" type="submit">
            Confirm
          </Button>
        </div>
      </form>
    </FormProvider>
  );
};

export const OAuthLink = () => {
  const [step, setStep] = useState<Step>('confirm');
  const [provider, setProvider] = useState('');
  const [email, setEmail] = useState('');
  const { classes } = useStyles();

  useEffect(() => {
    const usedOauthProvider = cookies.get('oauth');
    if (usedOauthProvider) {
      const matched = OAuth2Providers.find(({ id }) => id === usedOauthProvider);
      setProvider(matched?.name ?? usedOauthProvider);
      setEmail(cookies.get('email') ?? '');
    }
  }, []);
  return (
    <>
      <div className="margin-top-large" />
      <Paper elevation={0} className={`flexbox margin-top-large padding-small column ${classes.loginBox}`}>
        <LoginLogo className="margin-bottom-large" style={{ width: 210 }} />
        {step === 'confirm' ? (
          <ConfirmStep provider={provider} onContinue={() => setStep('verify')} />
        ) : (
          <VerifyStep classes={classes} provider={provider} email={email} onCancel={() => setStep('confirm')} />
        )}
      </Paper>
      <div className={`absolute ${classes.ntBrandingLeft}`}>
        <DotsGradient className={`absolute ${classes.dotsOverlayTop}`} />
        <DotsWhite className={`absolute ${classes.dotsOverlayBottom}`} />
      </div>
      <div className={`absolute ${classes.ntBrandingRight}`} />
    </>
  );
};
