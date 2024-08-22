// Copyright 2022 Northern.tech AS
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
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

// material ui
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControl,
  FormHelperText,
  InputLabel,
  MenuItem,
  Select,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TextField
} from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { generateToken, getTokens, revokeToken } from '../../actions/userActions';
import { canAccess as canShow } from '../../constants/appConstants';
import { customSort, toggle } from '../../helpers';
import { getCurrentUser, getIsEnterprise } from '../../selectors';
import CopyCode from '../common/copy-code';
import Time, { RelativeTime } from '../common/time';

const useStyles = makeStyles()(theme => ({
  accessTokens: {
    minWidth: 900
  },
  creationDialog: {
    minWidth: 500
  },
  formEntries: {
    minWidth: 270
  },
  warning: {
    color: theme.palette.warning.main
  }
}));

const creationTimeAttribute = 'created_ts';
const columnData = [
  { id: 'token', label: 'Token', canShow, render: ({ token }) => token.name },
  { id: creationTimeAttribute, label: 'Date created', canShow, render: ({ token }) => <Time value={token[creationTimeAttribute]} /> },
  {
    id: 'expiration_date',
    label: 'Expires',
    canShow,
    render: ({ token }) => <RelativeTime updateTime={token.expiration_date} shouldCount="up" />
  },
  {
    id: 'last_used',
    label: 'Last used',
    canShow: ({ hasLastUsedInfo }) => hasLastUsedInfo,
    render: ({ token }) => <RelativeTime updateTime={token.last_used} />
  },
  {
    id: 'actions',
    label: 'Manage',
    canShow,
    render: ({ onRevokeTokenClick, token }) => <Button onClick={() => onRevokeTokenClick(token)}>Revoke</Button>
  }
];

const A_DAY = 24 * 60 * 60;
const expirationTimes = {
  'never': {
    value: 0,
    hint: (
      <>
        The token will never expire.
        <br />
        WARNING: Never-expiring tokens are against security best practices. We highly suggest setting a token expiration date and rotating the secret at least
        yearly.
      </>
    )
  },
  '7 days': { value: 7 * A_DAY },
  '30 days': { value: 30 * A_DAY },
  '90 days': { value: 90 * A_DAY },
  'a year': { value: 365 * A_DAY }
};

export const AccessTokenCreationDialog = ({ onCancel, generateToken, isEnterprise, rolesById, token, userRoles }) => {
  const [name, setName] = useState('');
  const [expirationTime, setExpirationTime] = useState(expirationTimes['a year'].value);
  const [expirationDate, setExpirationDate] = useState(new Date());
  const [hint, setHint] = useState('');
  const { classes } = useStyles();

  useEffect(() => {
    const date = new Date();
    date.setSeconds(date.getSeconds() + expirationTime);
    setExpirationDate(date);
    const hint = Object.values(expirationTimes).find(({ value }) => value === expirationTime)?.hint ?? '';
    setHint(hint);
  }, [expirationTime]);

  const onGenerateClick = useCallback(() => generateToken({ name, expiresIn: expirationTime }), [generateToken, name, expirationTime]);

  const onChangeExpirationTime = ({ target: { value } }) => setExpirationTime(value);

  const generationHandler = token ? onCancel : onGenerateClick;

  const generationLabel = token ? 'Close' : 'Create token';

  const nameUpdated = ({ target: { value } }) => setName(value);

  const tokenRoles = useMemo(() => userRoles.map(roleId => rolesById[roleId]?.name).join(', '), [rolesById, userRoles]);

  return (
    <Dialog open>
      <DialogTitle>Create new token</DialogTitle>
      <DialogContent className={classes.creationDialog}>
        <form>
          <TextField className={`${classes.formEntries} required`} disabled={!!token} onChange={nameUpdated} placeholder="Name" value={name} />
        </form>
        <div>
          <FormControl className={classes.formEntries}>
            <InputLabel>Expiration</InputLabel>
            <Select disabled={!!token} onChange={onChangeExpirationTime} value={expirationTime}>
              {Object.entries(expirationTimes).map(([title, item]) => (
                <MenuItem key={item.value} value={item.value}>
                  {title}
                </MenuItem>
              ))}
            </Select>
            {hint ? (
              <FormHelperText className={classes.warning}>{hint}</FormHelperText>
            ) : (
              <FormHelperText title={expirationDate.toISOString().slice(0, 10)}>
                expires on <Time format="YYYY-MM-DD" value={expirationDate} />
              </FormHelperText>
            )}
          </FormControl>
        </div>
        {token && (
          <div className="margin-top margin-bottom">
            <CopyCode code={token} />
            <p className="warning">This is the only time you will be able to see the token, so make sure to store it in a safe place.</p>
          </div>
        )}
        {isEnterprise && (
          <FormControl className={classes.formEntries}>
            <TextField label="Permission level" id="role-name" value={tokenRoles} disabled />
            <FormHelperText>The token will have the same permissions as your user</FormHelperText>
          </FormControl>
        )}
      </DialogContent>
      <DialogActions>
        {!token && <Button onClick={onCancel}>Cancel</Button>}
        <Button disabled={!name.length} variant="contained" onClick={generationHandler}>
          {generationLabel}
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export const AccessTokenRevocationDialog = ({ onCancel, revokeToken, token }) => (
  <Dialog open>
    <DialogTitle>Revoke token</DialogTitle>
    <DialogContent>
      Are you sure you want to revoke the token <b>{token?.name}</b>?
    </DialogContent>
    <DialogActions>
      <Button onClick={onCancel}>Cancel</Button>
      <Button onClick={() => revokeToken(token)}>Revoke Token</Button>
    </DialogActions>
  </Dialog>
);

export const AccessTokenManagement = () => {
  const [showGeneration, setShowGeneration] = useState(false);
  const [showRevocation, setShowRevocation] = useState(false);
  const [currentToken, setCurrentToken] = useState(null);
  const isEnterprise = useSelector(getIsEnterprise);
  const { tokens = [], roles: userRoles = [] } = useSelector(getCurrentUser);
  const rolesById = useSelector(state => state.users.rolesById);
  const dispatch = useDispatch();

  const { classes } = useStyles();

  useEffect(() => {
    dispatch(getTokens());
  }, [dispatch]);

  const toggleGenerateClick = () => {
    setCurrentToken(null);
    setShowGeneration(toggle);
  };

  const toggleRevocationClick = () => {
    setCurrentToken(null);
    setShowRevocation(toggle);
  };

  const onRevokeClick = token => dispatch(revokeToken(token)).then(() => toggleRevocationClick());

  const onRevokeTokenClick = token => {
    toggleRevocationClick();
    setCurrentToken(token);
  };

  const onGenerateClick = config => dispatch(generateToken(config)).then(results => setCurrentToken(results[results.length - 1]));

  const hasLastUsedInfo = useMemo(() => tokens.some(token => !!token.last_used), [tokens]);

  const columns = useMemo(
    () =>
      columnData.reduce((accu, column) => {
        if (!column.canShow({ hasLastUsedInfo })) {
          return accu;
        }
        accu.push(column);
        return accu;
      }, []),
    [hasLastUsedInfo]
  );

  return (
    <>
      <div className={`flexbox space-between margin-top-small ${tokens.length ? classes.accessTokens : ''}`}>
        <p className="help-content">Personal access token management</p>
        <Button onClick={toggleGenerateClick}>Generate a token</Button>
      </div>
      {!!tokens.length && (
        <Table className={classes.accessTokens}>
          <TableHead>
            <TableRow>
              {columns.map(column => (
                <TableCell key={column.id} padding={column.disablePadding ? 'none' : 'normal'}>
                  {column.label}
                </TableCell>
              ))}
            </TableRow>
          </TableHead>
          <TableBody>
            {tokens.sort(customSort(true, creationTimeAttribute)).map(token => (
              <TableRow key={token.id} hover>
                {columns.map(column => (
                  <TableCell key={column.id}>{column.render({ onRevokeTokenClick, token })}</TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
      {showGeneration && (
        <AccessTokenCreationDialog
          onCancel={toggleGenerateClick}
          generateToken={onGenerateClick}
          isEnterprise={isEnterprise}
          rolesById={rolesById}
          token={currentToken}
          userRoles={userRoles}
        />
      )}
      {showRevocation && <AccessTokenRevocationDialog onCancel={toggleRevocationClick} revokeToken={onRevokeClick} token={currentToken} />}
    </>
  );
};

export default AccessTokenManagement;
