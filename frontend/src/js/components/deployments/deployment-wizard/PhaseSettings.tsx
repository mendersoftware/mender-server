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
import React, { useCallback, useState } from 'react';

import { Add as AddIcon, Cancel as CancelIcon } from '@mui/icons-material';
import {
  Checkbox,
  Chip,
  Collapse,
  FormControl,
  FormControlLabel,
  IconButton,
  InputAdornment,
  ListSubheader,
  MenuItem,
  OutlinedInput,
  Select,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow
} from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { DOCSTIPS, DocsTooltip } from '@northern.tech/common-ui/DocsLink';
import EnterpriseNotification from '@northern.tech/common-ui/EnterpriseNotification';
import { InfoHintContainer } from '@northern.tech/common-ui/InfoHint';
import Time from '@northern.tech/common-ui/Time';
import { BENEFITS } from '@northern.tech/store/constants';
import dayjs from 'dayjs';
import pluralize from 'pluralize';

// use this to get remaining percent of final phase so we don't set a hard number
export const getRemainderPercent = phases =>
  phases.reduce((accu, phase, index, source) => {
    // ignore final phase size if set
    if (index === source.length - 1) {
      return accu;
    }
    return phase.batch_size ? accu - phase.batch_size : accu;
  }, 100);

export const validatePhases = (phases, deploymentDeviceCount) => {
  if (!phases?.length) {
    return true;
  }
  const remainder = getRemainderPercent(phases);
  const { isValid } = phases.reduce(
    (accu, { batch_size = 0 }) => {
      if (!accu.isValid) {
        return accu;
      }
      const deviceCount = Math.floor((deploymentDeviceCount / 100) * (batch_size || remainder));
      const totalSize = accu.totalSize + batch_size;
      return { isValid: deviceCount >= 1 && totalSize <= 100, totalSize };
    },
    { isValid: true, totalSize: 0 }
  );
  return isValid;
};

export const getPhaseDeviceCount = (numberDevices = 1, batchSize, remainder, isLastPhase) =>
  isLastPhase ? Math.ceil((numberDevices / 100) * (batchSize || remainder)) : Math.floor((numberDevices / 100) * (batchSize || remainder));

const useStyles = makeStyles()(theme => ({
  chip: { marginTop: theme.spacing(2) },
  delayInputWrapper: { display: 'grid', gridTemplateColumns: 'max-content max-content', columnGap: theme.spacing() },
  row: { whiteSpace: 'nowrap' },
  input: { minWidth: 400 },
  patternSelection: { marginTop: theme.spacing(2), maxWidth: 515, width: 'min-content' }
}));

const timeframes = ['minutes', 'hours', 'days'];
const tableHeaders = ['', 'Batch size', 'Phase begins', 'Delay before next phase', ''];

export const getPhaseStartTime = (phases, index, startDate) => {
  if (index < 1) {
    return startDate?.toISOString ? startDate.toISOString() : startDate;
  } else if (phases[index].start_ts) {
    // if displaying an ongoing deployment we can rely on the timing info from the backend
    return phases[index].start_ts;
  }
  // since we don't want to get stale phase start times when the creation dialog is open for a long time
  // we have to ensure start times are based on delay from previous phases
  // since there likely won't be 1000s of phases this should still be fine to recalculate
  const newStartTime = phases.slice(0, index).reduce((accu, phase) => dayjs(accu).add(phase.delay, phase.delayUnit), startDate);
  return newStartTime.toISOString();
};

export const PhaseSettings = ({ classNames, deploymentObject, disabled, numberDevices, setDeploymentSettings }) => {
  const { classes } = useStyles();

  const { filter, phases = [] } = deploymentObject;
  const updateDelay = (value, index) => {
    let newPhases = [...phases];
    // value must be at least 1
    value = Math.max(1, value);
    newPhases[index] = { ...newPhases[index], delay: value };

    setDeploymentSettings({ phases: newPhases });
    // logic for updating time stamps should be in parent - only change delays here
  };

  const updateBatchSize = (value, index) => {
    let newPhases = [...phases];
    value = Math.min(100, Math.max(1, value));
    newPhases[index] = {
      ...newPhases[index],
      batch_size: value
    };
    // When phase's batch size changes, check for new 'remainder'
    const remainder = getRemainderPercent(newPhases);
    // if new remainder will be 0 or negative remove phase leave last phase to get remainder
    if (remainder < 1) {
      newPhases.pop();
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { batch_size, ...newFinalPhase } = newPhases[newPhases.length - 1];
      newPhases[newPhases.length - 1] = newFinalPhase;
    }
    setDeploymentSettings({ phases: newPhases });
  };

  const addPhase = () => {
    let newPhases = [...phases];
    // assign new batch size to *previous* last batch
    const remainder = getRemainderPercent(newPhases);
    newPhases[newPhases.length - 1] = {
      ...newPhases[newPhases.length - 1],
      // make it default 10, unless remainder is <=10 in which case make it half remainder
      batch_size: remainder > 10 ? 10 : Math.floor(remainder / 2),
      // check for previous phase delay or set 2hr default
      delay: newPhases[newPhases.length - 1].delay || 2,
      delayUnit: newPhases[newPhases.length - 1].delayUnit || 'hours'
    };
    newPhases.push({});
    // use function to set new phases incl start time of new phase
    setDeploymentSettings({ phases: newPhases });
  };

  const removePhase = index => {
    let newPhases = [...phases];
    newPhases.splice(index, 1);
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    let { batch_size, delay, ...newPhase } = newPhases[newPhases.length - 1]; // remove batch size from new last phase, use remainder
    if (newPhases.length > 1) {
      newPhase.delay = delay;
    }
    newPhases[newPhases.length - 1] = newPhase;
    setDeploymentSettings({ phases: newPhases });
  };

  const handleDelayToggle = (value, index) => {
    let newPhases = [...phases];
    newPhases[index] = {
      ...newPhases[index],
      delayUnit: value
    };
    setDeploymentSettings({ phases: newPhases });
  };

  const remainder = getRemainderPercent(phases);

  // disable 'add phase' button if last phase/remainder has only 1 device left
  const disableAdd = !filter && (remainder / 100) * numberDevices <= 1;
  const startTime = phases.length ? phases[0].start_ts || new Date() : new Date();
  const mappedPhases = phases.map((phase, index) => {
    let max = index > 0 ? 100 - phases[index - 1].batch_size : 100;
    const deviceCount = getPhaseDeviceCount(numberDevices, phase.batch_size, remainder, index === phases.length - 1);
    const isEmptyPhase = deviceCount < 1;
    return (
      <TableRow className={classes.row} key={index}>
        <TableCell component="td" scope="row">
          <Chip size="small" label={`Phase ${index + 1}`} />
        </TableCell>
        <TableCell>
          <div className="flexbox center-aligned">
            {phase.batch_size && phase.batch_size < 100 ? (
              <OutlinedInput
                value={phase.batch_size}
                onChange={event => updateBatchSize(event.target.value, index)}
                endAdornment={
                  <InputAdornment className={isEmptyPhase ? 'warning' : ''} position="end">
                    %
                  </InputAdornment>
                }
                disabled={disabled && deviceCount >= 1}
                inputProps={{
                  step: 1,
                  min: 1,
                  max: max,
                  type: 'number'
                }}
              />
            ) : (
              phase.batch_size || remainder
            )}
            <span className={isEmptyPhase ? 'warning info' : 'info'} style={{ marginLeft: '5px' }}>{`(${deviceCount} ${pluralize(
              'device',
              deviceCount
            )})`}</span>
          </div>
          {isEmptyPhase && <div className="warning">Phases must have at least 1 device</div>}
        </TableCell>
        <TableCell>
          <Time value={getPhaseStartTime(phases, index, startTime)} />
        </TableCell>
        <TableCell>
          {phase.delay && index !== phases.length - 1 ? (
            <div className={classes.delayInputWrapper}>
              <OutlinedInput
                value={phase.delay}
                onChange={event => updateDelay(event.target.value, index)}
                inputProps={{ step: 1, min: 1, max: 720, type: 'number' }}
              />
              <Select onChange={event => handleDelayToggle(event.target.value, index)} value={phase.delayUnit || 'hours'}>
                {timeframes.map(value => (
                  <MenuItem key={value} value={value}>
                    <div className="capitalized-start">{value}</div>
                  </MenuItem>
                ))}
              </Select>
            </div>
          ) : (
            '-'
          )}
        </TableCell>
        <TableCell>
          {index >= 1 ? (
            <IconButton onClick={() => removePhase(index)} size="large">
              <CancelIcon />
            </IconButton>
          ) : null}
        </TableCell>
      </TableRow>
    );
  });

  return (
    <div className={classNames}>
      <Table size="small">
        <TableHead>
          <TableRow>
            {tableHeaders.map((content, index) => (
              <TableCell key={index}>{content}</TableCell>
            ))}
          </TableRow>
        </TableHead>
        <TableBody>{mappedPhases}</TableBody>
      </Table>

      {!disableAdd ? <Chip className={classes.chip} color="primary" clickable={!disableAdd} icon={<AddIcon />} label="Add a phase" onClick={addPhase} /> : null}
    </div>
  );
};

export const RolloutPatternSelection = props => {
  const { setDeploymentSettings, deploymentObject = {}, disableSchedule, isEnterprise, open = false, previousPhases = [] } = props;
  const { deploymentDeviceCount = 0, deploymentDeviceIds = [], filter, phases = [] } = deploymentObject;

  const [usesPattern, setUsesPattern] = useState(open || phases.some(i => i));
  const { classes } = useStyles();

  const handlePatternChange = ({ target: { value } }) => {
    let updatedPhases = [];
    // check if a start time already exists from props and if so, use it
    const phaseStart = phases.length ? { start_ts: phases[0].start_ts } : {};
    // if setting new custom pattern we use default 2 phases
    // for small groups get minimum batch size containing at least 1 device
    const minBatch = deploymentDeviceCount < 10 && !filter ? Math.ceil((1 / deploymentDeviceCount) * 100) : 10;
    switch (value) {
      case 0:
        updatedPhases = [{ batch_size: 100, ...phaseStart }];
        break;
      case 1:
        updatedPhases = [{ batch_size: minBatch, delay: 2, delayUnit: 'hours', ...phaseStart }, {}];
        break;
      default:
        // have to create a deep copy of the array to prevent overwriting, due to nested objects in the array
        updatedPhases = JSON.parse(JSON.stringify(value));
        break;
    }
    setDeploymentSettings({ phases: updatedPhases });
  };

  const onUsesPatternClick = useCallback(() => {
    if (usesPattern) {
      setDeploymentSettings({ phases: phases.slice(0, 1) });
    }
    setUsesPattern(!usesPattern);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [usesPattern, JSON.stringify(phases), setDeploymentSettings, setUsesPattern]);

  const numberDevices = deploymentDeviceCount ? deploymentDeviceCount : deploymentDeviceIds ? deploymentDeviceIds.length : 0;
  const customPattern = phases && phases.length > 1 ? 1 : 0;

  const previousPhaseOptions =
    previousPhases.length > 0
      ? previousPhases.map((previousPhaseSetting, index) => {
          const remainder = getRemainderPercent(previousPhaseSetting);
          const phaseDescription = previousPhaseSetting.reduce(
            (accu, phase, _, source) => {
              const phaseDescription = phase.delay
                ? `${phase.batch_size}% > ${phase.delay} ${phase.delayUnit || 'hours'} >`
                : `${phase.batch_size || remainder || 100 / source.length}%`;
              return `${accu} ${phaseDescription}`;
            },
            `${previousPhaseSetting.length} ${pluralize('phase', previousPhaseSetting.length)}:`
          );
          return (
            <MenuItem key={`previousPhaseSetting-${index}`} value={previousPhaseSetting}>
              {phaseDescription}
            </MenuItem>
          );
        })
      : [
          <MenuItem key="noPreviousPhaseSetting" disabled={true} style={{ opacity: '0.4' }}>
            No recent patterns
          </MenuItem>
        ];
  return (
    <>
      <FormControlLabel
        control={<Checkbox color="primary" checked={usesPattern} disabled={!isEnterprise} onChange={onUsesPatternClick} size="small" />}
        label={
          <div className="flexbox center-aligned">
            <b className="margin-right-small">Select a rollout pattern</b> (optional)
            <InfoHintContainer>
              <EnterpriseNotification id={BENEFITS.phasedDeployments.id} />
              <DocsTooltip id={DOCSTIPS.phasedDeployments.id} />
            </InfoHintContainer>
          </div>
        }
      />
      <Collapse in={usesPattern}>
        <FormControl className={classes.patternSelection}>
          <Select className={classes.input} onChange={handlePatternChange} value={customPattern} disabled={!isEnterprise}>
            <MenuItem value={0}>Single phase: 100%</MenuItem>
            {(numberDevices > 1 || filter) && [
              <MenuItem key="customPhaseSetting" divider={true} value={1}>
                Custom
              </MenuItem>,
              <ListSubheader key="phaseSettingsSubheader">Recent patterns</ListSubheader>,
              ...previousPhaseOptions
            ]}
          </Select>
        </FormControl>
      </Collapse>
      {customPattern ? <PhaseSettings classNames="margin-bottom-small" disabled={disableSchedule} numberDevices={numberDevices} {...props} /> : null}
    </>
  );
};

export default PhaseSettings;
