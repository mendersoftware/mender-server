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
import { useEffect } from 'react';
import { useFormContext, useWatch } from 'react-hook-form';

import { Typography, inputLabelClasses, lighten } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { InfoHintContainer } from '@northern.tech/common-ui/InfoHint';
import ChipSelect from '@northern.tech/common-ui/forms/ChipSelect';
import TextInput from '@northern.tech/common-ui/forms/TextInput';
import { isDarkMode } from '@northern.tech/store/utils';

import { HELPTOOLTIPS } from '../../helptips/HelpTooltips';
import { MenderHelpTooltip } from '../../helptips/MenderTooltip';
import { FileInformation } from './FileInformation';

const useStyles = makeStyles()(theme => ({
  formWrapper: { display: 'flex', flexDirection: 'column', gap: theme.spacing(2) },
  hints: { pointerEvents: 'auto' },
  releaseName: {
    display: 'flex',
    overflow: 'visible',
    [`&.${inputLabelClasses.shrink}`]: {
      background: isDarkMode(theme.palette.mode) ? lighten(theme.palette.background.paper, 0.2) : theme.palette.background.default,
      paddingLeft: theme.spacing(0.5),
      height: 'inherit',
      paddingRight: theme.spacing(),
      marginTop: theme.spacing(-0.5)
    }
  }
}));

const versionFields = [
  { key: 'fileSystem', label: 'Software filesystem' },
  { key: 'softwareName', label: 'Software name' },
  { key: 'softwareVersion', label: 'Software version' }
];

export const VersionInformation = ({ file, onRemove, type }) => {
  const { classes } = useStyles();

  return (
    <div className={classes.formWrapper}>
      <FileInformation file={file} type={type} onRemove={onRemove} />
      <Typography variant="subtitle1">Version information</Typography>
      {versionFields.map(({ key, label }, index) => (
        <TextInput key={key} id={key} InputProps={{ autoFocus: !index }} label={label} required requiredRendered={false} width="100%" />
      ))}
    </div>
  );
};

const destinationHelperText = 'Where the file will be installed on your devices';

const checkDestinationValidity = destination => /^(?:\/|[a-z]+:\/\/)/.test(destination);

export const ArtifactInformation = ({ deviceTypes = [], file, onRemove, type }) => {
  const { classes } = useStyles();
  const {
    formState: { dirtyFields },
    setValue
  } = useFormContext();
  const releaseName = useWatch({ name: 'name' });

  useEffect(() => {
    if (dirtyFields.softwareName) {
      return;
    }
    setValue('softwareName', releaseName.replace('.', '-'));
  }, [dirtyFields.softwareName, releaseName, setValue]);

  const releaseNameLabel = (
    <>
      Release name
      <InfoHintContainer className={`margin-left-small margin-right-x-small ${classes.hints}`}>
        <MenderHelpTooltip small id={HELPTOOLTIPS.releaseName.id} placement="bottom-start" />
      </InfoHintContainer>
    </>
  );

  return (
    <div className={classes.formWrapper}>
      <FileInformation file={file} type={type} onRemove={onRemove} />
      <TextInput
        helperText={destinationHelperText}
        hint="Example: /opt/installed-by-single-file"
        id="destination"
        InputProps={{ autoFocus: true }}
        label="Destination directory"
        required
        requiredRendered={false}
        rules={{ validate: value => checkDestinationValidity(value) || 'Destination has to be an absolute path' }}
        width="100%"
      />
      <Typography variant="subtitle1">Artifact information</Typography>
      <TextInput
        hint="A descriptive name for the software"
        id="name"
        InputLabelProps={{ className: classes.releaseName, onClick: e => e.preventDefault() }}
        // the label carries the help tooltip, so it is rendered on top of the outline instead of in its notch - which requires suppressing the notch label
        InputProps={{ label: undefined }}
        label={releaseNameLabel}
        required
        requiredRendered={false}
        rules={{ required: 'Release name is required' }}
        width="100%"
      />
      <ChipSelect
        name="deviceTypes"
        helperText="Enter all device types this software is compatible with"
        label="Device types compatible"
        options={deviceTypes}
        rules={{ required: 'Device type is required' }}
      />
    </div>
  );
};

export const steps = [ArtifactInformation, VersionInformation];

export const ArtifactInformationForm = ({ activeStep, ...remainder }) => {
  const Component = steps[activeStep];
  return <Component {...remainder} />;
};

export default ArtifactInformationForm;
