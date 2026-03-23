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
import { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';

import { Button, DialogActions, DialogContent, FormControl, FormHelperText, MenuItem, Select, Typography } from '@mui/material';

import { DOCSTIPS, DocsTextLink } from '@northern.tech/common-ui/DocsLink';
import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import { getDeviceIdentityAttributes } from '@northern.tech/store/devicesSlice/selectors';
import { getDeviceAttributes } from '@northern.tech/store/devicesSlice/thunks';
import { useAppDispatch } from '@northern.tech/store/store';
import { getIdAttribute } from '@northern.tech/store/usersSlice/selectors';
import { saveGlobalSettings } from '@northern.tech/store/usersSlice/thunks';

export const DeviceIdentityDialog = ({ open, onClose }) => {
  const attributes = useSelector(getDeviceIdentityAttributes);
  const selectedAttribute = useSelector(getIdAttribute);
  const [attributeSelection, setAttributeSelection] = useState(selectedAttribute);
  const dispatch = useAppDispatch();

  useEffect(() => {
    setAttributeSelection(selectedAttribute);
  }, [selectedAttribute]);

  useEffect(() => {
    dispatch(getDeviceAttributes());
  }, [dispatch]);

  const onHandleSubmit = async () => {
    await dispatch(saveGlobalSettings({ id_attribute: selectedAttribute, notify: true })).unwrap();
    onClose();
  };

  const onChangeIdAttribute = ({ target: { value } }: { target: { value: string } }) => {
    const match = attributes.find(attr => attr.value === value);
    if (!match) {
      return;
    }
    setAttributeSelection({ attribute: match.value, scope: match.scope });
  };

  const isUnchanged = selectedAttribute.attribute === attributeSelection.attribute;

  return (
    <BaseDialog open={open} title="Default device identity" onClose={onClose}>
      <DialogContent dividers={false}>
        <FormControl className="margin-top-none">
          <Select className="margin-top-x-small" value={attributeSelection.attribute} onChange={onChangeIdAttribute}>
            {attributes.map(item => (
              <MenuItem key={item.value} value={item.value}>
                {item.label}
              </MenuItem>
            ))}
          </Select>
          <FormHelperText>Choose a device identity attribute to use to identify your devices throughout the UI.</FormHelperText>
        </FormControl>
        <Typography className="margin-top-x-small" variant="body2">
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
        <Button disabled={isUnchanged} variant="contained" onClick={onHandleSubmit}>
          Save
        </Button>
      </DialogActions>
    </BaseDialog>
  );
};
