// Copyright 2023 Northern.tech AS
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
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { ArrowDropDown as ArrowDropDownIcon, ArrowDropUp as ArrowDropUpIcon } from '@mui/icons-material';
// material ui
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  IconButton,
  InputLabel,
  LinearProgress,
  List,
  ListItem,
  ListItemText,
  ListSubheader,
  accordionClasses,
  listSubheaderClasses
} from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { getReportingLimits } from '../../actions/deviceActions';
import { toggle } from '../../helpers';
import { InfoHintContainer } from '../common/info-hint';
import { HELPTOOLTIPS, MenderHelpTooltip } from '../helptips/helptooltips';

const useStyles = makeStyles()(theme => ({
  accordion: {
    ul: {
      paddingInlineStart: 0
    },
    [`&.${accordionClasses.disabled}, &.${accordionClasses.expanded}`]: {
      backgroundColor: theme.palette.background.paper
    }
  },
  attributesList: {
    overflow: 'auto',
    maxHeight: 250,
    background: 'white',
    width: '100%',
    position: 'relative',
    [`.${listSubheaderClasses.root}`]: {
      top: -10
    },
    'li > ul': {
      overflow: 'initial'
    }
  },
  limitBar: { backgroundColor: theme.palette.grey[500], margin: '15px 0' },
  summary: { padding: 0, marginBottom: theme.spacing() }
}));

export const ReportingLimits = () => {
  const [open, setOpen] = useState(false);
  const { classes } = useStyles();
  const dispatch = useDispatch();
  const { attributes = {}, count = 0, limit = 100 } = useSelector(state => state.devices.filteringAttributesConfig);

  useEffect(() => {
    dispatch(getReportingLimits());
  }, [dispatch]);

  const toggleOpen = () => setOpen(toggle);

  return (
    <>
      <InputLabel className="margin-top-small padding-top-small flexbox" shrink id="filterable-attributes-usage-and-limit">
        Filterable attributes usage & limit ({count}/{limit})
        <InfoHintContainer>
          <MenderHelpTooltip id={HELPTOOLTIPS.attributeLimit.id} placement="top" />
        </InfoHintContainer>
      </InputLabel>
      <Accordion className={classes.accordion} square expanded={open} onChange={toggleOpen} disabled={!count}>
        <AccordionSummary className={classes.summary}>
          <LinearProgress className={classes.limitBar} variant="determinate" value={(count * 100) / limit} style={{ width: '100%' }} />
          <IconButton className="margin-left-small expandButton" size="large">
            {open ? <ArrowDropUpIcon /> : <ArrowDropDownIcon />}
          </IconButton>
        </AccordionSummary>
        <AccordionDetails>
          <List className={classes.attributesList}>
            {Object.entries(attributes).map(([scope, values = []]) => (
              <li key={scope}>
                <ul>
                  <ListSubheader>{scope}</ListSubheader>
                  {values.map(item => (
                    <ListItem key={`item-${scope}-${item}`}>
                      <ListItemText primary={item} />
                    </ListItem>
                  ))}
                </ul>
              </li>
            ))}
          </List>
        </AccordionDetails>
      </Accordion>
    </>
  );
};

export default ReportingLimits;
