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
import { useSelector } from 'react-redux';

// material ui
import { Button } from '@mui/material';

import { getCard, getOrganization } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { confirmCardUpdate, getCurrentCard, startCardUpdate } from '@northern.tech/store/thunks';

import CardSection from '../CardSection';
import { CardDetails } from './Billing';

interface OrganizationPaymentSettingsProps {
  onComplete?: () => void;
  isValid: boolean;
  updatingCard: boolean;
  setUpdatingCard: (updatingCard: boolean) => void;
}
export const OrganizationPaymentSettings = (props: OrganizationPaymentSettingsProps) => {
  const { onComplete, isValid, updatingCard, setUpdatingCard } = props;
  const card = useSelector(getCard);
  const organization = useSelector(getOrganization);
  const dispatch = useAppDispatch();

  const onCardConfirm = async () => {
    await dispatch(confirmCardUpdate());
    dispatch(getCurrentCard());
    setUpdatingCard(false);
    if (onComplete) {
      onComplete();
    }
  };

  return (
    <div className="flexbox margin-top">
      <div className="margin-top" />
      <div className="flexbox column">
        <div className="flexbox">
          <h5 className="margin-top-small margin-bottom-x-small margin-right-x-small">{updatingCard ? 'Edit payment card' : 'Payment card'}</h5>
          <Button className="align-self-start" onClick={() => setUpdatingCard(!updatingCard)}>
            {updatingCard ? 'cancel' : 'edit'}
          </Button>
        </div>
        {updatingCard ? (
          <CardSection
            isSignUp={false}
            organization={organization}
            onClose={() => setUpdatingCard(false)}
            onCardConfirmed={onCardConfirm}
            isValid={isValid}
            onSubmit={() => dispatch(startCardUpdate()).unwrap()}
            beforeCardSubmit={() => dispatch(startCardUpdate()).unwrap()}
          />
        ) : (
          <CardDetails card={card} />
        )}
      </div>
    </div>
  );
};

export default OrganizationPaymentSettings;
