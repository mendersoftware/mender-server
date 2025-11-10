// Copyright 2022 Northern.tech AS
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
import { useEffect, useMemo, useState } from 'react';
import { Controller, useFormContext } from 'react-hook-form';
import { useDispatch, useSelector } from 'react-redux';

// material ui
import { Button, Checkbox, Divider, Drawer, FormControlLabel, InputAdornment, TextField, Typography, drawerClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { DrawerTitle } from '@northern.tech/common-ui/DrawerTitle';
import Form from '@northern.tech/common-ui/forms/Form';
import { TIMEOUTS } from '@northern.tech/store/constants';
import { getDeploymentsConfig, saveDeltaDeploymentsConfig } from '@northern.tech/store/thunks';
import { useDebounce } from '@northern.tech/utils/debouncehook';

const useStyles = makeStyles()(theme => ({
  buttonWrapper: { '&.button-wrapper': { justifyContent: 'start', alignItems: 'center' } },
  drawer: { [`.${drawerClasses.paper}`]: { maxWidth: 'initial' } },
  formWrapper: { display: 'flex', flexDirection: 'column', gap: theme.spacing(2), marginLeft: theme.spacing(2), maxWidth: 300 },
  memoryFormWrapper: { gap: theme.spacing(4), marginLeft: 0 },
  info: { maxWidth: 750 },
  memoryTitle: { alignItems: 'baseline' }
}));

const formDefaults = {
  sourceWindow: 67108864,
  inputWindow: 8388608,
  instructionBuffer: 32768,
  duplicatesWindow: 262144,
  compressionLevel: 0,
  disableChecksum: false,
  disableDecompression: false
};

const numberFields = {
  compressionLevel: { key: 'compressionLevel', title: 'Compression level', hasAdornment: false },
  duplicatesWindow: { key: 'duplicatesWindow', title: 'Compression duplicates size', hasAdornment: true },
  inputWindow: { key: 'inputWindow', title: 'Input window size', hasAdornment: true },
  instructionBuffer: { key: 'instructionBuffer', title: 'Instruction buffer size', hasAdornment: true },
  sourceWindow: { key: 'sourceWindow', title: 'Source buffer size', hasAdornment: true }
};

const NumberInputLimited = ({ limit, onChange, value: propsValue, hasAdornment, ...remainder }) => {
  const [value, setValue] = useState(propsValue);
  const debouncedValue = useDebounce(value, TIMEOUTS.oneSecond);
  const { default: defaultValue, max, min } = limit;

  useEffect(() => {
    setValue(propsValue);
  }, [propsValue]);

  useEffect(() => {
    const minimum = Math.max(min, debouncedValue);
    const allowedValue = Math.min(max ?? minimum, minimum);
    if (allowedValue !== debouncedValue) {
      setValue(allowedValue);
      return;
    }
    onChange(allowedValue);
  }, [debouncedValue, max, min, onChange]);

  return (
    <TextField
      slotProps={{
        input: hasAdornment ? { endAdornment: <InputAdornment position="end">KB</InputAdornment> } : {},
        htmlInput: { step: 1, type: 'numeric', pattern: '[0-9]*', autoComplete: 'off' }
      }}
      error={min || max ? min > value || value > max : false}
      value={value}
      onChange={({ target: { value } }) => setValue(Number(value) || 0)}
      helperText={defaultValue !== undefined && defaultValue !== value ? `Defaults to: ${defaultValue}` : null}
      {...remainder}
    />
  );
};

const ArtifactGenerationSettingsForm = ({ deltaLimits, defaultValues }) => {
  const { classes } = useStyles();
  const { control, reset } = useFormContext();

  const onResetClick = () => reset({ ...defaultValues });

  const numberInputs = useMemo(
    () => [
      { ...numberFields.sourceWindow, ...deltaLimits.sourceWindow },
      { ...numberFields.inputWindow, ...deltaLimits.inputWindow },
      { ...numberFields.instructionBuffer, ...deltaLimits.instructionBuffer },
      { ...numberFields.duplicatesWindow, ...deltaLimits.duplicatesWindow },
      { ...numberFields.compressionLevel, ...deltaLimits.compressionLevel }
    ],
    [deltaLimits.sourceWindow, deltaLimits.inputWindow, deltaLimits.instructionBuffer, deltaLimits.duplicatesWindow, deltaLimits.compressionLevel]
  );

  return (
    <>
      <Typography className={classes.info} variant="body2">
        Before adjusting these parameters, we recommend experimenting with the{' '}
        <a href="https://docs.mender.io/artifact-creation/create-a-delta-update-artifact" target="_blank" rel="noopener noreferrer">
          mender-binary-delta-generator CLI tool
        </a>{' '}
        on your own workstation to find the optimal configuration for your Artifacts. You can learn more about these parameters on the{' '}
        <a href="https://github.com/jmacd/xdelta/blob/wiki/TuningMemoryBudget.md" target="_blank" rel="noopener noreferrer">
          xdelta wiki
        </a>
        .
      </Typography>
      <Typography className="margin-top-small" display="block" variant="subtitle1">
        Compression options
      </Typography>
      <div className={classes.formWrapper}>
        <Controller
          name="disableChecksum"
          control={control}
          render={({ field: { onChange, value } }) => (
            <FormControlLabel
              control={<Checkbox color="primary" checked={!!value} onChange={e => onChange(e.target.checked)} size="small" />}
              label="Disable checksum"
            />
          )}
        />
        <Controller
          name="disableDecompression"
          control={control}
          render={({ field: { onChange, value } }) => (
            <FormControlLabel
              className="margin-top-none"
              control={<Checkbox color="primary" checked={!!value} onChange={e => onChange(e.target.checked)} size="small" />}
              label="Disable external decompression"
            />
          )}
        />
      </div>
      <div className={`flexbox margin-top-small margin-bottom ${classes.memoryTitle}`}>
        <Typography display="block" variant="subtitle1">
          Memory options
        </Typography>
        <Button className="margin-left-small" onClick={onResetClick} variant="text">
          Reset to defaults
        </Button>
      </div>
      <div className={`${classes.formWrapper} ${classes.memoryFormWrapper}`}>
        {numberInputs.map(({ default: defaultValue, key, title, min = 0, max, hasAdornment }) => (
          <Controller
            key={key}
            name={key}
            control={control}
            render={({ field: { onChange, value } }) => (
              <NumberInputLimited limit={{ default: defaultValue, max, min }} hasAdornment={hasAdornment} label={title} value={value} onChange={onChange} />
            )}
          />
        ))}
      </div>
    </>
  );
};

export const ArtifactGenerationSettings = ({ onClose, open }) => {
  const { binaryDelta: deltaConfig = {}, binaryDeltaLimits: deltaLimits = {} } = useSelector(state => state.deployments.config) ?? {};
  const dispatch = useDispatch();
  const { classes } = useStyles();

  const initialValues = useMemo(() => ({ ...formDefaults, ...deltaConfig }), [deltaConfig]);

  const defaultValues = useMemo(
    () =>
      Object.entries(deltaLimits).reduce(
        (accu, [key, value]) => {
          if (accu[key]) {
            accu[key] = value.default;
          }
          return accu;
        },
        { ...formDefaults }
      ),
    [deltaLimits]
  );

  useEffect(() => {
    dispatch(getDeploymentsConfig());
  }, [dispatch]);

  const onSubmit = async formValues => {
    try {
      await dispatch(saveDeltaDeploymentsConfig(formValues)).unwrap();
    } catch (error) {
      console.log(error);
      return;
    }
    onClose();
  };

  return (
    <Drawer anchor="right" className={classes.drawer} open={open} onClose={onClose}>
      <DrawerTitle title="Delta artifacts generation configuration" onClose={onClose} />
      <Divider className="margin-bottom" />
      <Form
        classes={{ buttonWrapper: classes.buttonWrapper, cancelButton: '' }}
        defaultValues={defaultValues}
        handleCancel={onClose}
        initialValues={initialValues}
        onSubmit={onSubmit}
        showButtons
        submitLabel="Save"
      >
        <ArtifactGenerationSettingsForm defaultValues={defaultValues} deltaLimits={deltaLimits} />
      </Form>
    </Drawer>
  );
};

export default ArtifactGenerationSettings;
