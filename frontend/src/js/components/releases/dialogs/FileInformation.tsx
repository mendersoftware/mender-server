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
import { Delete as DeleteIcon, FileCopy as FileIcon } from '@mui/icons-material';
import { IconButton, ListItem, ListItemAvatar, ListItemText, Typography } from '@mui/material';

import FileSize from '@northern.tech/common-ui/FileSize';
import { InfoHintContainer } from '@northern.tech/common-ui/InfoHint';

import { HELPTOOLTIPS } from '../../helptips/HelpTooltips';
import { MenderHelpTooltip } from '../../helptips/MenderTooltip';

const fileInformationContent = {
  mender: {
    title: 'Mender Artifact',
    icon: FileIcon,
    infoId: 'menderArtifactUpload'
  },
  singleFile: {
    title: 'single file',
    icon: FileIcon,
    infoId: 'singleFileUpload'
  }
};

export const FileInformation = ({ file, type, onRemove }) => {
  if (!file) {
    return <div />;
  }
  const { icon: Icon, infoId, title } = fileInformationContent[type];
  return (
    <div>
      <div className="flexbox align-items-center margin-bottom-x-small">
        <Typography variant="subtitle1">Selected {title}</Typography>
        <InfoHintContainer>
          <MenderHelpTooltip id={HELPTOOLTIPS[infoId].id} placement="bottom-start" small />
        </InfoHintContainer>
      </div>
      <ListItem
        secondaryAction={
          <IconButton edge="end" onClick={onRemove}>
            <DeleteIcon />
          </IconButton>
        }
      >
        <ListItemAvatar className="align-self-center">
          <Icon color="action" />
        </ListItemAvatar>
        <ListItemText primary={file.name} secondary={<FileSize fileSize={file.size} />} />
      </ListItem>
    </div>
  );
};
