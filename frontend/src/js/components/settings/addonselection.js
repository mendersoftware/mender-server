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
import React, { useMemo } from 'react';

import { Checkbox } from '@mui/material';

import { ADDONS, PLANS } from '../../constants/appConstants';
import InfoText from '../common/infotext';
import { useStyles } from './planselection';

export const AddOnSelection = ({ addons = [], features, onChange, updatedPlan = PLANS.os.id }) => {
  const { classes } = useStyles();
  const onAddOnClick = (e, name, enabled) => {
    if (e.target.tagName === 'A') {
      return;
    }
    let changedAddOns = addons.filter(item => item.name !== name);
    if (enabled) {
      changedAddOns.push({ name, enabled });
    }
    onChange(changedAddOns);
  };

  const isUpgrade = Object.keys(PLANS).indexOf(updatedPlan) < Object.keys(PLANS).length - 1;
  const relevantAddons = useMemo(
    () =>
      Object.entries(ADDONS).reduce((accu, [addOnName, addOn]) => {
        if (!addOn.needs.every(need => features[need])) {
          return accu;
        }
        const isEnabled = addons.some(orgAddOn => orgAddOn.enabled && addOnName === orgAddOn.name);
        accu.push({ ...addOn, name: addOnName, isEnabled });
        return accu;
      }, []),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [JSON.stringify(addons), JSON.stringify(features)]
  );

  return (
    <>
      <h3 className="margin-top-large">Get more features with add-ons</h3>

      <div className="flexbox column">
        <p>
          Extend Mender features with our add-ons. Select one or more from the list and submit to request add-ons to be added to your plan. We&apos;ll adjust
          your subscription and confirm it with you.
        </p>
        {relevantAddons.map(addOn => {
          const isEligible = addOn.eligible.indexOf(updatedPlan) > -1;

          return (
            <div
              key={addOn.name}
              className={`planPanel ${classes.planPanel} addon ${isUpgrade ? 'upgrade' : ''} ${addOn.isEnabled ? 'active' : ''} ${isEligible ? '' : 'muted'}`}
              onClick={e => (isEligible ? onAddOnClick(e, addOn.name, !addOn.isEnabled) : () => false)}
            >
              <Checkbox disabled={!isEligible} checked={addOn.isEnabled} />
              <div className="bold">{addOn.title}</div>
              {isUpgrade && (
                <div className="flexbox column">
                  <div className={`link-color bold ${classes.price}`}>{addOn[updatedPlan].price}</div>
                  <div>{addOn[updatedPlan].deviceCount}</div>
                </div>
              )}
              <InfoText variant="dense">{addOn.description}</InfoText>
              {/* eslint-disable-next-line react/jsx-no-target-blank */}
              <a className="margin-left-small" href="https://mender.io/plans/features" target="_blank" rel="noopener">
                Learn more
              </a>
            </div>
          );
        })}
      </div>
    </>
  );
};

export default AddOnSelection;
