// Copyright 2019 Northern.tech AS
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
import React, { useState } from 'react';
import { useSelector } from 'react-redux';
import { Link } from 'react-router-dom';

import { Chip } from '@mui/material';
import { withStyles } from 'tss-react/mui';

import { ADDONS, BENEFITS, PLANS } from '@northern.tech/store/constants';
import { getTenantCapabilities } from '@northern.tech/store/selectors';

import MenderTooltip, { MenderTooltipClickable } from './mendertooltip';

const PlansTooltip = withStyles(MenderTooltip, ({ palette }) => ({
  arrow: {
    color: palette.tooltip.tierTipBackground
  },
  tooltip: {
    backgroundColor: palette.tooltip.tierTipBackground,
    maxWidth: 300
  }
}));

const PlanChip = withStyles(Chip, ({ palette }) => ({
  root: {
    backgroundColor: palette.tooltip.tierTipBackground,
    color: palette.text.disabled,
    textTransform: 'uppercase',
    '&:hover': {
      fontWeight: 'bold'
    }
  }
}));

export const DefaultUpgradeNotification = props => (
  <div {...props}>
    This feature is not available on your plan. <Link to="/settings/upgrade">Upgrade</Link> to enable it
  </div>
);

const EnterpriseNotification = ({ className = '', id = BENEFITS.default.id }) => {
  const [isOpen, setIsOpen] = useState(false);
  const tenantCapabilities = useSelector(getTenantCapabilities);
  const { isEnterprise, plan: currentPlan } = tenantCapabilities;
  const { benefit, requiredAddon = '', requiredPlan = PLANS.os.id } = BENEFITS[id];
  const hasAddon = requiredAddon ? ADDONS[requiredAddon].needs.every(need => tenantCapabilities[need]) : false;

  const currentPlanIndex = Object.keys(PLANS).indexOf(currentPlan);
  const requiredPlanIndex = Object.keys(PLANS).indexOf(requiredPlan);
  const shouldShow = requiredPlanIndex > currentPlanIndex;
  // we have to explicitly check for the plan requirement here, since the default value prevents us from relying on the `shouldShow` result
  if (isEnterprise || (BENEFITS[id].requiredPlan && !shouldShow) || (requiredAddon && hasAddon)) {
    return null;
  }
  const content = requiredAddon ? (
    <>
      Add the <b>{ADDONS[requiredAddon].title}</b> add-on to {benefit}.
    </>
  ) : (
    <>
      Upgrade to the <b>{PLANS[requiredPlan].name}</b> plan {requiredPlanIndex === Object.keys(PLANS).length - 1 ? '' : 'or higher '}to gain access to {benefit}
      .
    </>
  );
  return (
    <MenderTooltipClickable
      onOpenChange={setIsOpen}
      title={
        <div>
          {content}
          <div className="flexbox space-between margin-top-small">
            <Link to="/settings/upgrade">Upgrade now</Link>
            <span className="link" onClick={() => setIsOpen(false)}>
              Close
            </span>
          </div>
        </div>
      }
      tooltipComponent={PlansTooltip}
      visibility={isOpen}
    >
      <PlanChip className={className} label={PLANS[requiredPlan].name} />
    </MenderTooltipClickable>
  );
};

export default EnterpriseNotification;
