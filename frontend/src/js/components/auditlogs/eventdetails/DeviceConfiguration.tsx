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
import { ContentSection } from '@northern.tech/common-ui/ContentSection';
import { TwoColumnData } from '@northern.tech/common-ui/TwoColumnData';

import { DeviceDetailsSection, parseConfigChange } from './utils';

export const DeviceConfiguration = ({ item, onClose }) => {
  const { actor, change } = item;
  const config = parseConfigChange(change);

  return (
    <div className="flexbox column">
      <DeviceDetailsSection onClose={onClose} />
      <ContentSection title="Changed configuration details">
        <TwoColumnData data={config} />
      </ContentSection>
      <ContentSection title="Change details">
        <TwoColumnData data={{ User: actor.email }} />
      </ContentSection>
    </div>
  );
};

export default DeviceConfiguration;
