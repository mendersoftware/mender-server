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
import { InfoOutlined as InfoIcon } from '@mui/icons-material';
import type { SelectProps } from '@mui/material';
import { FormHelperText, InputAdornment, MenuItem, Select, Tooltip, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import type { NumberFieldRootProps } from '@base-ui/react/number-field';
import { NumberField } from '@northern.tech/common-ui/forms/NumberField';
import pluralize from 'pluralize';

import { delayUnits } from './constants';
import { type PhaseMessage, formatDeviceCount } from './utils';

const useStyles = makeStyles()(theme => ({
  batchInputWrapper: { alignItems: 'center', display: 'grid', gridTemplateColumns: '50% max-content max-content' },
  delayInputWrapper: { display: 'grid', gridTemplateColumns: 'min-content min-content', columnGap: theme.spacing() }
}));

interface DelayInputProps {
  delay: number;
  delayUnit: keyof typeof delayUnits;
  id: string;
  onDelayChange: NumberFieldRootProps['onValueChange'];
  onDelayUnitChange: SelectProps['onChange'];
}

export const DelayInput = ({ id, delay, delayUnit, onDelayChange, onDelayUnitChange }: DelayInputProps) => {
  const { classes } = useStyles();
  return (
    <div className={classes.delayInputWrapper}>
      <NumberField id={id} value={Number(delay) || null} onValueChange={onDelayChange} min={1} max={720} />
      <Select onChange={onDelayUnitChange} value={delayUnit || delayUnits.hours} style={{ minWidth: 'initial' }}>
        {Object.keys(delayUnits).map(value => (
          <MenuItem key={value} value={value}>
            <div className="capitalized-start">{value}</div>
          </MenuItem>
        ))}
      </Select>
    </div>
  );
};

interface BatchSizeInputProps {
  deviceCount: number;
  disabled?: boolean;
  hasError?: boolean;
  isPercentageMode: boolean;
  max?: number;
  messages?: PhaseMessage[];
  min?: number;
  onChange: NumberFieldRootProps['onValueChange'];
  value: number | undefined;
}

export const BatchSizeInput = ({
  deviceCount,
  value,
  onChange,
  isPercentageMode,
  hasError = false,
  max,
  min = 1,
  disabled = false,
  messages = []
}: BatchSizeInputProps) => {
  const { classes } = useStyles();

  return (
    <>
      <div className={classes.batchInputWrapper}>
        <NumberField
          value={value}
          onValueChange={onChange}
          endAdornment={isPercentageMode ? <InputAdornment position="end">%</InputAdornment> : undefined}
          disabled={disabled}
          error={hasError}
          size="small"
          step={1}
          min={min}
          max={max}
        />
        <Typography className="margin-left-x-small nowrap" variant="caption">
          ({formatDeviceCount(deviceCount)} {pluralize('device', deviceCount)})
        </Typography>
        {!hasError && deviceCount >= 1000 && (
          <Tooltip arrow title={`${(deviceCount ?? 0).toLocaleString()} ${pluralize('device', deviceCount)}`}>
            <InfoIcon className="margin-left-x-small" fontSize="small" />
          </Tooltip>
        )}
      </div>
      {messages.map(({ message, severity }, i) => (
        <FormHelperText className="margin-top-x-small" key={i} error={severity === 'error'}>
          {message}
        </FormHelperText>
      ))}
    </>
  );
};
