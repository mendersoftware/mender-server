// Copyright 2016 Northern.tech AS
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
import React, { useCallback, useEffect } from 'react';
import { useSelector } from 'react-redux';
import { Link } from 'react-router-dom';

import { ChevronRight } from '@mui/icons-material';
import { Button } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import LinedHeader from '@northern.tech/common-ui/LinedHeader';
import storeActions from '@northern.tech/store/actions';
import { getToken } from '@northern.tech/store/auth';
import { TIMEOUTS, locations, useradmApiUrl } from '@northern.tech/store/constants';
import { getCurrentUser, getFeatures, getIsEnterprise } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { loginUser, logoutUser } from '@northern.tech/store/thunks';
import { clearAllRetryTimers } from '@northern.tech/utils/retrytimer';
import Cookies from 'universal-cookie';

import LoginLogo from '../../../assets/img/loginlogo.svg';
import VeryMuch from '../../../assets/img/verymuch.svg';
import { LoginForm } from './LoginForm';
import { OAuth2Providers } from './OAuth2Providers';

const { setSnackbar } = storeActions;

const cookies = new Cookies();

export const locationMap = {
  eu: { ...locations.eu, fallback: locations.us },
  us: { ...locations.us, fallback: locations.eu }
};

const useStyles = makeStyles()(theme => {
  const skew = 3;
  const backgroundRadius = 100;
  return {
    entryLink: {
      color: theme.palette.background.paper,
      a: { color: theme.palette.text.entryLink }
    },
    reset: {
      transform: `skew(0, ${skew}deg)`,
      'svg': { maxWidth: 200 },
      '#login-logo path': { fill: theme.palette.background.paper },
      '#login-box': {
        background: theme.palette.background.paper,
        minWidth: 'calc(100% + 20px)',
        maxWidth: 480,
        paddingBottom: 25,
        paddingRight: 50,
        paddingLeft: 50,
        borderRadius: 10
      }
    },
    background: {
      background: theme.palette.background.darkBlue,
      padding: '40px 65px',
      borderTopLeftRadius: backgroundRadius,
      borderBottomRightRadius: backgroundRadius,
      marginBottom: 60,
      marginTop: 30,
      transform: `skew(0, -${skew}deg)`,
      zIndex: 1
    },
    ntBranding: { bottom: `calc(${theme.mixins.toolbar.minHeight}px + 3vh)`, right: 0, zIndex: 0 }
  };
});

const entryText = {
  signup: { linkText: 'Sign up here', question: `Don't have an account?`, target: '/signup' },
  login: { linkText: 'Log in', question: `Already have an account?`, target: '/login' }
};

export const EntryLink = ({ className = '', target = 'signup' }) => (
  <div className={`margin-top margin-bottom flexbox centered ${className}`}>
    <div className="muted margin-right">{entryText[target].question}</div>
    <Link className="flexbox center-aligned" to={entryText[target].target}>
      {entryText[target].linkText} <ChevronRight fontSize="small" />
    </Link>
  </div>
);

export const LocationWarning = () => {
  const location = Object.entries(locations).reduce(
    (accu, [key, value]) => ([`staging.${value.location}`, value.location].includes(window.location.hostname) ? key : accu),
    locations.us.key
  );
  const { icon: Icon, title, fallback } = locationMap[location];
  return (
    <div className="flexbox centered margin-top-large">
      <Icon />
      <div className="margin-left-small">
        You are logging into the <b style={{ marginLeft: 4 }}>{title} server</b>.
      </div>
      <a className="flexbox center-aligned margin-left-small" href={`https://${fallback.location}/ui/`}>
        Change to {fallback.title} <ChevronRight fontSize="small" />
      </a>
    </div>
  );
};

export const OAuthHeader = ({ buttonProps, type }) => (
  <>
    <div className="flexbox centered margin-bottom">{type} with:</div>
    <div className="flexbox centered">
      {OAuth2Providers.map(provider => {
        const props = buttonProps ? buttonProps : { href: `${useradmApiUrl}/oauth2/${provider.id}` };
        return (
          <Button className="oauth-provider" variant="contained" key={provider.id} startIcon={provider.icon} {...props}>
            {provider.name}
          </Button>
        );
      })}
    </div>
    <LinedHeader className="margin-top-large flexbox centered" heading="or your email address" innerStyle={{ padding: 15, top: -24 }} />
  </>
);

export const Login = () => {
  const dispatch = useAppDispatch();
  const currentUser = useSelector(getCurrentUser);
  const { isHosted } = useSelector(getFeatures);
  const isEnterprise = useSelector(getIsEnterprise);
  const { classes } = useStyles();

  useEffect(() => {
    clearAllRetryTimers(message => dispatch(setSnackbar(message)));
    if (getToken()) {
      dispatch(logoutUser());
    }
    const loginError = cookies.get('error');
    if (loginError) {
      dispatch(setSnackbar({ message: loginError, autoHideDuration: TIMEOUTS.refreshDefault }));
      cookies.remove('error');
    }
    return () => {
      dispatch(setSnackbar(''));
    };
  }, [dispatch]);

  useEffect(() => {
    if (currentUser.id) {
      dispatch(setSnackbar(''));
    }
  }, [currentUser, dispatch]);

  const onLoginClick = useCallback(
    ({ noExpiry, ...loginData }) =>
      // set no expiry in localstorage to remember checkbox value and avoid any influence of expiration time that might occur with cookies
      dispatch(loginUser({ ...loginData, stayLoggedIn: noExpiry })).unwrap(),
    [dispatch]
  );

  const onOAuthClick = ({ target: { textContent } }) => {
    const providerId = OAuth2Providers.find(provider => provider.name === textContent).id;
    const oauthTimeout = new Date();
    oauthTimeout.setDate(oauthTimeout.getDate() + 7);
    window.localStorage.setItem('oauth', `${oauthTimeout.getTime()}`);
    window.location.replace(`${useradmApiUrl}/oauth2/${providerId}`);
  };

  return (
    <>
      {isHosted ? <LocationWarning /> : <div />}
      <div className={`content ${classes.background}`}>
        <div className={`flexbox column centered ${classes.reset}`}>
          <LoginLogo alt="mender-logo" id="login-logo" className="margin-bottom" />
          <div className="flexbox column" id="login-box">
            <h1 className="flexbox centered">Welcome back!</h1>
            {isHosted && <OAuthHeader type="Log in" buttonProps={{ onClick: onOAuthClick }} />}
            <LoginForm isEnterprise={isEnterprise} isHosted={isHosted} onSubmit={onLoginClick} />
          </div>
          {isHosted ? <EntryLink className={classes.entryLink} target="signup" /> : <div className="padding" />}
        </div>
      </div>
      <VeryMuch className={`absolute ${classes.ntBranding}`} />
    </>
  );
};

export default Login;
