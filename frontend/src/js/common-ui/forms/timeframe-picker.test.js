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
import { FormProvider, useForm } from 'react-hook-form';

import { LocalizationProvider } from '@mui/x-date-pickers';
import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';

import { screen } from '@testing-library/react';

import { undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import { formRenderWrapper } from './form.test';
import TimeframePicker from './timeframe-picker';

const testRender = ui => {
  const Wrapper = ({ children }) => {
    const methods = useForm({ mode: 'onChange', defaultValues: { startDate: '', endDate: '2019-01-14T03:00:00.000' } });
    return <FormProvider {...methods}>{children}</FormProvider>;
  };
  return render(<Wrapper>{ui}</Wrapper>);
};

describe('TimeframePicker Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = formRenderWrapper(
      <LocalizationProvider dateAdapter={AdapterDayjs}>
        <TimeframePicker tonight={new Date().toISOString()} />
      </LocalizationProvider>
    );
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
  it('works in different timezones correctly', async () => {
    testRender(
      <LocalizationProvider dateAdapter={AdapterDayjs}>
        <TimeframePicker tonight={new Date().toISOString()} />
      </LocalizationProvider>
    );
    const endDatePicker = screen.getByLabelText(/to/i);
    expect(endDatePicker).toBeInTheDocument();
    expect(endDatePicker).toHaveValue('January 13th');
  });
});
