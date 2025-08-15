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

import { ControlledCountrySelect } from '@northern.tech/common-ui/forms/CountrySelect';
import TextInput from '@northern.tech/common-ui/forms/TextInput';

import { HELPTOOLTIPS } from '../helptips/HelpTooltips';
import { MenderHelpTooltip } from '../helptips/MenderTooltip';

interface PlanExpandedFormProp {
  className: string;
  setIsValid: (disabled: boolean) => void;
}

export const PlanExpandedForm = (props: PlanExpandedFormProp) => {
  const { className, setIsValid } = props;
  const { formState } = useFormContext();

  useEffect(() => {
    if (formState.isDirty) {
      setIsValid(formState.isValid);
    }
  }, [formState, setIsValid]);

  return (
    <div className={className}>
      <div className="flexbox center-aligned">
        <TextInput hint="Billing email" label="Billing email" id="email" required validations="isEmail,trim" />
        <MenderHelpTooltip className="required" id={HELPTOOLTIPS.planUpgradeEmail.id} />
      </div>
      <TextInput hint="Company name" label="Company name" id="name" required validations="isLength:2,trim" />

      <h4>Billing address</h4>
      <TextInput hint="Address line 1" label="Address line 1" required id="line1" validations="isLength:3,trim" />
      <TextInput hint="State" label="State" required id="state" validations="isLength:2,trim" />
      <TextInput hint="City" label="City" required id="city" validations="isLength:2,trim" />
      <TextInput hint="Zip or Postal code" required label="Zip or Postal code" id="postal_code" validations="isLength:4,trim" />
      <ControlledCountrySelect required id="country" />
    </div>
  );
};
