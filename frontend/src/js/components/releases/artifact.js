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
import React from 'react';

// material ui
import { ArrowDropDown as ArrowDropDownIcon, ArrowDropUp as ArrowDropUpIcon } from '@mui/icons-material';
import { Accordion, AccordionDetails, AccordionSummary, IconButton } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import ArtifactDetails from './artifactdetails';

const useStyles = makeStyles()(theme => ({
  accordion: {
    border: '1px solid',
    borderColor: theme.palette.grey[500],
    width: '100%'
  }
}));

export const Artifact = ({ artifact, columns, expanded, index, onRowSelection, showRemoveArtifactDialog }) => {
  const { classes } = useStyles();

  return (
    <div className="release-repo-item flexbox">
      <div className="muted">{index + 1}</div>
      <Accordion className={classes.accordion} square expanded={expanded} onChange={onRowSelection}>
        <AccordionSummary style={{ padding: '0 12px' }} classes={{ content: 'repo-item' }}>
          {columns.map(({ name, render: Component }) => (
            <Component key={name} artifact={artifact} />
          ))}
          <IconButton className="expandButton" size="large">
            {expanded ? <ArrowDropUpIcon /> : <ArrowDropDownIcon />}
          </IconButton>
        </AccordionSummary>
        <AccordionDetails>
          <ArtifactDetails artifact={artifact} open={expanded} showRemoveArtifactDialog={showRemoveArtifactDialog} />
        </AccordionDetails>
      </Accordion>
    </div>
  );
};

export default Artifact;
