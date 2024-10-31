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
import React, { useMemo } from 'react';

// material ui
import { MenuItem, Select } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { duplicateFilter } from '@northern.tech/utils/helpers';

const useStyles = makeStyles()(theme => ({
  selection: {
    fontSize: 13,
    marginLeft: theme.spacing(0.5),
    marginTop: 2,
    '>div': {
      paddingLeft: theme.spacing(0.5)
    }
  }
}));

export const DeviceStateSelection = ({ onStateChange, selectedState = '', states }) => {
  const { classes } = useStyles();
  const availableStates = useMemo(() => Object.values(states).filter(duplicateFilter), [states]);

  return (
    <div className="flexbox centered">
      Status:
      <Select className={classes.selection} disableUnderline onChange={e => onStateChange(e.target.value)} value={selectedState}>
        {availableStates.map(state => (
          <MenuItem key={state.key} value={state.key}>
            {state.title()}
          </MenuItem>
        ))}
      </Select>
    </div>
  );
};
