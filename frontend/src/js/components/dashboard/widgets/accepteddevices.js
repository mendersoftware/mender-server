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
import React from 'react';

// material ui
import { CheckCircle as CheckIcon } from '@mui/icons-material';

import { BaseWidget } from './baseWidget';

export const AcceptedDevices = props => {
  const { devicesCount, onClick } = props;
  const onWidgetClick = () => onClick({ route: '/devices/accepted' });

  const widgetMain = {
    counter: devicesCount
  };

  return (
    <BaseWidget
      {...props}
      header={<div className="flexbox center-aligned">Accepted devices {!!devicesCount && <CheckIcon className="margin-left-small green" />}</div>}
      main={widgetMain}
      onClick={onWidgetClick}
    />
  );
};

export default AcceptedDevices;
