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
import { useSelector } from 'react-redux';

import { LinearProgress, Typography } from '@mui/material';

import Link from '@northern.tech/common-ui/Link';
import { getTestDeviceCount, getTestDeviceLimit } from '@northern.tech/store/selectors';

export const useTestDeviceLimit = () => {
  const used = useSelector(getTestDeviceCount);
  const limit = useSelector(getTestDeviceLimit);
  const unlimited = limit === -1;
  return { used, limit, unlimited, isAtLimit: !unlimited && limit > 0 && used >= limit };
};
export const TestDeviceLimit = (props: { className?: string; onNavigate?: () => void }) => {
  const { className = '', onNavigate } = props;
  const { used, limit, unlimited } = useTestDeviceLimit();
  return (
    <div className={className}>
      <div className="flexbox space-between margin-bottom-x-small">
        <Typography>
          {used}
          {!unlimited && `/${limit}`} test devices set
        </Typography>
        <Link to="/devices/accepted?system=test_device:eq:true" onClick={onNavigate}>
          View all
        </Link>
      </div>
      {!unlimited && <LinearProgress variant="determinate" value={limit ? Math.min(100, (used / limit) * 100) : 0} />}
    </div>
  );
};
