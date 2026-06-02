// Copyright 2025 Northern.tech AS
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
import { Chip, Paper, Typography, alpha } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import TextOverflowMultiline from '@northern.tech/common-ui/TextOverflowMultiline';
import Time from '@northern.tech/common-ui/Time';
import type { Software } from '@northern.tech/types/MenderTypes';

interface SoftwareItemProps {
  onClick: (item: Software) => void;
  selected: boolean;
  software: Software;
}
const useStyles = makeStyles()(theme => ({
  hoverPaper: {
    padding: theme.spacing(1),
    transition: 'background-color 0.3s ease',
    '&:hover': {
      backgroundColor: alpha(theme.palette.primary.light, 0.08),
      cursor: 'pointer'
    }
  },
  selectedPaper: {
    backgroundColor: alpha(theme.palette.primary.light, 0.08),
    border: `1px solid ${theme.palette.primary.main}`
  }
}));

export const SoftwareItem = (props: SoftwareItemProps) => {
  const { classes } = useStyles();
  const { software, onClick, selected } = props;
  const { name, tags, modified, notes, kind } = software;
  const kindLabel = kind === 'manifest' ? 'Manifest' : 'Release';
  return (
    <Paper
      elevation={0}
      variant="outlined"
      className={`margin-top-small padding-small ${classes.hoverPaper} ${selected ? classes.selectedPaper : ''}`}
      onClick={() => onClick(software)}
    >
      <TextOverflowMultiline variant="subtitle2" className="margin-bottom-x-small" lines={2}>
        {name}
      </TextOverflowMultiline>
      {notes && (
        <TextOverflowMultiline variant="body2" className="margin-bottom-x-small" lines={2}>
          {notes}
        </TextOverflowMultiline>
      )}
      <div className="margin-top-x-small">
        <Chip className="margin-right-x-small" color={kind === 'manifest' ? 'secondary' : 'primary'} label={kindLabel} size="small" variant="outlined" />
        {tags?.map(tag => (
          <Chip className="margin-right-x-small" key={tag} label={tag} size="small" />
        ))}
      </div>
      <Typography variant="caption">Updated {<Time value={modified} />}</Typography>
    </Paper>
  );
};
