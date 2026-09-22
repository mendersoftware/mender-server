// Copyright 2026 Northern.tech AS
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
import { useSelector } from 'react-redux';

import { Button, DialogActions, DialogContent, FormControl, FormHelperText, TextField, Typography } from '@mui/material';

import { DOCSTIPS, DocsTextLink } from '@northern.tech/common-ui/DocsLink';
import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import { ControlledAutoComplete } from '@northern.tech/common-ui/forms/Autocomplete';
import { getDeviceIdentityAttributes } from '@northern.tech/store/devicesSlice/selectors';
import { getDeviceAttributes } from '@northern.tech/store/devicesSlice/thunks';
import { useAppDispatch } from '@northern.tech/store/store';
import { getIdAttribute } from '@northern.tech/store/usersSlice/selectors';
import { saveGlobalSettings } from '@northern.tech/store/usersSlice/thunks';

export const DeviceIdentityDialog = ({ open, onClose }) => {
  const attributes = useSelector(getDeviceIdentityAttributes);
  const selectedAttribute = useSelector(getIdAttribute);
  const dispatch = useAppDispatch();
  const selectedOption = attributes.find(({ value }) => value === selectedAttribute.attribute) ?? null;
  const methods = useForm({ defaultValues: { attribute: selectedOption } });
  const {
    formState: { isDirty },
    handleSubmit,
    reset
  } = methods;

  useEffect(() => {
    reset({ attribute: selectedOption });
  }, [reset, selectedOption]);

  useEffect(() => {
    dispatch(getDeviceAttributes());
  }, [dispatch]);

  const onHandleSubmit = async ({ attribute }) => {
    if (!attribute) {
      return;
    }
    await dispatch(saveGlobalSettings({ id_attribute: { attribute: attribute.value, scope: attribute.scope }, notify: true })).unwrap();
    onClose();
  };

  return (
    <BaseDialog open={open} title="Default device identity" onClose={onClose}>
      <FormProvider {...methods}>
        <DialogContent dividers={false}>
          <FormControl className="margin-top-none">
            <ControlledAutoComplete
              name="attribute"
              id="device-identity-attribute-selection"
              className="margin-top-x-small"
              autoHighlight
              disableClearable
              getOptionLabel={option => option.label}
              options={attributes}
              renderInput={params => <TextField {...params} />}
            />
            <FormHelperText>Choose a device identity attribute to use to identify your devices throughout the UI.</FormHelperText>
          </FormControl>
          <Typography className="margin-top-x-small" variant="body2" component="div">
            Add custom identity attributes to your devices.{' '}
            <DocsTextLink id={DOCSTIPS.deviceIdentity.id} typographyProps={{ variant: 'body2' }}>
              Learn how
            </DocsTextLink>
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button variant="text" onClick={onClose}>
            Cancel
          </Button>
          <Button disabled={!isDirty} variant="contained" onClick={handleSubmit(onHandleSubmit)}>
            Save
          </Button>
        </DialogActions>
      </FormProvider>
    </BaseDialog>
  );
};
