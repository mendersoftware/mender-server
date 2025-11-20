// Copyright 2022 Northern.tech AS
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

import { ArrowDropDown as ArrowDropDownIcon, ArrowDropUp as ArrowDropUpIcon } from '@mui/icons-material';
// material ui
import { Button, Menu, MenuItem } from '@mui/material';

const ListOptions = ({ options, title = 'Table options' }) => {
  const [anchorEl, setAnchorEl] = useState(null);

  return (
    <>
      <Button variant="text" color="inherit" onClick={e => setAnchorEl(e.currentTarget)} endIcon={anchorEl ? <ArrowDropUpIcon /> : <ArrowDropDownIcon />}>
        {title}
      </Button>
      <Menu
        anchorEl={anchorEl}
        anchorOrigin={{ horizontal: 'right', vertical: 'top' }}
        onClose={() => setAnchorEl(null)}
        open={Boolean(anchorEl)}
        transformOrigin={{ horizontal: 'right', vertical: 'bottom' }}
      >
        {options.map(({ key, title, onClick }) => (
          <MenuItem key={key} onClick={onClick} value={key}>
            {title}
          </MenuItem>
        ))}
      </Menu>
    </>
  );
};

export default ListOptions;
