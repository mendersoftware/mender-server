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
import { useEffect, useMemo, useState } from 'react';
import { Controller, FormProvider, useForm, useWatch } from 'react-hook-form';
import { useSelector } from 'react-redux';

// material ui
import {
  Alert,
  Button,
  DialogActions,
  DialogContent,
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
  TextField,
  Typography
} from '@mui/material';

import { ConfirmModal } from '@northern.tech/common-ui/ConfirmModal';
import CopyCode from '@northern.tech/common-ui/CopyCode';
import { SettingsItem } from '@northern.tech/common-ui/SettingsItem';
import Time, { RelativeTime } from '@northern.tech/common-ui/Time';
import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import type { Role } from '@northern.tech/store/constants';
import { canAccess as canShow } from '@northern.tech/store/constants';
import { getCurrentUser, getIsEnterprise, getRolesById } from '@northern.tech/store/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import { generateToken, getTokens, revokeToken } from '@northern.tech/store/thunks';
import type { PersonalAccessToken } from '@northern.tech/types/MenderTypes';
import { customSort, toggle } from '@northern.tech/utils/helpers';
import dayjs from 'dayjs';

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
  'never': 0,
  '7 days': 7 * A_DAY,
  '30 days': 30 * A_DAY,
  '90 days': 90 * A_DAY,
  'a year': 365 * A_DAY
};

interface TokenFormValues {
  expiresIn: number;
  name: string;
}

interface AccessTokenCreationDialogProps {
  isEnterprise?: boolean;
  onCancel: () => void;
  onGenerate: (values: TokenFormValues) => void;
  rolesById?: Record<string, Role>;
  token?: string;
  userRoles?: string[];
}

export const AccessTokenCreationDialog = ({ onCancel, onGenerate, isEnterprise, rolesById = {}, token, userRoles = [] }: AccessTokenCreationDialogProps) => {
  const methods = useForm<TokenFormValues>({ mode: 'onSubmit', defaultValues: { expiresIn: expirationTimes['a year'], name: '' } });
  const { control, handleSubmit } = methods;
  const expiresIn = useWatch({ control, name: 'expiresIn' });

  const expirationDate = useMemo(() => dayjs().add(expiresIn, 'seconds'), [expiresIn]);
  const neverExpires = expiresIn === expirationTimes.never;

  const tokenRoles = useMemo(() => userRoles.map(roleId => rolesById[roleId]?.name).join(', '), [rolesById, userRoles]);

  return (
    <BaseDialog title="Create new token" open onClose={onCancel} fullWidth maxWidth="sm">
      <FormProvider {...methods}>
        <form noValidate onSubmit={handleSubmit(onGenerate)}>
          <DialogContent className="flexbox column">
            <Typography className="margin-bottom-small" variant="subtitle1">
              Name
            </Typography>
            <TextInput id="name" label="Token name" disabled={!!token} validations="trim,isLength:1" required requiredRendered={false} width="100%" />
            <Typography className="margin-top-medium margin-bottom-small" variant="subtitle1">
              Expiration
            </Typography>
            <FormControl className="full-width">
              <InputLabel id="token-expiration-label">Expiration</InputLabel>
              <Controller
                name="expiresIn"
                control={control}
                render={({ field: { onChange, value } }) => (
                  <Select labelId="token-expiration-label" label="Expiration" disabled={!!token} onChange={onChange} value={value}>
                    {Object.entries(expirationTimes).map(([title, value]) => (
                      <MenuItem key={value} value={value}>
                        {title}
                      </MenuItem>
                    ))}
                  </Select>
                )}
              />
              {neverExpires ? (
                <FormHelperText>The token will never expire.</FormHelperText>
              ) : (
                <FormHelperText title={expirationDate.format('YYYY-MM-DD HH:mm')}>
                  expires on <Time format="YYYY-MM-DD" value={expirationDate} />
                </FormHelperText>
              )}
            </FormControl>
            {neverExpires && (
              <Alert className="margin-top-small" severity="warning">
                Never-expiring tokens are a security risk. We recommend to set an expiration date and rotate the secret at least yearly.
              </Alert>
            )}
            {token && (
              <div className="margin-top-medium">
                <CopyCode code={token} />
                <Alert className="margin-top-small" severity="error">
                  This is the only time you will be able to see the token, so make sure to store it in a safe place.
                </Alert>
              </div>
            )}
            {isEnterprise && (
              <TextField
                className="margin-top-medium"
                label="Permission level"
                id="role-name"
                value={tokenRoles}
                disabled
                helperText="The token will have the same permissions as your user"
              />
            )}
          </DialogContent>
          <DialogActions>
            {token ? (
              <Button variant="contained" onClick={onCancel}>
                Close
              </Button>
            ) : (
              <>
                <Button onClick={onCancel}>Cancel</Button>
                <Button variant="contained" type="submit">
                  Create token
                </Button>
              </>
            )}
          </DialogActions>
        </form>
      </FormProvider>
    </BaseDialog>
  );
};

export const AccessTokenManagement = () => {
  const [showGeneration, setShowGeneration] = useState(false);
  const [newToken, setNewToken] = useState<string>();
  const [tokenToRevoke, setTokenToRevoke] = useState<PersonalAccessToken>();
  const isEnterprise = useSelector(getIsEnterprise);
  const { tokens = [], roles: userRoles = [], id } = useSelector(getCurrentUser);
  const rolesById = useSelector(getRolesById);
  const dispatch = useAppDispatch();

  useEffect(() => {
    if (!id) {
      return;
    }
    dispatch(getTokens());
  }, [dispatch, id]);

  const toggleGenerateClick = () => {
    setNewToken(undefined);
    setShowGeneration(toggle);
  };

  const onGenerate = (values: TokenFormValues) =>
    dispatch(generateToken(values))
      .unwrap()
      .then(setNewToken)
      .catch(() => {});

  const hasLastUsedInfo = useMemo(() => tokens.some(token => !!token.last_used), [tokens]);

  const columns = useMemo(() => columnData.filter(column => column.canShow({ hasLastUsedInfo })), [hasLastUsedInfo]);

  return (
    <>
      <SettingsItem
        title="Personal access token management"
        description={<Button onClick={toggleGenerateClick}>Generate a token</Button>}
        secondary={
          !!tokens.length && (
            <Table>
              <TableHead>
                <TableRow>
                  {columns.map(column => (
                    <TableCell key={column.id}>{column.label}</TableCell>
                  ))}
                </TableRow>
              </TableHead>
              <TableBody>
                {tokens
                  .slice()
                  .sort(customSort(true, creationTimeAttribute))
                  .map(token => (
                    <TableRow key={token.id} hover>
                      {columns.map(column => (
                        <TableCell key={column.id}>{column.render({ onRevokeTokenClick: setTokenToRevoke, token })}</TableCell>
                      ))}
                    </TableRow>
                  ))}
              </TableBody>
            </Table>
          )
        }
      />
      {showGeneration && (
        <AccessTokenCreationDialog
          onCancel={toggleGenerateClick}
          onGenerate={onGenerate}
          isEnterprise={isEnterprise}
          rolesById={rolesById}
          token={newToken}
          userRoles={userRoles}
        />
      )}
      {!!tokenToRevoke && (
        <ConfirmModal
          header="Revoke token"
          description={
            <>
              Are you sure you want to revoke the token <b>{tokenToRevoke.name}</b>?
            </>
          }
          confirmButtonText="Revoke token"
          open
          close={() => setTokenToRevoke(undefined)}
          onConfirm={() => dispatch(revokeToken(tokenToRevoke))}
        />
      )}
    </>
  );
};

export default AccessTokenManagement;
