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
import React, { useCallback, useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Navigate, useParams } from 'react-router-dom';

import { formControlClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import Cookies from 'universal-cookie';

import LoginLogo from '../../../assets/img/loginlogo.svg';
import SignupHero from '../../../assets/img/signuphero.svg';
import { setSnackbar } from '../../actions/appActions';
import { createOrganizationTrial } from '../../actions/organizationActions';
import { TIMEOUTS, locations } from '../../constants/appConstants';
import { stringToBoolean } from '../../helpers';
import Form from '../common/forms/form';
import Loader from '../common/loader';
import { EntryLink } from './login';
import OrgDataEntry from './signup-steps/orgdata-entry';
import UserDataEntry from './signup-steps/userdata-entry';

const cookies = new Cookies();
const useStyles = makeStyles()(theme => ({
  background: {
    width: '100%',
    marginTop: -(50 + 45),
    height: `calc(100vh - ${theme.mixins.toolbar.minHeight}px)`,
    [`.${formControlClasses.root}`]: {
      marginTop: 0,
      marginBottom: theme.spacing(2)
    },
    '> div': {
      display: 'grid',
      gridTemplateColumns: 'minmax(min-content, 500px)',
      placeContent: 'center'
    }
  },
  locationSelect: { minWidth: 150, alignSelf: 'flex-start' },
  locationIcon: { marginLeft: theme.spacing(1.5), transform: 'scale(0.75)' },
  userData: {
    display: 'grid',
    justifyContent: 'center',
    alignContent: 'center',
    '> button': { justifySelf: 'flex-start' }
  },
  orgData: { display: 'grid', placeContent: 'center', gridTemplateColumns: 'min-content' },
  promo: {
    background: theme.palette.grey[400],
    gridTemplateRows: 'min-content min-content min-content',
    padding: '80px 0'
  },
  logo: { marginLeft: '5vw', marginTop: 45, maxHeight: 50 }
}));

export const Signup = () => {
  const [step, setStep] = useState(1);
  const [email, setEmail] = useState('');
  const [emailVerified, setEmailVerified] = useState(false);
  const [loading, setLoading] = useState(false);
  const [oauthProvider, setOauthProvider] = useState(undefined);
  const [oauthId, setOauthId] = useState('');
  const [marketing, setMarketing] = useState(false);
  const [organization, setOrganization] = useState('');
  const [tos, setTos] = useState(false);
  const [redirectOnLogin, setRedirectOnLogin] = useState(false);
  const [captchaTimestamp, setCaptchaTimestamp] = useState(0);
  const [recaptcha, setRecaptcha] = useState('');
  const [location, setLocation] = useState(locations.us.key);
  const { campaign = '' } = useParams();
  const currentUserId = useSelector(state => state.users.currentUserId);
  const recaptchaSiteKey = useSelector(state => state.app.recaptchaSiteKey);
  const dispatch = useDispatch();
  const { classes } = useStyles();

  const dispatchedSetSnackbar = useCallback(message => dispatch(setSnackbar(message)), [dispatch]);

  useEffect(() => {
    const usedOauthProvider = cookies.get('oauth');
    if (usedOauthProvider) {
      setOauthProvider(usedOauthProvider);
      setOauthId(`${cookies.get('externalID')}`);
      setEmail(cookies.get('email'));
      setEmailVerified(stringToBoolean(cookies.get('emailVerified')));
      setStep(2);
    }
  }, []);

  useEffect(() => {
    if (currentUserId) {
      dispatchedSetSnackbar('');
      setRedirectOnLogin(true);
    }
  }, [currentUserId, dispatchedSetSnackbar]);

  const handleSignup = formData => {
    if (recaptchaSiteKey !== '' && recaptcha === '') {
      return setSnackbar('Please complete the reCAPTCHA test before proceeding!', TIMEOUTS.fiveSeconds, '');
    }
    setLoading(true);
    const { name, marketing, password, ...remainder } = formData;
    const actualEmail = formData.email != null ? formData.email : email;
    const credentials = oauthProvider ? { email: actualEmail, login: { [oauthProvider]: oauthId } } : { email: actualEmail, password };
    const signup = {
      ...remainder,
      ...credentials,
      'g-recaptcha-response': recaptcha || 'empty',
      campaign,
      emailVerified,
      location,
      marketing: marketing == 'true',
      organization: name,
      plan: 'enterprise',
      ts: captchaTimestamp
    };
    return dispatch(createOrganizationTrial(signup)).catch(() => {
      setStep(1);
      setOrganization(formData.name);
      setTos(formData.tos);
      setMarketing(formData.marketing);
      setLoading(false);
    });
  };

  const onProgessClick = () => {
    setEmailVerified(true);
    setStep(2);
  };

  if (redirectOnLogin) {
    return <Navigate to="/" replace />;
  }

  const steps = {
    1: <UserDataEntry classes={classes} onProgessClick={onProgessClick} />,
    2: (
      <OrgDataEntry
        classes={classes}
        emailVerified={emailVerified}
        location={location}
        recaptchaSiteKey={recaptchaSiteKey}
        setCaptchaTimestamp={setCaptchaTimestamp}
        setLocation={setLocation}
        setRecaptcha={setRecaptcha}
      />
    )
  };
  const isStarting = step === 1;
  return (
    <>
      <LoginLogo className={classes.logo} />
      <div className={`${classes.background} ${isStarting ? 'two-columns' : classes.orgData}`} id="signup-box">
        <div>
          <Form
            buttonColor="primary"
            defaultValues={{ email: '', tos: false, marketing: false, name: '', captcha: '' }}
            initialValues={{ email, tos, marketing, name: organization, captcha: '' }}
            onSubmit={handleSignup}
            showButtons={!(isStarting || loading)}
            submitLabel={isStarting ? 'Sign up' : 'Complete signup'}
          >
            {loading ? <Loader show style={{ marginTop: '40vh' }} /> : steps[step]}
          </Form>
          {!loading && <EntryLink target="login" />}
        </div>
        {isStarting && (
          <div className={classes.promo}>
            <h2>Connect up to 10 devices free for 12 months – no credit card required.</h2>
            <p>
              Mender provides a complete over-the-air update infrastructure for all device software. Whether in the field or the factory, you can remotely and
              easily manage device software without the need for manual labor.
            </p>
            <div className="svg-container margin-top">
              <SignupHero />
            </div>
          </div>
        )}
      </div>
    </>
  );
};

export default Signup;
