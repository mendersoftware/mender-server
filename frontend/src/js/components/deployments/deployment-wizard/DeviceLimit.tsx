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
import { useEffect, useState } from 'react';
import { useFormContext } from 'react-hook-form';

import { Checkbox, Collapse, FormControlLabel } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { DOCSTIPS, DocsTooltip } from '@northern.tech/common-ui/DocsLink';
import { InfoHintContainer } from '@northern.tech/common-ui/InfoHint';
import { NumberInput } from '@northern.tech/common-ui/forms/NumberInput';
import type { Filter } from '@northern.tech/types/MenderTypes';

import { deploymentFormSections } from './utils';

const useStyles = makeStyles()(theme => ({
  limitSelection: {
    alignItems: 'baseline',
    display: 'flex',
    marginTop: theme.spacing(2),
    marginLeft: `calc(1em + ${theme.spacing(1.5)})`
  }
}));

export const DeviceLimit = ({
  deploymentDeviceCount = 0,
  deploymentDeviceIds = [],
  filter
}: {
  deploymentDeviceCount?: number;
  deploymentDeviceIds?: string[];
  filter?: Filter;
}) => {
  const numberDevices = deploymentDeviceCount ? deploymentDeviceCount : deploymentDeviceIds ? deploymentDeviceIds.length : 0;

  const { setValue } = useFormContext();
  const [shouldLimit, setShouldLimit] = useState(false);

  const { classes } = useStyles();

  useEffect(() => {
    if (!filter) {
      setValue(deploymentFormSections.maxDevices, 0);
      setShouldLimit(false);
    }
  }, [filter, setValue]);

  const onToggleLimit = (_, checked) => {
    setShouldLimit(checked);
    if (checked) {
      setValue(deploymentFormSections.maxDevices, numberDevices);
    } else {
      setValue(deploymentFormSections.maxDevices, 0);
    }
  };

  return (
    <>
      <FormControlLabel
        control={<Checkbox color="primary" checked={shouldLimit} disabled={!filter} onChange={onToggleLimit} size="small" />}
        label={
          <div className="flexbox align-items-center">
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
          <NumberInput
            id={deploymentFormSections.maxDevices}
            min={1}
            width={100}
            rules={{
              validate: value => !shouldLimit || (Number(value) >= 1 && !isNaN(Number(value))) || 'Please enter a valid number.'
            }}
          />{' '}
          devices have attempted to apply the update
        </div>
      </Collapse>
    </>
  );
};

export default DeviceLimit;
