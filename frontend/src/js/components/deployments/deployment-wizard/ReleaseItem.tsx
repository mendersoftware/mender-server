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

import Time from '@northern.tech/common-ui/Time';
import { Release } from '@northern.tech/store/releasesSlice';

interface ReleaseItemProps {
  onClick: (rel: Release) => void;
  release: Release;
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
  linesOverflow: {
    display: '-webkit-box',
    '-webkit-box-orient': 'vertical',
    '-webkit-line-clamp': '2',
    overflow: 'hidden'
  }
}));

export const ReleaseItem = (props: ReleaseItemProps) => {
  const { classes } = useStyles();
  const { release, onClick } = props;
  const { name, artifacts, tags, modified } = release;
  const description = artifacts[0]?.description;
  return (
    <Paper elevation={0} variant="outlined" className={`margin-top-small padding-small ${classes.hoverPaper}`} onClick={() => onClick(release)}>
      <Typography variant="subtitle2" className={`margin-bottom-x-small ${classes.linesOverflow}`}>
        {name}
      </Typography>
      {description && (
        <Typography variant="body2" className={`margin-bottom-x-small ${classes.linesOverflow}`}>
          {description}
        </Typography>
      )}
      <div className="margin-top-x-small">
        {tags?.map(tag => (
          <Chip className="margin-right-x-small" key={tag} label={tag} size="small" />
        ))}
      </div>
      <Typography variant="caption">Updated {<Time value={modified} />}</Typography>
    </Paper>
  );
};
