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
import { FormProvider, useForm } from 'react-hook-form';

import { Button } from '@mui/material';

import PasswordInput from '@northern.tech/common-ui/forms/PasswordInput';
import TextInput from '@northern.tech/common-ui/forms/TextInput';

import { OAuthHeader } from '../Login';

const defaultValues = { email: '', password_confirmation: '', password: '' };

export type UserData = {
  email: string;
  password: string;
  password_confirmation: string;
};

export const UserDataEntry = ({ classes, onSubmit }) => {
  const methods = useForm({ mode: 'onSubmit', defaultValues });
  const { handleSubmit, trigger, watch } = methods;
  const email = watch('email');
  const password = watch('password');
  const passwordConfirmation = watch('password_confirmation');
  const isNotDefined = !(email && password && passwordConfirmation);

  const onFormBlur = ({ target: { id } }) => {
    if (id !== 'password_confirmation') {
      return;
    }
    return trigger();
  };

  const commonProps = {
    InputLabelProps: { size: 'medium' },
    InputProps: { size: 'medium' },
    required: true
  };

  return (
    <FormProvider {...methods}>
      <form className={classes.userData} noValidate onBlur={onFormBlur} onSubmit={handleSubmit(onSubmit)}>
        <h1 className="flexbox centered">Create your account</h1>
        <OAuthHeader type="Sign up" />
        <TextInput {...commonProps} hint="Email *" label="Email *" id="email" validations="isLength:1,isEmail,trim" />
        <PasswordInput
          {...commonProps}
          id="password"
          label="Password *"
          validations={`isLength:8,isNot:${email}`}
          create={true}
          generate={false}
          className="margin-bottom-small"
        />
        <PasswordInput {...commonProps} id="password_confirmation" label="Confirm password *" validations={`isLength:8,isNot:${email}`} />
        <Button variant="contained" type="submit" disabled={isNotDefined}>
          Sign up
        </Button>
      </form>
    </FormProvider>
  );
};

export default UserDataEntry;
