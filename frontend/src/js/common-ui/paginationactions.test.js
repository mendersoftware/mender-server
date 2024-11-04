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
import React from 'react';

import { act, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { render } from '../../../tests/setupTests';
import { TablePaginationActions } from './pagination';

describe('TablePaginationActions Component', () => {
  it('paginates properly', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const changeListener = jest.fn();
    render(<TablePaginationActions count={42} page={0} onPageChange={changeListener} />);
    expect(screen.getByText('1-20 of 42')).toBeInTheDocument();
    expect(screen.getAllByRole('button')[0]).toBeDisabled();
    await user.click(screen.getAllByRole('button')[1]);
    act(() => jest.advanceTimersByTime(400));
    expect(changeListener).toHaveBeenCalled();
    expect(screen.getByText('21-40 of 42')).toBeInTheDocument();
  });
  it('paginates properly backwards', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const changeListener = jest.fn();
    render(<TablePaginationActions count={42} page={1} onPageChange={changeListener} />);
    expect(screen.getAllByRole('button')[0]).not.toBeDisabled();
    expect(screen.getAllByRole('button')[1]).not.toBeDisabled();
    await user.click(screen.getAllByRole('button')[0]);
    act(() => jest.advanceTimersByTime(400));
    expect(changeListener).toHaveBeenCalled();
  });
  it('prevents going too far', async () => {
    const changeListener = jest.fn();
    render(<TablePaginationActions count={42} page={2} onPageChange={changeListener} />);
    expect(screen.getByText('41-42 of 42')).toBeInTheDocument();
    expect(screen.getAllByRole('button')[1]).toBeDisabled();
  });
  it('prevents exceeding the pagination limit', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const changeListener = jest.fn();
    render(<TablePaginationActions count={2000000} page={498} onPageChange={changeListener} />);
    expect(screen.getAllByRole('button')[1]).not.toBeDisabled();
    await user.click(screen.getAllByRole('button')[1]);
    act(() => jest.advanceTimersByTime(400));
    expect(changeListener).toHaveBeenCalled();
    expect(screen.getByText('9981-10000 of 2000000')).toBeInTheDocument();
    expect(screen.getAllByRole('button')[1]).toBeDisabled();
  });
});
