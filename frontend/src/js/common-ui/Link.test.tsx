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
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { Link } from './Link';

describe('Link Component', () => {
  it('renders an anchor with target="_blank" and rel when external', () => {
    render(
      <Link href="https://example.com" external>
        External
      </Link>
    );
    const link = screen.getByRole('link', { name: 'External' });
    expect(link).toHaveAttribute('href', 'https://example.com');
    expect(link).toHaveAttribute('target', '_blank');
    expect(link).toHaveAttribute('rel', 'noopener noreferrer');
  });

  it('does not add target/rel when external is omitted', () => {
    render(<Link href="https://example.com">Plain href</Link>);
    const link = screen.getByRole('link', { name: 'Plain href' });
    expect(link).toHaveAttribute('href', 'https://example.com');
    expect(link).not.toHaveAttribute('target');
    expect(link).not.toHaveAttribute('rel');
  });

  it('respects an explicit target/rel override', () => {
    render(
      <Link href="https://example.com" external target="_self" rel="opener">
        Override
      </Link>
    );
    const link = screen.getByRole('link', { name: 'Override' });
    expect(link).toHaveAttribute('target', '_self');
    expect(link).toHaveAttribute('rel', 'opener');
  });

  it('renders a router anchor when to is provided', () => {
    render(<Link to="/devices">Devices</Link>);
    const link = screen.getByRole('link', { name: 'Devices' });
    expect(link).toHaveAttribute('href', '/devices');
  });

  it('forwards onClick on a to-link', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const onClick = vi.fn();
    render(
      <Link to="/devices" onClick={onClick}>
        Devices
      </Link>
    );
    await user.click(screen.getByRole('link', { name: 'Devices' }));
    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it('renders a button when only onClick is provided', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const onClick = vi.fn();
    render(<Link onClick={onClick}>Run action</Link>);
    const btn = screen.getByRole('button', { name: 'Run action' });
    expect(btn.tagName).toBe('BUTTON');
    await user.click(btn);
    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it('renders a span when nothing actionable is provided', () => {
    render(<Link>cosmetic</Link>);
    const node = screen.getByText('cosmetic');
    expect(node.tagName).toBe('SPAN');
    expect(node).not.toHaveAttribute('href');
  });

  it('respects explicit component override on cosmetic links', () => {
    render(<Link component="div">div link</Link>);
    const node = screen.getByText('div link');
    expect(node.tagName).toBe('DIV');
  });
});
