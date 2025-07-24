// Copyright 2025 Northern.tech AS
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
import { useEffect, useRef, useState } from 'react';
import { useSelector } from 'react-redux';

import { ErrorOutline as ErrorOutlineIcon } from '@mui/icons-material';
import { Alert, Button, CircularProgress, Divider, Drawer, Typography, buttonClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { DrawerTitle } from '@northern.tech/common-ui/DrawerTitle';
import Loader from '@northern.tech/common-ui/Loader';
import { SupportLink } from '@northern.tech/common-ui/SupportLink';
import Form from '@northern.tech/common-ui/forms/Form';
import storeActions from '@northern.tech/store/actions';
import { Address } from '@northern.tech/store/api/types/Address';
import { AvailableAddon, Plan } from '@northern.tech/store/constants';
import { Organization } from '@northern.tech/store/organizationSlice/types';
import { getBillingProfile, getCurrentUser, getSubscription } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import {
  confirmCardUpdate,
  createBillingProfile,
  editBillingProfile,
  getBillingPreview,
  requestPlanUpgrade,
  startCardUpdate
} from '@northern.tech/store/thunks';

import CardSection from '../settings/CardSection';
import { PlanExpandedForm } from '../settings/PlanExpandedForm';
import { BillingDetails } from '../settings/organization/BillingDetails';
import OrganizationPaymentSettings from '../settings/organization/OrganizationPaymentSettings';
import { PreviewPrice } from './SubscriptionPage';
import { SubscriptionSummary } from './SubscriptionSummary';
import { currencyFormatter } from './utils';

const { setSnackbar } = storeActions;

interface PlanProps {
  addons: Record<AvailableAddon, boolean>;
  isTrial?: boolean;
  onCloseClick: () => void;
  order?: any;
  organization: Organization;
  plan: Plan;
  previewPrice?: PreviewPrice;
}

const useStyles = makeStyles()(theme => ({
  formWrapper: {
    display: 'flex',
    flexDirection: 'column',
    gap: theme.spacing(2),
    maxWidth: 600,
    '.required .relative': { marginLeft: theme.spacing(10) }
  },
  buttonWrapper: {
    '&.button-wrapper': { justifyContent: 'initial' },
    [`.${buttonClasses.root}`]: { lineHeight: 'initial' }
  }
}));

const emptyAddress: Address = { city: '', country: '', line1: '', postal_code: '', state: '' };

const successMessage = (plan: string) =>
  `Thank you! You have successfully subscribed to the ${plan} plan.  You can view and edit your billing details on the Organization and billing page.`;

export const SubscriptionDrawer = (props: PlanProps) => {
  const { onCloseClick, previewPrice, order, isTrial, plan: selectedPlan, organization } = props;
  const { email } = useSelector(getCurrentUser);
  const billing = useSelector(getBillingProfile);
  const currentSubscription = useSelector(getSubscription);
  const initialValues = { email, name: organization?.name || '', line1: '', state: '', city: '', postal_code: '', country: '' };
  const [formInitialValues, setFormInitialValues] = useState(initialValues);
  const [isValid, setIsValid] = useState(false);
  const [isEdit, setIsEdit] = useState<boolean>(false);
  const [nextPayment, setNextPayment] = useState(0);
  const [updatingCard, setUpdatingCard] = useState(false);
  const dispatch = useAppDispatch();
  const formSubmitRef = useRef<() => void | null>(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const { classes } = useStyles();

  const handleUpgrade = () => {
    if (formSubmitRef.current) {
      formSubmitRef.current();
    }
  };
  const onInitEditProfile = () => {
    setIsEdit(true);
    setFormInitialValues({ ...(billing.address || emptyAddress), name: billing.name, email: billing.email });
  };
  useEffect(() => {
    if (!isTrial) {
      dispatch(getBillingPreview({ ...order, preview_mode: 'next' }))
        .unwrap()
        .then(next => {
          setNextPayment(next.total);
        });
    }
  }, [dispatch]);
  const handleBillingProfileEdit = async values => {
    const { email, name, state, city, line1, postal_code } = values;
    const code: string = values.country.code ? values.country.code : values.country;
    const billing_profile = { email, name, address: { country: code, state, city, line1, postal_code } };
    if (isEdit) {
      await dispatch(editBillingProfile({ billingProfile: billing_profile }));
    } else if (isTrial) {
      await dispatch(createBillingProfile({ billingProfile: billing_profile }));
    }
    setIsEdit(false);
  };
  const upgradePlanSubmit = async () => {
    try {
      setLoading(true);
      setError('');
      await dispatch(requestPlanUpgrade(order));
      dispatch(setSnackbar(successMessage((selectedPlan as Plan).name)));
      onCloseClick();
    } catch (e) {
      console.error(e);
      setError('There was an issue while processing your order. Please try again, or contact.');
    } finally {
      setLoading(false);
    }
  };
  const onCardConfirmed = async () => {
    await dispatch(confirmCardUpdate()).unwrap();
    await upgradePlanSubmit();
  };

  const summary = previewPrice && order && (
    <div style={{ maxWidth: '250px' }} className="margin-top-large">
      <SubscriptionSummary
        previewPrice={previewPrice}
        plan={props.plan}
        title="Your new subscription"
        isNew={false}
        addons={props.addons || {}}
        deviceLimit={order.products[0].quantity}
        readOnly
      />
    </div>
  );

  return (
    <Drawer anchor="right" open={true} PaperProps={{ style: { minWidth: '50vw' } }}>
      <DrawerTitle title={currentSubscription ? `Upgrade your subscription` : `Subscribe to Mender ${selectedPlan.name}`} onClose={onCloseClick} />
      <Divider className="margin-bottom" />
      {selectedPlan && (
        <div className="margin-bottom">
          Complete checkout to subscribe to <b>{selectedPlan.name}</b> at <b> {previewPrice ? currencyFormatter.format(previewPrice.total / 100) : ''}</b>
        </div>
      )}

      {isEdit || (isTrial && !billing) ? (
        <Form
          classes={classes}
          submitRef={formSubmitRef}
          onSubmit={handleBillingProfileEdit}
          handleCancel={!isTrial && (() => setIsEdit(false))}
          defaultValues={formInitialValues}
          submitLabel="Save Billing profile"
          showButtons={!updatingCard}
          autocomplete="off"
        >
          <PlanExpandedForm className={classes.formWrapper} setIsValid={setIsValid} />
        </Form>
      ) : (
        <>
          <Typography className="margin-top-small" variant="subtitle1">
            Billing details
          </Typography>
          <BillingDetails setChangeBilling={onInitEditProfile} hideCard={isTrial} />
        </>
      )}
      {isEdit ? (
        <OrganizationPaymentSettings
          className={classes.formWrapper}
          onComplete={handleUpgrade}
          updatingCard={updatingCard}
          setUpdatingCard={setUpdatingCard}
          isValid={isValid}
        />
      ) : (
        billing &&
        isTrial && (
          <div className={classes.formWrapper}>
            <h4 className="margin-top margin-bottom-none">Card Details</h4>
            <CardSection
              organization={organization}
              onCardConfirmed={onCardConfirmed}
              onSubmit={() => dispatch(startCardUpdate()).unwrap()}
              summary={summary}
              isSignUp
            />
          </div>
        )
      )}
      {(!isTrial || (isTrial && !billing)) && summary}
      {nextPayment > 0 && currentSubscription ? (
        <>
          <p>
            You’re currently subscribed to Mender Basic at {currencyFormatter.format(currentSubscription.total / 100)}/month. <br />
            On your next payment, you&#39;ll be charged for any days used under your current <br /> rate, and the rest will be billed at your new subscription
            rate. The total amount <br /> for your next payment will be {currencyFormatter.format(nextPayment / 100)}.
          </p>
          {error && (
            <Alert icon={<ErrorOutlineIcon />} severity="error">
              There was an error while creating the tenant. Please try again, or contact <SupportLink variant="email" />
            </Alert>
          )}
          <div className="margin-top flexbox">
            <Button className="margin-right-small" onClick={onCloseClick}>
              Cancel
            </Button>
            <Button
              className="margin-right-small"
              onClick={() => upgradePlanSubmit()}
              color="secondary"
              variant="contained"
              disabled={(isEdit && !isValid) || loading}
            >
              Confirm Subscription
            </Button>
            {loading && <CircularProgress />}
          </div>
        </>
      ) : (
        currentSubscription && <Loader show />
      )}
    </Drawer>
  );
};
