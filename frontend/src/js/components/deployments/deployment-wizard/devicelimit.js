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
import React, { useEffect, useState } from 'react';

import { Checkbox, Collapse, FormControl, FormControlLabel, FormHelperText, Input, formControlClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import validator from 'validator';

import { TIMEOUTS } from '../../../constants/appConstants';
import { useDebounce } from '../../../utils/debouncehook';
import { DOCSTIPS, DocsTooltip } from '../../common/docslink';
import { InfoHintContainer } from '../../common/info-hint';

const useStyles = makeStyles()(theme => ({
  limitSelection: {
    alignItems: 'baseline',
    display: 'flex',
    marginLeft: `calc(1em + ${theme.spacing(1.5)})`,
    [`.${formControlClasses.root}`]: { minWidth: 'unset', width: 100, marginLeft: theme.spacing(), marginRight: theme.spacing() }
  }
}));

export const DeviceLimit = props => {
  const { setDeploymentSettings, deploymentObject = {} } = props;
  const { deploymentDeviceCount = 0, deploymentDeviceIds = [], filter } = deploymentObject;
  const numberDevices = deploymentDeviceCount ? deploymentDeviceCount : deploymentDeviceIds ? deploymentDeviceIds.length : 0;

  const [shouldLimit, setShouldLimit] = useState(false);
  const [error, setError] = useState('');
  const [value, setValue] = useState(0);

  const { classes } = useStyles();

  const debouncedValue = useDebounce(value, TIMEOUTS.debounceDefault);

  useEffect(() => {
    if (debouncedValue > 1) {
      setDeploymentSettings({ maxDevices: debouncedValue });
    }
  }, [debouncedValue, setDeploymentSettings]);

  useEffect(() => {
    if (!shouldLimit) {
      return;
    }
    setValue(numberDevices);
  }, [numberDevices, shouldLimit]);

  useEffect(() => {
    if (!filter) {
      setDeploymentSettings({ maxDevices: 0 });
    }
  }, [filter, setDeploymentSettings]);

  const handleLimitChange = ({ target: { value } }) => {
    setError('');
    setValue(value);
    if (!validator.isNumeric(value) || value < 1) {
      setError('Please enter a valid number.');
    }
  };

  const onToggleLimit = (_, checked) => {
    setShouldLimit(checked);
    if (!checked) {
      setDeploymentSettings({ maxDevices: 0 });
    }
  };

  return (
    <>
      <FormControlLabel
        control={<Checkbox color="primary" checked={shouldLimit} disabled={!filter} onChange={onToggleLimit} size="small" />}
        label={
          <div className="flexbox center-aligned">
            <b className="margin-right-small">Limit deployment to a maximum number of devices</b> (optional)
            <InfoHintContainer>
              <DocsTooltip id={DOCSTIPS.limitedDeployments.id} />
            </InfoHintContainer>
          </div>
        }
      />
      <Collapse in={shouldLimit}>
        <div className={classes.limitSelection}>
          Finish deployment after{' '}
          <FormControl error={!!error}>
            <Input value={value} placeholder="Limit" onChange={handleLimitChange} type="text" hint={numberDevices} />
            <FormHelperText>{error}</FormHelperText>
          </FormControl>
          devices have attempted to apply the update
        </div>
      </Collapse>
    </>
  );
};

export default DeviceLimit;
