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
import React, { useEffect } from 'react';
import { FormProvider, useForm } from 'react-hook-form';

import { Button } from '@mui/material';

import validator from 'validator';

const getErrorMsg = (validateMethod, args) => {
  switch (validateMethod) {
    case 'isLength':
      if (args[0] === 1) {
        return 'This field is required';
      } else if (args[0] > 1) {
        return `Must be at least ${args[0]} characters long`;
      }
      break;
    case 'isAlpha':
      return 'This field must contain only letters';
    case 'isAlphanumeric':
      return 'This field must contain only letters or numbers';
    case 'isNumeric':
      return 'Please enter a valid code';
    case 'isEmail':
      return 'Please enter a valid email address';
    case 'isUUID':
      return 'Please enter a valid ID';
    case 'isNot':
      if (args[0] === args[1]) {
        return `This field should have a value other than ${args[0]}`;
      }
      break;
    default:
      return 'There is an error with this field';
  }
};

const tryApplyValidationEntry = (value, validations = [], validationResults = []) => {
  const validation = validations.shift();
  if (!validation) {
    return validationResults.pop();
  }
  let args = validation.split(':');
  const validateMethod = args.shift();
  // We use JSON.parse to convert the string values passed to the
  // correct type. Ex. 'isLength:1' will make '1' actually a number
  args = args.map(arg => JSON.parse(JSON.stringify(arg)));

  const tmpArgs = args;
  // We then merge two arrays, ending up with the value
  // to pass first, then options, if any. ['valueFromInput', 5]
  args = [value].concat(args);
  try {
    // So the next line of code is actually:
    // validator.isLength('valueFromInput', 5)
    if (!validator[validateMethod].apply(validator, args)) {
      return tryApplyValidationEntry(value, validations, [...validationResults, { errortext: getErrorMsg(validateMethod, tmpArgs), isValid: false }]);
    }
  } catch {
    const errortext = getErrorMsg(validateMethod, args) || '';
    return tryApplyValidationEntry(value, validations, [...validationResults, { errortext, isValid: !errortext }]);
  }
  return { errortext: '', isValid: true };
};

const tryApplyValidations = (value, validations, initialValidationResult) =>
  validations.split(',').reduce((accu, validation) => {
    if (!accu.isValid || !validation) {
      return accu;
    }
    const alternatives = validation.split('||');
    return tryApplyValidationEntry(value, alternatives, [accu]);
  }, initialValidationResult);

const runPasswordValidations = ({ required, value, validations, isValid, errortext }) => {
  if (required && !value) {
    return { isValid: false, errortext: 'Password is required' };
  } else if (required || value) {
    isValid = tryApplyValidations(value, validations, { isValid, errortext }).isValid;
    return { isValid, errortext: !isValid ? 'Password too weak' : errortext };
  }
  return { isValid, errortext };
};

export const runValidations = ({ required, value, id, validations }) => {
  let isValid = true;
  let errortext = '';
  if (id && id.includes('password')) {
    return runPasswordValidations({ required, value, validations, isValid, errortext });
  } else {
    if (value || required) {
      return tryApplyValidations(validations.includes('trim') ? value.trim() : value, validations, { isValid, errortext });
    }
  }
  return { isValid, errortext };
};

export const Form = ({
  autocomplete,
  buttonColor,
  children,
  className = '',
  defaultValues = {},
  handleCancel,
  id,
  initialValues = {},
  onSubmit,
  showButtons,
  submitLabel
}) => {
  const methods = useForm({ mode: 'onChange', defaultValues });
  const {
    handleSubmit,
    formState: { isValid },
    setValue
  } = methods;

  useEffect(() => {
    Object.entries(initialValues).map(([key, value]) => setValue(key, value));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(initialValues), setValue]);

  return (
    <FormProvider {...methods}>
      <form autoComplete={autocomplete} className={className} id={id} noValidate onSubmit={handleSubmit(onSubmit)}>
        {children}
        {!!showButtons && (
          <div className="flexbox" style={{ justifyContent: 'flex-end', height: 'min-content', marginTop: 32 }}>
            {!!handleCancel && (
              <Button key="cancel" onClick={handleCancel} style={{ marginRight: 10, display: 'inline-block' }}>
                Cancel
              </Button>
            )}
            <Button variant="contained" type="submit" disabled={!isValid} color={buttonColor}>
              {submitLabel}
            </Button>
          </div>
        )}
      </form>
    </FormProvider>
  );
};

export default Form;
