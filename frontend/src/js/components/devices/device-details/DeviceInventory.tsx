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
import { InventoryTable } from '@northern.tech/common-ui/InventoryTable';
import Time from '@northern.tech/common-ui/Time';
import { extractSoftware } from '@northern.tech/utils/helpers';

import DeviceDataCollapse from './DeviceDataCollapse';
import DeviceInventoryLoader from './DeviceInventoryLoader';

export const Title = ({ updateTime }) => (
  <div className="flexbox" style={{ alignItems: 'baseline' }}>
    <h4 className="margin-right">Device Inventory</h4>
    <div className="muted slightly-smaller" style={{ marginTop: 2 }}>
      Last changed: <Time value={updateTime} />
    </div>
  </div>
);

export const DeviceInventory = ({ device, setSnackbar }) => {
  const { attributes = {}, updated_ts: updateTime } = device;
  const { nonSoftware } = extractSoftware(attributes);
  const deviceInventory = nonSoftware.reduce((accu, attribute) => {
    const attributeValue = Array.isArray(attribute[1]) ? attribute[1].join(',') : attribute[1];
    accu[attribute[0]] = attributeValue;
    return accu;
  }, {});

  const waiting = !Object.values(attributes).some(i => i);
  return (
    <DeviceDataCollapse header={null} title={<Title updateTime={updateTime} />}>
      {waiting ? <DeviceInventoryLoader /> : <InventoryTable config={deviceInventory} setSnackbar={setSnackbar} />}
    </DeviceDataCollapse>
  );
};

export default DeviceInventory;
