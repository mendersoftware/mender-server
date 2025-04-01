// Copyright 2025 Northern.tech AS
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
import { useCallback, useEffect, useRef, useState } from 'react';
import { FormProvider, useForm } from 'react-hook-form';
import { Link } from 'react-router-dom';

import { Edit as EditIcon } from '@mui/icons-material';
import { Alert, Button, Collapse, IconButton, InputAdornment, Theme } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import FormCheckbox from '@northern.tech/common-ui/forms/FormCheckbox';
import PasswordInput from '@northern.tech/common-ui/forms/PasswordInput';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import { HELPTOOLTIPS, MenderHelpTooltip } from '@northern.tech/helptips/HelpTooltips';
import { TIMEOUTS } from '@northern.tech/store/constants';
import { toggle } from '@northern.tech/utils/helpers';

const useStyles = makeStyles()((theme: Theme) => ({
  alert: { fontWeight: theme.typography.fontWeightMedium },
  gapRemover: { marginTop: theme.spacing(-1.5) },
  formWrapper: { display: 'flex', flexDirection: 'column', gap: theme.spacing(1.5), position: 'relative', '.required:after': { content: 'none' } },
  passwordWrapper: { '.password-wrapper': { gridTemplateColumns: '1fr' } },
  tfaNote: { maxWidth: 300 }
}));

interface LoginFormState {
  email: string;
  noExpiry: boolean;
  password: string;
  token2fa: string;
}

export const LoginForm = ({ isHosted, isEnterprise, onSubmit }) => {
  const [emailEditingDisabled, setEmailEditingDisabled] = useState<boolean>(false);
  const [showPassword, setShowPassword] = useState<boolean>(false);
  const [has2FA, setHas2FA] = useState<boolean>(false);
  const [hasError, setHasError] = useState<boolean>(false);
  const twoFARef = useRef<HTMLInputElement | undefined>(undefined);
  // can't use the existing Form component due to the validation mode that's targeted
  const methods = useForm<LoginFormState>({ mode: 'onSubmit', defaultValues: { email: '', password: '', noExpiry: false, token2fa: '' } });
  const { handleSubmit, trigger, setFocus } = methods;
  const isOsInstallation = !(isEnterprise || isHosted);

  const { classes } = useStyles();

  useEffect(() => {
    setShowPassword(isOsInstallation);
    setFocus('email');
  }, [isOsInstallation, setFocus]);

  useEffect(() => {
    if (isOsInstallation) {
      return;
    }
    setEmailEditingDisabled(showPassword);
  }, [isOsInstallation, showPassword]);

  const maybeShowPassword = useCallback(async () => {
    const isValidEmail = await trigger('email');
    setShowPassword(isValidEmail);
  }, [trigger]);

  const onSubmitClick = (formData: LoginFormState) =>
    onSubmit(formData).catch(err => {
      // don't reset the state once it was set - thus not setting `has2FA` solely based on the existence of 2fa in the error
      if (err?.error?.includes('2fa')) {
        setShowPassword(true);
        return setHas2FA(true);
      } else if (!showPassword) {
        return maybeShowPassword();
      }
      setHasError(true);
    });

  const onShowPassword = () => {
    if (isOsInstallation) {
      return setFocus('email');
    }
    setFocus('password');
  };

  const onEditEmailClick = () => {
    setEmailEditingDisabled(toggle);
    setShowPassword(false);
  };

  const onShow2fa = () => {
    setFocus('token2fa');
    setTimeout(() => window.dispatchEvent(new Event('resize')), TIMEOUTS.oneSecond); // since there is no state change associated here, the timeout can be skipped from clearing on unmount
  };

  return (
    <FormProvider {...methods}>
      <form autoComplete="off" className={classes.formWrapper} noValidate onSubmit={handleSubmit(onSubmitClick)}>
        {hasError && (
          <Alert className={classes.alert} severity="error">
            {has2FA ? 'Incorrect email address, password and / or two factor authentication code.' : 'Incorrect email address and / or password.'}
          </Alert>
        )}
        <TextInput
          disabled={emailEditingDisabled}
          hint="Your email"
          label="Your email"
          id="email"
          required
          validations="isLength:1,isEmail,trim"
          InputProps={{
            endAdornment: emailEditingDisabled ? (
              <InputAdornment position="end">
                <IconButton onClick={onEditEmailClick} size="large">
                  <EditIcon />
                </IconButton>
              </InputAdornment>
            ) : undefined
          }}
        />
        <Collapse className={showPassword ? '' : classes.gapRemover} in={showPassword} onEntering={onShowPassword} timeout={isOsInstallation ? 0 : 'auto'}>
          <PasswordInput className={classes.passwordWrapper} id="password" label="Password" required={isOsInstallation} />
        </Collapse>
        {isHosted && (
          <div>
            <Link to="/password">Forgot your password?</Link>
          </div>
        )}
        <Collapse className={has2FA ? '' : classes.gapRemover} in={has2FA} onEntering={onShow2fa}>
          <TextInput
            controlRef={twoFARef}
            hint="Two Factor Authentication Code"
            id="token2fa"
            label="Two Factor Authentication Code"
            required={has2FA}
            validations="isLength:6,isNumeric"
          />
        </Collapse>
        <FormCheckbox className="margin-top-none" id="noExpiry" label="Stay logged in" />
        <Button className="full-width" variant="contained" type="submit">
          Next
        </Button>
        {has2FA && twoFARef.current && (
          <MenderHelpTooltip
            id={HELPTOOLTIPS.twoFactorNote.id}
            disableHoverListener={false}
            placement="right"
            className="absolute"
            style={{ top: twoFARef.current.parentElement.parentElement.offsetTop + twoFARef.current.parentElement.parentElement.offsetHeight / 4, right: -35 }}
            contentProps={{ className: classes.tfaNote }}
          />
        )}
      </form>
    </FormProvider>
  );
};
