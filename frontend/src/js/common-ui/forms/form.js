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
import { makeStyles } from 'tss-react/mui';

import validator from 'validator';

const validationMethods = {
  isAlpha: 'This field must contain only letters',
  isAlphanumeric: 'This field must contain only letters or numbers',
  isEmail: 'Please enter a valid email address',
  isHexadecimal: 'The secret has to be entered as a hexadecimal string',
  isNumeric: 'Please enter a valid code',
  isURL: 'Please enter a valid URL',
  isUUID: 'Please enter a valid ID'
};

const getErrorMsg = (validateMethod, args) => {
  if (validationMethods[validateMethod]) {
    return validationMethods[validateMethod];
  }
  switch (validateMethod) {
    case 'isLength':
      if (Number(args[0]) === 1) {
        return 'This field is required';
      } else if (args[0] > 1) {
        return `Must be at least ${args[0]} characters long`;
      }
      break;
    case 'isAlphanumericLocator':
      if (args[0] && validator.matches(args[0], /^[a-zA-Z0-9_-]+$/)) {
        return '';
      } else {
        return 'This please only enter valid characters. Valid characters are a-z, A-Z, 0-9, _ and -';
      }
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

export const runValidations = ({ required, value, id, validations, wasMaybeTouched }) => {
  let isValid = true;
  let errortext = '';
  if (id && id.includes('password')) {
    return runPasswordValidations({ required, value, validations, isValid, errortext });
  } else {
    if (value || required || (wasMaybeTouched && validations.includes('isLength:1'))) {
      return tryApplyValidations(validations.includes('trim') ? value.trim() : value, validations, { isValid, errortext });
    }
  }
  return { isValid, errortext };
};

const useStyles = makeStyles()(theme => ({
  buttonWrapper: { display: 'flex', justifyContent: 'flex-end', height: 'min-content', marginTop: theme.spacing(4) },
  cancelButton: { marginRight: theme.spacing() }
}));

export const Form = ({
  autocomplete,
  buttonColor,
  children,
  className = '',
  classes = { buttonWrapper: '', cancelButton: '' },
  defaultValues = {},
  handleCancel,
  id,
  initialValues = {},
  onSubmit,
  showButtons,
  submitLabel
}) => {
  const { classes: internalClasses } = useStyles();
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
          <div className={`button-wrapper ${internalClasses.buttonWrapper} ${classes.buttonWrapper}`}>
            {!!handleCancel && (
              <Button className={`${internalClasses.cancelButton} ${classes.cancelButton}`} key="cancel" onClick={handleCancel}>
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
