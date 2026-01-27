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
import { Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

const useStyles = makeStyles()(theme => ({
  border: {
    borderBottom: `1px solid ${theme.palette.divider}`,
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
    <div className={`dashboard-header ${classes.border} ${className}`} ref={innerRef} style={style}>
      <Typography variant="body2" style={innerStyle} component="span">
        {heading}
      </Typography>
    </div>
  );
};

export const LinedGroupHeader = ({ heading }) => {
  const { classes } = useStyles();
  return (
    <>
      <span className={classes.groupHeading}>{heading}</span>
      <div className={classes.groupBorder} />
    </>
  );
};

export default LinedHeader;
