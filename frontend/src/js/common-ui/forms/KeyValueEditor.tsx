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
import type { CSSProperties, ComponentType } from 'react';
import { useEffect, useState } from 'react';
import { useFieldArray, useFormContext } from 'react-hook-form';

import { AddCircle as AddIcon, Clear as ClearIcon } from '@mui/icons-material';
import { Button, FormControl, FormHelperText, IconButton, OutlinedInput, outlinedInputClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import Form from './Form';

type HelptipProps = {
  [key: string]: any;
  style?: CSSProperties;
};

type InputHelptip = {
  component: ComponentType<HelptipProps>;
  position?: string;
  props?: HelptipProps;
};

type InputLineItem = {
  helptip: InputHelptip | null;
  key: string;
  value: string;
};

const emptyInput: InputLineItem = { helptip: null, key: '', value: '' };

const reducePairs = (pairs: InputLineItem[]) => (pairs || []).reduce((accu, item) => ({ ...accu, ...(item.value ? { [item.key]: item.value } : {}) }), {});

const inputWidth = 240;

const useStyles = makeStyles()(theme => ({
  helptip: { left: -35, top: theme.spacing(), position: 'absolute !important' },
  keyValueContainer: {
    display: 'grid',
    gridTemplateColumns: `${inputWidth}px ${inputWidth}px max-content`,
    columnGap: theme.spacing(),
    alignItems: 'baseline',
    [`.${outlinedInputClasses.root}`]: { minWidth: inputWidth }
  },
  lineAdditionButton: { marginLeft: theme.spacing(-1) },
  lineRemovalButton: { marginBottom: 2 }
}));

interface KeyValueFieldsProps {
  disabled?: boolean;
  initialValues: InputLineItem[];
  inputHelpTipsMap: Record<string, { component: React.ComponentType<any>; props: any }>;
  onInputChange: (value: Record<string, string>) => void;
}

const KeyValueFields = ({ disabled, initialValues, inputHelpTipsMap, onInputChange }: KeyValueFieldsProps) => {
  const { classes } = useStyles();
  const {
    control,
    watch,
    setValue,
    formState: { errors },
    trigger
  } = useFormContext();

  const { fields, append, remove, replace } = useFieldArray<{ inputs: InputLineItem[] }>({
    control,
    name: 'inputs',
    rules: {
      validate: {
        noDuplicates: (inputs?: InputLineItem[]) => {
          const keys = (inputs || []).map(item => item.key).filter(Boolean);
          return new Set(keys).size === keys.length || 'Duplicate keys exist, only the last set value will be submitted';
        }
      }
    }
  });

  const inputs = watch('inputs') as InputLineItem[];

  useEffect(() => {
    const inputObject = reducePairs(inputs);
    onInputChange(inputObject);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(inputs), onInputChange]);

  useEffect(() => {
    replace(initialValues);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(initialValues)]);

  const onClearClick = () => replace([{ ...emptyInput }]);

  const addKeyValue = () => append({ ...emptyInput });

  const updateField = (index: number, field: 'key' | 'value', value: string) => {
    setValue(`inputs.${index}.${field}`, value);
    if (field === 'key') {
      const normalizedKey = value.toLowerCase();
      setValue(`inputs.${index}.helptip`, inputHelpTipsMap[normalizedKey]);
    }
    trigger();
  };

  return (
    <>
      {fields.map((field, index) => {
        const errorMessage = index === fields.length - 1 ? errors?.inputs?.root?.message : undefined;
        const hasRemovalDisabled = !(inputs?.[index]?.key && inputs?.[index]?.value);
        const { component: Helptip = null, props: helptipProps = {} } = (inputs[index].helptip ?? {}) as InputHelptip;
        return (
          <div className={`${classes.keyValueContainer} relative margin-bottom-x-small`} key={field.id}>
            <FormControl>
              <OutlinedInput
                disabled={disabled}
                value={inputs?.[index]?.key || ''}
                placeholder="Key"
                onChange={e => updateField(index, 'key', e.target.value)}
                type="text"
              />
              {!!errorMessage && <FormHelperText>{errorMessage}</FormHelperText>}
            </FormControl>
            <FormControl>
              <OutlinedInput
                disabled={disabled}
                value={inputs?.[index]?.value || ''}
                placeholder="Value"
                onChange={e => updateField(index, 'value', e.target.value)}
                type="text"
              />
            </FormControl>
            <div>
              {fields.length > 1 && !hasRemovalDisabled ? (
                <IconButton className={classes.lineRemovalButton} disabled={disabled} onClick={() => remove(index)}>
                  <ClearIcon fontSize="small" />
                </IconButton>
              ) : (
                <span />
              )}
            </div>
            {Helptip && <Helptip className={classes.helptip} {...helptipProps} />}
          </div>
        );
      })}
      <div className={`margin-top-x-small ${classes.keyValueContainer} align-items-center`}>
        <div>
          <IconButton
            aria-label="add-editor-line-button"
            className={classes.lineAdditionButton}
            disabled={disabled || !inputs?.[fields.length - 1]?.key || !inputs?.[fields.length - 1]?.value}
            onClick={addKeyValue}
            size="small"
          >
            <AddIcon fontSize="large" />
          </IconButton>
        </div>
        <div />
        {inputs.length > 1 ? (
          <Button variant="text" onClick={onClearClick} color="inherit">
            Clear all
          </Button>
        ) : (
          <div />
        )}
      </div>
    </>
  );
};

export const KeyValueEditor = ({ disabled, initialInput = {}, inputHelpTipsMap = {}, onInputChange }) => {
  const defaultValues = {
    inputs: Object.keys(initialInput).length
      ? Object.entries(initialInput).map(([key, value]) => ({ helptip: inputHelpTipsMap[key.toLowerCase()], key, value }) as InputLineItem)
      : [{ ...emptyInput }]
  };
  const [initialValues, setInitialValues] = useState(defaultValues);

  useEffect(() => {
    setInitialValues(defaultValues);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(initialInput)]);

  const onFormSubmit = data => onInputChange(reducePairs(data.inputs));

  return (
    <Form autocomplete="off" defaultValues={defaultValues} id="key-value-editor" initialValues={initialValues} onSubmit={onFormSubmit}>
      <KeyValueFields disabled={disabled} initialValues={defaultValues.inputs} inputHelpTipsMap={inputHelpTipsMap} onInputChange={onInputChange} />
    </Form>
  );
};

export default KeyValueEditor;
