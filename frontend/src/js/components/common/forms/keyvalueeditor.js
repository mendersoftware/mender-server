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
import React, { createRef, useEffect, useState } from 'react';

import { Clear as ClearIcon, Add as ContentAddIcon } from '@mui/icons-material';
import { Fab, FormControl, FormHelperText, IconButton, Input } from '@mui/material';
import { useTheme } from '@mui/material/styles';

const emptyInput = { helptip: null, key: '', value: '' };

export const KeyValueEditor = ({ disabled, errortext, initialInput = {}, inputHelpTipsMap = {}, onInputChange, reset }) => {
  const theme = useTheme();
  const [inputs, setInputs] = useState([{ ...emptyInput }]);
  const [error, setError] = useState('');

  useEffect(() => {
    const newInputs = Object.keys(initialInput).length
      ? Object.entries(initialInput).map(([key, value]) => ({ helptip: inputHelpTipsMap[key.toLowerCase()], key, ref: createRef(), value }))
      : [{ ...emptyInput, ref: createRef() }];
    setInputs(newInputs);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(initialInput), JSON.stringify(inputHelpTipsMap), reset]);

  const onClearClick = () => {
    const changedInputs = [{ ...emptyInput }];
    setInputs(changedInputs);
    const inputObject = reducePairs(changedInputs);
    onInputChange(inputObject);
  };

  const updateInputs = (key, index, event) => {
    let changedInputs = [...inputs];
    const {
      target: { value }
    } = event;
    changedInputs[index][key] = value;
    changedInputs[index].helptip = null;
    const normalizedKey = changedInputs[index].key.toLowerCase();
    if (inputHelpTipsMap[normalizedKey]) {
      changedInputs[index].helptip = inputHelpTipsMap[normalizedKey];
    }
    setInputs(changedInputs);
    const inputObject = reducePairs(changedInputs);
    if (changedInputs.every(item => item.key && item.value) && changedInputs.length !== Object.keys(inputObject).length) {
      setError('Duplicate keys exist, only the last set value will be submitted');
    } else {
      setError('');
    }
    onInputChange(inputObject);
  };

  const reducePairs = listOfPairs => listOfPairs.reduce((accu, item) => ({ ...accu, ...(item.value ? { [item.key]: item.value } : {}) }), {});

  const addKeyValue = () => {
    const changedInputs = [...inputs, { ...emptyInput, ref: createRef() }];
    setInputs(changedInputs);
    setError('');
  };

  const removeInput = index => {
    let changedInputs = [...inputs];
    changedInputs.splice(index, 1);
    setInputs(changedInputs);
    const inputObject = reducePairs(changedInputs);
    onInputChange(inputObject);
    setError('');
  };

  return (
    <div>
      {inputs.map((input, index) => {
        const hasError = Boolean(index === inputs.length - 1 && (errortext || error));
        const hasRemovalDisabled = !(inputs[index].key && inputs[index].value);
        const Helptip = inputs[index].helptip?.component;
        return (
          <div className="key-value-container relative" key={index}>
            <FormControl>
              <Input disabled={disabled} value={input.key} placeholder="Key" onChange={e => updateInputs('key', index, e)} type="text" />
              {hasError && <FormHelperText>{errortext || error}</FormHelperText>}
            </FormControl>
            <FormControl>
              <Input disabled={disabled} value={`${input.value}`} placeholder="Value" onChange={e => updateInputs('value', index, e)} type="text" />
            </FormControl>
            {inputs.length > 1 && !hasRemovalDisabled ? (
              <IconButton disabled={disabled} onClick={() => removeInput(index)} size="large">
                <ClearIcon fontSize="small" />
              </IconButton>
            ) : (
              <span />
            )}
            {Helptip && <Helptip anchor={{ left: -35, top: 5 }} {...inputs[index].helptip.props} />}
          </div>
        );
      })}
      <div className="key-value-container">
        <div style={{ minWidth: theme.spacing(30) }}>
          <Fab
            disabled={disabled || !inputs[inputs.length - 1].key || !inputs[inputs.length - 1].value}
            style={{ marginBottom: 10 }}
            color="secondary"
            size="small"
            onClick={addKeyValue}
          >
            <ContentAddIcon />
          </Fab>
        </div>
        <div style={{ minWidth: theme.spacing(30) }} />
        {inputs.length > 1 ? (
          <a className="margin-left-small" onClick={onClearClick}>
            clear all
          </a>
        ) : (
          <div />
        )}
      </div>
    </div>
  );
};

export default KeyValueEditor;
