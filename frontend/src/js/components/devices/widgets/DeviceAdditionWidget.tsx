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
import { useState } from 'react';

import { ArrowDropDown as ArrowDropDownIcon, Launch as LaunchIcon } from '@mui/icons-material';
import { Button, ButtonGroup, Menu, MenuItem } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import DocsLink from '@northern.tech/common-ui/DocsLink';
import { ALL_DEVICES, canAccess, uiPermissionsById } from '@northern.tech/store/constants';

const useStyles = makeStyles()(() => ({
  buttonStyle: { textTransform: 'none' }
}));

export const DeviceAdditionWidget = ({ innerRef, onConnectClick, onPreauthClick, userCapabilities }) => {
  const [anchorEl, setAnchorEl] = useState();
  const [selectedIndex, setSelectedIndex] = useState(0);
  const { classes } = useStyles();

  const options = [
    {
      action: onConnectClick,
      title: 'Connect a new device',
      value: 'connect',
      canAccess: ({ userCapabilities: { groupsPermissions } }) => groupsPermissions[ALL_DEVICES]?.includes(uiPermissionsById.manage.value)
    },
    { action: onPreauthClick, title: 'Preauthorize a device', value: 'preauth', canAccess },
    {
      component: DocsLink,
      path: 'client-installation/overview',
      title: 'Learn how to connect devices',
      value: 'learntoconnect',
      canAccess
    }
  ].filter(({ canAccess }) => canAccess({ userCapabilities }));

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
      <ButtonGroup variant="outlined" ref={innerRef}>
        <Button className={classes.buttonStyle} onClick={options[selectedIndex].action}>
          {options[selectedIndex].title}
        </Button>
        <Button className={`${classes.buttonStyle} padding-left-small padding-right-small`} size="small" onClick={handleToggle}>
          <ArrowDropDownIcon />
        </Button>
      </ButtonGroup>
      <Menu id="device-connection-menu" anchorEl={anchorEl} keepMounted open={Boolean(anchorEl)} onClose={handleToggle} variant="menu">
        {options.map((option, index) => {
          // eslint-disable-next-line @typescript-eslint/no-unused-vars
          const { canAccess, component, title, value, ...optionProps } = option;
          return component ? (
            <MenuItem {...optionProps} key={value} component={component}>
              {title}
              <LaunchIcon style={{ fontSize: '10pt' }} />
            </MenuItem>
          ) : (
            <MenuItem className={classes.buttonStyle} key={value} onClick={() => handleSelection(index)}>
              {title}
            </MenuItem>
          );
        })}
      </Menu>
    </>
  );
};

export default DeviceAdditionWidget;
