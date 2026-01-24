// Copyright 2015 Northern.tech AS
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
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

// material ui
import {
  Delete as DeleteIcon,
  ExpandLess,
  ExpandMore,
  Launch as LaunchIcon,
  SaveAlt as SaveAltIcon,
  GppGoodOutlined as SignedIcon,
  GppBadOutlined as UnsignedIcon
} from '@mui/icons-material';
import { Accordion, AccordionDetails, AccordionSummary, Button, Divider, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { EditableLongText } from '@northern.tech/common-ui/EditableLongText';
import { SynchronizedTwoColumnData } from '@northern.tech/common-ui/TwoColumnData';
import { getUserCapabilities } from '@northern.tech/store/selectors';
import { editArtifact, getArtifactInstallCount, getArtifactUrl } from '@northern.tech/store/thunks';
import { extractSoftware, extractSoftwareItem, isEmpty, toggle } from '@northern.tech/utils/helpers';
import pluralize from 'pluralize';

import ArtifactMetadataList from './ArtifactMetadataList';
import ArtifactPayload from './ArtifactPayload';

const useStyles = makeStyles()(() => ({
  accordPanel1: {
    padding: '0 15px',
    marginBottom: 30,
    [`&.Mui-expanded`]: {
      marginBottom: 30
    }
  }
}));

export const transformArtifactCapabilities = (capabilities = {}) =>
  Object.entries(capabilities).reduce((accu, [key, value]) => {
    if (!Array.isArray(value)) {
      accu[key] = value;
    } else if (!key.startsWith('device_type')) {
      // we can expect this to be an array of artifacts or artifact groups this artifact depends on
      accu = value.reduce((dependenciesAccu, dependency, index) => {
        const dependencyKey = value.length > 1 ? `${key}-${index + 1}` : key;
        dependenciesAccu[dependencyKey] = dependency;
        return dependenciesAccu;
      }, accu);
    }
    return accu;
  }, {});

export const transformArtifactMetadata = (metadata = {}) =>
  Object.entries(metadata).reduce((accu, [key, value]) => {
    if (Array.isArray(value)) {
      accu[key] = value.length ? value.join(',') : '-';
    } else if (value instanceof Object) {
      accu[key] = JSON.stringify(value) || '-';
    } else {
      accu[key] = value || '-';
    }
    return accu;
  }, {});

const DevicesLink = ({ artifact: { installCount }, softwareItem: { key, name, version }, title = '' }) => {
  const text = `${installCount} ${pluralize('device', installCount)}`;
  if (!installCount) {
    return <div title={title}>{text}</div>;
  }
  const attribute = `${key}${name ? `.${name}` : ''}.version`;
  return (
    <a
      className="flexbox center-aligned"
      href={`${window.location.origin}/ui/devices/accepted?inventory=${attribute}:eq:${version}`}
      target="_blank"
      rel="noreferrer"
      title={title}
    >
      {text}
      <LaunchIcon className="margin-left-small" fontSize="small" />
    </a>
  );
};

export const ArtifactDetails = ({ artifact, open, showRemoveArtifactDialog }) => {
  const { classes } = useStyles();
  const [showPayloads, setShowPayloads] = useState(false);
  const [showProvidesDepends, setShowProvidesDepends] = useState(false);

  const dispatch = useDispatch();

  const { canManageReleases } = useSelector(getUserCapabilities);

  const softwareVersions = useMemo(() => {
    const { software } = extractSoftware(artifact.artifact_provides);
    return software.reduce((accu, item) => {
      const infoItems = item[0].split('.');
      if (infoItems[infoItems.length - 1] !== 'version') {
        return accu;
      }
      accu.push({ key: infoItems[0], name: infoItems.slice(1, infoItems.length - 1).join('.'), version: item[1], nestingLevel: infoItems.length });
      return accu;
    }, []);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(artifact.artifact_provides)]);

  useEffect(() => {
    if (artifact.url || !open) {
      return;
    }
    dispatch(getArtifactUrl(artifact.id));
  }, [artifact.id, artifact.url, dispatch, open]);

  useEffect(() => {
    if (artifact.installCount || !open || softwareVersions.length > 1) {
      return;
    }
    const { version } = softwareVersions.sort((a, b) => a.nestingLevel - b.nestingLevel).reduce((accu, item) => accu ?? item, undefined) ?? {};
    if (version) {
      dispatch(getArtifactInstallCount(artifact.id));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [artifact.id, artifact.installCount, dispatch, open, softwareVersions.length]);

  const onDescriptionChanged = useCallback(description => dispatch(editArtifact({ id: artifact.id, body: { description } })), [artifact.id, dispatch]);

  const softwareItem = extractSoftwareItem(artifact.artifact_provides);
  const softwareInformation = softwareItem
    ? {
        title: 'Software versioning information',
        content: {
          'Software filesystem': softwareItem.key,
          'Software name': softwareItem.name,
          'Software version': softwareItem.version
        }
      }
    : { title: '', content: {} };

  const artifactMetaInfo = [
    { key: 'depends', title: 'Depends', content: transformArtifactCapabilities(artifact.artifact_depends) },
    { key: 'clears', title: 'Clears', content: transformArtifactCapabilities(artifact.artifact_clears) },
    { key: 'provides', title: 'Provides', content: transformArtifactCapabilities(artifact.artifact_provides) },
    { key: 'metadata', title: 'Artifact metadata', content: transformArtifactMetadata(artifact.metaData) }
  ];
  const hasMetaInfo = artifactMetaInfo.some(item => !isEmpty(item.content));
  const { installCount } = artifact;
  return (
    <div className={artifact.name == null ? 'muted' : null}>
      <SynchronizedTwoColumnData
        className="margin-bottom-small"
        data={{
          'Description': <EditableLongText fullWidth original={artifact.description} onChange={onDescriptionChanged} />,
          'Signed': artifact.signed ? (
            <div className="flexbox center-aligned">
              <SignedIcon className="green margin-right-x-small" />
              <Typography variant="body2">Signed</Typography>
            </div>
          ) : (
            <div className="flexbox center-aligned">
              <UnsignedIcon className="red margin-right-x-small" />
              <Typography variant="body2">Unsigned</Typography>
            </div>
          )
        }}
      />

      {installCount !== undefined && softwareVersions.length === 1 && (
        <SynchronizedTwoColumnData
          className="margin-bottom-small"
          data={{
            'Installed on': (
              <DevicesLink artifact={artifact} softwareItem={softwareItem} title={`installed on ${installCount} ${pluralize('device', installCount)}`} />
            )
          }}
        />
      )}
      <ArtifactMetadataList metaInfo={softwareInformation} />
      <Accordion square expanded={showPayloads} onChange={() => setShowPayloads(toggle)} className={classes.accordPanel1}>
        <AccordionSummary className="flexbox center-aligned">
          <Typography>Artifact contents</Typography>
          <div style={{ marginLeft: 'auto' }}>{showPayloads ? <ExpandLess /> : <ExpandMore />}</div>
        </AccordionSummary>
        <AccordionDetails>
          {showPayloads &&
            !!artifact.updates.length &&
            artifact.updates.map((update, index) => <ArtifactPayload index={index} payload={update} key={`artifact-update-${index}`} />)}
        </AccordionDetails>
      </Accordion>
      {hasMetaInfo && (
        <Accordion square expanded={showProvidesDepends} onChange={() => setShowProvidesDepends(!showProvidesDepends)} className={classes.accordPanel1}>
          <AccordionSummary className="flexbox center-aligned">
            <Typography>Provides and Depends</Typography>
            <div style={{ marginLeft: 'auto' }}>{showProvidesDepends ? <ExpandLess /> : <ExpandMore />}</div>
          </AccordionSummary>
          <AccordionDetails>
            {artifactMetaInfo
              .filter(({ content }) => !isEmpty(content))
              .map(({ key, title, content }) => (
                <div key={key}>
                  <Typography variant="subtitle2">{title}</Typography>
                  <Divider className="margin-top-small margin-bottom-small" />
                  <SynchronizedTwoColumnData data={content} />
                </div>
              ))}
          </AccordionDetails>
        </Accordion>
      )}
      <div className="two-columns margin-top-small" style={{ maxWidth: 'fit-content' }}>
        {canManageReleases && (
          <>
            <Button
              href={artifact.url}
              color="neutral"
              variant="outlined"
              target="_blank"
              disabled={!artifact.url}
              download={artifact.name ? `${artifact.name}.mender` : true}
              startIcon={<SaveAltIcon />}
            >
              Download Artifact
            </Button>
            <Button onClick={showRemoveArtifactDialog} variant="outlined" color="error" startIcon={<DeleteIcon className="red auth" />}>
              Remove this Artifact?
            </Button>
          </>
        )}
      </div>
    </div>
  );
};

export default ArtifactDetails;
