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
import { useDispatch } from 'react-redux';
import { Link, useParams } from 'react-router-dom';

import { Mail as MailIcon } from '@mui/icons-material';
import { Button, CircularProgress, IconButton, Typography, alpha } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { verifyEmailComplete } from '@northern.tech/store/thunks';

import { PasswordScreenContainer } from './Password';

const useStyles = makeStyles()(theme => ({
  iconButton: {
    width: theme.spacing(9),
    height: theme.spacing(9),
    cursor: 'default'
  },
  iconButtonSuccess: {
    backgroundColor: alpha(theme.palette.success.main, 0.08)
  },
  iconButtonError: {
    backgroundColor: alpha(theme.palette.error.main, 0.08)
  },
  icon: {
    fontSize: '2rem'
  }
}));

export const Activate = () => {
  const [isVerifying, setIsVerifying] = useState(true);
  const [error, setError] = useState('');
  const [errorDetails, setErrorDetails] = useState('');
  const { code } = useParams();
  const dispatch = useDispatch();

  const { classes } = useStyles();

  useEffect(() => {
    if (!code) {
      setIsVerifying(false);
      setError('Verification code is missing');
      return;
    }
    dispatch(verifyEmailComplete(code))
      .unwrap()
      .then(() => setIsVerifying(false))
      .catch(e => {
        if (e.status === 410) {
          setError('This link has expired');
        } else if (e.status === 400) {
          setError('This link is invalid');
        } else {
          setError('Unknown error verifying email');
          setErrorDetails(`Error code: ${e.status} details: ${e.data?.error}`);
        }
        setIsVerifying(false);
      });
  }, [code, dispatch]);

  return (
    <PasswordScreenContainer hasLocationWarning={false} hasReturn={false} title="">
      {isVerifying ? (
        <div className="flexbox centered">
          <CircularProgress />
          <Typography className="margin-left-small">Verifying your email address...</Typography>
        </div>
      ) : (
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
            {error ? error : 'Email verified'}
          </Typography>
          {error ? (
            <>
              {errorDetails && <Typography>{errorDetails}</Typography>}
              <Typography>
                Back to <Link to="/settings">Settings page</Link>{' '}
              </Typography>
            </>
          ) : (
            <>
              <Typography>Your new email address has been successfully confirmed.</Typography>
              <Button className="margin-top-small" variant="contained" component={Link} to="/dashboard">
                Continue to dashboard
              </Button>
            </>
          )}
        </div>
      )}
    </PasswordScreenContainer>
  );
};

export default Activate;
