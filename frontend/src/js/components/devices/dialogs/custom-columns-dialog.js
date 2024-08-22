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
import React, { Suspense, lazy, useState } from 'react';

// material ui
import { Button, Dialog, DialogActions, DialogTitle } from '@mui/material';

import Loader from '../../common/loader';

const Content = lazy(() => import('./custom-columns-dialog-content'));

export const ColumnCustomizationDialog = ({ customColumnSizes, open, onCancel, onSubmit, ...props }) => {
  const [selectedAttributes, setSelectedAttributes] = useState([]);

  const onHandleSubmit = () => {
    const attributes = selectedAttributes.map(attribute => ({
      id: attribute.id,
      key: attribute.key,
      name: attribute.key,
      scope: attribute.scope,
      title: attribute.title || attribute.key
    }));
    onSubmit(attributes, customColumnSizes);
  };

  return (
    <Dialog open={open}>
      <DialogTitle>Customize Columns</DialogTitle>
      <Suspense fallback={<Loader show />}>
        <Content selectedAttributes={selectedAttributes} setSelectedAttributes={setSelectedAttributes} {...props} />
      </Suspense>
      <DialogActions className="space-between">
        <Button variant="text" onClick={onCancel}>
          Cancel
        </Button>
        <Button variant="contained" onClick={onHandleSubmit} color="secondary">
          Save
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default ColumnCustomizationDialog;
