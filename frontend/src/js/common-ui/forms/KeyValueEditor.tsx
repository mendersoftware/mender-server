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
import { CSSProperties, ComponentType, useEffect, useState } from 'react';
import { useFieldArray, useFormContext } from 'react-hook-form';

import { Clear as ClearIcon, Add as ContentAddIcon } from '@mui/icons-material';
import { Fab, FormControl, FormHelperText, IconButton, OutlinedInput } from '@mui/material';
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

const useStyles = makeStyles()(theme => ({
  spacer: { minWidth: theme.spacing(30) },
  helptip: { left: -35, top: theme.spacing(2), '&.relative': { position: 'absolute' } },
  keyValueContainer: {
    display: 'grid',
    gridTemplateColumns: 'min-content min-content max-content',
    columnGap: theme.spacing(2),
    alignItems: 'baseline',
    justifyItems: 'baseline',
    '> div': {
      marginTop: 10
    }
  }
}));

interface KeyValueFieldsProps {
  disabled?: boolean;
  errortext?: string;
  inputHelpTipsMap: Record<string, { component: React.ComponentType<any>; props: any }>;
  onInputChange: (value: Record<string, string>) => void;
}

const KeyValueFields = ({ disabled, errortext, inputHelpTipsMap, onInputChange }: KeyValueFieldsProps) => {
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
    <div>
      {fields.map((field, index) => {
        const hasError = Boolean(index === fields.length - 1 && (errortext || errors?.inputs?.root?.message));
        const hasRemovalDisabled = !(inputs?.[index]?.key && inputs?.[index]?.value);
        const { component: Helptip = null, props: helptipProps = {} } = (inputs[index].helptip ?? {}) as InputHelptip;
        return (
          <div className={`${classes.keyValueContainer} relative`} key={field.id}>
            <FormControl>
              <OutlinedInput
                disabled={disabled}
                value={inputs?.[index]?.key || ''}
                placeholder="Key"
                onChange={e => updateField(index, 'key', e.target.value)}
                type="text"
              />
              {hasError && <FormHelperText>{errortext || errors?.inputs?.root?.message}</FormHelperText>}
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
            {fields.length > 1 && !hasRemovalDisabled ? (
              <IconButton disabled={disabled} onClick={() => remove(index)} size="large">
                <ClearIcon fontSize="small" />
              </IconButton>
            ) : (
              <span />
            )}
            {Helptip && <Helptip className={classes.helptip} {...helptipProps} />}
          </div>
        );
      })}
      <div className={classes.keyValueContainer}>
        <div className={classes.spacer}>
          <Fab
            disabled={disabled || !inputs?.[fields.length - 1]?.key || !inputs?.[fields.length - 1]?.value}
            style={{ marginBottom: 10 }}
            size="small"
            onClick={addKeyValue}
          >
            <ContentAddIcon />
          </Fab>
        </div>
        <div className={classes.spacer} />
        {inputs.length > 1 ? <a onClick={onClearClick}>clear all</a> : <div />}
      </div>
    </div>
  );
};

export const KeyValueEditor = ({ disabled, errortext, initialInput = {}, inputHelpTipsMap = {}, onInputChange }) => {
  const defaultValues = {
    inputs: Object.keys(initialInput).length
      ? Object.entries(initialInput).map(([key, value]) => ({ helptip: inputHelpTipsMap[key.toLowerCase()], key, value }))
      : [{ ...emptyInput }]
  };
  const [initialValues] = useState(defaultValues);

  const onFormSubmit = data => onInputChange(reducePairs(data.inputs));

  return (
    <Form autocomplete="off" defaultValues={defaultValues} id="key-value-editor" initialValues={initialValues} onSubmit={onFormSubmit}>
      <KeyValueFields disabled={disabled} errortext={errortext} inputHelpTipsMap={inputHelpTipsMap} onInputChange={onInputChange} />
    </Form>
  );
};

export default KeyValueEditor;
