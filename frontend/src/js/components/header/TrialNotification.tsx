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
import { useNavigate } from 'react-router';

import { InfoOutlined as InfoOutlinedIcon } from '@mui/icons-material';
import { Chip, Tooltip } from '@mui/material';

import dayjs from 'dayjs';
import durationDayJs from 'dayjs/plugin/duration';
import pluralize from 'pluralize';

dayjs.extend(durationDayJs);

const TrialInformation = ({ expiresSoon, label }) =>
  expiresSoon ? (
    <>
      You’re using the trial version of Mender. Your trial ends in <b>{label}</b>. Upgrade now to keep managing your devices without interruption.
    </>
  ) : (
    <>You’re using the trial version of Mender. It’s free to try for 12 months.</>
  );
const TrialNotification = ({ sectionClassName, expiration }) => {
  const navigate = useNavigate();
  const expirationDate = dayjs(expiration);
  const duration = dayjs.duration(expirationDate.diff(dayjs()));
  const days = Math.floor(duration.asDays());
  const months = Math.floor(duration.asMonths());
  const expiresSoon = months < 4;

  const label = days > 30 ? `${months} ${pluralize('months', months)}` : `${days} ${pluralize('day', days)}`;
  return (
    <div className={`flexbox centered ${sectionClassName}`}>
      <Tooltip arrow title={<TrialInformation expiresSoon={expiresSoon} label={label} />}>
        <Chip
          color={expiresSoon ? 'error' : 'default'}
          icon={expiresSoon ? <InfoOutlinedIcon /> : undefined}
          label="Trial plan"
          size="small"
          onClick={() => navigate('/subscription')}
          variant="outlined"
        />
      </Tooltip>
    </div>
  );
};

export default TrialNotification;
