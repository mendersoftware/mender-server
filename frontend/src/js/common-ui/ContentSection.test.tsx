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

import { ContentSection } from './ContentSection';

describe('ContentSection Component', () => {
  it('renders correctly with string title', async () => {
    const { baseElement } = render(<ContentSection title="Device Identity">Content here</ContentSection>);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('renders correctly with title and postTitle', async () => {
    const { baseElement } = render(
      <ContentSection title="Some Title" postTitle={<>Online</>} isAddOn>
        Content here
      </ContentSection>
    );
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
    expect(screen.getByText('Online')).toBeInTheDocument();
    expect(screen.getByText('Add-on')).toBeInTheDocument();
  });
});
