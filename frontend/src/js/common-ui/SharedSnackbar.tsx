// Copyright 2018 Northern.tech AS
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
import React, { memo } from 'react';

import { Snackbar } from '@mui/material';

import copy from 'copy-to-clipboard';

export const SharedSnackbar = ({ setSnackbar, snackbar }) => {
  const handleActionClick = () => {
    copy(snackbar.message);
    setSnackbar('Copied to clipboard');
  };

  const onCloseSnackbar = (_, reason) => {
    const { onClose = false } = snackbar;
    if (onClose && reason === 'clickaway') {
      return;
    }
    setSnackbar('');
  };

  const { maxWidth, onClick = handleActionClick, ...snackProps } = snackbar;
  return (
    <Snackbar
      {...snackProps}
      anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      style={{ maxWidth, height: 'auto', lineHeight: '28px', padding: 24, whiteSpace: 'pre-line' }}
      onClick={onClick}
      onClose={onCloseSnackbar}
    />
  );
};

const areEqual = (prevProps, nextProps) => {
  if (prevProps.snackbar.open != nextProps.snackbar.open || prevProps.snackbar.message != nextProps.snackbar.message) {
    return false;
  }
  return prevProps.snackbar.children != nextProps.snackbar.children;
};

export default memo(SharedSnackbar, areEqual);
