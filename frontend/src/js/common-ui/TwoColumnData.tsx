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
import React, { Fragment, useState } from 'react';

// material ui
import { FileCopyOutlined as CopyToClipboardIcon } from '@mui/icons-material';
import { Chip, Tooltip, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import copy from 'copy-to-clipboard';

const getGridTemplateColumnSizing = columnWidth => `${columnWidth} 650px`;

const useStyles = makeStyles()(theme => ({
  copyIconOverride: {
    '&.copy-to-clipboard svg': {
      fill: theme.palette.action.active
    }
  },
  copyIconVisible: {
    opacity: 1
  },
  root: {
    '&.two-columns.column-data': {
      gridTemplateColumns: getGridTemplateColumnSizing('max-content'),
      maxWidth: 'initial',
      rowGap: theme.spacing()
    }
  }
}));

const cutoffLength = 100;
const ValueColumn = ({ value = '', setSnackbar }) => {
  const { classes } = useStyles();
  const [tooltipVisible, setTooltipVisible] = useState(false);
  const isComponent = React.isValidElement(value);
  const onClick = () => {
    if (setSnackbar) {
      let copyable = value;
      if (isComponent) {
        copyable = value.props.value;
      }
      copy(copyable);
      setSnackbar('Value copied to clipboard');
    }
  };
  let shownValue = value;
  if (!isComponent) {
    shownValue = value.length > cutoffLength ? `${value.substring(0, cutoffLength - 3)}...` : value;
  }
  return (
    <Typography
      className={`flexbox copy-to-clipboard ${setSnackbar ? 'clickable' : ''} ${classes.copyIconOverride}`}
      component="div"
      onClick={onClick}
      title={value}
      onMouseEnter={() => setTooltipVisible(true)}
      onMouseLeave={() => setTooltipVisible(false)}
      variant="body2"
    >
      {shownValue}
      {setSnackbar && (
        <Tooltip title="Copy to clipboard" placement="top">
          <CopyToClipboardIcon color="action" fontSize="small" className={`margin-left-x-small ${tooltipVisible ? classes.copyIconVisible : ''}`} />
        </Tooltip>
      )}
    </Typography>
  );
};

const KeyColumn = ({ value, chipLikeKey }) =>
  chipLikeKey ? (
    <Chip label={value} size="small" style={{ justifySelf: 'end' }} />
  ) : (
    <Typography className="key" variant="subtitle2">
      {value}
    </Typography>
  );

export const TwoColumnData = ({ className = '', chipLikeKey = false, columnSize, config = {}, setSnackbar, style = {} }) => {
  const { classes } = useStyles();
  return (
    <div
      className={`break-all two-columns ${classes.root} column-data ${className}`}
      style={{ ...style, gridTemplateColumns: columnSize ? getGridTemplateColumnSizing(`${columnSize}px`) : undefined }}
    >
      {Object.entries(config).map(([key, value]) => (
        <Fragment key={key}>
          <KeyColumn chipLikeKey={chipLikeKey} value={key} />
          <ValueColumn setSnackbar={setSnackbar} value={value} />
        </Fragment>
      ))}
    </div>
  );
};
