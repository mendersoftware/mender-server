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
import { useRef, useState } from 'react';
import { useSelector } from 'react-redux';

import { Divider, Drawer } from '@mui/material';

import { DrawerTitle } from '@northern.tech/common-ui/DrawerTitle';
import Form from '@northern.tech/common-ui/forms/form';
import storeActions from '@northern.tech/store/actions';
import { Plan } from '@northern.tech/store/constants';
import { Organization } from '@northern.tech/store/organizationSlice/types';
import { getCurrentUser } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { completeUpgrade } from '@northern.tech/store/thunks';

import { PlanExpandedForm } from './PlanExpandedForm';
import CardSection from './cardsection';

const { setSnackbar } = storeActions;

interface PlanExpandedProps {
  plan: Plan;
  organization: Organization;
  onCloseClick: () => void;
}

export const PlanExpanded = (props: PlanExpandedProps) => {
  const { plan: selectedPlan, onCloseClick, organization } = props;
  const [isValid, setIsValid] = useState(false);
  const dispatch = useAppDispatch();
  const formSubmitRef = useRef<() => void | null>(null);
  const { email } = useSelector(getCurrentUser);
  const handleUpgrade = () => {
    if (formSubmitRef.current) {
      formSubmitRef.current();
    }
  };

  const handleSubmit = async values => {
    const {
      email,
      name,
      country: { code },
      state,
      city,
      line1,
      postal_code
    } = values;
    const billing_profile = { email, name, address: { country: code, state, city, line1, postal_code } };
    await dispatch(completeUpgrade({ tenantId: organization.id, plan: selectedPlan.id, billing_profile }));
    onCloseClick();
  };
  return (
    <Drawer anchor="right" open={true} PaperProps={{ style: { minWidth: '75vw' } }}>
      <DrawerTitle title={<>Subscribe to {selectedPlan.name}</>} onClose={onCloseClick} />
      <Divider className="margin-bottom" />
      <div>
        Complete checkout to subscribe to <b>{selectedPlan.name}</b> at <b>{selectedPlan.price}</b>
      </div>
      <Form
        submitRef={formSubmitRef}
        onSubmit={handleSubmit}
        initialValues={{ email, name: organization.name || '', line1: '', state: '', city: '', postal_code: '', country: '' }}
        showButtons={false}
        autocomplete="off"
      >
        <PlanExpandedForm setIsValid={setIsValid} />
      </Form>
      {isValid && (
        <>
          <h4>Card Details</h4>
          <CardSection
            organization={organization}
            onComplete={handleUpgrade}
            setSnackbar={(message: string) => dispatch(setSnackbar(message))}
            isSignUp={true}
          />
        </>
      )}
    </Drawer>
  );
};
