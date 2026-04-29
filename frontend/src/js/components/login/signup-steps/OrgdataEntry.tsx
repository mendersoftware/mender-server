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
import type { Dispatch, SetStateAction } from 'react';
import { useEffect, useRef } from 'react';
import ReCAPTCHA from 'react-google-recaptcha';
import { useFormContext } from 'react-hook-form';

import { Typography } from '@mui/material';

import LinedHeader from '@northern.tech/common-ui/LinedHeader';
import { Link } from '@northern.tech/common-ui/Link';
import Form from '@northern.tech/common-ui/forms/Form';
import FormCheckbox from '@northern.tech/common-ui/forms/FormCheckbox';
import TextInput from '@northern.tech/common-ui/forms/TextInput';

export type OrgData = {
  captcha: string | null;
  location: string;
  marketing?: boolean;
  name: string;
  tos: boolean;
};

const OrgDataContent = ({
  emailVerified,
  recaptchaSiteKey = '',
  setCaptchaTimestamp
}: Pick<OrgDataProps, 'emailVerified' | 'recaptchaSiteKey' | 'setCaptchaTimestamp'>) => {
  const { register, setValue, trigger } = useFormContext();
  const inputRef = useRef<HTMLInputElement | undefined>();
  const captchaFieldName = 'captcha';

  useEffect(() => {
    if (inputRef.current) {
      inputRef.current.focus();
    }
  }, []);

  const handleCaptchaChange = value => {
    setCaptchaTimestamp(new Date().getTime());
    setValue(captchaFieldName, value ? value : '');
    trigger(captchaFieldName);
  };

  return (
    <>
      <Typography variant="subtitle1" className="margin-bottom-x-small">
        Organization name
      </Typography>
      <TextInput
        className="margin-bottom-small"
        controlRef={inputRef}
        id="name"
        InputLabelProps={{ size: 'medium' }}
        InputProps={{ size: 'medium' }}
        label="Name*"
        helperText="Set an organization name for your account"
        required
        requiredRendered={false}
        validations="isLength:1:256,trim"
      />
      {!emailVerified && <TextInput className="margin-bottom-small" hint="Email *" label="Email *" id="email" required validations="isLength:1,isEmail,trim" />}
      <FormCheckbox
        id="tos"
        label={
          <label htmlFor="tos" style={{ fontSize: 'smaller' }}>
            I have read and agreed to the Mender
            <Link href="https://northern.tech/legal/hosted-mender-agreement-northern-tech-as.pdf" external>
              Terms of service
            </Link>{' '}
            and
            <Link href="https://northern.tech/legal/privacy-policy" external>
              Privacy Policy
            </Link>{' '}
          </label>
        }
        required={true}
      />
      <FormCheckbox
        id="marketing"
        label={
          <label htmlFor="marketing" style={{ fontSize: 'smaller' }}>
            Keep me updated on Mender news and releases. You can unsubscribe at any time.
          </label>
        }
      />
      {recaptchaSiteKey && (
        <div className="margin-top">
          <input type="hidden" {...register(captchaFieldName, { required: 'reCAPTCHA is not completed' })} />
          <ReCAPTCHA sitekey={recaptchaSiteKey} onChange={handleCaptchaChange} />
        </div>
      )}
    </>
  );
};

const defaultValues: OrgData = { tos: false, location: '', marketing: false, name: '', captcha: '' };

type OrgDataProps = {
  classes: Record<string, string>;
  emailVerified: boolean;
  handleSignup: (formData: OrgData) => null;
  initialValues: OrgData;
  loading: boolean;
  recaptchaSiteKey: string;
  setCaptchaTimestamp: Dispatch<SetStateAction<number>>;
};

export const OrgDataEntry = (props: OrgDataProps) => {
  const { classes, initialValues, loading, handleSignup, ...remainder } = props;
  return (
    <>
      <Form
        className={classes.orgData}
        id="signup-org-data"
        defaultValues={defaultValues}
        initialValues={initialValues}
        onSubmit={handleSignup}
        showButtons={!loading}
        submitLabel="Complete sign-up"
      >
        <Typography variant="h4" className="flexbox centered margin-bottom-medium">
          Sign up for Mender
        </Typography>
        <Typography className="margin-bottom-small">Complete the options below to finish creating your Mender account</Typography>
        <OrgDataContent {...remainder} />
      </Form>
      <LinedHeader className="margin-top-large flexbox centered" heading="OR" />
    </>
  );
};

export default OrgDataEntry;
