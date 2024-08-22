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
import React from 'react';

import { Chip } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

const useStyles = makeStyles()(theme => ({
  addOnPill: { marginRight: theme.spacing(2) }
}));

export const DeviceDataCollapse = ({ children, className = '', header, isAddOn = false, title }) => {
  const { classes } = useStyles();

  return (
    <div className={`margin-bottom ${className}`}>
      <div className="flexbox space-between center-aligned">
        {typeof title === 'string' ? <h4 className="margin-bottom-small">{title}</h4> : title}
        <div className="flexbox centered">{isAddOn && <Chip className={classes.addOnPill} label="Add-on" />}</div>
      </div>
      <div>{header}</div>
      {children}
    </div>
  );
};

export default DeviceDataCollapse;
