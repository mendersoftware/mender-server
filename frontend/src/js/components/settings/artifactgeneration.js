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
import React, { useEffect, useMemo, useRef, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

// material ui
import { InfoOutlined as InfoOutlinedIcon } from '@mui/icons-material';
import { Checkbox, FormControlLabel, TextField, Typography, formControlLabelClasses, textFieldClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import DeltaIcon from '../../../assets/img/deltaicon.svg';
import { getDeploymentsConfig, saveDeltaDeploymentsConfig } from '../../actions/deploymentActions';
import { BENEFITS, TIMEOUTS } from '../../constants/appConstants';
import { useDebounce } from '../../utils/debouncehook';
import EnterpriseNotification from '../common/enterpriseNotification';
import InfoText from '../common/infotext';
import Loader from '../common/loader';

const useStyles = makeStyles()(theme => ({
  deviceLimitBar: { backgroundColor: theme.palette.grey[500], margin: '15px 0' },
  wrapper: {
    backgroundColor: theme.palette.background.lightgrey,
    display: 'flex',
    flexDirection: 'column',
    marginTop: theme.spacing(6),
    padding: theme.spacing(2),
    paddingTop: 0,
    '&>h5': { marginTop: 0, marginBottom: 0 },
    '.flexbox>span': { alignSelf: 'flex-end' },
    [`.${textFieldClasses.root}`]: { maxWidth: 200, minWidth: 100 },
    [`.${formControlLabelClasses.root}`]: { marginTop: theme.spacing() }
  }
}));

const numberFields = {
  compressionLevel: { key: 'compressionLevel', title: 'Compression level' },
  sourceWindow: { key: 'sourceWindow', title: 'Source window size' },
  inputWindow: { key: 'inputWindow', title: 'Input window size' },
  duplicatesWindow: { key: 'duplicatesWindow', title: 'Compression duplicates window' },
  instructionBuffer: { key: 'instructionBuffer', title: 'Instruction buffer size' }
};

const NumberInputLimited = ({ limit, onChange, value: propsValue, ...remainder }) => {
  const [value, setValue] = useState(propsValue);
  const debouncedValue = useDebounce(value, TIMEOUTS.oneSecond);
  const { default: defaultValue, max, min } = limit;

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
      inputProps={{ step: 1, type: 'numeric', pattern: '[0-9]*', autoComplete: 'off' }}
      InputLabelProps={{ shrink: true }}
      error={min || max ? min > value || value > max : false}
      value={value}
      onChange={({ target: { value } }) => setValue(Number(value) || 0)}
      helperText={defaultValue !== undefined ? `Defaults to: ${defaultValue}` : null}
      {...remainder}
    />
  );
};

export const ArtifactGenerationSettings = () => {
  const { binaryDelta: deltaConfig = {}, binaryDeltaLimits: deltaLimits = {}, hasDelta: deltaEnabled } = useSelector(state => state.deployments.config) ?? {};
  const dispatch = useDispatch();
  const [timeoutValue, setTimeoutValue] = useState(deltaConfig.timeout);
  const [disableChecksum, setDisableChecksum] = useState(deltaConfig.disableChecksum);
  const [disableDecompression, setDisableDecompression] = useState(deltaConfig.disableChecksum);
  const [compressionLevel, setCompressionLevel] = useState(deltaConfig.compressionLevel);
  const [sourceWindow, setSourceWindow] = useState(deltaConfig.sourceWindow);
  const [inputWindow, setInputWindow] = useState(deltaConfig.inputWindow);
  const [duplicatesWindow, setDuplicatesWindow] = useState(deltaConfig.duplicatesWindow);
  const [instructionBuffer, setInstructionBuffer] = useState(deltaConfig.instructionBuffer);
  const isInitialized = useRef(false);

  const { classes } = useStyles();

  useEffect(() => {
    dispatch(getDeploymentsConfig());
  }, [dispatch]);

  useEffect(() => {
    if (deltaConfig.timeout === -1) {
      return;
    }
    const { timeout, duplicatesWindow, compressionLevel, disableChecksum, disableDecompression, inputWindow, instructionBuffer, sourceWindow } = deltaConfig;
    setDisableChecksum(disableChecksum);
    setDisableDecompression(disableDecompression);
    setCompressionLevel(compressionLevel);
    setTimeoutValue(timeout);
    setSourceWindow(sourceWindow);
    setInputWindow(inputWindow);
    setDuplicatesWindow(duplicatesWindow);
    setInstructionBuffer(instructionBuffer);
    setTimeout(() => (isInitialized.current = true), TIMEOUTS.fiveSeconds);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(deltaConfig), JSON.stringify(deltaLimits)]);

  const debouncedConfig = useDebounce(
    {
      timeout: timeoutValue,
      duplicatesWindow,
      compressionLevel,
      disableChecksum,
      disableDecompression,
      inputWindow,
      instructionBuffer,
      sourceWindow
    },
    TIMEOUTS.threeSeconds
  );

  useEffect(() => {
    if (!isInitialized.current) {
      return;
    }
    dispatch(saveDeltaDeploymentsConfig(debouncedConfig));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dispatch, JSON.stringify(debouncedConfig)]);

  const numberInputs = useMemo(() => {
    return [
      { ...numberFields.compressionLevel, setter: setCompressionLevel, value: compressionLevel, ...deltaLimits.compressionLevel },
      { ...numberFields.sourceWindow, setter: setSourceWindow, value: sourceWindow, ...deltaLimits.sourceWindow },
      { ...numberFields.inputWindow, setter: setInputWindow, value: inputWindow, ...deltaLimits.inputWindow },
      { ...numberFields.duplicatesWindow, setter: setDuplicatesWindow, value: duplicatesWindow, ...deltaLimits.duplicatesWindow },
      { ...numberFields.instructionBuffer, setter: setInstructionBuffer, value: instructionBuffer, ...deltaLimits.instructionBuffer }
    ];
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    compressionLevel,
    setCompressionLevel,
    setSourceWindow,
    sourceWindow,
    inputWindow,
    setInputWindow,
    setDuplicatesWindow,
    duplicatesWindow,
    setInstructionBuffer,
    instructionBuffer,
    // eslint-disable-next-line react-hooks/exhaustive-deps
    JSON.stringify(deltaLimits)
  ]);

  return (
    <div className={`flexbox column ${classes.wrapper}`}>
      <div className="flexbox center-aligned">
        <DeltaIcon />
        <h5 className="margin-left-small">Delta artifacts generation configuration</h5>
        <EnterpriseNotification className="margin-left-small" id={BENEFITS.deltaGeneration.id} />
      </div>
      {deltaEnabled ? (
        <>
          {isInitialized.current ? (
            <div className="margin-small margin-top-none">
              <div className="flexbox">
                <NumberInputLimited
                  limit={{ default: deltaLimits.timeout.default, min: deltaLimits.timeout.min, max: deltaLimits.timeout.max }}
                  label="Timeout"
                  onChange={setTimeoutValue}
                  value={timeoutValue}
                />
                <span className="margin-left-small muted slightly-smaller">seconds</span>
              </div>
              <Typography className="margin-top-small" display="block" variant="caption">
                xDelta3 arguments
              </Typography>
              <div className="flexbox column margin-left">
                <FormControlLabel
                  control={
                    <Checkbox color="primary" checked={disableChecksum} onChange={({ target: { checked } }) => setDisableChecksum(checked)} size="small" />
                  }
                  label="Disable checksum"
                />
                <FormControlLabel
                  control={
                    <Checkbox
                      color="primary"
                      checked={disableDecompression}
                      onChange={({ target: { checked } }) => setDisableDecompression(checked)}
                      size="small"
                    />
                  }
                  label="Disable external decompression"
                />
                {numberInputs.map(({ default: defaultValue, key, setter, title, value, min = 0, max }) => (
                  <NumberInputLimited key={key} limit={{ default: defaultValue, max, min }} label={title} value={value} onChange={setter} />
                ))}
              </div>
            </div>
          ) : (
            <Loader show />
          )}
        </>
      ) : (
        <InfoText>
          <InfoOutlinedIcon style={{ fontSize: 14, margin: '0 4px 4px 0', verticalAlign: 'middle' }} />
          Automatic delta artifacts generation is not enabled in your account. If you want to start using this feature,{' '}
          <a href="mailto:support@mender.io" target="_blank" rel="noopener noreferrer">
            contact our team
          </a>
          .
        </InfoText>
      )}
    </div>
  );
};

export default ArtifactGenerationSettings;
