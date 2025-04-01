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
import { FormProvider, useForm } from 'react-hook-form';
import { useDispatch, useSelector } from 'react-redux';
import { Navigate, useParams } from 'react-router-dom';

import { Button, formControlClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import Loader from '@northern.tech/common-ui/Loader';
import storeActions from '@northern.tech/store/actions';
import { TIMEOUTS, locations } from '@northern.tech/store/constants';
import { getRecaptchaKey } from '@northern.tech/store/selectors';
import { createOrganizationTrial } from '@northern.tech/store/thunks';
import { stringToBoolean } from '@northern.tech/utils/helpers';
import Cookies from 'universal-cookie';

import LoginLogo from '../../../assets/img/loginlogo.svg';
import SignupHero from '../../../assets/img/signuphero.svg';
import { EntryLink } from './Login';
import { OrgDataEntry } from './signup-steps/OrgdataEntry';
import { UserDataEntry } from './signup-steps/UserdataEntry';

const { setSnackbar } = storeActions;

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

const getCurrentLocation = (location: Location): string => {
  const currentLocation = Object.values(locations).find(value => [`staging.${value.location}`, value.location].includes(location.hostname));
  return currentLocation ? currentLocation.key : locations.us.key;
};

const defaultValues = { email: '', tos: false, marketing: false, name: '', location: '', captcha: '', password_confirmation: '', password: '' };

export const Signup = () => {
  const [isStarting, setIsStarting] = useState(true);
  const [emailVerified, setEmailVerified] = useState(false);
  const [loading, setLoading] = useState(false);
  const [oauthProvider, setOauthProvider] = useState(undefined);
  const [oauthId, setOauthId] = useState('');
  const [redirectOnLogin, setRedirectOnLogin] = useState(false);
  const [captchaTimestamp, setCaptchaTimestamp] = useState(0);
  const [location, setLocation] = useState<string>(getCurrentLocation(window.location));
  const [initialValues, setInitialValues] = useState({ ...defaultValues });
  const { campaign = '' } = useParams();
  const currentUserId = useSelector(state => state.users.currentUserId);
  const recaptchaSiteKey = useSelector(getRecaptchaKey);
  const dispatch = useDispatch();
  const { classes } = useStyles();

  const methods = useForm({ mode: 'onSubmit', defaultValues });
  const { handleSubmit, setValue, trigger, watch, getFieldState } = methods;
  const password = watch('password');
  const email = watch('email');
  const passwordConfirmation = watch('password_confirmation');
  const isNotDefined = !(email && password && passwordConfirmation);

  const dispatchedSetSnackbar = useCallback(message => dispatch(setSnackbar(message)), [dispatch]);

  useEffect(() => {
    const usedOauthProvider = cookies.get('oauth');
    if (usedOauthProvider) {
      setOauthProvider(usedOauthProvider);
      setOauthId(`${cookies.get('externalID')}`);
      setFormValues(current => ({ ...current, email: cookies.get('email') }));
      setEmailVerified(stringToBoolean(cookies.get('emailVerified')));
      setIsStarting(false);
    }
    const location = getCurrentLocation(window.location);
    setFormValues(current => ({ ...current, location }));
  }, []);

  useEffect(() => {
    Object.entries(initialValues).forEach(([key, value]) => setValue(key, value));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(initialValues), setValue]);

  useEffect(() => {
    if (currentUserId) {
      dispatchedSetSnackbar('');
      setRedirectOnLogin(true);
    }
  }, [currentUserId, dispatchedSetSnackbar]);

  const handleSignup = async formData => {
    if (isStarting) {
      await trigger();
      const { invalid: emailInvalid } = getFieldState('email');
      const { invalid: passwordInvalid } = getFieldState('password');
      const { invalid: passwordConfirmationInvalid } = getFieldState('password_confirmation');
      if (emailInvalid || passwordInvalid || passwordConfirmationInvalid) {
        return;
      }
      return onProgressClick();
    }
    if (recaptchaSiteKey !== '' && formData.captcha === '') {
      return setSnackbar({ message: 'Please complete the reCAPTCHA test before proceeding!', autoHideDuration: TIMEOUTS.fiveSeconds, action: '' });
    }
    setLoading(true);
    const { email, name, marketing, password, captcha, ...remainder } = formData;
    const credentials = oauthProvider ? { email, login: { [oauthProvider]: oauthId } } : { email, password };
    const signup = {
      ...remainder,
      ...credentials,
      'g-recaptcha-response': captcha || 'empty',
      campaign,
      emailVerified,
      location,
      marketing: marketing == 'true',
      organization: name,
      plan: 'enterprise',
      ts: captchaTimestamp
    };
    return dispatch(createOrganizationTrial(signup)).catch(() => {
      setInitialValues({ ...formData, captcha: '' });
      setIsStarting(true);
      setLoading(false);
    });
  };

  const onProgressClick = async () => {
    const canProgress = await trigger();
    if (canProgress) {
      setEmailVerified(true);
      setIsStarting(false);
    }
  };

  const onFormBlur = ({ target: { id } }) => {
    if (id !== 'password_confirmation') {
      return;
    }
    return trigger();
  };

  if (redirectOnLogin) {
    return <Navigate to="/" replace />;
  }

  return (
    <>
      <LoginLogo className={classes.logo} />
      <div className={`${classes.background} ${isStarting ? 'two-columns' : classes.orgData}`} id="signup-box">
        <div>
          <FormProvider {...methods}>
            <form noValidate onSubmit={handleSubmit(handleSignup)} onBlur={onFormBlur}>
              {loading ? (
                <Loader show style={{ marginTop: '40vh' }} />
              ) : isStarting ? (
                <>
                  <UserDataEntry classes={classes} onProgessClick={onProgressClick} />
                  <div className={`flexbox align-self-end margin-top`}>
                    <Button variant="contained" disabled={isNotDefined} onClick={onProgressClick}>
                      Sign up
                    </Button>
                  </div>
                </>
              ) : (
                <>
                  <OrgDataEntry
                    classes={classes}
                    emailVerified={emailVerified}
                    location={location}
                    recaptchaSiteKey={recaptchaSiteKey}
                    setCaptchaTimestamp={setCaptchaTimestamp}
                    setLocation={setLocation}
                  />
                  <div className={`flexbox align-self-end margin-top`}>
                    <Button variant="contained" type="submit">
                      Complete signup
                    </Button>
                  </div>
                </>
              )}
            </form>
          </FormProvider>
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
