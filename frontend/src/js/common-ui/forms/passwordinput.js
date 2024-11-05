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
import React, { useEffect, useRef, useState } from 'react';
import { Controller, useFormContext, useWatch } from 'react-hook-form';

import { CheckCircle as CheckIcon, Visibility as VisibilityIcon, VisibilityOff as VisibilityOffIcon } from '@mui/icons-material';
import { Button, FormControl, FormHelperText, IconButton, Input, InputAdornment, InputLabel } from '@mui/material';

import { TIMEOUTS } from '@northern.tech/store/constants';
import { toggle } from '@northern.tech/utils/helpers';
import copy from 'copy-to-clipboard';
import generator from 'generate-password';

import { runValidations } from './form';

const PasswordGenerateButtons = ({ clearPass, edit, generatePass }) => (
  <div className="pass-buttons">
    <Button color="primary" onClick={generatePass}>
      Generate
    </Button>
    {edit ? <Button onClick={clearPass}>Cancel</Button> : null}
  </div>
);

const SCORE_THRESHOLD = 3;

const PasswordGenerationControls = ({ score, feedback }) => (
  <>
    <div className="help-text" id="pass-strength">
      Strength: <meter max={4} min={0} value={score} high={3.9} optimum={4} low={2.5} />
      {score > SCORE_THRESHOLD ? <CheckIcon className="fadeIn green" style={{ height: 18, marginTop: -3, marginBottom: -3 }} /> : null}
    </div>
    {!!feedback.length && (
      <p className="help-text">
        {feedback.map((message, index) => (
          <React.Fragment key={`feedback-${index}`}>
            <span>{message}</span>
            <br />
          </React.Fragment>
        ))}
      </p>
    )}
  </>
);

export const PasswordInput = ({
  autocomplete,
  className,
  control,
  create,
  defaultValue,
  disabled,
  edit,
  generate,
  id,
  InputLabelProps = {},
  label,
  onClear,
  placeholder,
  required,
  validations = ''
}) => {
  const [score, setScore] = useState(0);
  const [visible, setVisible] = useState(false);
  const [copied, setCopied] = useState(false);
  const [feedback, setFeedback] = useState([]);
  const [confirmationId] = useState(id.includes('current') ? '' : ['password', 'password_confirmation'].find(thing => thing !== id));
  const timer = useRef();
  const {
    clearErrors,
    formState: { errors },
    setError,
    setValue,
    trigger,
    getValues
  } = useFormContext();
  const confirmation = useWatch({ name: confirmationId });
  const errorKey = `${id}`;
  const { message } = errors[errorKey] ?? {};

  useEffect(() => {
    if (confirmationId === 'password' && !message) {
      trigger(confirmationId);
    }
  }, [confirmationId, message, trigger]);

  useEffect(() => {
    return () => {
      clearTimeout(timer.current);
    };
  });

  const clearPassClick = () => {
    setValue(id, '');
    onClear();
    setCopied(false);
  };

  const generatePassClick = () => {
    const password = generator.generate({ length: 16, numbers: true });
    setValue(id, password);
    const form = getValues();
    if (form.hasOwnProperty(`${id}_confirmation`)) {
      setValue(`${id}_confirmation`, password);
    }
    copy(password);
    setCopied(true);
    setVisible(true);
    timer.current = setTimeout(() => setCopied(false), TIMEOUTS.fiveSeconds);
    trigger();
  };

  const validate = async (value = '') => {
    let { isValid, errortext } = runValidations({ id, required, validations, value });
    if (confirmation && value !== confirmation) {
      isValid = false;
      errortext = 'The passwords you provided do not match, please check again.';
    }
    if (isValid) {
      clearErrors(errorKey);
    } else {
      setError(errorKey, { type: 'validate', message: errortext });
    }
    if (!create || (!required && !value)) {
      return isValid;
    }
    const { default: zxcvbn } = await import(/* webpackChunkName: "zxcvbn" */ 'zxcvbn');
    const strength = zxcvbn(value);
    const score = strength.score;
    setFeedback(strength.feedback.suggestions || []);
    setScore(score);
    return score > SCORE_THRESHOLD && isValid;
  };

  return (
    <div className={className}>
      <div className="password-wrapper">
        <Controller
          name={id}
          control={control}
          rules={{ required, validate }}
          render={({ field: { value, onChange, onBlur, ref }, fieldState: { error } }) => (
            <FormControl className={required ? 'required' : ''} error={Boolean((error || errors[errorKey])?.message)} style={{ width: 400 }}>
              <InputLabel htmlFor={id} {...InputLabelProps}>
                {label}
              </InputLabel>
              <Input
                autoComplete={autocomplete}
                id={id}
                name={id}
                type={visible ? 'text' : 'password'}
                defaultValue={defaultValue}
                placeholder={placeholder}
                value={value ?? ''}
                disabled={disabled}
                inputRef={ref}
                required={required}
                onChange={({ target: { value } }) => {
                  setValue(id, value);
                  onChange(value);
                }}
                onBlur={onBlur}
                endAdornment={
                  <InputAdornment position="end">
                    <IconButton onClick={() => setVisible(toggle)} size="large">
                      {visible ? <VisibilityIcon /> : <VisibilityOffIcon />}
                    </IconButton>
                  </InputAdornment>
                }
              />
              <FormHelperText>{(errors[errorKey] || error)?.message}</FormHelperText>
            </FormControl>
          )}
        />
        {generate && !required && <PasswordGenerateButtons clearPass={clearPassClick} edit={edit} generatePass={generatePassClick} />}
      </div>
      {copied ? <div className="green fadeIn margin-bottom-small">Copied to clipboard</div> : null}
      {create && (
        <>
          <PasswordGenerationControls feedback={feedback} score={score} />
          {generate && required && <PasswordGenerateButtons clearPass={clearPassClick} edit={edit} generatePass={generatePassClick} />}
        </>
      )}
    </div>
  );
};

export default PasswordInput;
