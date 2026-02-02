// Copyright 2019 Northern.tech AS
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
// material ui
import { useSelector } from 'react-redux';

import { CheckCircle as CheckIcon } from '@mui/icons-material';
import { Skeleton, Tooltip, Typography } from '@mui/material';

import { getDeviceLimitStats } from '@northern.tech/store/selectors';
import pluralize from 'pluralize';

import { DeviceLimit } from '../../header/DeviceNotifications';
import { BaseWidget } from './BaseWidget';

const LoadingSkeleton = () => <Skeleton className="margin-bottom-x-small" variant="rectangular" height={47} />;
const numberLocale = 'en-US';
export const AcceptedDevices = props => {
  const { devicesCount, onClick } = props;
  const onWidgetClick = () => onClick({ route: '/devices/accepted' });

  const mappedLimits = useSelector(getDeviceLimitStats);

  const limits = mappedLimits.map(({ type, total, limit }) => (
    <Tooltip
      key={type}
      title={
        limit > 0
          ? `${Math.round(total / limit)}% used. To increase limits, go to Settings > Billing`
          : `You have ${total} accepted ${type} ${pluralize('device', total)}`
      }
      slotProps={{
        popper: {
          modifiers: [
            {
              name: 'offset',
              options: {
                offset: [0, -15]
              }
            }
          ]
        }
      }}
      arrow
    >
      <div className="margin-bottom-x-small">
        <DeviceLimit limit={limit} type={type} total={total} compact />
      </div>
    </Tooltip>
  ));

  const main = (
    <div className="full-width">
      <Typography variant="h5">{devicesCount.toLocaleString(numberLocale)}</Typography>
      <div className="flexbox column full-width margin-top-small">{limits.length > 0 ? limits : <LoadingSkeleton />}</div>
    </div>
  );

  return (
    <BaseWidget
      {...props}
      header={<div className="flexbox center-aligned">Accepted devices {!!devicesCount && <CheckIcon className="margin-left-small green" />}</div>}
      main={main}
      onClick={onWidgetClick}
    />
  );
};

export default AcceptedDevices;
