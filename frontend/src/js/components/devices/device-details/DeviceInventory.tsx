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
import { Typography } from '@mui/material';

import { ContentSection } from '@northern.tech/common-ui/ContentSection';
import { InventoryTable } from '@northern.tech/common-ui/InventoryTable';
import Time from '@northern.tech/common-ui/Time';
import { extractSoftware } from '@northern.tech/utils/helpers';

import DeviceInventoryLoader from './DeviceInventoryLoader';

export const LastChangedNote = ({ updateTime }) => (
  <Typography variant="body2">
    Last changed: <Time value={updateTime} />
  </Typography>
);

export const DeviceInventory = ({ device, setSnackbar }) => {
  const { attributes = {}, updated_ts: updateTime } = device;
  const { nonSoftware } = extractSoftware(attributes);
  const deviceInventory = nonSoftware.reduce((accu, [key, value]) => {
    const attributeValue = Array.isArray(value) ? value.join(',') : value;
    accu[key] = attributeValue;
    return accu;
  }, {});

  const waiting = !Object.values(attributes).some(i => i);
  return (
    <ContentSection postTitle={<LastChangedNote updateTime={updateTime} />} title="Device Inventory">
      {waiting ? <DeviceInventoryLoader /> : <InventoryTable config={deviceInventory} setSnackbar={setSnackbar} />}
    </ContentSection>
  );
};

export default DeviceInventory;
