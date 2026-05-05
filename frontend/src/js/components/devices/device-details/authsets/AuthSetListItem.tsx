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

import { InfoOutlined as InfoOutlinedIcon } from '@mui/icons-material';
// material ui
import { Accordion, AccordionActions, AccordionDetails, AccordionSummary, Button, Chip, Divider, Tooltip, Typography } from '@mui/material';

import CopyCode from '@northern.tech/common-ui/CopyCode';
import { Link } from '@northern.tech/common-ui/Link';
import Loader from '@northern.tech/common-ui/Loader';
import Time from '@northern.tech/common-ui/Time';
import { DEVICE_DISMISSAL_STATE, DEVICE_STATES } from '@northern.tech/store/constants';
import { formatTime } from '@northern.tech/utils/helpers';

const padder = <div key="padder" style={{ flexGrow: 1 }} />;

const tierChangeTooltip = (newTier: string) =>
  `This has a different tier than the currently active authentication set. If you accept this authentication set, this device will become a ${newTier} device and contribute to your ${newTier} device tier limit.`;

const getDismissalConfirmation = (device, authset) => {
  switch (authset.status) {
    case DEVICE_STATES.preauth:
      return 'The device authentication set will be removed from the preauthorization list.';
    case DEVICE_STATES.accepted:
      if (device.auth_sets.length > 1) {
        // if there are other authsets, device will still be in UI
        return 'The device with this public key will no longer be accepted, and this authorization request will be removed from the UI.';
      } else {
        return 'The device with this public key will no longer be accepted, and will be removed from the UI. If it makes another request in the future, it will show again as pending for you to accept or reject at that time.';
      }
    case DEVICE_STATES.pending: {
      const message = 'You can dismiss this authentication request for now.';
      if (device.auth_sets.length > 1) {
        // it has other authsets
        return `${message} This will remove this request from the UI, but won’t affect the device.`;
      }
      return `${message} The device will be removed from the UI, but if the same device asks for authentication again in the future, it will show again as pending.`;
    }
    case DEVICE_STATES.rejected:
      return 'This request will be removed from the UI, but if the device asks for authentication again in the future, it will show as pending for you to accept or reject it at that time.';
    default:
      break;
  }
  return '';
};

export const getConfirmationMessage = (status, device, authset) => {
  let message = '';
  switch (status) {
    case DEVICE_STATES.accepted:
      message = 'By accepting, the device with this identity data and public key will be granted authentication by the server.';
      if (device.status === DEVICE_STATES.accepted) {
        // if device already accepted, and you are accepting a different authset:
        return `${message} The previously accepted public key will be rejected automatically in favor of this new key.`;
      }
      return message;
    case DEVICE_STATES.rejected:
      message = 'The device with this identity data and public key will be rejected, and blocked from communicating with the Mender server.';
      if (device.status === DEVICE_STATES.accepted && authset.status !== DEVICE_STATES.accepted) {
        // if device is accepted but you are rejecting an authset that is not accepted, device status is unaffected:
        return `${message} Rejecting this request will not affect the device status as it is using a different key. `;
      }
      return message;
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
    return <div className="capitalized">Active</div>;
  }
  if (authset.status === DEVICE_STATES.pending) {
    return <Chip size="small" label="new" color="primary" style={{ justifySelf: 'flex-start' }} />;
  }
  return <div />;
};

const ActionButtons = ({ authset, confirmMessage, newStatus, limitMaxed, onAcceptClick, onDismissClick, onRequestConfirm, userCapabilities }) => {
  const { canManageDevices } = userCapabilities;
  if (!canManageDevices) {
    return null;
  }
  return confirmMessage.length ? (
    <div>Set to: {newStatus}?</div>
  ) : (
    <div className="action-buttons flexbox">
      {authset.status !== DEVICE_STATES.accepted && authset.status !== DEVICE_STATES.preauth && !limitMaxed ? (
        <Link component="div" onClick={onAcceptClick}>
          Accept
        </Link>
      ) : (
        <div>Accept</div>
      )}
      {authset.status !== DEVICE_STATES.rejected && authset.status !== DEVICE_STATES.preauth ? (
        <Link component="div" onClick={() => onRequestConfirm(DEVICE_STATES.rejected)}>
          Reject
        </Link>
      ) : (
        <div>Reject</div>
      )}
      <Link component="div" onClick={onDismissClick}>
        Dismiss
      </Link>
    </div>
  );
};

const AuthsetListItem = ({ authset, classes, columns, confirm, device, isExpanded, limitMaxed, loading, onExpand, total, userCapabilities }) => {
  const [showKey, setShowKey] = useState(false);
  const [confirmMessage, setConfirmMessage] = useState('');
  const [newStatus, setNewStatus] = useState('');
  const [keyHash, setKeyHash] = useState('');
  const [endKey, setEndKey] = useState('');
  const { tier } = device;

  const newTier = authset.tier !== tier;

  useEffect(() => {
    if (!isExpanded) {
      setShowKey(false);
      setConfirmMessage('');
      setNewStatus('');
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
    setConfirmMessage(false);
  };

  const onCancelConfirm = () => {
    onExpand(false);
    setConfirmMessage('');
  };

  const onRequestConfirm = status => {
    const message = getConfirmationMessage(status, device, authset);
    setConfirmMessage(message);
    setNewStatus(status);
    setShowKey(false);
    onExpand(authset.id);
  };

  const onConfirm = confirmedState => confirm(device.id, authset.id, confirmedState).then(onCancelConfirm);

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

  let key = (
    <Link component="span" onClick={onShowKey}>
      show key
    </Link>
  );
  let content = [
    padder,
    <p className="bold expanded" key="content">
      {loading === authset.id ? 'Updating status' : `${confirmMessage} Are you sure you want to continue?`}
    </p>
  ];

  if (showKey) {
    content = [
      <div key="content">
        <CopyCode code={endKey} />
        <Divider className={classes.divider} />
        <div title="SHA256">
          <Typography>Checksum</Typography>
          <Typography variant="body2">{keyHash}</Typography>
        </div>
      </div>,
      padder
    ];
    key = (
      <Link component="span" onClick={() => onShowKey(false)}>
        hide key
      </Link>
    );
  }
  return (
    <Accordion className={classes.accordion} square expanded={isExpanded}>
      <AccordionSummary className={`columns-${columns.length}`}>
        <AuthSetStatus authset={authset} device={device} />
        <div className="capitalized">{authset.status}</div>
        <Tooltip arrow title={newTier ? tierChangeTooltip(authset.tier) : undefined}>
          <div className={`capitalized clickable flexbox align-items-center ${classes.fitContent}`}>
            {newTier && <InfoOutlinedIcon className="margin-right-x-small muted" fontSize="small" />}
            <div>{authset.tier}</div>
          </div>
        </Tooltip>
        {key}
        <Time value={formatTime(authset.ts)} />
        {loading === authset.id ? (
          <div>
            Updating status <Loader table={true} waiting={true} show={true} style={{ height: '4px', marginLeft: '10px' }} />
          </div>
        ) : (
          <ActionButtons
            authset={authset}
            confirmMessage={confirmMessage}
            newStatus={newStatus}
            limitMaxed={limitMaxed}
            onAcceptClick={onAcceptClick}
            onDismissClick={onDismissClick}
            onRequestConfirm={onRequestConfirm}
            userCapabilities={userCapabilities}
          />
        )}
      </AccordionSummary>
      <AccordionDetails>{content}</AccordionDetails>
      {isExpanded && !showKey && (
        <AccordionActions className="margin-right-small">
          {loading === authset.id ? (
            <Loader table={true} waiting={true} show={true} style={{ height: '4px' }} />
          ) : (
            <>
              <Button className="margin-right-small" onClick={onCancelConfirm}>
                Cancel
              </Button>
              <Button variant="contained" onClick={() => onConfirm(newStatus)}>
                Confirm
              </Button>
            </>
          )}
        </AccordionActions>
      )}
    </Accordion>
  );
};

export default AuthsetListItem;
