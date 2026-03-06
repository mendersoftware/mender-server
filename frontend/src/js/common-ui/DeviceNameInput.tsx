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
import { useDispatch } from 'react-redux';

import { setDeviceTags } from '@northern.tech/store/thunks';

import { EditableNameInput } from './EditableNameInput';

export const DeviceNameInput = ({ device, isHovered }) => {
  const dispatch = useDispatch();

  const { id = '', tags = {} } = device;
  const { name = '' } = tags;

  const onSave = (value: string) => dispatch(setDeviceTags({ deviceId: id, tags: { ...tags, name: value } }));

  return <EditableNameInput id={`${device.id}-id-input`} name={name} placeholder={`${id.substring(0, 6)}...`} isHovered={isHovered} onSave={onSave} />;
};

export default DeviceNameInput;
