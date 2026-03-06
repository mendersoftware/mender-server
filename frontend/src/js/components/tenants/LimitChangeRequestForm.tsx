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
import { useEffect } from 'react';
import { useFormContext } from 'react-hook-form';
import { useSelector } from 'react-redux';

import { Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import Form from '@northern.tech/common-ui/forms/Form';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import { getOrganization } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { requestPlanChange } from '@northern.tech/store/thunks';

const useStyles = makeStyles()(() => ({
  container: { maxWidth: '550px' },
  buttonWrapper: {
    '&.button-wrapper': {
      justifyContent: 'start'
    }
  }
}));
const RequestForm = () => {
  const { setFocus } = useFormContext();
  useEffect(() => {
    setFocus('enterpriseMessage');
  }, [setFocus]);

  return (
    <>
      <TextInput
        id="enterpriseMessage"
        label="Your message"
        InputLabelProps={{ shrink: true }}
        required
        requiredRendered={false}
        InputProps={{
          rows: 3,
          multiline: true,
          placeholder: 'Tell us about your requirements and device fleet details to help us provide you with an accurate quote. '
        }}
        width="100%"
      />
    </>
  );
};

export const LimitChangeRequestForm = () => {
  const { id: tenantId } = useSelector(getOrganization);
  const { classes } = useStyles();

  const dispatch = useAppDispatch();

  const onEnterpriseRequest = ({ enterpriseMessage }: { enterpriseMessage: string }) =>
    dispatch(
      requestPlanChange({
        tenantId,
        content: {
          user_message: `This request to modify device limits was initiated via the Provider Tenant UI; \n\n Customer message: ${enterpriseMessage}`
        }
      })
    );
  return (
    <div className={classes.container}>
      <Typography variant="h5">Request a change to your device limits</Typography>
      <Typography className="padding-bottom-medium margin-top-medium">
        To request an adjustment to your overall device limits, please fill out the form below with details about your desired device tiers and quantities.
      </Typography>
      <Typography>If you’d like to report a bug or need technical support, please reach out via our support portal instead.</Typography>
      <Form
        className="margin-top-small"
        initialValues={{ enterpriseMessage: '' }}
        buttonColor="secondary"
        showButtons
        classes={classes}
        submitLabel="Submit request"
        validationMode="onSubmit"
        onSubmit={onEnterpriseRequest}
        autocomplete="off"
      >
        <RequestForm />
      </Form>
    </div>
  );
};
