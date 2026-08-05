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
import { useEffect } from 'react';
import { useFormContext } from 'react-hook-form';

import { Checkbox, Collapse, FormControlLabel, FormHelperText } from '@mui/material';

import { DOCSTIPS, DocsTextLink } from '@northern.tech/common-ui/DocsLink';
import { InfoHintContainer } from '@northern.tech/common-ui/InfoHint';
import { NumberInput } from '@northern.tech/common-ui/forms/NumberInput';

import type { DeploymentFormValues } from './types';
import { deploymentFormSections, useDerivedData, useValidatedSetValue } from './utils';

export const DeviceLimit = () => {
  const { watch } = useFormContext<DeploymentFormValues>();
  const setValue = useValidatedSetValue();
  const { deploymentDeviceCount, deploymentDeviceIds, filter } = useDerivedData(watch);
  const numberDevices = deploymentDeviceCount ? deploymentDeviceCount : deploymentDeviceIds ? deploymentDeviceIds.length : 0;
  const shouldLimit = watch(deploymentFormSections.shouldLimit);

  useEffect(() => {
    if (!filter) {
      setValue(deploymentFormSections.maxDevices, 0);
      setValue(deploymentFormSections.shouldLimit, false);
    }
  }, [filter, setValue]);

  const onToggleLimit = (_, checked) => {
    setValue(deploymentFormSections.shouldLimit, checked);
    if (checked) {
      setValue(deploymentFormSections.maxDevices, numberDevices);
    } else {
      setValue(deploymentFormSections.maxDevices, 0);
    }
  };

  return (
    <>
      <FormControlLabel
        control={<Checkbox className="margin-left-small" color="primary" checked={shouldLimit} disabled={!filter} onChange={onToggleLimit} size="small" />}
        label={
          <div className="flexbox align-items-center">
            Limit deployment to a maximum number of devices
            <InfoHintContainer>
              <DocsTextLink id={DOCSTIPS.limitedDeployments.id} />
            </InfoHintContainer>
          </div>
        }
      />
      <Collapse in={shouldLimit}>
        <NumberInput id={deploymentFormSections.maxDevices} min={1} width={120} showSteps size="small" />
        <FormHelperText className="margin-left-small margin-top-x-small margin-bottom-small">
          The deployment will automatically finish after this many devices have attempted to update.
        </FormHelperText>
      </Collapse>
    </>
  );
};

export default DeviceLimit;
