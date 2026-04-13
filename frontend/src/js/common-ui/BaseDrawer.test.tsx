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
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import BaseDrawer from './BaseDrawer';

describe('BaseDrawer Component', () => {
  it('renders correctly with string title', async () => {
    const { baseElement } = render(
      <BaseDrawer open onClose={vi.fn()} slotProps={{ header: { title: 'Test Drawer' } }}>
        <div>Drawer content</div>
      </BaseDrawer>
    );
    const view = baseElement;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('renders correctly with ReactNode title', async () => {
    const { baseElement } = render(
      <BaseDrawer open onClose={vi.fn()} slotProps={{ header: { title: <div data-testid="custom-title">Custom Title</div> } }}>
        <div>Content</div>
      </BaseDrawer>
    );
    const view = baseElement;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
    expect(screen.getByTestId('custom-title')).toBeInTheDocument();
  });

  it('renders notification slot', async () => {
    const { baseElement } = render(
      <BaseDrawer open onClose={vi.fn()} slotProps={{ header: { title: 'Test' } }} notification={<div data-testid="notification">Alert!</div>}>
        <div>Content</div>
      </BaseDrawer>
    );
    expect(screen.getByTestId('notification')).toBeInTheDocument();
    expect(baseElement).toMatchSnapshot();
  });

  it('calls onClose with reason when close button is clicked', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const onClose = vi.fn();
    render(
      <BaseDrawer open onClose={onClose} slotProps={{ header: { title: 'Test Drawer' } }}>
        <div>Content</div>
      </BaseDrawer>
    );
    await user.click(screen.getByRole('button', { name: /close/i }));
    expect(onClose).toHaveBeenCalledTimes(1);
    expect(onClose).toHaveBeenCalledWith(expect.anything(), 'escapeKeyDown');
  });

  it('renders header preCloser, postTitle and onLinkCopy through slotProps', async () => {
    const onLinkCopy = vi.fn();
    render(
      <BaseDrawer
        open
        onClose={vi.fn()}
        slotProps={{
          header: {
            title: 'Test',
            onLinkCopy,
            postTitle: <span data-testid="post-title">post</span>,
            preCloser: <span data-testid="pre-closer">pre</span>
          }
        }}
      >
        <div>Content</div>
      </BaseDrawer>
    );
    expect(screen.getByTestId('post-title')).toBeInTheDocument();
    expect(screen.getByTestId('pre-closer')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /copy link/i })).toBeInTheDocument();
  });
});
