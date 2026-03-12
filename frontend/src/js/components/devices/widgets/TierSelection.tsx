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
import { MenuItem, Select } from '@mui/material';

export const TierSelection = ({ className = '', onChange, selectedTier = '', enabledTiers }) => (
  <div className="flexbox align-items-center margin-left">
    Tier:
    <Select className={`capitalized ${className}`} displayEmpty onChange={e => onChange(e.target.value)} value={selectedTier}>
      <MenuItem value="">Any</MenuItem>
      {enabledTiers.map(tier => (
        <MenuItem className="capitalized" key={tier} value={tier}>
          {tier}
        </MenuItem>
      ))}
    </Select>
  </div>
);
