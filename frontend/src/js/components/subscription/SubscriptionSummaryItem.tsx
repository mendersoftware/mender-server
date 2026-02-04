// Copyright 2026 Northern.tech AS
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
import { Skeleton, Typography } from '@mui/material';

import { AvailableAddon } from '@northern.tech/store/appSlice/constants';

import type { PlanPreviewPriceItem } from './SubscriptionPage';
import { formatPrice } from './utils';

const NumberSkeleton = () => <Skeleton width={35} height={26} />;

type SubscriptionSummaryItemProps = {
  addons: AvailableAddon[];
  isPreviewLoading: boolean;
  previewPriceItem: PlanPreviewPriceItem;
  summaryLabel: string;
};
export const SubscriptionSummaryItem = (props: SubscriptionSummaryItemProps) => {
  const { addons, previewPriceItem, isPreviewLoading, summaryLabel } = props;
  const disabled = previewPriceItem?.price === 0;
  const textColor = disabled ? 'text.disabled' : 'text.primary';
  return (
    previewPriceItem && (
      <>
        <div className="flexbox space-between margin-top-small margin-bottom-small">
          <div>
            <Typography color={textColor} variant="body1" className="capitalized-start">
              {summaryLabel}
            </Typography>
            <Typography color={textColor} variant="caption">
              x {previewPriceItem.quantity} devices
            </Typography>
          </div>
          {isPreviewLoading || !disabled ? (
            <Typography variant="subtitle1">{isPreviewLoading ? <NumberSkeleton /> : formatPrice(previewPriceItem.price || 0)}</Typography>
          ) : (
            <Typography color={textColor} variant="subtitle2">
              -
            </Typography>
          )}
        </div>
        {previewPriceItem.addons && Object.keys(previewPriceItem.addons).length > 0 && (
          <div className="margin-left-x-small margin-top-small margin-bottom-small">
            <Typography variant="body2" className="margin-bottom-x-small">
              Add-ons
            </Typography>
            {addons.map(addon => (
              <div key={addon} className="flexbox space-between margin-bottom-x-small">
                <div>
                  <Typography textTransform="capitalize" variant="body1">
                    {addon}
                  </Typography>
                  <Typography variant="body2">x {previewPriceItem.quantity} devices</Typography>
                </div>
                <Typography variant="subtitle1">
                  {isPreviewLoading || !previewPriceItem.addons[addon] ? <NumberSkeleton /> : formatPrice(previewPriceItem.addons[addon])}
                </Typography>
              </div>
            ))}
          </div>
        )}
      </>
    )
  );
};
