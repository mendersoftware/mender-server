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
import { DetailInformation } from './DeviceDetails';
import { DeviceDetailsSection } from './utils';

export const FileTransfer = ({ item, onClose }) => {
  const {
    actor,
    meta: { path = [] }
  } = item;

  const sessionMeta = {
    Path: path.join(','),
    User: actor.email
  };

  return (
    <div className="flexbox column">
      <DeviceDetailsSection onClose={onClose} />
      <DetailInformation title="file transfer" details={sessionMeta} />
    </div>
  );
};

export default FileTransfer;
