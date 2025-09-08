// Copyright 2024 Northern.tech AS
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

import { Checkbox, FormControl, FormHelperText, InputLabel, MenuItem, Select } from '@mui/material';

import { ADDONS, AddonId } from '@northern.tech/store/constants';

interface AddonSelectProps {
  initialState: AddonId[];
  onChange: (e: AddonId[]) => void;
}
export const AddonSelect = (props: AddonSelectProps) => {
  const { onChange, initialState } = props;
  const [selectedAddons, setSelectedAddons] = useState(initialState);
  const onSelectedChange = ({ target: { value } }) => {
    setSelectedAddons(value);
    onChange(value);
  };
  return (
    <FormControl id="addons-form" style={{ width: 550 }}>
      <InputLabel id="addons-selection-label">Select Add-ons (optional)</InputLabel>
      <Select
        multiple
        label="Select Add-ons (optional)"
        labelId="addons-selection-label"
        value={selectedAddons}
        onChange={onSelectedChange}
        renderValue={selected => selected.map((role: AddonId) => 'Mender ' + ADDONS[role].title).join(', ')}
      >
        {Object.values(ADDONS).map(addon => (
          <MenuItem id={addon.id} key={addon.id} value={addon.id}>
            <Checkbox id={`${addon.id}-checkbox`} checked={selectedAddons.includes(addon.id as AddonId)} />
            Mender {addon.title}
          </MenuItem>
        ))}
      </Select>
      <FormHelperText className="info">Select any Add-ons you are interested in including in your subscription</FormHelperText>
    </FormControl>
  );
};
