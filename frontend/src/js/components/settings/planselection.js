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
import React from 'react';

import { makeStyles } from 'tss-react/mui';

import { PLANS } from '../../constants/appConstants';
import { isDarkMode } from '../../helpers.js';
import InfoText from '../common/infotext';

export const useStyles = makeStyles()(theme => ({
  planNote: { marginBottom: -11, fontSize: 'smaller' },
  planPanel: {
    borderColor: theme.palette.background.lightgrey,
    ['&.active,&:hover']: {
      borderColor: theme.palette.grey[50],
      boxShadow: '0 1px 6px rgba(0, 0, 0, 0.15)'
    },
    ['&.active']: {
      backgroundColor: isDarkMode(theme.palette.mode) ? theme.palette.grey[50] : theme.palette.grey[400]
    },
    '&.addon': {
      alignItems: 'center',
      columnGap: 10,
      display: 'grid',
      gridTemplateColumns: 'max-content minmax(150px, min-content) 1fr max-content',
      height: 'initial',
      marginTop: 10,
      minWidth: 550,
      width: 'initial'
    },
    '&.addon.upgrade': {
      gridTemplateColumns: 'max-content minmax(120px, min-content) minmax(120px, min-content) 1fr max-content'
    }
  },
  price: { fontSize: '1rem' }
}));

export const PlanSelection = ({ currentPlan = PLANS.os.id, isTrial, offerValid, offerTag, setUpdatedPlan, updatedPlan }) => {
  const { classes } = useStyles();
  const canUpgrade = plan => Object.keys(PLANS).indexOf(plan) >= Object.keys(PLANS).indexOf(currentPlan);
  const onPlanSelect = plan => (isTrial || canUpgrade(plan) ? setUpdatedPlan(plan) : undefined);
  return (
    <>
      <h3 className="margin-top">{isTrial ? '1. Choose a plan' : 'Plans'}</h3>
      <div className="flexbox space-between" style={{ paddingBottom: 15 }}>
        {Object.values(PLANS).map(item => (
          <div
            key={item.id}
            className={`planPanel ${classes.planPanel} ${updatedPlan === item.id ? 'active' : ''} ${isTrial || canUpgrade(item.id) ? '' : 'muted'}`}
            onClick={() => onPlanSelect(item.id)}
          >
            {!isTrial && canUpgrade(item.id) && (
              <div className={`uppercased align-center muted ${classes.planNote}`}>{item.id === currentPlan ? 'current plan' : 'upgrade'}</div>
            )}
            <h4>
              {item.name} {item.offer && isTrial && offerValid ? offerTag : null}
            </h4>
            <div>
              {item.offer && isTrial && offerValid ? (
                <>
                  <div className={`link-color bold ${classes.price}`}>{item.offerprice}</div>
                  <div className="pre-line">{item.price2}</div>
                </>
              ) : (
                <>
                  <div className={`link-color bold ${classes.price}`}>{item.price}</div>
                  <div>{item.deviceCount}</div>
                </>
              )}
            </div>
            <ul className="unstyled">
              {item.features.map((feature, index) => (
                <li key={`${item.id}-feature-${index}`}>
                  <InfoText variant="dense">{feature}</InfoText>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
    </>
  );
};

export default PlanSelection;
