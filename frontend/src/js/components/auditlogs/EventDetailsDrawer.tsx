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
import BaseDrawer from '@northern.tech/common-ui/BaseDrawer';

export const EventDetailsDrawer = ({ eventItem = {}, onClose, open, mapChangeToContent, fallbackComponent }) => {
  const { title, content: Component } = mapChangeToContent(eventItem, fallbackComponent);
  return (
    <BaseDrawer open={open} onClose={onClose} slotProps={{ header: { title } }}>
      <Component item={eventItem} onClose={onClose} />
    </BaseDrawer>
  );
};

export default EventDetailsDrawer;
