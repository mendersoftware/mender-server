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
import React from 'react';
import { useFormState, useWatch } from 'react-hook-form';

import { Button } from '@mui/material';

import PasswordInput from '../../common/forms/passwordinput';
import TextInput from '../../common/forms/textinput';
import { OAuthHeader } from '../login';

export const UserDataEntry = ({ classes, onProgessClick }) => {
  const { isValid } = useFormState();
  const email = useWatch({ name: 'email' });

  const handleKeyPress = ({ key }) => {
    if (key === 'Enter') {
      onProgessClick();
    }
  };

  return (
    <div className={classes.userData} onKeyPress={handleKeyPress}>
      <h1 className="flexbox centered">Create your account</h1>
      <OAuthHeader type="Sign up" />
      <TextInput hint="Email *" label="Email *" id="email" required={true} validations="isLength:1,isEmail,trim" />
      <PasswordInput
        id="password"
        label="Password *"
        validations={`isLength:8,isNot:${email}`}
        create={true}
        generate={false}
        required={true}
        className="margin-bottom-small"
      />
      <PasswordInput id="password_confirmation" label="Confirm password *" validations={`isLength:8,isNot:${email}`} required={true} />
      <Button className="margin-top" color="primary" disabled={!isValid} variant="contained" onClick={onProgessClick}>
        Sign up
      </Button>
    </div>
  );
};

export default UserDataEntry;
