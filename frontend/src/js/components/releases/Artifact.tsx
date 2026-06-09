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
import { ExpandMore } from '@mui/icons-material';
import { Accordion, AccordionDetails, AccordionSummary, accordionSummaryClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import ArtifactDetails from './ArtifactDetails';

const useStyles = makeStyles()(() => ({
  summary: {
    minHeight: 40,
    padding: '0 8px',
    '&.Mui-expanded': { minHeight: 40 },
    [`& .${accordionSummaryClasses.content}`]: { margin: 0, '&.Mui-expanded': { margin: 0 } }
  }
}));

export const Artifact = ({ artifact, className, columns, expanded, index, onRowSelection, showRemoveArtifactDialog }) => {
  const { classes } = useStyles();
  return (
    <Accordion className={`${className} padding-none`} variant="outlined" expanded={expanded} onChange={onRowSelection}>
      <AccordionSummary className={classes.summary} expandIcon={<ExpandMore />} classes={{ content: 'repo-item' }}>
        {columns.map(({ name, render: Component }) => (
          <Component key={name} artifact={artifact} index={index + 1 + 8} />
        ))}
      </AccordionSummary>
      <AccordionDetails>
        <ArtifactDetails artifact={artifact} open={expanded} showRemoveArtifactDialog={showRemoveArtifactDialog} />
      </AccordionDetails>
    </Accordion>
  );
};
