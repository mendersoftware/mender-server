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
import { Typography } from '@mui/material';

import { Select } from '@northern.tech/common-ui/forms/Select';

export const TierSelection = ({ className = '', onChange, selectedTier = '', enabledTiers }) => (
  <div className="flexbox align-items-center margin-left">
    <Typography variant="body2" className="margin-right-x-small">
      Tier:
    </Typography>
    <Select
      className={`capitalized ${className}`}
      displayEmpty
      MenuItemProps={{ className: 'capitalized' }}
      onChange={e => onChange(e.target.value)}
      options={[{ id: '', title: 'Any' }, ...enabledTiers.map(tier => ({ id: tier, title: tier }))]}
      value={selectedTier}
    />
  </div>
);
