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
import { Delete as DeleteIcon, InsertDriveFile as InsertDriveFileIcon } from '@mui/icons-material';
import { Divider, IconButton } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import FileSize from '@northern.tech/common-ui/FileSize';

import { HELPTOOLTIPS } from '../../helptips/HelpTooltips';
import { MenderHelpTooltip } from '../../helptips/MenderTooltip';

const useStyles = makeStyles()(theme => ({
  fileInfo: {
    alignItems: 'center',
    columnGap: theme.spacing(4),
    display: 'grid',
    gridTemplateColumns: 'max-content 1fr max-content max-content',
    marginBottom: theme.spacing(2),
    marginRight: theme.spacing(4)
  },
  fileSizeWrapper: { marginTop: 5 }
}));

const fileInformationContent = {
  mender: {
    title: 'Mender Artifact',
    icon: InsertDriveFileIcon,
    infoId: 'menderArtifactUpload'
  },
  singleFile: {
    title: 'Single File',
    icon: InsertDriveFileIcon,
    infoId: 'singleFileUpload'
  }
};

export const FileInformation = ({ file, type, onRemove }) => {
  const { classes } = useStyles();
  if (!file) {
    return <div />;
  }
  const { icon: Icon, infoId, title } = fileInformationContent[type];
  return (
    <>
      <h4>Selected {title}</h4>
      <div className={classes.fileInfo}>
        <Icon size="large" />
        <div className="flexbox column">
          <div>{file.name}</div>
          <div className={`muted ${classes.fileSizeWrapper}`}>
            <FileSize fileSize={file.size} />
          </div>
        </div>
        <IconButton size="large" onClick={onRemove}>
          <DeleteIcon />
        </IconButton>
        <MenderHelpTooltip id={HELPTOOLTIPS[infoId].id} />
      </div>
      <Divider className="margin-right-large" />
    </>
  );
};
