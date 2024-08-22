// Copyright 2020 Northern.tech AS
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
import React, { useCallback, useState } from 'react';

import { Add as AddIcon } from '@mui/icons-material';
import { Button, FormControl, IconButton, InputLabel, ListSubheader, MenuItem, Select, iconButtonClasses, selectClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { BENEFITS, chartTypes, emptyChartSelection } from '../../../constants/appConstants';
import { toggle } from '../../../helpers';
import Confirm from '../../common/confirm';
import EnterpriseNotification from '../../common/enterpriseNotification';
import { InfoHintContainer } from '../../common/info-hint';
import { HELPTOOLTIPS, MenderHelpTooltip } from '../../helptips/helptooltips';
import { Header } from './distribution';

const fontSize = 'smaller';

const useStyles = makeStyles()(theme => ({
  additionButton: { fontSize: '1rem', cursor: 'pointer' },
  button: { marginLeft: theme.spacing(2), padding: '6px 8px', fontSize },
  buttonWrapper: { display: 'flex', justifyContent: 'flex-end', alignContent: 'center' },
  iconButton: {
    [`&.${iconButtonClasses.root}`]: {
      borderRadius: 5,
      border: `1px solid ${theme.palette.primary.main}`,
      marginRight: theme.spacing(),
      '&.selected': {
        background: theme.palette.primary.main,
        color: theme.palette.background.paper
      }
    }
  },
  formWrapper: {
    alignItems: 'baseline',
    columnGap: theme.spacing(3),
    display: 'grid',
    fontSize,
    gridTemplateColumns: 'max-content 1fr',
    gridTemplateRows: 'auto',
    rowGap: theme.spacing(0.5),
    marginTop: theme.spacing(),
    [`.${selectClasses.select}`]: { paddingBottom: theme.spacing(0.5), paddingTop: 0, fontSize }
  }
}));

const GroupSelect = ({ groups, selection, setSelection }) => (
  <FormControl className="margin-top-none">
    <InputLabel id="group-select-label">Device group</InputLabel>
    <Select labelId="group-select-label" value={selection.group || true} onChange={e => setSelection({ group: e.target.value })}>
      <MenuItem value={true}>
        <em>All Devices</em>
      </MenuItem>
      {Object.keys(groups).map(group => (
        <MenuItem key={group} value={group}>
          {group}
        </MenuItem>
      ))}
    </Select>
  </FormControl>
);

const themeSpacing = 8;
const basePadding = 2 * themeSpacing;
const getIndentation = level => ({ paddingLeft: basePadding + level * themeSpacing });

const SoftwareSelect = ({ selection, setSelection, software }) => (
  <FormControl className="margin-top-none">
    <InputLabel id="software-select-label">Software</InputLabel>
    <Select labelId="software-select-label" value={selection.software} onBlur={undefined} onChange={e => setSelection({ software: e.target.value })}>
      {software.map(({ subheader, title, value, nestingLevel }) =>
        subheader ? (
          <ListSubheader key={value} style={getIndentation(nestingLevel)}>
            {subheader}
          </ListSubheader>
        ) : (
          <MenuItem key={value} style={getIndentation(nestingLevel)} value={value}>
            {title}
          </MenuItem>
        )
      )}
    </Select>
  </FormControl>
);

const ChartSelect = ({ classes, selection, setSelection }) => (
  <div>
    {Object.values(chartTypes).map(type => {
      const { Icon, key } = type;
      return (
        <IconButton
          className={`${classes.iconButton} ${selection.chartType === key ? 'selected' : ''}`}
          key={key}
          size="small"
          onClick={() => setSelection({ chartType: key })}
        >
          <Icon fontSize="small" />
        </IconButton>
      );
    })}
  </div>
);

const chartOptions = [
  { key: 'software', title: 'Software', Selector: SoftwareSelect },
  { key: 'group', title: 'Device group', Selector: GroupSelect },
  { key: 'type', title: 'Display', Selector: ChartSelect }
];

export const ChartEditWidget = ({ groups, onSave, onCancel, selection: selectionProp = {}, software = [] }) => {
  const [selection, setSelection] = useState({ ...emptyChartSelection, ...selectionProp });
  const { classes } = useStyles();

  const addCurrentSelection = useCallback(
    () => onSave({ ...emptyChartSelection, ...selection, group: typeof selection.group === 'string' ? selection.group : null }),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [JSON.stringify(selection), onSave]
  );

  const onSelectionChange = changedSelection => setSelection(current => ({ ...current, ...changedSelection }));

  return (
    <div className="widget chart-widget">
      <Header chartType={selection.chartType} />
      <div className={classes.formWrapper}>
        {chartOptions.map(({ key, title, Selector }) => (
          <React.Fragment key={key}>
            <div>{title}</div>
            <Selector classes={classes} groups={groups} software={software} selection={selection} setSelection={onSelectionChange} />
          </React.Fragment>
        ))}
      </div>
      <div className={classes.buttonWrapper}>
        <Button className={classes.button} size="small" onClick={onCancel}>
          Cancel
        </Button>
        <Button className={classes.button} size="small" onClick={addCurrentSelection} variant="contained" disabled={!selection}>
          Save
        </Button>
      </div>
    </div>
  );
};

export const RemovalWidget = ({ onCancel, onClick }) => (
  <div className="widget chart-widget">
    <Confirm classes="flexbox centered confirmation-overlay" cancel={onCancel} action={onClick} style={{ justifyContent: 'center' }} type="chartRemoval" />
  </div>
);

export const WidgetAdditionWidget = ({ onAdditionClick, ...remainder }) => {
  const [adding, setAdding] = useState(false);
  const { classes } = useStyles();

  const addCurrentSelection = selection => {
    onAdditionClick(selection);
    setAdding(false);
  };

  const onCancelClick = () => setAdding(toggle);

  return adding ? (
    <ChartEditWidget {...remainder} onSave={addCurrentSelection} onCancel={onCancelClick} />
  ) : (
    <div className="widget">
      <InfoHintContainer className="" style={{ alignItems: 'end' }}>
        <EnterpriseNotification id={BENEFITS.dashboard.id} />
        <MenderHelpTooltip id={HELPTOOLTIPS.dashboardWidget.id} />
      </InfoHintContainer>
      <div className={`flexbox centered muted ${classes.additionButton}`} onClick={() => setAdding(true)}>
        <AddIcon />
        <span className={classes.additionButton}>add a widget</span>
      </div>
    </div>
  );
};

export default WidgetAdditionWidget;
