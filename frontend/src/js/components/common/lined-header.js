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
import React from 'react';

import { makeStyles } from 'tss-react/mui';

const useStyles = makeStyles()(theme => ({
  border: {
    borderBottom: `1px solid ${theme.palette.grey[50]}`,
    span: {
      background: theme.palette.background.default
    }
  },
  groupBorder: {
    background: theme.palette.grey[50]
  },
  groupHeading: {
    background: theme.palette.background.default
  }
}));

const LinedHeader = ({ className = '', heading, innerStyle = {}, innerRef, style = {} }) => {
  const { classes } = useStyles();
  return (
    <h4 className={`dashboard-header ${classes.border} ${className}`} ref={innerRef} style={style}>
      <span style={innerStyle}>{heading}</span>
    </h4>
  );
};

export const LinedGroupHeader = ({ heading }) => {
  const { classes } = useStyles();
  return (
    <>
      <span className={classes.groupHeading}>{heading}</span>
      <div className={classes.groupBorder}></div>
    </>
  );
};

export default LinedHeader;
