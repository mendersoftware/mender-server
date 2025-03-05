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
import { useEffect, useRef, useState } from 'react';
import { Form, Link } from 'react-router-dom';

import { Checkbox, Collapse, FormControlLabel, Theme } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import PasswordInput from '@northern.tech/common-ui/forms/PasswordInput';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import { HELPTOOLTIPS, MenderHelpTooltip } from '@northern.tech/helptips/HelpTooltips';

const useStyles = makeStyles()((theme: Theme) => ({
  form: { maxWidth: 400 },
  link: { marginTop: theme.spacing(0.5) },
  tfaTip: { position: 'absolute', right: -120 },
  tfaNote: { maxWidth: 300 }
}));

export const LoginForm = ({ isHosted, isEnterprise, onSubmit }) => {
  const [noExpiry, setNoExpiry] = useState<boolean>(false);
  const [showPassword, setShowPassword] = useState<boolean>(!(isEnterprise || isHosted));
  const [has2FA, setHas2FA] = useState<boolean>(false);
  const twoFARef = useRef<HTMLInputElement | undefined>(undefined);

  const { classes } = useStyles();

  useEffect(() => {
    if (isEnterprise || isHosted) {
      setShowPassword(false);
    }
  }, [isEnterprise, isHosted]);

  const onLoginClick = ({ noExpiry, ...loginData }) =>
    onSubmit({ ...loginData, stayLoggedIn: noExpiry }).catch(err => {
      // don't reset the state once it was set - thus not setting `has2FA` solely based on the existence of 2fa in the error
      if (err?.error?.includes('2fa')) {
        setHas2FA(true);
      }
    });

  const onNoExpiryClick = ({ target: { checked } }) => setNoExpiry(checked);

  return (
    <Form className={classes.form} showButtons={true} buttonColor="primary" onSubmit={onLoginClick} submitLabel="Log in">
      <TextInput hint="Your email" label="Your email" id="email" required={true} validations="isLength:1,isEmail,trim" />
      <Collapse in={showPassword}>
        <PasswordInput className="margin-top-small" id="password" label="Password" required={showPassword} />
      </Collapse>
      {isHosted ? (
        <div className="flexbox">
          <Link className={classes.link} to="/password">
            Forgot your password?
          </Link>
        </div>
      ) : (
        <div />
      )}
      <Collapse in={has2FA}>
        <TextInput
          hint="Two Factor Authentication Code"
          label="Two Factor Authentication Code"
          id="token2fa"
          validations="isLength:6,isNumeric"
          required={has2FA}
          controlRef={twoFARef}
        />
      </Collapse>
      <FormControlLabel control={<Checkbox color="primary" checked={noExpiry} onChange={onNoExpiryClick} />} label="Stay logged in" />
      {has2FA && twoFARef.current && (
        <MenderHelpTooltip
          id={HELPTOOLTIPS.twoFactorNote.id}
          disableHoverListener={false}
          placement="right"
          className={classes.tfaTip}
          style={{ top: twoFARef.current.parentElement.parentElement.offsetTop + twoFARef.current.parentElement.parentElement.offsetHeight / 2 }}
          contentProps={{ className: classes.tfaNote }}
        />
      )}
    </Form>
  );
};
