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
import { useEffect } from 'react';
import { FormProvider, useForm } from 'react-hook-form';

import { Button } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

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
  resetOnSubmit = false,
  showButtons,
  submitLabel,
  submitRef,
  validationMode = 'onChange'
}) => {
  const { classes: internalClasses } = useStyles();
  const methods = useForm({ mode: validationMode, defaultValues });
  const {
    handleSubmit,
    formState: { isValid },
    reset,
    setValue
  } = methods;

  useEffect(() => {
    if (submitRef) {
      submitRef.current = handleSubmit(onSubmit);
    }
  }, [handleSubmit, onSubmit, submitRef]);
  useEffect(() => {
    Object.entries(initialValues).forEach(([key, value]) => setValue(key, value));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(initialValues), setValue]);
  const handleFormSubmit = async data => {
    await onSubmit(data);
    if (resetOnSubmit) {
      reset();
    }
  };
  return (
    <FormProvider {...methods}>
      <form autoComplete={autocomplete} className={className} id={id} noValidate onSubmit={handleSubmit(handleFormSubmit)}>
        {children}
        {!!showButtons && (
          <div className={`button-wrapper ${internalClasses.buttonWrapper} ${classes.buttonWrapper}`}>
            {!!handleCancel && (
              <Button className={`${internalClasses.cancelButton} ${classes.cancelButton}`} key="cancel" onClick={handleCancel}>
                Cancel
              </Button>
            )}
            <Button variant="contained" type="submit" disabled={!isValid && validationMode !== 'onSubmit'} color={buttonColor}>
              {submitLabel}
            </Button>
          </div>
        )}
      </form>
    </FormProvider>
  );
};

export default Form;
