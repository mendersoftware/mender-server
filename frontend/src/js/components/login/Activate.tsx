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
import { useSelector } from 'react-redux';
import { Link as RouterLink, useParams } from 'react-router';

import { Mail as MailIcon } from '@mui/icons-material';
import { Button, CircularProgress, IconButton, Typography, alpha } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { Link } from '@northern.tech/common-ui/Link';
import { getSessionInfo } from '@northern.tech/store/auth';
import { getCurrentSession } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { completeEmailChange, verifyEmailComplete } from '@northern.tech/store/thunks';

import { PasswordScreenContainer } from './Password';

const useStyles = makeStyles()(theme => ({
  iconButton: {
    width: theme.spacing(12),
    height: theme.spacing(12),
    cursor: 'default'
  },
  iconButtonSuccess: {
    backgroundColor: alpha(theme.palette.success.main, 0.08)
  },
  iconButtonError: {
    backgroundColor: alpha(theme.palette.error.main, 0.08)
  },
  icon: {
    fontSize: '3rem'
  }
}));

const errorsByStatus: Record<number, string> = {
  410: 'This link has expired',
  400: 'This link is invalid',
  409: 'This email change request can no longer be completed'
};

const getErrorForStatus = (status: number, data?: { error?: string }) => {
  const error = errorsByStatus[status] ?? 'Unknown error verifying email';
  const details = errorsByStatus[status] ? '' : `Error code: ${status} details: ${data?.error}`;
  return { error, details };
};

const ActivateError = ({ errorDetails, isLoggedIn }: { errorDetails: string; isLoggedIn: boolean }) => (
  <>
    {errorDetails && <Typography>{errorDetails}</Typography>}
    <Typography>
      Go to <Link to={isLoggedIn ? '/settings' : '/login'}>{isLoggedIn ? 'Settings ' : 'Login '}page</Link>
    </Typography>
  </>
);

const ActivateSuccess = ({ isLoggedIn }: { isLoggedIn: boolean }) => (
  <>
    <Typography>Your new email address has been successfully confirmed.</Typography>
    <Button className="margin-top-small" variant="contained" component={RouterLink} to={isLoggedIn ? '/dashboard' : '/login'}>
      Continue to {isLoggedIn ? 'dashboard' : 'login'}
    </Button>
    <Typography className="margin-top">If you didn&#39;t make this change, please contact support.</Typography>
  </>
);

export const Activate = () => {
  const [isVerifying, setIsVerifying] = useState(true);
  const [error, setError] = useState('');
  const [errorDetails, setErrorDetails] = useState('');
  const { code, secretHash } = useParams();
  const dispatch = useAppDispatch();
  const { token: storedToken } = getSessionInfo();
  const { token = storedToken } = useSelector(getCurrentSession);
  const isLoggedIn = !!token;
  const { classes } = useStyles();

  useEffect(() => {
    if (!code && !secretHash) {
      setIsVerifying(false);
      setError('Verification code is missing');
      return;
    }
    dispatch(secretHash ? completeEmailChange(secretHash) : verifyEmailComplete(code))
      .unwrap()
      .then(() => setIsVerifying(false))
      .catch(e => {
        const { error: msg, details } = getErrorForStatus(e.status, e.data);
        setError(msg);
        setErrorDetails(details);
        setIsVerifying(false);
      });
  }, [code, dispatch, secretHash]);

  if (isVerifying) {
    return (
      <PasswordScreenContainer hasLocationWarning={false} hasReturn={false} title="">
        <div className="flexbox centered">
          <CircularProgress />
          <Typography className="margin-left-small">Verifying your email address...</Typography>
        </div>
      </PasswordScreenContainer>
    );
  }

  return (
    <PasswordScreenContainer hasLocationWarning={false} hasReturn={false} title="">
      <div className="flexbox column centered">
        <IconButton
          className="align-center"
          size="large"
          color={error ? 'error' : 'success'}
          classes={{ root: `${classes.iconButton} ${error ? classes.iconButtonError : classes.iconButtonSuccess}` }}
          disableRipple
        >
          <MailIcon className={classes.icon} />
        </IconButton>
        <Typography className="align-center margin-top-medium margin-bottom-medium" variant="h4">
          {error || 'Email verified'}
        </Typography>
        {error ? <ActivateError errorDetails={errorDetails} isLoggedIn={isLoggedIn} /> : <ActivateSuccess isLoggedIn={isLoggedIn} />}
      </div>
    </PasswordScreenContainer>
  );
};

export default Activate;
