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
import { Divider, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

const useStyles = makeStyles()(theme => ({
  header: {
    alignItems: 'center',
    gap: theme.spacing(2)
  },
  divider: { flex: 1 }
}));

const LinedHeader = ({ centered = false, className = '', heading, ref }) => {
  const { classes } = useStyles();
  return (
    <div className={`flexbox margin-bottom-small ${classes.header} ${className}`} ref={ref}>
      {centered && <Divider className={classes.divider} />}
      <Typography variant="subtitle1">{heading}</Typography>
      <Divider className={classes.divider} />
    </div>
  );
};

export default LinedHeader;
