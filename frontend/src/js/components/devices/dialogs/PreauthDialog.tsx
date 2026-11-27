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
import { useCallback, useMemo, useState } from 'react';
import { useSelector } from 'react-redux';

// material ui
import { Alert, Button, DialogActions, DialogContent, Typography } from '@mui/material';

import { ContentSection } from '@northern.tech/common-ui/ContentSection';
import { Link } from '@northern.tech/common-ui/Link';
import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import FileUpload from '@northern.tech/common-ui/forms/FileUpload';
import KeyValueEditor from '@northern.tech/common-ui/forms/KeyValueEditor';
import { Select } from '@northern.tech/common-ui/forms/Select';
import type { DeviceTier } from '@northern.tech/store/constants';
import { DEVICE_TIERS } from '@northern.tech/store/constants';
import { getEnabledTiers, getReachedTierLimits } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { preauthDevice } from '@northern.tech/store/thunks';
import { isEmpty } from '@northern.tech/utils/helpers';

const publicKeyEnvelope = /^-----BEGIN PUBLIC KEY-----([A-Za-z0-9+/=\s]+)-----END PUBLIC KEY-----$/;

const generalFileHint = 'Please upload a PEM encoded public key, starting with "-----BEGIN PUBLIC KEY-----".';
const invalidKeyError = `This file does not contain a public key. ${generalFileHint}`;
const emptyKeyError = `This file is empty. ${generalFileHint}`;

export const DeviceLimitWarning = ({ tier }: { tier?: string }) => {
  const reachedTierLimits = useSelector(getReachedTierLimits).filter(({ type }) => !tier || type === tier);
  if (!reachedTierLimits.length) {
    return null;
  }
  if (tier) {
    const [{ limit, total }] = reachedTierLimits;
    return (
      <Alert severity="warning">
        You have reached your {tier} device limit ({total} of {limit}). Preauthorized devices won&apos;t be accepted when they connect until you increase your
        limit.
      </Alert>
    );
  }
  const reachedLimits = reachedTierLimits.map(({ type, limit, total }) => `${total} of ${limit} ${type}`).join(', ');
  return <Alert severity="warning">You have reached your limit of authorized devices: {reachedLimits}</Alert>;
};

export const PreauthDialog = ({ onCancel, onSubmit }) => {
  const enabledTiers = useSelector(getEnabledTiers);
  const [jsonIdentity, setJsonIdentity] = useState<Record<string, string>>({});
  const [keyError, setKeyError] = useState<string>('');
  const [publicKey, setPublicKey] = useState<string>('');
  const [submitError, setSubmitError] = useState<string>('');
  const [selectedTier, setSelectedTier] = useState<DeviceTier>();
  const dispatch = useAppDispatch();
  const defaultTier = enabledTiers.includes(DEVICE_TIERS.standard) || !enabledTiers.length ? DEVICE_TIERS.standard : enabledTiers[0];
  const tier = selectedTier && enabledTiers.includes(selectedTier) ? selectedTier : defaultTier;
  const tierOptions = useMemo(() => enabledTiers.map(type => ({ id: type, title: type })), [enabledTiers]);

  const convertIdentityToJSON = useCallback(jsonIdentity => {
    setSubmitError('');
    setJsonIdentity(jsonIdentity);
  }, []);

  const onKeyChange = (content?: string) => {
    setSubmitError('');
    setPublicKey('');
    if (content === undefined) {
      setKeyError('');
      return;
    }
    // PEM files usually end with a newline
    const key = content.trim();
    if (!key) {
      setKeyError(emptyKeyError);
      return;
    }
    if (!publicKeyEnvelope.test(key)) {
      setKeyError(invalidKeyError);
      return;
    }
    setKeyError('');
    setPublicKey(key);
  };

  const onHandleSubmit = async shouldClose => {
    setSubmitError('');
    try {
      await dispatch(preauthDevice({ pubkey: publicKey, identity_data: jsonIdentity, tier })).unwrap();
      onSubmit(shouldClose);
    } catch (error) {
      setSubmitError(typeof error === 'string' ? error : (error as Error)?.message);
    }
  };

  const isSubmitDisabled = !publicKey || isEmpty(jsonIdentity);
  return (
    <BaseDialog open title="Preauthorize devices" onClose={onCancel}>
      <DialogContent>
        <Typography>Add a device&apos;s authentication set to authorize it automatically as soon as it connects.</Typography>
        <ContentSection title="Public key">
          <FileUpload
            isValid={!!publicKey}
            placeholder={
              <>
                Drag and drop or <Link>browse</Link> to upload a file
              </>
            }
            onFileChange={onKeyChange}
          />
          {!!keyError && (
            <Alert className="margin-top-small" severity="error">
              {keyError}
            </Alert>
          )}
        </ContentSection>
        <ContentSection title="Identity data">
          <KeyValueEditor onInputChange={convertIdentityToJSON} />
        </ContentSection>
        {tierOptions.length > 1 && (
          <ContentSection title="Device tier">
            <Select
              className="capitalized"
              MenuItemProps={{ className: 'capitalized' }}
              onChange={({ target: { value } }) => setSelectedTier(value as DeviceTier)}
              options={tierOptions}
              value={tier}
            />
            <Alert className="margin-top-small" severity="info">
              If the device reports a different tier, it won’t be accepted automatically and will appear as pending.
            </Alert>
          </ContentSection>
        )}
        <DeviceLimitWarning tier={tier} />
        {!!submitError && <Alert severity="error">The device could not be added: {submitError}</Alert>}
      </DialogContent>
      <DialogActions>
        <Button onClick={onCancel}>Cancel</Button>
        <Button variant="outlined" disabled={isSubmitDisabled} onClick={() => onHandleSubmit(false)}>
          Save and add another
        </Button>
        <Button variant="contained" disabled={isSubmitDisabled} onClick={() => onHandleSubmit(true)}>
          Save
        </Button>
      </DialogActions>
    </BaseDialog>
  );
};

export default PreauthDialog;
