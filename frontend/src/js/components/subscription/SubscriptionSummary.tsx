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
import { Button, Card, CardContent, CardHeader, Chip, Divider, Typography } from '@mui/material';

import { AvailableAddon, Plan } from '@northern.tech/store/appSlice/constants';

interface SubscriptionSummaryProps {
  addons: Record<AvailableAddon, boolean>;
  deviceLimit: number;
  isNew: boolean;
  plan: Plan;
  title: string;
}
const amountFormatter = new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 0 });

export const SubscriptionSummary = (props: SubscriptionSummaryProps) => {
  const { plan, deviceLimit, addons, title, isNew } = props;
  const enabledAddons = Object.entries(addons)
    .filter(([addon, enabled]) => enabled && addon)
    .map(([addon]) => addon);
  const devicePrice = deviceLimit * plan.name.length;
  const addonPrice = deviceLimit * 10;
  return (
    <Card variant="outlined" className="padding margin-left-x-large" sx={{ minWidth: '320px' }}>
      <CardHeader
        className="padding-none"
        title={
          <div className="flexbox space-between">
            <Typography variant="subtitle1">{title}</Typography>
            {isNew && <Chip variant="outlined" size="small" label="New" color="primary" />}
          </div>
        }
      />
      <CardContent className="padding-none">
        <div className="flexbox space-between margin-top-small margin-bottom-small">
          <div>
            <Typography variant="body2">Plan: {plan.name}</Typography>
            <Typography variant="body1">Devices: x {deviceLimit}</Typography>
          </div>
          <Typography variant="subtitle1">{amountFormatter.format(devicePrice)}</Typography>
        </div>
        {enabledAddons.length > 0 && (
          <div className="margin-top-small margin-bottom-small">
            <Typography variant="body2">Add-ons</Typography>
            {enabledAddons.map(addon => (
              <div key={addon} className="flexbox space-between">
                <div>
                  <Typography textTransform="capitalize" variant="body1">
                    {addon}
                  </Typography>
                  <Typography variant="body2">x {deviceLimit} devices</Typography>
                </div>
                <Typography variant="subtitle1">{amountFormatter.format(addonPrice)}</Typography>
              </div>
            ))}
          </div>
        )}
        <Divider variant="middle" className="margin-none" />
        <div className="flexbox space-between margin-top-small margin-bottom-small">
          <Typography variant="subtitle1">Monthly price</Typography>
          <Typography variant="subtitle1">{amountFormatter.format(addonPrice * enabledAddons.length + devicePrice)} </Typography>
        </div>
      </CardContent>
      <Button sx={{ textTransform: 'none' }} disabled={!isNew} variant="contained" fullWidth>
        Upgrade now
      </Button>
    </Card>
  );
};
