// Copyright 2024 Northern.tech AS
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
import React, { useCallback, useEffect, useState } from 'react';

// material ui
import { TextField } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { toggle } from '@northern.tech/utils/helpers';

import { ConfirmationButtons, EditButton } from './confirm';
import ExpandableAttribute from './expandable-attribute';

const useStyles = makeStyles()(theme => ({
  notes: { display: 'block', whiteSpace: 'pre-wrap' },
  notesWrapper: { minWidth: theme.components?.MuiFormControl?.styleOverrides?.root?.minWidth }
}));

export const EditableLongText = ({ contentFallback = '', fullWidth, original, onChange, placeholder = '-' }) => {
  const [isEditing, setIsEditing] = useState(false);
  const [value, setValue] = useState(original);
  const { classes } = useStyles();

  useEffect(() => {
    setValue(original);
  }, [original]);

  const onCancelClick = () => {
    setValue(original);
    setIsEditing(false);
  };

  const onEdit = ({ target: { value } }) => setValue(value);

  const onEditClick = () => setIsEditing(true);

  const onToggleEditing = useCallback(
    event => {
      event.stopPropagation();
      if (event.key && (event.key !== 'Enter' || event.shiftKey)) {
        return;
      }
      if (isEditing) {
        // save change
        onChange(value);
      }
      setIsEditing(toggle);
    },
    [isEditing, onChange, value]
  );

  const fullWidthClass = fullWidth ? 'full-width' : '';

  return (
    <div className="flexbox" style={{ alignItems: 'end' }}>
      {isEditing ? (
        <>
          <TextField
            className={`margin-right ${fullWidthClass}`}
            multiline
            onChange={onEdit}
            onKeyDown={onToggleEditing}
            placeholder={placeholder}
            value={value}
          />
          <ConfirmationButtons onCancel={onCancelClick} onConfirm={onToggleEditing} />
        </>
      ) : (
        <>
          <ExpandableAttribute
            className={`${fullWidthClass} margin-right ${classes.notesWrapper}`}
            component="div"
            dense
            disableGutters
            primary=""
            secondary={original || value || contentFallback}
            textClasses={{ secondary: classes.notes }}
          />
          <EditButton onClick={onEditClick} />
        </>
      )}
    </div>
  );
};
