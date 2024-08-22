// Copyright 2019 Northern.tech AS
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
import React, { memo, useCallback, useState } from 'react';

// material ui
import { Checkbox } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { DEVICE_STATES } from '../../constants/deviceConstants';
import { deepCompare } from '../../helpers';
import DeviceIdentityDisplay from '../common/deviceidentity';
import { DefaultAttributeRenderer } from './base-devices';

const useStyles = makeStyles()(theme => ({
  active: {
    [`> *`]: {
      backgroundColor: theme.palette.background.light
    }
  }
}));

const DeviceListItem = ({ columnHeaders, device, deviceListState, idAttribute, index, onClick, onRowSelect, selectable, selected }) => {
  const [isHovering, setIsHovering] = useState(false);
  const { classes } = useStyles();

  const onMouseOut = () => setIsHovering(false);
  const onMouseOver = () => setIsHovering(true);

  const handleOnClick = useCallback(
    event => {
      if (event && event.target.closest('input')?.hasOwnProperty('checked')) {
        return;
      }
      onClick(device);
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [device.id, onClick, deviceListState.selectedId]
  );

  const handleRowSelect = () => onRowSelect(index);

  return (
    <div
      className={`deviceListRow deviceListItem clickable ${isHovering ? classes.active : ''} ${device.status === DEVICE_STATES.pending ? classes.active : ''}`}
      onClick={handleOnClick}
      onMouseEnter={onMouseOver}
      onMouseLeave={onMouseOut}
    >
      {/*
        we need to wrap the checkbox into a div here to ensure the bottom border etc. works as intended since the outer div will
        not create an own box due to "display: contents" being needed until subgrid support lands in browsers
      */}
      {selectable && (
        <div>
          <Checkbox checked={selected} onChange={handleRowSelect} />
        </div>
      )}
      <DeviceIdentityDisplay device={device} isHovered={isHovering} />
      {/* we'll skip the first column, since this is the id and that gets resolved differently in the lines above */}
      {columnHeaders.slice(1).map((column, index) => {
        let Component = column.component ? column.component : DefaultAttributeRenderer;
        return <Component column={column} device={device} idAttribute={idAttribute} key={`column-${index}`} />;
      })}
    </div>
  );
};

const areEqual = (prevProps, nextProps) => {
  if (
    prevProps.idAttribute != nextProps.idAttribute ||
    prevProps.selected != nextProps.selected ||
    !deepCompare(prevProps.columnHeaders, nextProps.columnHeaders) ||
    !deepCompare(prevProps.device, nextProps.device)
  ) {
    return false;
  }
  return deepCompare(prevProps.deviceListState, nextProps.deviceListState);
};

export default memo(DeviceListItem, areEqual);
