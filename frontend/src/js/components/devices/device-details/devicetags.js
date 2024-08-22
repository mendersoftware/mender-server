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
import React, { useEffect, useState } from 'react';
import { useDispatch } from 'react-redux';

import { Button } from '@mui/material';
import { useTheme } from '@mui/material/styles';

import { getDeviceAttributes, setDeviceTags } from '../../../actions/deviceActions';
import { toggle } from '../../../helpers';
import Tracking from '../../../tracking';
import ConfigurationObject from '../../common/configurationobject';
import { EditButton } from '../../common/confirm';
import KeyValueEditor from '../../common/forms/keyvalueeditor';
import { HELPTOOLTIPS, MenderHelpTooltip } from '../../helptips/helptooltips';
import DeviceDataCollapse from './devicedatacollapse';

const NameTipComponent = props => <MenderHelpTooltip id={HELPTOOLTIPS.nameTagTip.id} {...props} />;

const configHelpTipsMap = {
  name: { component: NameTipComponent, position: 'right' }
};

export const DeviceTags = ({ device, setSnackbar, userCapabilities }) => {
  const { canWriteDevices } = userCapabilities;
  const theme = useTheme();
  const [changedTags, setChangedTags] = useState({});
  const [editableTags, setEditableTags] = useState();
  const [isEditDisabled, setIsEditDisabled] = useState(!canWriteDevices);
  const [isEditing, setIsEditing] = useState(false);
  const [shouldUpdateEditor, setShouldUpdateEditor] = useState(false);
  const dispatch = useDispatch();

  const { tags = {} } = device;
  const hasTags = !!Object.keys(tags).length;

  useEffect(() => {
    setShouldUpdateEditor(toggle);
  }, [isEditing]);

  useEffect(() => {
    if (canWriteDevices) {
      setIsEditing(!hasTags);
    }
  }, [hasTags, canWriteDevices]);

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
    return dispatch(setDeviceTags(device.id, changedTags))
      .then(() => {
        dispatch(getDeviceAttributes());
        setIsEditing(false);
      })
      .finally(() => setIsEditDisabled(false));
  };

  const helpTipsMap = Object.entries(configHelpTipsMap).reduce((accu, [key, value]) => {
    accu[key] = {
      ...value,
      props: { deviceId: device.id }
    };
    return accu;
  }, {});
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
              inputHelpTipsMap={helpTipsMap}
              onInputChange={setChangedTags}
              reset={shouldUpdateEditor}
            />
            <div className="flexbox center-aligned margin-bottom-small" style={{ justifyContent: 'flex-end' }}>
              <Button color="primary" onClick={onSubmit} variant="contained" style={{ marginRight: theme.spacing(2) }}>
                Save
              </Button>
              <Button onClick={onCancel}>Cancel</Button>
            </div>
          </>
        ) : (
          hasTags && <ConfigurationObject config={tags} setSnackbar={setSnackbar} />
        )}
      </div>
    </DeviceDataCollapse>
  );
};

export default DeviceTags;
