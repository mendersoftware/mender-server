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
import React, { useState } from 'react';

import { ArrowDropDown as ArrowDropDownIcon, Launch as LaunchIcon } from '@mui/icons-material';
import { Button, ButtonGroup, Menu, MenuItem } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { canAccess } from '../../../constants/appConstants';
import DocsLink from '../../common/docslink';

const useStyles = makeStyles()(() => ({
  buttonStyle: { textTransform: 'none' }
}));

export const DeviceAdditionWidget = ({ features, innerRef, onConnectClick, onMakeGatewayClick, onPreauthClick, tenantCapabilities }) => {
  const [anchorEl, setAnchorEl] = useState();
  const [selectedIndex, setSelectedIndex] = useState(0);
  const { classes } = useStyles();

  const options = [
    { action: onConnectClick, title: 'Connect a new device', value: 'connect', canAccess },
    { action: onPreauthClick, title: 'Preauthorize a device', value: 'preauth', canAccess },
    {
      action: onMakeGatewayClick,
      title: 'Promote a device to gateway',
      value: 'makegateway',
      canAccess: ({ features, tenantCapabilities }) => features.isHosted && tenantCapabilities.isEnterprise
    },
    {
      component: DocsLink,
      path: 'client-installation/overview',
      title: 'Learn how to connect devices',
      value: 'learntoconnect',
      canAccess
    }
  ];

  const handleToggle = event => {
    const anchor = anchorEl ? null : event?.currentTarget.parentElement;
    setAnchorEl(anchor);
  };

  const handleSelection = index => {
    setSelectedIndex(index);
    handleToggle();
    options[index].action(true);
  };

  return (
    <>
      <ButtonGroup className="muted device-addition-widget" ref={innerRef} style={{ height: 'max-content' }}>
        <Button className={classes.buttonStyle} onClick={options[selectedIndex].action} variant="text">
          {options[selectedIndex].title}
        </Button>
        <Button className={classes.buttonStyle} size="small" onClick={handleToggle} variant="text">
          <ArrowDropDownIcon />
        </Button>
      </ButtonGroup>
      <Menu id="device-connection-menu" anchorEl={anchorEl} keepMounted open={Boolean(anchorEl)} onClose={handleToggle} variant="menu">
        {options.reduce((accu, option, index) => {
          if (!option.canAccess({ features, tenantCapabilities })) {
            return accu;
          }
          // eslint-disable-next-line no-unused-vars
          const { canAccess, component, title, value, ...optionProps } = option;
          const item = component ? (
            <MenuItem {...optionProps} key={value} component={component}>
              {title}
              <LaunchIcon style={{ fontSize: '10pt' }} />
            </MenuItem>
          ) : (
            <MenuItem className={classes.buttonStyle} key={value} onClick={() => handleSelection(index)}>
              {title}
            </MenuItem>
          );
          accu.push(item);
          return accu;
        }, [])}
      </Menu>
    </>
  );
};

export default DeviceAdditionWidget;
