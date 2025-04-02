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
import { useSelector } from 'react-redux';

import { Snackbar } from '@mui/material';

import storeActions from '@northern.tech/store/actions';
import { getSnackbar } from '@northern.tech/store/appSlice/selectors';
import { useAppDispatch } from '@northern.tech/store/store';
import copy from 'copy-to-clipboard';

const { setSnackbar } = storeActions;

export const SharedSnackbar = () => {
  const dispatch = useAppDispatch();
  const snackbar = useSelector(getSnackbar);
  const handleActionClick = () => {
    if (typeof snackbar.message === 'string') {
      copy(snackbar.message);
      dispatch(setSnackbar('Copied to clipboard'));
    }
  };

  const { preventClickToCopy, ...snackProps } = snackbar;
  return (
    <Snackbar
      {...snackProps}
      anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      style={{ maxWidth: 900, height: 'auto', lineHeight: '28px', padding: 24, whiteSpace: 'pre-line' }}
      onClick={preventClickToCopy ? undefined : handleActionClick}
      onClose={() => dispatch(setSnackbar(''))}
    />
  );
};

export default SharedSnackbar;
