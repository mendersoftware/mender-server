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
import { useCallback, useEffect, useRef, useState } from 'react';
import { Controller, useFormContext, useWatch } from 'react-hook-form';

import { Check as CheckIcon, Visibility as VisibilityIcon, VisibilityOff as VisibilityOffIcon, WarningAmber as WarningIcon } from '@mui/icons-material';
import {
  Button,
  FormControl,
  FormHelperText,
  IconButton,
  InputAdornment,
  InputLabel,
  OutlinedInput,
  inputLabelClasses,
  outlinedInputClasses
} from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { TIMEOUTS } from '@northern.tech/store/constants';
import { toggle } from '@northern.tech/utils/helpers';
import copy from 'copy-to-clipboard';
import generator from 'generate-password-browser';

import { runValidations } from './Form';
import type { CommonTextInputProps } from './TextInput';
import { checkPasswordStrength } from './passwordStrength';

const PasswordGenerateButtons = ({
  clearPass,
  edit,
  generatePass,
  disabled
}: {
  clearPass: () => void;
  disabled?: boolean;
  edit?: boolean;
  generatePass: () => void;
}) => (
  <div className="pass-buttons">
    <Button onClick={generatePass} disabled={disabled}>
      Generate
    </Button>
    {edit ? <Button onClick={clearPass}>Cancel</Button> : null}
  </div>
);

const SCORE_THRESHOLD = 3;
const STRONG_SCORE = 4;

const useStyles = makeStyles()(theme => ({
  icon: { marginRight: theme.spacing(0.5) },
  success: {
    [`& .${outlinedInputClasses.notchedOutline}`]: { borderColor: theme.palette.success.main },
    [`& .${inputLabelClasses.root}`]: { color: theme.palette.success.main }
  }
}));

type PasswordInputProps = {
  create?: boolean;
  defaultValue?: string;
  edit?: boolean;
  generate?: boolean;
  id: string;
  onClear?: () => void;
  placeholder?: string;
} & Partial<CommonTextInputProps>;

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
  InputProps = {},
  label,
  onClear,
  placeholder,
  required,
  validations = '',
  width = 400
}: PasswordInputProps) => {
  const { classes } = useStyles();
  const [visible, setVisible] = useState(false);
  const [copied, setCopied] = useState(false);
  const [strong, setStrong] = useState(false);
  const [warningIcon, setWarningIcon] = useState(false);
  const [confirmationId] = useState(id.includes('current') ? '' : ['password', 'password_confirmation'].find(thing => thing !== id));
  const timer = useRef();
  const {
    formState: { errors },
    setValue,
    trigger,
    getValues
  } = useFormContext();
  const confirmation = useWatch({ name: confirmationId });
  const confirmationRef = useRef(confirmation);
  confirmationRef.current = confirmation;
  const errorKey = id;

  const validate = useCallback(
    async (value = '') => {
      if ((!validations && !required) || disabled) {
        return true;
      }
      let { isValid, errortext } = runValidations({ id, required, validations, value });
      if (value && confirmationId === 'password' && confirmationRef.current && value !== confirmationRef.current) {
        isValid = false;
        errortext = `Passwords don't match. Please try again.`;
      }
      let isStrong = false;
      let isWarningIcon = false;
      if (create && value && isValid) {
        const { score, feedback } = await checkPasswordStrength(value);
        const suggestions = feedback.suggestions || [];
        if (score <= SCORE_THRESHOLD || feedback.warning || suggestions.length) {
          isValid = false;
          isWarningIcon = true;
          // show a single message, most actionable first: a suggestion, else the warning, else the generic fallback
          errortext = suggestions[0] || feedback.warning || 'Password is too guessable';
        }
        isStrong = isValid && score >= STRONG_SCORE;
      }
      setStrong(isStrong);
      setWarningIcon(isWarningIcon);
      return isValid || errortext;
    },
    [confirmationId, create, disabled, id, required, validations]
  );

  const currentValue = getValues(id);
  // Revalidate if mismatch and user change password to match confirm_password
  useEffect(() => {
    if (errors.password_confirmation && create && !generate && id == 'password') {
      trigger('password_confirmation');
    }
  }, [create, currentValue, errors.password_confirmation, generate, id, trigger]);

  useEffect(
    () => () => {
      clearTimeout(timer.current);
    },
    []
  );

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

  const showAsNotched = label && typeof label !== 'string' ? { notched: true } : {};
  return (
    <div className={className}>
      <div className="password-wrapper">
        <Controller
          name={id}
          control={control}
          rules={{ validate }}
          render={({ field: { value, onChange, onBlur, ref }, fieldState: { error } }) => {
            const errorMessage = (errors[errorKey] || error)?.message;
            const showSuccess = strong && !errorMessage && Boolean(value);
            return (
              <FormControl
                className={`${required ? 'required' : ''} ${showSuccess ? classes.success : ''}`.trim()}
                error={Boolean(errorMessage)}
                color={showSuccess ? 'success' : undefined}
                style={{ width }}
              >
                <InputLabel htmlFor={id} {...InputLabelProps}>
                  {label}
                </InputLabel>
                <OutlinedInput
                  autoComplete={autocomplete}
                  id={id}
                  label={label}
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
                    if (create) {
                      trigger(id);
                    }
                  }}
                  onBlur={() => {
                    if (id === 'password_confirmation') {
                      trigger(id);
                    }
                    onBlur();
                  }}
                  endAdornment={
                    <InputAdornment position="end">
                      <IconButton onClick={() => setVisible(toggle)} size="large">
                        {visible ? <VisibilityIcon /> : <VisibilityOffIcon />}
                      </IconButton>
                    </InputAdornment>
                  }
                  {...showAsNotched}
                  {...InputProps}
                />
                <FormHelperText component="div" className={`flexbox align-items-center ${showSuccess ? 'green' : ''}`}>
                  {!!errorMessage && warningIcon && <WarningIcon fontSize="small" className={classes.icon} />}
                  {showSuccess && <CheckIcon fontSize="small" className={classes.icon} />}
                  {errorMessage || (showSuccess ? 'Strong password' : '')}
                </FormHelperText>
              </FormControl>
            );
          }}
        />
        {generate && !required && <PasswordGenerateButtons disabled={disabled} clearPass={clearPassClick} edit={edit} generatePass={generatePassClick} />}
      </div>
      {copied ? <div className="green fadeIn margin-bottom-small">Copied to clipboard</div> : null}
      {create && generate && required && (
        <PasswordGenerateButtons disabled={disabled} clearPass={clearPassClick} edit={edit} generatePass={generatePassClick} />
      )}
    </div>
  );
};

export default PasswordInput;
