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
import { useEffect, useState } from 'react';

import { Check as CheckIcon, FileCopyOutlined as CopyPasteIcon, InfoOutlined as InfoOutlinedIcon } from '@mui/icons-material';
// material ui
import { Accordion, AccordionDetails, AccordionSummary, Button, Chip, CircularProgress, DialogContentText, Divider, Tooltip, Typography } from '@mui/material';

import { ConfirmModal } from '@northern.tech/common-ui/ConfirmModal';
import { Code } from '@northern.tech/common-ui/CopyCode';
import Time from '@northern.tech/common-ui/Time';
import { DEVICE_DISMISSAL_STATE, DEVICE_STATES, TIMEOUTS } from '@northern.tech/store/constants';
import { formatTime } from '@northern.tech/utils/helpers';
import copy from 'copy-to-clipboard';

const tierChangeTooltip = (newTier: string) =>
  `This has a different tier than the currently active authentication set. If you accept this authentication set, this device will become a ${newTier} device and contribute to your ${newTier} device tier limit.`;

const isRequestOnAcceptedDevice = (device, authset) => device.status === DEVICE_STATES.accepted && authset.status !== DEVICE_STATES.accepted;

const isRequestAction = (status, device, authset) =>
  isRequestOnAcceptedDevice(device, authset) || (status === DEVICE_DISMISSAL_STATE && device.status === DEVICE_STATES.pending);

const getDismissalConfirmation = (device, authset) => {
  if (authset.status === DEVICE_STATES.preauth) {
    return 'The device authentication set will be removed from the preauthorization list.';
  }
  if (isRequestOnAcceptedDevice(device, authset)) {
    // if device is accepted but you are dismissing an authset that is not accepted, device status is unaffected:
    return (
      <>
        The request will be dismissed, and removed from the list. Dismissing this request will <b>not</b> affect the device status as it is using a different
        authentication set. Do you want to continue?
      </>
    );
  }
  if (device.status === DEVICE_STATES.pending) {
    return 'The request will be dismissed, and removed from the list. If the device connects again with the same key, it will appear as a new pending request. Do you want to continue?';
  }
  return 'The device will no longer be accepted using this public key, and this authentication set will be removed. If the device connects again with the same key, it will appear as a new pending request. Are you sure you want to continue?';
};

export const getConfirmationMessage = (status, device, authset) => {
  let message = '';
  switch (status) {
    case DEVICE_STATES.accepted:
      message = 'The request will be accepted, and the device with this authentication set will be authorized by the server.';
      if (device.status === DEVICE_STATES.accepted) {
        // if device already accepted, and you are accepting a different authset:
        return `${message} The previously accepted authentication set will be rejected automatically. Do you want to continue?`;
      }
      return message;
    case DEVICE_STATES.rejected:
      if (isRequestOnAcceptedDevice(device, authset)) {
        // if device is accepted but you are rejecting an authset that is not accepted, device status is unaffected:
        return (
          <>
            This request will be rejected, and further requests from this authentication set will be blocked. Rejecting this request will <b>not</b> affect the
            device status as it is using a different authentication set. Do you want to continue?
          </>
        );
      }
      return 'The device with this authentication set will be rejected, and blocked from communicating with the Mender server. Are you sure you want to continue?';
    case DEVICE_DISMISSAL_STATE:
      message = getDismissalConfirmation(device, authset);
      break;
    default:
      break;
  }
  return message;
};

const LF = '\n';

const AuthSetStatus = ({ authset, device }) => {
  if (authset.status === device.status) {
    return (
      <Typography className="capitalized" variant="body2">
        Active
      </Typography>
    );
  }
  if (authset.status === DEVICE_STATES.pending) {
    return <Chip size="small" label="New" style={{ justifySelf: 'flex-start' }} />;
  }
  return <div />;
};

const confirmModalConfig = {
  [DEVICE_STATES.accepted]: { header: 'Accept authentication set?', confirmButtonText: 'Accept request', isDanger: false },
  [DEVICE_STATES.rejected]: { header: 'Reject device?', confirmButtonText: 'Reject device', isDanger: true },
  [DEVICE_DISMISSAL_STATE]: { header: 'Dismiss device?', confirmButtonText: 'Dismiss device', isDanger: true }
};

const requestModalConfig = {
  [DEVICE_STATES.rejected]: { header: 'Reject request?', confirmButtonText: 'Reject request' },
  [DEVICE_DISMISSAL_STATE]: { header: 'Dismiss request?', confirmButtonText: 'Dismiss request' }
};

const ActionButtons = ({ authset, limitMaxed, onAcceptClick, onDismissClick, onRequestConfirm, userCapabilities }) => {
  const { canManageDevices } = userCapabilities;
  if (!canManageDevices) {
    return null;
  }
  return (
    <div className="action-buttons flexbox">
      <Button onClick={onAcceptClick} disabled={authset.status === DEVICE_STATES.accepted || authset.status === DEVICE_STATES.preauth || limitMaxed}>
        Accept
      </Button>
      <Button
        onClick={() => onRequestConfirm(DEVICE_STATES.rejected)}
        disabled={authset.status === DEVICE_STATES.rejected || authset.status === DEVICE_STATES.preauth}
      >
        Reject
      </Button>
      <Button onClick={onDismissClick}>Dismiss</Button>
    </div>
  );
};

const AuthsetListItem = ({ authset, classes, columns, confirm, device, isExpanded, limitMaxed, loading, onExpand, total, userCapabilities }) => {
  const [showKey, setShowKey] = useState(false);
  const [keyHash, setKeyHash] = useState('');
  const [endKey, setEndKey] = useState('');
  const [copied, setCopied] = useState(false);
  const [isConfirmOpen, setIsConfirmOpen] = useState(false);
  const [confirmStatus, setConfirmStatus] = useState(DEVICE_STATES.accepted);
  const { tier } = device;

  const newTier = authset.tier !== tier;

  useEffect(() => {
    if (!isExpanded) {
      setShowKey(false);
    }
  }, [isExpanded]);

  useEffect(() => {
    const data = new TextEncoder().encode(authset.pubkey);
    if (crypto?.subtle) {
      crypto.subtle.digest('SHA-256', data).then(hashBuffer => {
        const hashHex = Array.from(new Uint8Array(hashBuffer))
          .map(b => b.toString(16).padStart(2, '0'))
          .join('');
        setKeyHash(hashHex);
      });
    } else {
      setKeyHash('SHA calculation is not supported by this browser');
    }
    // to ensure the pubkey is copied with the new line at the end we have to double it at the end, as one of the endings gets trimmed in the process of copying
    const key = authset.pubkey.endsWith(LF) ? `${authset.pubkey}${LF}` : authset.pubkey;
    setEndKey(key);
  }, [authset.pubkey]);

  const onShowKey = show => {
    onExpand(show && authset.id);
    setShowKey(show);
  };

  const onRequestConfirm = status => {
    setConfirmStatus(status);
    setIsConfirmOpen(true);
  };

  const onConfirm = confirmedState => confirm(device.id, authset.id, confirmedState);

  const onDismissClick = () => {
    if (total > 1 || device.status !== DEVICE_STATES.pending) {
      return onRequestConfirm(DEVICE_DISMISSAL_STATE);
    }
    return onConfirm(DEVICE_DISMISSAL_STATE);
  };

  const onAcceptClick = () => {
    if (total > 1) {
      return onRequestConfirm(DEVICE_STATES.accepted);
    }
    return onConfirm(DEVICE_STATES.accepted);
  };

  const onCopyKey = () => {
    copy(endKey);
    setCopied(true);
    setTimeout(() => setCopied(false), TIMEOUTS.fiveSeconds);
  };

  const toggleKey = () => onShowKey(!showKey);
  const key = <Button onClick={toggleKey}>{showKey ? 'Hide' : 'Show'} key</Button>;
  const content = (
    <div className="full-width">
      <div className={`${classes.fitContent} flexbox centered`}>
        <Code style={{ padding: 0 }} noBackground>
          <Typography variant="code1">{endKey.trim()}</Typography>
        </Code>
        {copied ? (
          <Typography className="green fadeIn flexbox align-items-center margin-left-small">
            <CheckIcon className="margin-right-x-small" fontSize="small" />
            copied
          </Typography>
        ) : (
          <CopyPasteIcon className="margin-left-small clickable" onClick={onCopyKey} />
        )}
      </div>
      <Divider className="margin-bottom-x-small" />
      <div title="SHA256">
        <Typography>Checksum</Typography>
        <Typography variant="body2">{keyHash}</Typography>
      </div>
    </div>
  );
  return (
    <>
      <Accordion className={`${classes.accordion} padding-none`} square expanded={isExpanded}>
        <AccordionSummary component="div" className={`columns-${columns.length}`}>
          <AuthSetStatus authset={authset} device={device} />
          <Typography className="capitalized" variant="body2">
            {authset.status}
          </Typography>
          <Tooltip arrow title={newTier ? tierChangeTooltip(authset.tier) : undefined}>
            <Typography className={`capitalized clickable flexbox align-items-center ${classes.fitContent}`} component="div" variant="body2">
              {newTier && <InfoOutlinedIcon className="margin-right-x-small muted" fontSize="small" />}
              <div>{authset.tier}</div>
            </Typography>
          </Tooltip>
          <div>{key}</div>
          <Time Component={Typography} component="time" variant="body2" value={formatTime(authset.ts)} />
          {loading === authset.id ? (
            <div className="flexbox align-items-center">
              <Typography variant="body2">Updating...</Typography>
              <CircularProgress className="margin-left-x-small" size={20} />
            </div>
          ) : (
            <ActionButtons
              authset={authset}
              limitMaxed={limitMaxed}
              onAcceptClick={onAcceptClick}
              onDismissClick={onDismissClick}
              onRequestConfirm={onRequestConfirm}
              userCapabilities={userCapabilities}
            />
          )}
        </AccordionSummary>
        <AccordionDetails>
          <div className="margin-top-small">{content}</div>
        </AccordionDetails>
      </Accordion>
      <ConfirmModal
        open={isConfirmOpen}
        close={() => setIsConfirmOpen(false)}
        description={<DialogContentText>{getConfirmationMessage(confirmStatus, device, authset)}</DialogContentText>}
        onConfirm={() => onConfirm(confirmStatus)}
        isCritical={authset.status === DEVICE_STATES.accepted}
        {...confirmModalConfig[confirmStatus]}
        {...(isRequestAction(confirmStatus, device, authset) ? requestModalConfig[confirmStatus] : {})}
      />
    </>
  );
};

export default AuthsetListItem;
