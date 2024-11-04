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
import { Tooltip } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import copy from 'copy-to-clipboard';

const useStyles = makeStyles()(theme => ({
  root: {
    ['.key > b']: {
      backgroundColor: theme.palette.grey[400],
      color: theme.palette.getContrastText(theme.palette.grey[400])
    }
  }
}));

const cutoffLength = 100;
const ValueColumn = ({ value = '', setSnackbar }) => {
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
    shownValue = <div title={value}>{value.length > cutoffLength ? `${value.substring(0, cutoffLength - 3)}...` : value}</div>;
  }
  return (
    <div
      className={`flexbox ${setSnackbar ? 'clickable' : ''}`}
      onClick={onClick}
      onMouseEnter={() => setTooltipVisible(true)}
      onMouseLeave={() => setTooltipVisible(false)}
    >
      {shownValue}
      {setSnackbar && (
        <Tooltip title={'Copy to clipboard'} placement="top" open={tooltipVisible}>
          <CopyToClipboardIcon color="primary" className={`margin-left-small ${tooltipVisible ? 'fadeIn' : 'fadeOut'}`} fontSize="small"></CopyToClipboardIcon>
        </Tooltip>
      )}
    </div>
  );
};

const KeyColumn = ({ value, chipLikeKey }) => (
  <div className={`align-right ${chipLikeKey ? 'key' : ''} muted`}>
    <b>{value}</b>
  </div>
);

export const TwoColumns = ({
  className = '',
  children = undefined,
  chipLikeKey = true,
  compact = false,
  items = {},
  KeyComponent = KeyColumn,
  KeyProps = {},
  setSnackbar,
  style = {},
  ValueComponent = ValueColumn,
  ValueProps = {}
}) => {
  const { classes } = useStyles();
  return (
    <div className={`break-all two-columns ${classes.root} ${compact ? 'compact' : ''} ${className}`} style={style}>
      {children
        ? children
        : Object.entries(items).map(([key, value]) => (
            <Fragment key={key}>
              <KeyComponent chipLikeKey={chipLikeKey} value={key} {...KeyProps} />
              <ValueComponent setSnackbar={setSnackbar} value={value} {...ValueProps} />
            </Fragment>
          ))}
    </div>
  );
};

export const TwoColumnData = ({ className = '', config, ...props }) => <TwoColumns className={`column-data ${className}`} items={config} {...props} />;

export const TwoColumnDataMultiple = ({ className = '', config, style, ...props }) => (
  <div className={`two-columns-multiple ${className}`} style={{ ...style }}>
    {Object.entries(config).map(([key, value]) => (
      <TwoColumnData className="multiple" config={{ [key]: value }} key={key} compact {...props} />
    ))}
  </div>
);

export const ConfigurationObject = ({ config, ...props }) => {
  const content = Object.entries(config).reduce((accu, [key, value]) => {
    accu[key] = `${value}`;
    return accu;
  }, {});
  return <TwoColumnData {...props} config={content} />;
};

export default ConfigurationObject;
