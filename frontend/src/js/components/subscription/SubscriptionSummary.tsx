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
import { Button, Card, CardContent, CardHeader, Chip, Divider, Skeleton, Typography } from '@mui/material';

import { AvailableAddon, Plan } from '@northern.tech/store/appSlice/constants';

import { PreviewPrice } from './SubscriptionPage';
import { formatPrice } from './utils';

interface SubscriptionSummaryProps {
  addons: AvailableAddon[];
  deviceLimit: number;
  isNew: boolean;
  isPreviewLoading?: boolean;
  onAction?: () => void;
  plan: Plan;
  previewPrice: PreviewPrice;
  readOnly?: boolean;
  title: string;
}
const NumberSkeleton = () => <Skeleton width={35} height={26} />;

export const SubscriptionSummary = (props: SubscriptionSummaryProps) => {
  const { plan, deviceLimit, addons: enabledAddons, title, isNew, isPreviewLoading, readOnly, onAction, previewPrice } = props;
  const outlinedProps = { variant: 'outlined' as const, className: 'padding' };
  return (
    <Card style={{ minWidth: '320px' }} {...(readOnly ? { elevation: 0 } : outlinedProps)}>
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
          <Typography variant="subtitle1">{isPreviewLoading ? <NumberSkeleton /> : formatPrice(previewPrice.plan)}</Typography>
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
                <Typography variant="subtitle1">
                  {isPreviewLoading || !previewPrice.addons[addon] ? <NumberSkeleton /> : formatPrice(previewPrice.addons[addon])}
                </Typography>
              </div>
            ))}
          </div>
        )}
        <Divider variant="middle" className="margin-none" />
        <div className="flexbox space-between margin-top-small">
          <Typography variant="subtitle1">Monthly price</Typography>
          <Typography variant="h5">{isPreviewLoading ? <NumberSkeleton /> : formatPrice(previewPrice.total)} </Typography>
        </div>
      </CardContent>
      {!readOnly && (
        <Button className="margin-top-small" disabled={!isNew} variant="contained" onClick={onAction} fullWidth>
          Upgrade now
        </Button>
      )}
    </Card>
  );
};
