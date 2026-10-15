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
import { useCallback, useEffect, useState } from 'react';

// material ui
import { CheckCircle, Edit } from '@mui/icons-material';
import { Button, List, ListItem, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { TwoColumnData } from '@northern.tech/common-ui/TwoColumnData';
import storeActions from '@northern.tech/store/actions';
import { SSO_TYPES } from '@northern.tech/store/constants';
import { getSsoEndpoint } from '@northern.tech/store/selectors';
import { useAppDispatch, useAppSelector } from '@northern.tech/store/store';
import { getSsoEndpoints } from '@northern.tech/store/thunks';

import SSOEditor from './SSOEditor';

const { setSnackbar } = storeActions;

const getEndpointDetails = (type, ssoEndpoint) =>
  Object.fromEntries(SSO_TYPES[type].configDetails.map(({ key, label }) => [label, ssoEndpoint?.[type]?.[key]]));

const useStyles = makeStyles()(theme => ({
  list: {
    listStyleType: 'decimal',
    paddingLeft: theme.spacing(4)
  },
  listItem: {
    display: 'list-item'
  },
  setup: {
    fontWeight: 600
  }
}));

export const SSOConfig = ({ ssoItem, config, onSave, readOnly = false }) => {
  const [configDetails, setConfigDetails] = useState({});
  const [fileContent, setFileContentState] = useState('');
  const [isEditing, setIsEditing] = useState(false);
  const { id, type, ...content } = config || {};
  const { type: selectedType } = ssoItem;
  const configContent = content.config || '';
  const ssoEndpoint = useAppSelector(getSsoEndpoint);
  const dispatch = useAppDispatch();
  const dispatchedSetSnackbar = useCallback((...args) => dispatch(setSnackbar(...args)), [dispatch]);
  // file content should be text, otherwise editor will fail
  const setFileContent = content => setFileContentState(typeof content === 'object' ? JSON.stringify(content) : content);

  const { classes } = useStyles();

  useEffect(() => {
    dispatch(getSsoEndpoints());
  }, [dispatch]);

  useEffect(() => {
    setFileContent(configContent);
    if (config?.id) {
      setConfigDetails({
        Type: SSO_TYPES[type].title,
        ...getEndpointDetails(type, ssoEndpoint),
        Metadata: (
          <Typography variant="body2" className="flexbox align-items-center gap-x-small green">
            Uploaded
            <CheckCircle fontSize="small" />
          </Typography>
        )
      });
    }
  }, [config, configContent, ssoEndpoint, type]);

  const onSaveSSOSettings = async () => {
    await onSave(id, fileContent);
    setIsEditing(false);
  };
  return (
    <>
      {readOnly ? (
        <TwoColumnData className="margin-top-small" data={configDetails} setSnackbar={dispatchedSetSnackbar} />
      ) : (
        <>
          <Typography className={classes.setup}>Setup:</Typography>
          <List component="ol" className={classes.list}>
            <ListItem className={classes.listItem}>
              <Typography className="margin-bottom-small">Copy the details to your identity provider:</Typography>
              <TwoColumnData data={getEndpointDetails(selectedType, ssoEndpoint)} setSnackbar={dispatchedSetSnackbar} />
            </ListItem>
            <ListItem className={classes.listItem}>Provide the connection metadata from your identity provider:</ListItem>
          </List>
          <div>
            <Button onClick={() => setIsEditing(true)} endIcon={<Edit />}>
              Upload or edit metadata
            </Button>
          </div>
        </>
      )}
      <SSOEditor
        open={isEditing}
        ssoItem={ssoItem}
        config={configContent}
        fileContent={fileContent}
        onClose={() => setIsEditing(false)}
        onSave={onSaveSSOSettings}
        setFileContent={setFileContent}
      />
    </>
  );
};

export default SSOConfig;
