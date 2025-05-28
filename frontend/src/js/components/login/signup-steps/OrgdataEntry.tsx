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
import ReCAPTCHA from 'react-google-recaptcha';
import { Controller, useFormContext } from 'react-hook-form';

import { MenuItem, Select } from '@mui/material';

import DocsLink from '@northern.tech/common-ui/DocsLink';
import Form from '@northern.tech/common-ui/forms/Form';
import FormCheckbox from '@northern.tech/common-ui/forms/FormCheckbox';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import { locations } from '@northern.tech/store/constants';

export type OrgData = {
  captcha: string | null;
  location: string;
  marketing?: boolean;
  name: string;
  tos: boolean;
};

const OrgDataContent = ({
  classes,
  emailVerified,
  recaptchaSiteKey = '',
  setCaptchaTimestamp
}: Pick<OrgDataProps, 'classes' | 'emailVerified' | 'recaptchaSiteKey' | 'setCaptchaTimestamp'>) => {
  const { control, register, setValue, trigger } = useFormContext();
  const captchaFieldName = 'captcha';

  const handleCaptchaChange = value => {
    setCaptchaTimestamp(new Date().getTime());
    setValue(captchaFieldName, value ? value : '');
    trigger(captchaFieldName);
  };

  return (
    <>
      <TextInput hint="Company or organization name *" label="Company or organization name *" id="name" required validations="isLength:1,trim" />
      {!emailVerified && <TextInput hint="Email *" label="Email *" id="email" required validations="isLength:1,isEmail,trim" />}
      <div className={classes.locationSelect}>
        <div className="flexbox center-aligned slightly-smaller margin-bottom-x-small">
          <p className="margin-bottom-none margin-top-none muted" style={{ marginRight: 4 }}>
            Choose a hosting region for your account.
          </p>
          <DocsLink path="general/hosted-mender-regions" title="Learn more" />
        </div>
        <Controller
          name="location"
          control={control}
          render={({ field }) => (
            <Select
              renderValue={selected => {
                const { icon: Icon, title } = locations[selected];
                return (
                  <div className="flexbox center-aligned">
                    {title} <Icon className={classes.locationIcon} />
                  </div>
                );
              }}
              {...field}
            >
              {Object.entries(locations).map(([key, { icon: Icon, title }]) => (
                <MenuItem key={key} value={key}>
                  {title} <Icon className={classes.locationIcon} />
                </MenuItem>
              ))}
            </Select>
          )}
        />
      </div>
      <FormCheckbox
        id="tos"
        label={
          <label htmlFor="tos">
            By checking this you agree to our {/* eslint-disable-next-line react/jsx-no-target-blank */}
            <a href="https://northern.tech/legal/hosted-mender-agreement-northern-tech-as.pdf" target="_blank" rel="noopener">
              Terms of service
            </a>{' '}
            and {/* eslint-disable-next-line react/jsx-no-target-blank */}
            <a href="https://northern.tech/legal/privacy-policy" target="_blank" rel="noopener">
              Privacy Policy
            </a>{' '}
            *
          </label>
        }
        required={true}
      />
      <FormCheckbox
        id="marketing"
        label="By checking this you agree that we can send you occasional email updates about Mender. You can unsubscribe from these emails at any time"
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
    <Form
      className={`flexbox column centered ${classes.orgData}`}
      id="signup-org-data"
      buttonColor="primary"
      defaultValues={defaultValues}
      initialValues={initialValues}
      onSubmit={handleSignup}
      showButtons={!loading}
      submitLabel="Complete signup"
    >
      <h1>Setting up your Account</h1>
      <h2 className="margin-bottom-large">
        To finish creating your account,
        <br />
        please fill in a few details
      </h2>
      <OrgDataContent classes={classes} {...remainder} />
    </Form>
  );
};

export default OrgDataEntry;
