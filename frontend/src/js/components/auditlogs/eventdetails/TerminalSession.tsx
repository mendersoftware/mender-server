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
import { useSelector } from 'react-redux';

import Loader from '@northern.tech/common-ui/Loader';
import { getCurrentSession } from '@northern.tech/store/selectors';

import TerminalPlayer from './TerminalPlayer';
import type { EventDetailsProps } from './utils';
import { SessionInfo, useSessionDetails } from './utils';

export const TerminalSession = ({ item, onClose }: EventDetailsProps) => {
  const { sessionDetails, isLoading, sessionMeta } = useSessionDetails(item);
  const { token } = useSelector(getCurrentSession);

  if (isLoading) {
    return <Loader show={true} />;
  }

  return (
    <div className="flexbox" style={{ flexWrap: 'wrap' }}>
      <TerminalPlayer className="flexbox column margin-top-small" item={item} sessionInitialized={!!sessionDetails} token={token} />
      <SessionInfo sessionMeta={sessionMeta} onClose={onClose} title="Session details" />
    </div>
  );
};

export default TerminalSession;
