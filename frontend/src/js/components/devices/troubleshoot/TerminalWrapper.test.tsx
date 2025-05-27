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
import { vi } from 'vitest';

import { defaultState, undefineds } from '../../../../../tests/mockData';
import { render } from '../../../../../tests/setupTests';
import { TroubleshootContent as TroubleshootDialog } from './TerminalWrapper';

describe('TroubleshootDialog Component', () => {
  let socketSpyFactory;

  beforeEach(() => {
    socketSpyFactory = vi.spyOn(window, 'WebSocket');
    socketSpyFactory.mockImplementation(() => ({
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      close: () => {},
      send: () => {}
    }));
  });

  afterEach(() => {
    socketSpyFactory.mockReset();
  });

  it('renders correctly', async () => {
    const { baseElement } = render(
      <TroubleshootDialog
        device={defaultState.devices.byId.a1}
        onDownload={vi.fn()}
        setSocketClosed={vi.fn()}
        setUploadPath={vi.fn()}
        setFile={vi.fn()}
        setSnackbar={vi.fn()}
        setSocketInitialized={vi.fn()}
        socketInitialized
      />
    );
    expect(baseElement).toMatchSnapshot();
    expect(baseElement).toEqual(expect.not.stringMatching(undefineds));
  });
});
