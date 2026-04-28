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
import { render } from '@/testUtils';
import { undefineds } from '@northern.tech/testing/mockData';
import { vi } from 'vitest';

import TextOverflowMultiline from './TextOverflowMultiline';

describe('TextOverflowMultiline Component', () => {
  it('renders correctly with short text', async () => {
    const { baseElement } = render(<TextOverflowMultiline>Short text</TextOverflowMultiline>);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('calls onOverflowChange prop', () => {
    const scrollSpy = vi.spyOn(HTMLElement.prototype, 'scrollHeight', 'get').mockReturnValue(200);
    const clientSpy = vi.spyOn(HTMLElement.prototype, 'clientHeight', 'get').mockReturnValue(40);
    const onOverflowChange = vi.fn();
    const { rerender } = render(<TextOverflowMultiline onOverflowChange={onOverflowChange}>Long text</TextOverflowMultiline>);
    expect(onOverflowChange).toHaveBeenCalledWith(true);
    onOverflowChange.mockClear();

    scrollSpy.mockReturnValue(40);
    rerender(<TextOverflowMultiline onOverflowChange={onOverflowChange}>Short</TextOverflowMultiline>);
    expect(onOverflowChange).toHaveBeenCalledWith(false);
    scrollSpy.mockRestore();
    clientSpy.mockRestore();
  });

  it('renders with custom lines and component', () => {
    const { baseElement } = render(
      <TextOverflowMultiline lines={5} component="div">
        Content
      </TextOverflowMultiline>
    );
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view.tagName).toBe('DIV');
  });
});
