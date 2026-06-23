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
import { useFormContext } from 'react-hook-form';
import { useDispatch } from 'react-redux';

import { Button, DialogActions, DialogContent, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import ChipSelect from '@northern.tech/common-ui/ChipSelect';
import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import Form from '@northern.tech/common-ui/forms/Form';
import { setReleaseTags, setReleasesListState } from '@northern.tech/store/thunks';

const useStyles = makeStyles()(theme => ({
  tagSelect: { marginRight: theme.spacing(2), maxWidth: 350 }
}));

const AddTagsDialogContent = ({ onClose }) => {
  const { watch } = useFormContext();
  const inputName = 'tags';
  const tags = watch(inputName);
  const { classes } = useStyles();

  return (
    <>
      <DialogContent>
        <Typography className="margin-bottom">Add tags to the selected Releases. If a Release already has the tag, it won’t be added again.</Typography>
        <ChipSelect className={classes.tagSelect} label="" name={inputName} placeholder="Add release tags" />
      </DialogContent>
      <DialogActions>
        <Button style={{ marginRight: 10 }} onClick={onClose}>
          Cancel
        </Button>
        <Button variant="contained" disabled={!tags?.length} type="submit">
          Add tags
        </Button>
      </DialogActions>
    </>
  );
};

export const AddTagsDialog = ({ selectedReleases, onClose }) => {
  const dispatch = useDispatch();

  const onSubmit = ({ tags }) => {
    dispatch(setReleasesListState({ loading: true })).then(() => {
      const addRequests = selectedReleases.reduce((accu, release) => {
        accu.push(dispatch(setReleaseTags({ name: release.name, tags: [...new Set([...release.tags, ...tags])] })));
        return accu;
      }, []);
      return Promise.all(addRequests).then(onClose);
    });
  };

  return (
    <BaseDialog open title="Add tags to Releases" fullWidth maxWidth="sm" onClose={onClose}>
      <Form onSubmit={onSubmit} defaultValues={{ tags: [] }}>
        <AddTagsDialogContent onClose={onClose} />
      </Form>
    </BaseDialog>
  );
};

export default AddTagsDialog;
