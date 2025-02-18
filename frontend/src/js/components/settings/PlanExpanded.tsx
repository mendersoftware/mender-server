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

import { Button, Divider, Drawer } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { DrawerTitle } from '@northern.tech/common-ui/DrawerTitle';
import Form from '@northern.tech/common-ui/forms/Form';
import storeActions from '@northern.tech/store/actions';
import { Plan } from '@northern.tech/store/constants';
import { BillingProfile, Card, Organization } from '@northern.tech/store/organizationSlice/types';
import { getCurrentUser } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { completeUpgrade, editBillingProfile, startUpgrade } from '@northern.tech/store/thunks';

import CardSection from './CardSection';
import { PlanExpandedForm } from './PlanExpandedForm';
import OrganizationPaymentSettings from './organization/OrganizationPaymentSettings';

const { setSnackbar } = storeActions;

interface PlanExpandedPropsBase {
  isEdit: boolean;
  onCloseClick: () => void;
  organization: Organization;
}
interface ProfileEditProps extends PlanExpandedPropsBase {
  card: Card;
  currentBillingProfile: BillingProfile;
  isEdit: true;
}

interface PlanProps extends PlanExpandedPropsBase {
  isEdit: false;
  plan: Plan;
}

const useStyles = makeStyles()(theme => ({
  formWrapper: {
    display: 'flex',
    flexDirection: 'column',
    gap: theme.spacing(2),
    maxWidth: 600,
    '.required .relative': { marginLeft: theme.spacing(10) }
  }
}));

const successMessage = (plan: string) =>
  `Thank you! You have successfully subscribed to the ${plan} plan.  You can view and edit your billing details on the Organization and billing page.`;

export const PlanExpanded = (props: ProfileEditProps | PlanProps) => {
  const { onCloseClick, isEdit } = props;
  const organization = !isEdit ? props.organization : null;
  const [isValid, setIsValid] = useState(isEdit);
  const [updatingCard, setUpdatingCard] = useState(false);
  const selectedPlan = isEdit ? null : props.plan;
  const dispatch = useAppDispatch();
  const formSubmitRef = useRef<() => void | null>(null);
  const { email } = useSelector(getCurrentUser);
  const { classes } = useStyles();

  const handleUpgrade = () => {
    if (formSubmitRef.current) {
      formSubmitRef.current();
    }
  };
  const initialValues = isEdit
    ? { ...props.currentBillingProfile.address, name: props.currentBillingProfile.name, email: props.currentBillingProfile.email }
    : { email, name: organization?.name || '', line1: '', state: '', city: '', postal_code: '', country: '' };
  const handleSubmit = async values => {
    const { email, name, state, city, line1, postal_code } = values;
    const code: string = values.country.code ? values.country.code : values.country;
    const billing_profile = { email, name, address: { country: code, state, city, line1, postal_code } };
    if (isEdit) {
      await dispatch(editBillingProfile({ billingProfile: billing_profile }));
    } else {
      await dispatch(completeUpgrade({ tenantId: (organization as Organization).id, plan: (selectedPlan as Plan).id, billing_profile }));
      dispatch(setSnackbar(successMessage((selectedPlan as Plan).name)));
    }
    onCloseClick();
  };

  return (
    <Drawer anchor="right" open={true} PaperProps={{ style: { minWidth: '50vw' } }}>
      <DrawerTitle title={selectedPlan ? `Subscribe to Mender ${selectedPlan.name}` : 'Edit billing details'} onClose={onCloseClick} />
      <Divider className="margin-bottom" />
      {selectedPlan && (
        <div className="margin-bottom">
          Complete checkout to subscribe to <b>{selectedPlan.name}</b> at <b>{selectedPlan.price}</b>
        </div>
      )}
      <Form submitRef={formSubmitRef} onSubmit={handleSubmit} defaultValues={initialValues} showButtons={false} autocomplete="off">
        <PlanExpandedForm className={classes.formWrapper} setIsValid={setIsValid} />
      </Form>
      {isEdit ? (
        <OrganizationPaymentSettings
          className={classes.formWrapper}
          onComplete={handleUpgrade}
          updatingCard={updatingCard}
          setUpdatingCard={setUpdatingCard}
          isValid={isValid}
        />
      ) : (
        isValid &&
        organization && (
          <div className={classes.formWrapper}>
            <h4 className="margin-top margin-bottom-none">Card Details</h4>
            <CardSection
              organization={organization}
              onCardConfirmed={handleUpgrade}
              onSubmit={() => dispatch(startUpgrade(organization.id)).unwrap()}
              isSignUp
            />
          </div>
        )
      )}
      {isEdit && !updatingCard && (
        <div className="margin-top">
          <Button className="margin-right-small" onClick={onCloseClick}>
            Cancel
          </Button>
          <Button onClick={() => handleUpgrade()} color="secondary" variant="contained" disabled={!isValid}>
            Save
          </Button>
        </div>
      )}
    </Drawer>
  );
};
