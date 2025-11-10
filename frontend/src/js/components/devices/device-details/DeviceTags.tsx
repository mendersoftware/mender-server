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
import { useDispatch } from 'react-redux';

import { Button, Typography } from '@mui/material';

import ConfigurationObject from '@northern.tech/common-ui/ConfigurationObject';
import { EditButton } from '@northern.tech/common-ui/Confirm';
import KeyValueEditor from '@northern.tech/common-ui/forms/KeyValueEditor';
import { getDeviceAttributes, setDeviceTags } from '@northern.tech/store/thunks';

import Tracking from '../../../tracking';
import { HELPTOOLTIPS } from '../../helptips/HelpTooltips';
import { MenderHelpTooltip } from '../../helptips/MenderTooltip';
import DeviceDataCollapse from './DeviceDataCollapse';

const NameTipComponent = props => <MenderHelpTooltip id={HELPTOOLTIPS.nameTagTip.id} {...props} />;

const configHelpTipsMap = {
  name: { component: NameTipComponent }
};

export const DeviceTags = ({ device, setSnackbar, userCapabilities }) => {
  const { canWriteDevices } = userCapabilities;
  const [changedTags, setChangedTags] = useState({});
  const [editableTags, setEditableTags] = useState();
  const [isEditDisabled, setIsEditDisabled] = useState(!canWriteDevices);
  const [isEditing, setIsEditing] = useState(false);
  const dispatch = useDispatch();

  const { tags = {} } = device;
  const hasTags = !!Object.keys(tags).length;

  const onCancel = () => {
    setIsEditing(false);
    setChangedTags(tags);
  };

  const onStartEdit = e => {
    e.stopPropagation();
    setEditableTags(tags);
    setChangedTags(tags);
    setIsEditing(true);
  };

  const onSubmit = () => {
    Tracking.event({ category: 'devices', action: 'modify_tags' });
    setIsEditDisabled(true);
    return dispatch(setDeviceTags({ deviceId: device.id, tags: changedTags }))
      .then(() => {
        dispatch(getDeviceAttributes());
        setIsEditing(false);
      })
      .finally(() => setIsEditDisabled(false));
  };

  const isFullyDefined = Object.entries(changedTags).every(([key, value]) => !!key && !!value);

  return (
    <DeviceDataCollapse
      title={
        <div className="two-columns">
          <div className="flexbox center-aligned">
            <h4 className="margin-right">Tags</h4>
            {!isEditing && canWriteDevices && <EditButton onClick={onStartEdit} />}
          </div>
        </div>
      }
    >
      <div className="relative" style={{ maxWidth: 700 }}>
        {isEditing ? (
          <>
            <KeyValueEditor
              disabled={isEditDisabled}
              errortext=""
              initialInput={editableTags}
              inputHelpTipsMap={configHelpTipsMap}
              onInputChange={setChangedTags}
            />
            <div className="flexbox center-aligned margin-bottom-small" style={{ justifyContent: 'flex-end' }}>
              <Button className="margin-right-small" disabled={!isFullyDefined} onClick={onSubmit} variant="contained">
                Save
              </Button>
              <Button onClick={onCancel}>Cancel</Button>
            </div>
          </>
        ) : hasTags ? (
          <ConfigurationObject config={tags} setSnackbar={setSnackbar} />
        ) : (
          <Typography variant="subtitle2">No tags have been set for this device.</Typography>
        )}
      </div>
    </DeviceDataCollapse>
  );
};

export default DeviceTags;
