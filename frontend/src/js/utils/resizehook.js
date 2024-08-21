// Copyright 2021 Northern.tech AS
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
import { useLayoutEffect, useRef, useState } from 'react';

import { TIMEOUTS } from '@northern.tech/store/constants';

export const useWindowSize = () => {
  const [size, setSize] = useState({ height: window.innerHeight, width: window.innerWidth });
  const timer = useRef();
  useLayoutEffect(() => {
    const handleResize = () => {
      timer.current = setTimeout(() => setSize({ height: window.innerHeight, width: window.innerWidth }), TIMEOUTS.halfASecond);
    };
    window.addEventListener('resize', handleResize);
    handleResize();
    return () => {
      clearTimeout(timer.current);
      window.removeEventListener('resize', handleResize);
    };
  }, []);

  return size;
};

export default useWindowSize;
