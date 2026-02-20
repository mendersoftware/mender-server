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
import { useMemo } from 'react';

// material ui
import { MenuItem, Select } from '@mui/material';

import { duplicateFilter } from '@northern.tech/utils/helpers';

export const DeviceStateSelection = ({ className = '', onStateChange, selectedState = '', states }) => {
  const availableStates = useMemo(() => Object.values(states).filter(duplicateFilter), [states]);

  return (
    <div className="flexbox align-items-center">
      Status:
      <Select className={`capitalized ${className}`} onChange={e => onStateChange(e.target.value)} value={selectedState}>
        {availableStates.map(state => (
          <MenuItem className="capitalized" key={state.key} value={state.key}>
            {state.title()}
          </MenuItem>
        ))}
      </Select>
    </div>
  );
};
