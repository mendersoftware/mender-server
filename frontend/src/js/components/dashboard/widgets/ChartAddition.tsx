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
import React, { useState } from 'react';
import { Controller, useFormContext } from 'react-hook-form';

import { Add as AddIcon } from '@mui/icons-material';
import {
  FormControl,
  IconButton,
  InputLabel,
  ListSubheader,
  MenuItem,
  Select,
  iconButtonClasses,
  selectClasses,
  svgIconClasses,
  useTheme
} from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import Confirm from '@northern.tech/common-ui/Confirm';
import EnterpriseNotification from '@northern.tech/common-ui/EnterpriseNotification';
import { InfoHintContainer } from '@northern.tech/common-ui/InfoHint';
import Form from '@northern.tech/common-ui/forms/Form';
import { BENEFITS, chartTypes, emptyChartSelection } from '@northern.tech/store/constants';
import { toggle } from '@northern.tech/utils/helpers';

import { HELPTOOLTIPS } from '../../helptips/HelpTooltips';
import { MenderHelpTooltip } from '../../helptips/MenderTooltip';

const fontSize = 'smaller';

const useStyles = makeStyles()(theme => ({
  additionButton: { fontSize: '1rem', cursor: 'pointer' },
  button: { marginLeft: theme.spacing(2), padding: '6px 8px', fontSize },
  buttonWrapper: { display: 'flex', justifyContent: 'flex-end', alignContent: 'center' },
  header: { minHeight: 30, [`.${svgIconClasses.root}`]: { marginLeft: theme.spacing() } },
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

export const Header = ({ chartType }) => {
  const { classes } = useStyles();
  const { Icon } = chartTypes[chartType];
  return (
    <div className={`flexbox align-items-center ${classes.header}`}>
      Software distribution
      <Icon />
    </div>
  );
};

const GroupSelect = ({ groups, name }) => {
  const { control } = useFormContext();
  return (
    <FormControl className="margin-top-none">
      <InputLabel id="group-select-label" shrink>
        Device group
      </InputLabel>
      <Controller
        name={name}
        control={control}
        defaultValue=""
        render={({ field }) => (
          <Select labelId="group-select-label" displayEmpty label="Device group" {...field}>
            <MenuItem value="">
              <em>All Devices</em>
            </MenuItem>
            {Object.keys(groups).map(group => (
              <MenuItem key={group} value={group}>
                {group}
              </MenuItem>
            ))}
          </Select>
        )}
      />
    </FormControl>
  );
};

const getIndentation = (level, theme) => ({ paddingLeft: theme.spacing(2) + level * theme.spacing() });

const SoftwareSelect = ({ software, name }) => {
  const { control } = useFormContext();
  const theme = useTheme();
  return (
    <FormControl className="margin-top-none">
      <InputLabel shrink id="software-select-label">
        Software
      </InputLabel>
      <Controller
        name={name}
        control={control}
        defaultValue=""
        render={({ field }) => (
          <Select labelId="software-select-label" label="Software" displayEmpty {...field}>
            {software.map(({ subheader, title, value, nestingLevel }) =>
              subheader ? (
                <ListSubheader key={value} style={getIndentation(nestingLevel, theme)}>
                  {subheader}
                </ListSubheader>
              ) : (
                <MenuItem key={value} style={getIndentation(nestingLevel, theme)} value={value}>
                  {title}
                </MenuItem>
              )
            )}
          </Select>
        )}
      />
    </FormControl>
  );
};

const ChartSelect = ({ classes, name }) => {
  const { control, watch } = useFormContext();
  const selectedType = watch(name);
  return (
    <Controller
      name={name}
      control={control}
      defaultValue={Object.keys(chartTypes)[0]}
      render={({ field: { onChange } }) => (
        <div>
          {Object.values(chartTypes).map(type => {
            const { Icon, key } = type;
            return (
              <IconButton className={`${classes.iconButton} ${selectedType === key ? 'selected' : ''}`} key={key} size="small" onClick={() => onChange(key)}>
                <Icon fontSize="small" />
              </IconButton>
            );
          })}
        </div>
      )}
    />
  );
};

const chartOptions = [
  { key: 'software', title: 'Software', Selector: SoftwareSelect },
  { key: 'group', title: 'Device group', Selector: GroupSelect },
  { key: 'chartType', title: 'Display', Selector: ChartSelect }
];

export const ChartEditWidget = ({ groups, onSave, onCancel, selection: selectionProp = {}, software = [] }) => {
  const { classes } = useStyles();

  return (
    <Form
      defaultValues={emptyChartSelection}
      initialValues={selectionProp}
      onSubmit={onSave}
      handleCancel={onCancel}
      showButtons
      submitLabel="Save"
      className="widget chart-widget"
    >
      <Header chartType={emptyChartSelection.chartType} />
      <div className={classes.formWrapper}>
        {chartOptions.map(({ key, title, Selector }) => (
          <React.Fragment key={key}>
            <div>{title}</div>
            <Selector classes={classes} groups={groups} software={software} name={key} />
          </React.Fragment>
        ))}
      </div>
    </Form>
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
