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
import Cookies from 'universal-cookie';

import { TIMEOUTS } from './constants';

const cookies = new Cookies();

const emptySession = Object.freeze({ token: '', exiresAt: undefined });

let tokenCache = '';

export const getSessionInfo = () => {
  let sessionInfo = { ...emptySession };
  try {
    sessionInfo = JSON.parse(window.localStorage.getItem('JWT') ?? '');
  } catch (error) {
    // most likely not logged in - nothing to do here
  }
  if (sessionInfo.expiresAt && new Date(sessionInfo.expiresAt) < new Date()) {
    cleanUp();
    return { ...emptySession };
  }
  if (!sessionInfo.token) {
    let jwtTokenFromCookie = cookies.get('JWT', { doNotParse: true }) ?? '';
    if (jwtTokenFromCookie) {
      setSessionInfo({ token: jwtTokenFromCookie, undefined });
      sessionInfo.token = jwtTokenFromCookie;
      cookies.remove('JWT');
      cookies.remove('JWT', { path: '/' });
    }
  }
  sessionInfo.token = sessionInfo.token || undefined;
  tokenCache = sessionInfo.token;
  return sessionInfo;
};

export const getToken = () => (tokenCache ? tokenCache : getSessionInfo().token);

export const setSessionInfo = ({ token, expiresAt }) => {
  tokenCache = token;
  window.localStorage.setItem('JWT', JSON.stringify({ token, expiresAt }));
};

export const cleanUp = () => {
  tokenCache = '';
  cookies.remove('JWT');
  cookies.remove('JWT', { path: '/' });
  window.localStorage.removeItem('JWT');
  window.localStorage.removeItem('oauth');
};

export const maxSessionAge = 900;

export const updateMaxAge = ({ expiresAt, token }) => {
  const oAuthExpiration = Number(window.localStorage.getItem('oauth'));
  let updateWithOAuth = false;
  if (oAuthExpiration) {
    const soon = Date.now() + maxSessionAge * TIMEOUTS.oneSecond;
    updateWithOAuth = oAuthExpiration <= soon;
    if (updateWithOAuth) {
      window.localStorage.removeItem('oauth');
    }
  }
  if (token && expiresAt && (!oAuthExpiration || updateWithOAuth)) {
    const expiration = new Date();
    expiration.setSeconds(expiration.getSeconds() + maxSessionAge);
    setSessionInfo({ token, expiresAt: expiration.toISOString() });
  }
};
