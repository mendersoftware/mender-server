// Copyright 2021 Northern.tech AS
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
// material ui
import { Checkbox, MenuItem, Select, Typography } from '@mui/material';

import { DEVICE_ISSUE_OPTIONS } from '@northern.tech/store/constants';

const menuProps = {
  anchorOrigin: {
    vertical: 'bottom',
    horizontal: 'left'
  },
  transformOrigin: {
    vertical: 'top',
    horizontal: 'left'
  }
};
const getSelectionDisplayValue = (selected, options) => {
  if (!selected.length) {
    return 'all';
  }
  const titles = selected.map(issue => DEVICE_ISSUE_OPTIONS[issue].title).join(', ');
  const sum = selected.reduce((accu, issue) => accu + (options.find(option => option.key === issue)?.count || 0), 0);
  return `${titles} (${sum})`;
};

const DeviceIssuesSelection = ({ className = '', onChange, options, selection }) => (
  <div className="flexbox align-items-center margin-left">
    <Typography variant="body2" className="margin-right-x-small">
      Show:
    </Typography>
    <Select
      className={`capitalized ${className}`}
      displayEmpty
      MenuProps={menuProps}
      multiple
      onChange={onChange}
      renderValue={selected => getSelectionDisplayValue(selected, options)}
      value={selection}
      autoWidth={false}
    >
      {options.map(({ count, key, title }) => (
        <MenuItem className="capitalized-start padding-left-x-small" key={key} value={key} size="small">
          <Checkbox className="padding-none margin-right-x-small" checked={selection.includes(key)} />
          {title} ({count})
        </MenuItem>
      ))}
    </Select>
  </div>
);

export default DeviceIssuesSelection;
