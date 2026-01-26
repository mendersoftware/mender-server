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
import { Button, Fab, FormControl, FormHelperText, IconButton, OutlinedInput } from '@mui/material';
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
  formReset: { justifyContent: 'end' },
  helptip: { left: -35, top: theme.spacing(), position: 'absolute !important' },
  keyValueContainer: {
    display: 'grid',
    gridTemplateColumns: 'minmax(200px, min-content) minmax(200px, min-content) max-content',
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
  initialValues: InputLineItem[];
  inputHelpTipsMap: Record<string, { component: React.ComponentType<any>; props: any }>;
  onInputChange: (value: Record<string, string>) => void;
}

const KeyValueFields = ({ disabled, errortext, initialValues, inputHelpTipsMap, onInputChange }: KeyValueFieldsProps) => {
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
      <div className={`margin-top-small ${classes.keyValueContainer}`}>
        <div className="margin-left-x-small">
          <Fab disabled={disabled || !inputs?.[fields.length - 1]?.key || !inputs?.[fields.length - 1]?.value} size="small" onClick={addKeyValue}>
            <ContentAddIcon />
          </Fab>
        </div>
        <div className={`flexbox center-aligned full-width ${classes.formReset}`}>
          {inputs.length > 1 ? (
            <Button className="align-self-end" variant="text" onClick={onClearClick}>
              Clear all
            </Button>
          ) : (
            <div />
          )}
        </div>
      </div>
    </div>
  );
};

export const KeyValueEditor = ({ disabled, errortext, initialInput = {}, inputHelpTipsMap = {}, onInputChange }) => {
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
      <KeyValueFields
        disabled={disabled}
        errortext={errortext}
        initialValues={defaultValues.inputs}
        inputHelpTipsMap={inputHelpTipsMap}
        onInputChange={onInputChange}
      />
    </Form>
  );
};

export default KeyValueEditor;
