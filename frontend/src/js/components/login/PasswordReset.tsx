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
import { useState } from 'react';
import { useDispatch } from 'react-redux';
import { useParams } from 'react-router-dom';

import Form from '@northern.tech/common-ui/forms/Form';
import PasswordInput from '@northern.tech/common-ui/forms/PasswordInput';
import { passwordResetComplete } from '@northern.tech/store/thunks';

import { PasswordScreenContainer } from './Password';

export const PasswordReset = () => {
  const [confirm, setConfirm] = useState(false);
  const { secretHash } = useParams();
  const dispatch = useDispatch();

  const handleSubmit = formData =>
    dispatch(passwordResetComplete({ secretHash, newPassword: formData.password }))
      .unwrap()
      .then(() => setConfirm(true));

  return (
    <PasswordScreenContainer title="Change your password">
      {confirm ? (
        <p className="margin-bottom align-center">Your password has been updated.</p>
      ) : (
        <>
          <p className="margin-bottom align-center">
            You requested to change your password.
            <br />
            Enter a new, secure password of your choice below.
          </p>
          <Form showButtons={true} onSubmit={handleSubmit} submitLabel="Save password">
            <PasswordInput id="password" label="Password *" validations="isLength:8" create={true} generate={false} required={true} />
            <PasswordInput id="password_confirmation" label="Confirm password *" validations="isLength:8" required={true} />
          </Form>
        </>
      )}
    </PasswordScreenContainer>
  );
};

export default PasswordReset;
