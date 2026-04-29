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
import { locations } from '@northern.tech/store/constants';

import FlagCN from '../../../assets/img/flag-cn.svg';
import FlagEU from '../../../assets/img/flag-eu.svg';
import FlagUS from '../../../assets/img/flag-us.svg';

export const locationMap = {
  cn: { ...locations.cn, icon: FlagCN, fallback: locations.us },
  eu: { ...locations.eu, icon: FlagEU, fallback: locations.us },
  us: { ...locations.us, icon: FlagUS, fallback: locations.eu }
};

export const getCurrentLocation = (location: Location): string => {
  const currentLocation = Object.values(locations).find(value => [`staging.${value.location}`, value.location].includes(location.hostname));
  return currentLocation ? currentLocation.key : locations.us.key;
};
