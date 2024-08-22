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

import { InfoOutlined as InfoOutlinedIcon } from '@mui/icons-material';
import { makeStyles } from 'tss-react/mui';

import InfoText from './infotext';

const useStyles = makeStyles()(theme => ({
  container: { gap: theme.spacing(2) },
  default: { gap: theme.spacing() }
}));

export const InfoHint = ({ className = '', content, ...props }) => {
  const { classes } = useStyles();
  return (
    <InfoText className={`icon flexbox center-aligned ${classes.default} ${className}`} {...props}>
      <InfoOutlinedIcon fontSize="small" />
      {content}
    </InfoText>
  );
};

export const InfoHintContainer = ({ children, className = 'margin-left-small', ...props }) => {
  const { classes } = useStyles();
  return (
    <div className={`flexbox center-aligned ${className} ${classes.container}`} {...props}>
      {children}
    </div>
  );
};

export default InfoHint;
