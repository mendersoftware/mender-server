// Copyright 2019 Northern.tech AS
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
// material ui
import { ExpandLess, ExpandMore } from '@mui/icons-material';
import { Accordion, AccordionDetails, AccordionSummary } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import ArtifactDetails from './ArtifactDetails';

const useStyles = makeStyles()(theme => ({
  accordion: {
    border: '1px solid',
    borderColor: theme.palette.grey[500],
    width: '100%'
  },
  index: {
    color: theme.palette.grey[500],
    marginTop: theme.spacing(2.5)
  }
}));

export const Artifact = ({ artifact, className, columns, expanded, index, onRowSelection, showRemoveArtifactDialog }) => {
  const { classes } = useStyles();

  return (
    <div className="flexbox">
      <div className={`${classes.index} margin-right-small`}>{index + 1}</div>
      <Accordion className={`${classes.accordion} ${className}`} square expanded={expanded} onChange={onRowSelection}>
        <AccordionSummary classes={{ content: 'repo-item' }}>
          {columns.map(({ name, render: Component }) => (
            <Component key={name} artifact={artifact} />
          ))}
          {expanded ? <ExpandLess className="expandButton" /> : <ExpandMore className="expandButton" />}
        </AccordionSummary>
        <AccordionDetails>
          <ArtifactDetails artifact={artifact} open={expanded} showRemoveArtifactDialog={showRemoveArtifactDialog} />
        </AccordionDetails>
      </Accordion>
    </div>
  );
};

export default Artifact;
