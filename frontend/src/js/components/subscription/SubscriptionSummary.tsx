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

import { PreviewPrice, deviceTypes } from './SubscriptionPage';
import { formatPrice } from './utils';

interface SubscriptionSummaryProps {
  addons: AvailableAddon[];
  isEnabled: boolean;
  isPreviewLoading?: boolean;
  onAction?: () => void;
  plan: Plan;
  previewPrice: PreviewPrice;
  readOnly?: boolean;
  title: string;
}
const NumberSkeleton = () => <Skeleton width={35} height={26} />;

export const SubscriptionSummary = (props: SubscriptionSummaryProps) => {
  const { plan, addons: enabledAddons, title, isEnabled, isPreviewLoading, readOnly, onAction, previewPrice } = props;
  const disabled = previewPrice.total === 0;
  const textColor = disabled ? 'text.disabled' : 'text.primary';
  const outlinedProps = { variant: 'outlined' as const, className: 'padding' };
  return (
    <Card style={{ minWidth: '320px' }} {...(readOnly ? { elevation: 0 } : outlinedProps)}>
      <CardHeader
        className="padding-none"
        title={
          <div className="flexbox space-between">
            <Typography variant="subtitle1">{title}</Typography>
            {isEnabled && <Chip variant="outlined" size="small" label="New" color="primary" />}
          </div>
        }
      />
      <CardContent className="padding-none">
        <Typography variant="body2" className="margin-top-small">
          Plan: {plan.name}
        </Typography>
        {Object.values(deviceTypes).map(({ id, summaryLabel }) => {
          const disabled = previewPrice[id].price === 0;
          const textColor = disabled ? 'text.disabled' : 'text.primary';
          return (
            <div key={id} className="flexbox space-between margin-top-small margin-bottom-small">
              <div>
                <Typography color={textColor} variant="body1">
                  {summaryLabel}
                </Typography>
                <Typography color={textColor} variant="caption">
                  x {previewPrice[id].quantity} devices
                </Typography>
              </div>
              {isPreviewLoading || !disabled ? (
                <Typography variant="subtitle1">{isPreviewLoading ? <NumberSkeleton /> : formatPrice(previewPrice[id].price)}</Typography>
              ) : (
                <Typography color={textColor} variant="subtitle2">
                  -
                </Typography>
              )}
            </div>
          );
        })}
        {enabledAddons.length > 0 && (
          <div className="margin-left-x-small margin-top-small margin-bottom-small">
            <Typography variant="body2" className="margin-bottom-x-small">
              Add-ons
            </Typography>
            {enabledAddons.map(addon => (
              <div key={addon} className="flexbox space-between margin-bottom-x-small">
                <div>
                  <Typography textTransform="capitalize" variant="body1">
                    {addon}
                  </Typography>
                  <Typography variant="body2">x {previewPrice['standard'].quantity} devices</Typography>
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
          <Typography color={textColor} variant="subtitle1">
            Monthly price
          </Typography>
          {isPreviewLoading || !disabled ? (
            <Typography variant="h5">{isPreviewLoading ? <NumberSkeleton /> : formatPrice(previewPrice.total)} </Typography>
          ) : (
            <Typography variant="subtitle2">-</Typography>
          )}
        </div>
      </CardContent>
      {!readOnly && (
        <Button className="margin-top-small" disabled={!isEnabled || disabled} variant="contained" onClick={onAction} fullWidth>
          Upgrade now
        </Button>
      )}
    </Card>
  );
};
