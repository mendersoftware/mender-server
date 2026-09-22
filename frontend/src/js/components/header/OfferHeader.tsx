// Copyright 2020 Northern.tech AS
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
import { Close } from '@mui/icons-material';
import { Chip, IconButton, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { Link } from '@northern.tech/common-ui/Link';

const useStyles = makeStyles()(theme => ({
  container: {
    height: theme.mixins.toolbar.minHeight,
    backgroundColor: theme.palette.primary.main,
    color: theme.palette.primary.contrastText
  },
  note: { gap: theme.spacing(2) },
  closeButton: {
    color: 'inherit',
    marginRight: theme.spacing(3)
  }
}));

const OfferHeader = ({ onHide }) => {
  const { classes } = useStyles();

  return (
    <div className={`flexbox align-items-center ${classes.container}`}>
      <div className={`flexbox centered full-width ${classes.note}`}>
        <Chip label="New" color="secondary" size="small" />
        <Typography variant="body2">More accurate device search: Find exact devices faster using your device identity attribute.</Typography>
        <Link href="https://mender.io/blog/simpler-more-predictable-device-search-in-the-mender-ui" external onClick={onHide} color="inherit" variant="body2">
          Learn more
        </Link>
      </div>

      <IconButton size="small" onClick={onHide} className={classes.closeButton}>
        <Close fontSize="small" />
      </IconButton>
    </div>
  );
};

export default OfferHeader;
