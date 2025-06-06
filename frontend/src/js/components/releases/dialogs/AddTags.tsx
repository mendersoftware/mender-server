// Copyright 2024 Northern.tech AS
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
import { FormProvider, useForm } from 'react-hook-form';
import { useDispatch } from 'react-redux';

import { Button, DialogActions, DialogContent } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import ChipSelect from '@northern.tech/common-ui/ChipSelect';
import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import { setReleaseTags, setReleasesListState } from '@northern.tech/store/thunks';

const useStyles = makeStyles()(theme => ({
  DialogContent: {
    padding: 0,
    margin: 30
  },
  DialogActions: {
    padding: `${theme.spacing()} ${theme.spacing(3)}`
  },
  tagSelect: { marginRight: theme.spacing(2), maxWidth: 350 }
}));

export const AddTagsDialog = ({ selectedReleases, onClose }) => {
  const [initialValues] = useState({ tags: [] });
  const [disableSave, setDisableSave] = useState(true);
  const inputName = 'tags';
  const { classes } = useStyles();

  const methods = useForm({ mode: 'onChange', defaultValues: initialValues });
  const { watch, getValues } = methods;
  const watchTagsInput = watch([inputName]);
  const dispatch = useDispatch();

  const addTagsToReleases = () => {
    const tags = getValues(inputName);
    dispatch(setReleasesListState({ loading: true })).then(() => {
      const addRequests = selectedReleases.reduce((accu, release) => {
        accu.push(dispatch(setReleaseTags({ name: release.name, tags: [...new Set([...release.tags, ...tags])] })));
        return accu;
      }, []);
      return Promise.all(addRequests).then(onClose);
    });
  };

  useEffect(() => {
    setDisableSave(!getValues('tags').length);
  }, [getValues, watchTagsInput]);

  return (
    <BaseDialog open title="Add tags to Releases" fullWidth maxWidth="sm" onClose={onClose}>
      <DialogContent className={`${classes.DialogContent}`}>
        <div className="margin-bottom">Add tags to the selected Releases. If a Release already has the tag, it won’t be added again.</div>
        <FormProvider {...methods}>
          <form noValidate>
            <ChipSelect className={classes.tagSelect} label="" name={inputName} placeholder="Enter release tags" />
          </form>
        </FormProvider>
      </DialogContent>
      <DialogActions className={`flexbox space-between margin-top-none ${classes.DialogActions}`}>
        <Button style={{ marginRight: 10 }} onClick={onClose}>
          Cancel
        </Button>
        <Button variant="contained" color="primary" disabled={disableSave} onClick={addTagsToReleases}>
          Add tags
        </Button>
      </DialogActions>
    </BaseDialog>
  );
};

export default AddTagsDialog;
