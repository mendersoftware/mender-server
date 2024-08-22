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
import React, { useEffect, useState } from 'react';
import { FormProvider, useForm } from 'react-hook-form';

import { FormControl, Input, InputLabel, TextField } from '@mui/material';

import ChipSelect from '../../common/chipselect';
import { DOCSTIPS, DocsTooltip } from '../../common/docslink';
import { InfoHintContainer } from '../../common/info-hint';
import { HELPTOOLTIPS, MenderHelpTooltip } from '../../helptips/helptooltips';
import { FileInformation } from './addartifact';

const defaultVersion = '1.0.0';

export const VersionInformation = ({ creation = {}, onRemove, updateCreation }) => {
  const { file, fileSystem: propFs, name, softwareName: propName, softwareVersion: version = '', type } = creation;
  const [fileSystem, setFileSystem] = useState(propFs);
  const [softwareName, setSoftwareName] = useState(propName || name.replace('.', '-'));
  const [softwareVersion, setSoftwareVersion] = useState(version || defaultVersion);

  useEffect(() => {
    updateCreation({ finalStep: true });
  }, [updateCreation]);

  useEffect(() => {
    updateCreation({ fileSystem, softwareName, softwareVersion, isValid: fileSystem && softwareName && softwareVersion });
  }, [fileSystem, softwareName, softwareVersion, updateCreation]);

  return (
    <>
      <FileInformation file={file} type={type} onRemove={onRemove} />
      <h4>Version information</h4>
      <div className="flexbox column">
        {[
          { key: 'fileSystem', title: 'Software filesystem', setter: setFileSystem, value: fileSystem },
          { key: 'softwareName', title: 'Software name', setter: setSoftwareName, value: softwareName },
          { key: 'softwareVersion', title: 'Software version', setter: setSoftwareVersion, value: softwareVersion }
        ].map(({ key, title, setter, value: currentValue }, index) => (
          <TextField autoFocus={!index} fullWidth key={key} label={title} onChange={({ target: { value } }) => setter(value)} value={currentValue} />
        ))}
      </div>
    </>
  );
};

const checkDestinationValidity = destination => (destination.length ? /^(?:\/|[a-z]+:\/\/)/.test(destination) : true);

export const ArtifactInformation = ({ creation = {}, deviceTypes = [], onRemove, updateCreation }) => {
  const { destination = '', file, name = '', selectedDeviceTypes = [], type } = creation;

  const methods = useForm({ mode: 'onChange', defaultValues: { deviceTypes: selectedDeviceTypes } });
  const { watch } = methods;
  const formDeviceTypes = watch('deviceTypes');

  useEffect(() => {
    updateCreation({ selectedDeviceTypes: formDeviceTypes });
  }, [formDeviceTypes, updateCreation]);

  useEffect(() => {
    updateCreation({
      destination,
      isValid: checkDestinationValidity(destination) && selectedDeviceTypes.length && name,
      finalStep: false
    });
  }, [destination, name, selectedDeviceTypes.length, updateCreation]);

  const onDestinationChange = ({ target: { value } }) =>
    updateCreation({ destination: value, isValid: checkDestinationValidity(value) && selectedDeviceTypes.length && name });

  const isValidDestination = checkDestinationValidity(destination);
  return (
    <div className="flexbox column">
      <FileInformation file={file} type={type} onRemove={onRemove} />
      <TextField
        autoFocus={true}
        error={!isValidDestination}
        fullWidth
        helperText={!isValidDestination && <div className="warning">Destination has to be an absolute path</div>}
        inputProps={{ style: { marginTop: 16 } }}
        InputLabelProps={{ shrink: true }}
        label="Destination directory where the file will be installed on your devices"
        onChange={onDestinationChange}
        placeholder="Example: /opt/installed-by-single-file"
        value={destination}
      />
      <h4>Artifact information</h4>
      <FormControl>
        <InputLabel htmlFor="release-name" className="flexbox center-aligned" onClick={e => e.preventDefault()}>
          Release name
          <InfoHintContainer>
            <MenderHelpTooltip id={HELPTOOLTIPS.releaseName.id} />
            <DocsTooltip id={DOCSTIPS.releases.id} />
          </InfoHintContainer>
        </InputLabel>
        <Input
          defaultValue={name}
          className="release-name-input"
          id="release-name"
          placeholder="A descriptive name for the software"
          onChange={e => updateCreation({ name: e.target.value })}
        />
      </FormControl>
      <FormProvider {...methods}>
        <form noValidate>
          <ChipSelect
            name="deviceTypes"
            label="Device types compatible"
            placeholder="Enter all device types this software is compatible with"
            options={deviceTypes}
          />
        </form>
      </FormProvider>
    </div>
  );
};

const steps = [ArtifactInformation, VersionInformation];

export const ArtifactInformationForm = ({ activeStep, ...remainder }) => {
  const Component = steps[activeStep];
  return <Component {...remainder} />;
};

export default ArtifactInformationForm;
