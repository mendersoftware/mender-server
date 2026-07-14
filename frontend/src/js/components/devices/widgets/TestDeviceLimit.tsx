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
import { LinearProgress, Typography } from '@mui/material';

import Link from '@northern.tech/common-ui/Link';

export const MAX_TEST_DEVICES = 10;

export const TestDeviceLimit = (props: { className?: string; onNavigate?: () => void; testDeviceUsed: number }) => {
  const { testDeviceUsed, className = '', onNavigate } = props;
  return (
    <div className={className}>
      <div className="flexbox space-between margin-bottom-x-small">
        <Typography>
          {testDeviceUsed}/{MAX_TEST_DEVICES} test devices set
        </Typography>
        <Link to="/devices/accepted?system=test_device:eq:true" onClick={onNavigate}>
          View all
        </Link>
      </div>
      <LinearProgress variant="determinate" value={(testDeviceUsed / MAX_TEST_DEVICES) * 100} />
    </div>
  );
};
