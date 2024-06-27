// Copyright 2024 Northern.tech AS
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
import React, { useEffect, useRef, useState } from 'react';

// material ui
import { Sort } from '@mui/icons-material';

import { TIMEOUTS } from '../../constants/appConstants';

const SortIcon = ({ columnKey, disabled = false, sortDown = false }) => {
  const timer = useRef();
  const [fadeIcon, setFadeIcon] = useState(true);

  useEffect(() => {
    if (disabled) {
      timer.current = setTimeout(() => setFadeIcon(true), TIMEOUTS.oneSecond);
    } else {
      timer.current = setTimeout(() => setFadeIcon(false), TIMEOUTS.debounceShort);
    }
    return () => {
      clearTimeout(timer.current);
    };
  }, [disabled]);

  return <Sort className={`sortIcon ${columnKey && !disabled ? 'selected' : ''} ${fadeIcon ? 'fadeOut' : ''} ${sortDown}`} style={{ fontSize: 16 }} />;
};

export default SortIcon;
