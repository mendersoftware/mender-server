// Copyright 2020 Northern.tech AS
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
import { useSelector } from 'react-redux';

import { ConfirmUpgrade } from '@northern.tech/common-ui/dialogs/ConfirmUpgrade';
import { AvailablePlans, PLANS, Plan } from '@northern.tech/store/constants';
import { Addon } from '@northern.tech/store/organizationSlice/types';
import { getFeatures, getOrganization } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { requestPlanChange } from '@northern.tech/store/thunks';

import AddOnSelection from './AddonSelection';
import { EnterpriseRequestExpanded } from './EnterpriseRequestExpanded';
import { PlanExpanded } from './PlanExpanded';
import PlanSelection from './PlanSelection';

export const updateLocalStorage = (orgId: string, name: string, isAdd: boolean) => {
  const currentState = JSON.parse(localStorage.getItem(orgId + '_upgrades') || '{}') || {};
  currentState[name] = { pending: true, isAdd };
  localStorage.setItem(orgId + '_upgrades', JSON.stringify(currentState));
};
export const clearLocalStorageEntry = (orgId: string, name: string) => {
  const currentState = JSON.parse(localStorage.getItem(orgId + '_upgrades') || '{}') || {};
  delete currentState[name];
  localStorage.setItem(orgId + '_upgrades', JSON.stringify(currentState));
};

export const isUpgrade = (orgId: string): Partial<{ plan: AvailablePlans }> => {
  const currentState = JSON.parse(localStorage.getItem(orgId + '_upgrades') || '{}') || {};
  let upgrade = {};
  Object.keys(PLANS).forEach(key => {
    if (currentState[key]) {
      upgrade = { plan: PLANS[key].id };
    }
  });
  return upgrade;
};
export const addOnsToString = (addons: Addon[] = []) =>
  addons
    .reduce((accu: string[], item) => {
      if (item.enabled) {
        accu.push(item.name);
      }
      return accu;
    }, [])
    .join(', ');
export const PricingContactNote = () => (
  <p>
    * Device limits can be adjusted on request; prices can change for larger limits. If you have any questions about the plan pricing or device limits,{' '}
    <a href="mailto:support@mender.io" target="_blank" rel="noopener noreferrer">
      contact our team.
    </a>
  </p>
);

const upgradeNotes = {
  default: {
    title: 'Upgrades and add-ons',
    description: 'Upgrade your plan or purchase an add-on to connect more devices, access more features and advanced support.'
  },
  trial: {
    title: 'Upgrade now',
    description: 'Upgrade to one of our plans to connect more devices, continue using advanced features, and get access to support.'
  }
};

export const Upgrade = () => {
  const [addOns, setAddOns] = useState<Addon[]>([]);
  const [updatedPlan, setUpdatedPlan] = useState<string>(PLANS.os.id);
  const [selectedPlan, setSelectedPlan] = useState<Plan | null>(null);

  const dispatch = useAppDispatch();
  const features = useSelector(getFeatures);
  const org = useSelector(getOrganization);
  const { addons: orgAddOns = [], plan: currentPlan = PLANS.os.id as AvailablePlans, trial: isTrial = true } = org;

  useEffect(() => {
    const currentAddOns = orgAddOns.reduce((accu: Addon[], addon) => {
      if (addon.enabled) {
        accu.push(addon);
      }
      return accu;
    }, []);
    const plan = Object.values(PLANS).find(plan => plan.id === (isTrial ? PLANS.os.id : currentPlan));
    setAddOns(currentAddOns);
    if (plan) {
      setUpdatedPlan(plan.id);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentPlan, isTrial, JSON.stringify(orgAddOns)]);

  const selectPlan = (plan: string) => {
    setUpdatedPlan(plan);
    setSelectedPlan(PLANS[plan]);
  };
  const onEnterpriseRequest = ({ message, selectedAddons }) => {
    onSendRequest(message, selectedAddons.join(', '));
  };

  const onSendRequest = async (message = '', requestedAddons = '') => {
    try {
      await dispatch(
        requestPlanChange({
          tenantId: org.id,
          content: {
            current_plan: PLANS[org.plan || PLANS.os.id].name,
            requested_plan: PLANS[updatedPlan].name,
            current_addons: addOnsToString(org.addons) || '-',
            requested_addons: requestedAddons || addOnsToString(org.addons) || '-',
            user_message: message
          }
        })
      ).unwrap();
    } catch (error) {
      console.error(error);
      return;
    }
    updateLocalStorage(org.id, PLANS[updatedPlan].id, true);
    setSelectedPlan(null);
  };
  const { description, title } = isTrial ? upgradeNotes.trial : upgradeNotes.default;
  return (
    <div style={{ maxWidth: 750 }}>
      <h2>{title}</h2>
      <p className="margin-bottom-small">{description}</p>
      <p className="margin-bottom-large">
        See the full details of plans and features at {/* eslint-disable-next-line react/jsx-no-target-blank */}
        <a href="https://mender.io/plans/pricing" target="_blank" rel="noopener">
          mender.io/plans/pricing
        </a>
      </p>
      <PlanSelection currentPlan={currentPlan} isTrial={isTrial} setUpdatedPlan={selectPlan} updatedPlan={updatedPlan} orgId={org.id} />

      <PricingContactNote />
      {!isTrial && <AddOnSelection org={org} currentPlan={currentPlan} addons={orgAddOns} features={features} isTrial={isTrial} />}

      {isTrial && selectedPlan && updatedPlan !== PLANS.enterprise.id ? (
        <PlanExpanded isEdit={false} plan={selectedPlan} organization={org} onCloseClick={() => setSelectedPlan(null)} />
      ) : updatedPlan === PLANS.enterprise.id && selectedPlan ? (
        <EnterpriseRequestExpanded addons={addOns} onClose={() => setSelectedPlan(null)} onSendRequest={onEnterpriseRequest} />
      ) : (
        selectedPlan &&
        selectedPlan.id !== currentPlan && (
          <ConfirmUpgrade
            currentPlan={PLANS[currentPlan]}
            onClose={() => setSelectedPlan(null)}
            onConfirm={() => onSendRequest()}
            addOns={addOns}
            newPlan={selectedPlan}
          />
        )
      )}
    </div>
  );
};

export default Upgrade;
