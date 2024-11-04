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
import React, { useEffect, useRef, useState } from 'react';
import { useDispatch } from 'react-redux';

// material ui
import { Input, InputAdornment } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { setDeviceTags } from '@northern.tech/store/thunks';

import { ConfirmationButtons, EditButton } from './confirm';

const useStyles = makeStyles()(theme => ({
  icon: {
    fontSize: '1.25rem'
  },
  input: {
    color: theme.palette.text.primary,
    fontSize: '0.8125rem'
  }
}));

export const DeviceNameInput = ({ device, isHovered }) => {
  const [isEditing, setIsEditing] = useState(false);
  const [value, setValue] = useState('');
  const { classes } = useStyles();
  const inputRef = useRef();

  const dispatch = useDispatch();

  const { id = '', tags = {} } = device;
  const { name = '' } = tags;

  useEffect(() => {
    if (!isEditing && name !== value) {
      setValue(name);
    }
  }, [device, isEditing, name, value]);

  useEffect(() => {
    if (!isEditing || !inputRef.current) {
      return;
    }
    inputRef.current.focus();
  }, [isEditing]);

  const onSubmit = () => dispatch(setDeviceTags({ deviceId: id, tags: { ...tags, name: value } })).then(() => setIsEditing(false));

  const onCancel = () => {
    setValue(name);
    setIsEditing(false);
  };

  const onStartEdit = e => {
    e.stopPropagation();
    setIsEditing(true);
  };

  const onInputClick = e => e.stopPropagation();

  return (
    <Input
      id={`${device.id}-id-input`}
      className={classes.input}
      disabled={!isEditing}
      inputRef={inputRef}
      value={value}
      placeholder={`${id.substring(0, 6)}...`}
      onClick={onInputClick}
      onChange={({ target: { value } }) => setValue(value)}
      type="text"
      endAdornment={
        (isHovered || isEditing) && (
          <InputAdornment position="end">
            {isEditing ? <ConfirmationButtons onCancel={onCancel} onConfirm={onSubmit} /> : <EditButton onClick={onStartEdit} />}
          </InputAdornment>
        )
      }
    />
  );
};

export default DeviceNameInput;
