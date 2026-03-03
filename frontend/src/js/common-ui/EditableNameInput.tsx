// Copyright 2026 Northern.tech AS
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
import { useEffect, useRef, useState } from 'react';
import type { MouseEvent } from 'react';

// material ui
import { InputAdornment, OutlinedInput } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { ConfirmationButtons, EditButton } from './Confirm';

const useStyles = makeStyles()(theme => ({
  input: {
    color: theme.palette.text.primary,
    fontSize: '0.8125rem'
  }
}));

interface EditableNameInputProps {
  id: string;
  isHovered: boolean;
  name: string;
  onSave: (value: string) => Promise<unknown>;
  placeholder: string;
}

export const EditableNameInput = ({ id, name, placeholder, isHovered, onSave }: EditableNameInputProps) => {
  const [isEditing, setIsEditing] = useState(false);
  const [value, setValue] = useState(name);
  const { classes } = useStyles();
  const inputRef = useRef<HTMLInputElement>();

  useEffect(() => {
    if (!isEditing) {
      setValue(current => (name !== current ? name : current));
    }
  }, [name, isEditing]);

  useEffect(() => {
    if (!isEditing || !inputRef.current) {
      return;
    }
    inputRef.current.focus();
  }, [isEditing]);

  const onSubmit = () => onSave(value).then(() => setIsEditing(false));

  const onCancel = () => {
    setValue(name);
    setIsEditing(false);
  };

  const onStartEdit = (e: MouseEvent) => {
    e.stopPropagation();
    setIsEditing(true);
  };

  const onInputClick = (e: MouseEvent) => e.stopPropagation();

  return (
    <OutlinedInput
      id={id}
      className={classes.input}
      disabled={!isEditing}
      inputRef={inputRef}
      value={value}
      placeholder={placeholder}
      onClick={onInputClick}
      onChange={({ target: { value } }) => setValue(value)}
      type="text"
      size="small"
      endAdornment={
        (isHovered || isEditing) && (
          <InputAdornment position="end">
            {isEditing ? <ConfirmationButtons onCancel={onCancel} onConfirm={onSubmit} /> : <EditButton label="" onClick={onStartEdit} />}
          </InputAdornment>
        )
      }
    />
  );
};

export default EditableNameInput;
