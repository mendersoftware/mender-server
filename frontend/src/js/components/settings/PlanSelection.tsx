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
import React, { useEffect } from 'react';

import { Check as CheckIcon, DeveloperBoard as DeveloperBoardIcon } from '@mui/icons-material';
import { Button } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { AvailablePlans, PLANS, Plan } from '@northern.tech/store/constants';
import { isEmpty } from '@northern.tech/utils/helpers';

import { clearLocalStorageEntry, isUpgrade } from './Upgrade';

const useStyles = makeStyles()(theme => ({
  planIcon: { color: theme.palette.primary.main, fontSize: '16px', alignSelf: 'center' },
  planFeature: { fontSize: '12px' },
  planButton: { alignSelf: 'center' },
  planPanel: {
    padding: '22px 16px',
    borderRadius: '5px',
    verticalAlign: 'top',
    width: '32%',
    display: 'flex',
    flexDirection: 'column',
    justifyContent: 'space-between',
    marginRight: theme.spacing(2),
    border: `1px solid ${theme.palette.grey[600]}`,
    fontSize: '14px',
    fontWeight: 500,
    '&.muted, &.muted:hover': {
      cursor: 'default'
    }
  }
}));
const canUpgrade = (plan: AvailablePlans, currentPlan: AvailablePlans) => Object.keys(PLANS).indexOf(plan) > Object.keys(PLANS).indexOf(currentPlan);

const isPlanDisabled = (item: Plan, currentPlan: AvailablePlans, isTrial: boolean, isUpgradePending: Partial<{ plan: AvailablePlans }>) =>
  (isUpgradePending.plan && !canUpgrade(item.id, isUpgradePending.plan)) || (!isTrial && !canUpgrade(item.id, currentPlan));
const isPlanMuted = (item: Plan, currentPlan: AvailablePlans, isTrial: boolean, isUpgradePending: Partial<{ plan: AvailablePlans }>) =>
  isPlanDisabled(item, currentPlan, isTrial, isUpgradePending) &&
  (item.id !== currentPlan || isTrial) &&
  !(isUpgradePending.plan && isUpgradePending.plan === item.id);
export const PlanSelection = ({ orgId, currentPlan = PLANS.os.id, isTrial, setUpdatedPlan, updatedPlan }) => {
  const { classes } = useStyles();
  const onPlanSelect = (plan: AvailablePlans) => (isTrial || canUpgrade(plan, currentPlan) ? setUpdatedPlan(plan) : undefined);
  const isUpgradePending = isUpgrade(orgId);

  useEffect(() => {
    if (!isEmpty(isUpgradePending) && currentPlan === isUpgradePending.plan) {
      clearLocalStorageEntry(orgId, isUpgradePending.plan);
    }
  }, [currentPlan, isUpgradePending, orgId]);
  const getPlanButtonLabel = (planId: AvailablePlans) => {
    if (!isEmpty(isUpgradePending) && isUpgradePending.plan === planId) return 'pending';
    if (isTrial) return planId === PLANS.enterprise.id ? 'contact us' : 'subscribe';
    if (currentPlan === planId) return 'current plan';
    return planId === PLANS.enterprise.id ? 'contact us' : 'upgrade';
  };

  return (
    <>
      <p className="margin-top">Your current plan: {isTrial ? ' Free trial' : `Mender ${PLANS[currentPlan].name}`}</p>
      <div className="flexbox margin-bottom-small">
        {Object.values(PLANS).map(item => (
          <div
            key={item.id}
            className={`planPanel ${classes.planPanel} ${updatedPlan === item.id ? 'active' : ''} ${isPlanMuted(item, currentPlan, isTrial, isUpgradePending) ? 'muted' : ''}`}
          >
            <div>
              <p className="margin-none">Mender {item.name}</p>
              <h2>{item.price}</h2>
              <div className="margin-bottom">
                <div className="flexbox align-center">
                  <DeveloperBoardIcon fontSize="medium" className={classes.planIcon} />
                  <div className="margin-left-x-small">
                    {item.deviceCount}
                    {item.id === PLANS.enterprise.id ? '' : '*'}
                  </div>
                </div>
              </div>
              <ul className="unstyled">
                {item.features.map((feature, index) => (
                  <li className="flexbox margin-bottom-x-small slightly-smaller" key={`${item.id}-feature-${index}`}>
                    <CheckIcon className={classes.planIcon} />
                    <div className="margin-left-x-small">{feature}</div>{' '}
                  </li>
                ))}
              </ul>
            </div>
            <Button
              className={classes.planButton}
              onClick={() => onPlanSelect(item.id)}
              variant="contained"
              color={item.id === PLANS.enterprise.id ? 'secondary' : 'primary'}
              disabled={isPlanDisabled(item, currentPlan, isTrial, isUpgradePending)}
            >
              {getPlanButtonLabel(item.id)}
            </Button>
          </div>
        ))}
      </div>
    </>
  );
};

export default PlanSelection;
