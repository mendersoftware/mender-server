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

import { defaultState } from '../../../tests/mockData';
import { render } from '../../../tests/setupTests';
import { HELPTOOLTIPS } from './helptooltips';

const preloadedState = {
  ...defaultState,
  app: {
    ...defaultState.app,
    features: {
      ...defaultState.app.features,
      hasMultitenancy: true,
      isHosted: true
    }
  },
  devices: {
    ...defaultState.devices,
    byId: {
      ...defaultState.devices.byId,
      a1: {
        ...defaultState.devices.byId.a1,
        attributes: {
          ...defaultState.devices.byId.a1.attributes,
          device_type: ['raspberrypi12']
        }
      }
    }
  }
};

describe('Helptooltips Components', () => {
  Object.values(HELPTOOLTIPS).forEach(async ({ SpecialComponent, Component = SpecialComponent }) => {
    it(`renders ${Component.displayName || Component.name} correctly`, () => {
      const { baseElement } = render(<Component deviceId={defaultState.devices.byId.a1.id} />, { preloadedState });
      const view = baseElement.firstChild.childNodes.length > 1 ? baseElement.firstChild.childNodes : baseElement.firstChild.firstChild;
      expect(view).toMatchSnapshot();
    });
  });
});
