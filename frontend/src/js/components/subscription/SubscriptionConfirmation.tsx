import { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';

import { TaskAlt as TaskAltIcon } from '@mui/icons-material';
import { Alert, AlertTitle, Button, DialogContent, Divider, IconButton, Typography } from '@mui/material';

import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import { cleanUp } from '@northern.tech/store/auth';
import { ADDONS, AvailableAddon, Plan } from '@northern.tech/store/constants';
import { getOrganization } from '@northern.tech/store/organizationSlice/selectors';

import { formatPrice } from './utils';

interface SubscriptionConfirmationProps {
  devices: number;
  onClose: () => void;
  orderedAddons: { name: AvailableAddon }[];
  plan: Plan;
  price: number;
}
export const SubscriptionConfirmation = (props: SubscriptionConfirmationProps) => {
  const { plan, devices, price, orderedAddons, onClose } = props;
  const { addons: enabledAddons, plan: currentPlan } = useSelector(getOrganization);

  const addonList = orderedAddons.map(addon => addon.name);

  const previousAddonsList = enabledAddons.filter(addon => addon.enabled);
  const willLogout = addonList.length > previousAddonsList.length || currentPlan !== plan.id;
  const [count, setCount] = useState<number>(60);
  const logOut = () => {
    cleanUp();
    window.location.replace('/ui/');
  };

  useEffect(() => {
    if (!willLogout) {
      return;
    }
    if (count === 0) {
      return logOut();
    }
    const timer = setInterval(() => {
      setCount(prevCount => prevCount - 1);
    }, 1000);

    return () => clearInterval(timer);
  }, [count, willLogout]);

  return (
    <BaseDialog
      open
      title={
        <div className="flexbox center-aligned">
          <IconButton size="large">
            <TaskAltIcon selected className="green" />
          </IconButton>
          <Typography variant="h6">Payment successful!</Typography>
        </div>
      }
    >
      <Divider />
      <DialogContent style={{ maxWidth: '440px' }}>
        <Typography variant="body1">
          {willLogout
            ? `Your subscription has been successfully updated${currentPlan !== plan.id ? ` to Mender ${plan.name}.` : '.'}`
            : 'Your device limit has been successfully updated.'}
        </Typography>
        <div className="margin-top-small margin-bottom-small">
          <Typography> Subscription details: </Typography>
          <Typography> Plan: {plan.name} </Typography>
          <Typography> Devices: {devices} </Typography>
          {addonList.length > 0 && <Typography>Add-ons: {addonList.map(addon => ADDONS[addon].title).join(', ')}</Typography>}
          <Typography>Monthly cost: {formatPrice(price)}</Typography>
        </div>
        {willLogout ? (
          <Alert severity="info" icon={false}>
            <AlertTitle textAlign="center">Automatic logout in {count} seconds</AlertTitle>
            <div className="flexbox column centered">
              <Typography className="margin-bottom-x-small" textAlign="center">
                You will be logged out automatically, for your new subscription to take effect.
              </Typography>
              <Button variant="contained" onClick={logOut}>
                Log out now
              </Button>
            </div>
          </Alert>
        ) : (
          <Alert
            severity="success"
            icon={false}
            sx={{
              '& .MuiAlert-message': {
                width: '100%'
              }
            }}
            className="flexbox space-between"
          >
            <AlertTitle textAlign="center">Subscription Active</AlertTitle>
            <div className="flexbox column centered">
              <Typography className="margin-bottom-x-small" textAlign="center">
                Your updated subscription is ready to use.{' '}
              </Typography>
              <Button variant="outlined" onClick={onClose}>
                Close
              </Button>
            </div>
          </Alert>
        )}
      </DialogContent>
    </BaseDialog>
  );
};
