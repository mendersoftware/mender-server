// Copyright 2024 Northern.tech AS
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
import { Drawer } from '@mui/material';

import { DrawerTitle } from '@northern.tech/common-ui/DrawerTitle';
import { AddonSelect } from '@northern.tech/common-ui/forms/AddonSelect';
import Form from '@northern.tech/common-ui/forms/Form';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import { AddonId } from '@northern.tech/store/constants';
import { Addon } from '@northern.tech/store/organizationSlice/types';

const note = 'Tell us a little about your requirements and device fleet size, so we can provide you with an accurate quote';

interface EnterpriseRequestExpandedProps {
  addons: Addon[];
  onClose: () => void;
  onSendRequest: (e: { message: string; selectedAddons: AddonId[] }) => void;
}

const defaultValues = { message: '', selectedAddons: [] };

export const EnterpriseRequestExpanded = (props: EnterpriseRequestExpandedProps) => {
  const { onSendRequest, onClose, addons } = props;
  return (
    <Drawer anchor="right" open={true} PaperProps={{ style: { minWidth: '75vw' } }}>
      <DrawerTitle onClose={onClose} title="Request for Mender Enterprise" />
      <div className="flexbox column margin-bottom-large">
        <div className="margin-top-large">
          Get in touch with our team to request a quote for <b>Mender Enterprise.</b>
        </div>
        <Form
          onSubmit={onSendRequest}
          handleCancel={onClose}
          submitLabel="Submit request"
          showButtons={true}
          initialValues={{ message: '', selectedAddons: addons.map(addon => addon.name) }}
          defaultValues={defaultValues}
          buttonColor="secondary"
        >
          <TextInput
            id="message"
            label="Your message"
            fullWidth
            multiline
            placeholder={note}
            validations="isLength:1"
            required
            className="margin-top-none margin-bottom"
            style={{ maxWidth: '550px' }}
          />
          <AddonSelect name="selectedAddons" />
        </Form>
      </div>
    </Drawer>
  );
};
