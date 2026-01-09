// Copyright 2020 Northern.tech AS
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
import { Link } from 'react-router-dom';

import { Chip } from '@mui/material';

import { MenderTooltipClickable } from '@northern.tech/common-ui/helptips/MenderTooltip';
import dayjs from 'dayjs';
import durationDayJs from 'dayjs/plugin/duration';
import pluralize from 'pluralize';

dayjs.extend(durationDayJs);

const TrialInformation = () => (
  <>
    <h3>Trial plan</h3>
    <p>You&apos;re using the trial version of Mender – it&apos;s free for up to 10 devices for 12 months.</p>
    <p>
      <Link to="/subscription">Upgrade to a plan</Link> to add more devices and continue using Mender after the trial expires.
    </p>
    <p>
      Or compare the plans at {/* eslint-disable-next-line react/jsx-no-target-blank */}
      <a href={`https://mender.io/plans/pricing`} target="_blank" rel="noopener">
        mender.io/plans/pricing
      </a>
      .
    </p>
  </>
);

const TrialNotification = ({ sectionClassName, expiration }) => {
  const expirationDate = dayjs(expiration);
  const duration = dayjs.duration(expirationDate.diff(dayjs()));
  const daysLeft = Math.floor(duration.asDays());
  return (
    <div className={`flexbox centered ${sectionClassName}`}>
      <MenderTooltipClickable className="flexbox center-aligned margin-right-small" disableHoverListener={false} title={<TrialInformation />}>
        <Chip className="clickable" component={Link} label="Trial plan" size="small" to="/subscription" variant="outlined" />
      </MenderTooltipClickable>

      {expiration && daysLeft <= 100 && daysLeft >= 0 && (
        <div className="muted">
          You have {daysLeft} {pluralize('day', daysLeft)} remaining on the trial plan
        </div>
      )}
    </div>
  );
};

export default TrialNotification;
