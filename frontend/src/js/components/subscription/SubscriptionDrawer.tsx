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
import { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';

import { CheckCircleOutlined as CheckCircleOutlinedIcon, ErrorOutline as ErrorOutlineIcon } from '@mui/icons-material';
import { Alert, Button, CircularProgress, Divider, Drawer, Typography, buttonClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { DrawerTitle } from '@northern.tech/common-ui/DrawerTitle';
import Loader from '@northern.tech/common-ui/Loader';
import { SupportLink } from '@northern.tech/common-ui/SupportLink';
import Form from '@northern.tech/common-ui/forms/Form';
import { AvailableAddon, PLANS, Plan } from '@northern.tech/store/constants';
import { Organization } from '@northern.tech/store/organizationSlice/types';
import { getBillingProfile, getCard, getCurrentUser, getSubscription } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import {
  confirmCardUpdate,
  createBillingProfile,
  editBillingProfile,
  getBillingPreview,
  requestPlanUpgrade,
  startCardUpdate
} from '@northern.tech/store/thunks';
import { Address } from '@northern.tech/types/MenderTypes';
import { isEmpty } from '@northern.tech/utils/helpers';

import CardSection from '../settings/CardSection';
import { PlanExpandedForm } from '../settings/PlanExpandedForm';
import { CardDetails } from '../settings/organization/Billing';
import { BillingDetails } from '../settings/organization/BillingDetails';
import OrganizationPaymentSettings from '../settings/organization/OrganizationPaymentSettings';
import { SubscriptionConfirmation } from './SubscriptionConfirmation';
import { PreviewPrice } from './SubscriptionPage';
import { SubscriptionSummary } from './SubscriptionSummary';
import { formatPrice } from './utils';

interface SubscriptionDrawerProps {
  addons: AvailableAddon[];
  currentPlanId?: string;
  isTrial?: boolean;
  onClose: () => void;
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

export const SubscriptionDrawer = (props: SubscriptionDrawerProps) => {
  const { onClose, previewPrice, order, isTrial, plan: selectedPlan, organization, currentPlanId } = props;
  const { email } = useSelector(getCurrentUser);
  const card = useSelector(getCard);
  const billing = useSelector(getBillingProfile);
  const currentSubscription = useSelector(getSubscription);
  const initialValues = { email, name: organization?.name || '', line1: '', state: '', city: '', postal_code: '', country: '' };
  const [formInitialValues, setFormInitialValues] = useState(initialValues);
  const [isValid, setIsValid] = useState(false);
  const [isEdit, setIsEdit] = useState<boolean>(false);
  const [successConfirmationShown, setSuccessConfirmationShown] = useState(false);

  const [nextPayment, setNextPayment] = useState(0);
  const [updatingCard, setUpdatingCard] = useState(false);
  const dispatch = useAppDispatch();
  const [error, setError] = useState(false);
  const [loading, setLoading] = useState(false);
  const [billingSaved, setBillingSaved] = useState(false);

  const orderedAddons = order?.products.filter(product => product.addons?.length).reduce((acc, curr) => [...acc, ...curr.addons], []);
  const orderedProducts = order?.products.map(product => ({ id: product.name.slice('mender_'.length), quantity: product.quantity }));
  const canShowConfirmation = successConfirmationShown && previewPrice && order;
  const { classes } = useStyles();

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
  }, [dispatch, isTrial, order]);
  const handleBillingProfileEdit = async values => {
    const { email, name, state, city, line1, postal_code } = values;
    const code: string = values.country.code ? values.country.code : values.country;
    const billing_profile = { email, name, address: { country: code, state, city, line1, postal_code } };
    if (isEdit) {
      await dispatch(editBillingProfile({ billingProfile: billing_profile }));
    } else if (isTrial) {
      await dispatch(createBillingProfile({ billingProfile: billing_profile }))
        .unwrap()
        .then(() => setBillingSaved(true));
    }
    setIsEdit(false);
  };
  const upgradePlanSubmit = async () => {
    try {
      setLoading(true);
      setError(false);
      await dispatch(requestPlanUpgrade(order));
      setSuccessConfirmationShown(true);
    } catch (e) {
      console.error(e);
      setError(true);
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
      <SubscriptionSummary previewPrice={previewPrice} plan={props.plan} title="Your new subscription" isEnabled={false} addons={props.addons} readOnly />
    </div>
  );
  const cardDetailsDisabled = isTrial && isEmpty(billing);
  return (
    <Drawer anchor="right" open={true} PaperProps={{ style: { minWidth: '50vw' } }}>
      <DrawerTitle title={currentSubscription ? `Upgrade your subscription` : `Subscribe to Mender ${selectedPlan.name}`} onClose={onClose} />
      <Divider className="margin-bottom-small" />
      {selectedPlan && (
        <div className="margin-bottom-large">
          Complete checkout to subscribe to Mender <b>{selectedPlan.name}</b> at <b> {previewPrice ? formatPrice(previewPrice.total) : ''}</b>
        </div>
      )}

      {isEdit || (isTrial && isEmpty(billing)) ? (
        <Form
          classes={classes}
          onSubmit={handleBillingProfileEdit}
          handleCancel={!isEmpty(billing) && (() => setIsEdit(false))}
          defaultValues={formInitialValues}
          submitLabel="Save Billing details"
          showButtons
          autocomplete="off"
          validationMode="onSubmit"
        >
          <PlanExpandedForm className={classes.formWrapper} setIsValid={setIsValid} />
        </Form>
      ) : (
        <>
          <Typography className="margin-top-small margin-bottom-x-small" variant="subtitle1">
            Your billing details
          </Typography>
          <BillingDetails setChangeBilling={onInitEditProfile} hideCard editDisabled={updatingCard} />

          {billingSaved && (
            <Alert className="margin-top-large" icon={<CheckCircleOutlinedIcon />}>
              Billing details saved
            </Alert>
          )}
        </>
      )}
      {isTrial ? (
        <div style={{ maxWidth: 600 }}>
          <Typography variant="subtitle1" color={cardDetailsDisabled ? 'textDisabled' : 'textPrimary'} className="margin-top-large margin-bottom-none">
            Card details
          </Typography>
          <CardSection
            organization={organization}
            onCardConfirmed={onCardConfirmed}
            onSubmit={() => dispatch(startCardUpdate()).unwrap()}
            summary={summary}
            isSignUp
            isValid={!isEdit}
            disabled={cardDetailsDisabled}
          />
        </div>
      ) : (
        <div>
          <Typography className="margin-top-large margin-bottom-x-small" variant="subtitle1">
            Card details
          </Typography>
          {updatingCard ? (
            <OrganizationPaymentSettings
              className={classes.formWrapper}
              updatingCard={updatingCard}
              setUpdatingCard={setUpdatingCard}
              isValid={isValid}
              omitHeader
            />
          ) : (
            <>
              <CardDetails card={card} />
              <Button disabled={isEdit} variant="outlined" className="margin-top-x-small" onClick={() => setUpdatingCard(true)} size="medium">
                Change card
              </Button>
            </>
          )}
        </div>
      )}
      {!isTrial && summary}
      {nextPayment > 0 && currentSubscription && currentPlanId ? (
        <div className={classes.formWrapper}>
          <Typography variant="body2">
            You’re currently subscribed to {PLANS[currentPlanId].name} at {formatPrice(currentSubscription.total)}/month. On your next payment, you&#39;ll be
            charged for any days used under your current rate, and the rest will be billed at your new subscription rate. The total amount for your next payment
            will be {formatPrice(nextPayment)}.
          </Typography>
          {error && (
            <Alert icon={<ErrorOutlineIcon />} severity="error">
              There was an issue while processing your order. Please try again, or contact <SupportLink variant="email" />.
            </Alert>
          )}
          <div className="margin-top flexbox">
            <Button className="margin-right-small" onClick={onClose}>
              Cancel
            </Button>
            <Button
              className="margin-right-small"
              onClick={() => upgradePlanSubmit()}
              color="secondary"
              variant="contained"
              disabled={isEdit || loading || updatingCard}
            >
              Confirm Subscription
            </Button>
            {loading && <CircularProgress />}
          </div>
        </div>
      ) : (
        currentSubscription && <Loader show />
      )}
      {canShowConfirmation && (
        <SubscriptionConfirmation products={orderedProducts} plan={selectedPlan} price={previewPrice?.total} orderedAddons={orderedAddons} />
      )}
    </Drawer>
  );
};
