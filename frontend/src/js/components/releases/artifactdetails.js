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
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

// material ui
import {
  Add as AddIcon,
  Cancel as CancelIcon,
  CancelOutlined as CancelOutlinedIcon,
  CheckCircleOutline as CheckCircleOutlineIcon,
  ExitToApp as ExitToAppIcon,
  Launch as LaunchIcon,
  Remove as RemoveIcon
} from '@mui/icons-material';
import { Accordion, AccordionDetails, AccordionSummary, Button, List, ListItem, ListItemText } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import pluralize from 'pluralize';

import { editArtifact, getArtifactInstallCount, getArtifactUrl } from '../../actions/releaseActions';
import { extractSoftware, extractSoftwareItem, toggle } from '../../helpers';
import { getUserCapabilities } from '../../selectors';
import ExpandableAttribute from '../common/expandable-attribute';
import ArtifactPayload from './artifactPayload';
import ArtifactMetadataList from './artifactmetadatalist';
import { EditableLongText } from './releasedetails';

const useStyles = makeStyles()(theme => ({
  link: { marginTop: theme.spacing() },
  listItemStyle: {
    bordered: {
      borderBottom: '1px solid',
      borderBottomColor: theme.palette.grey[500]
    },
    color: theme.palette.text.primary,
    fontSize: 13,
    marginRight: '2vw',
    minWidth: 200,
    padding: 0,
    width: 'initial'
  },
  paddingOverride: { paddingBottom: 4, paddingTop: 0 },
  accordPanel1: {
    background: theme.palette.grey[500],
    borderTop: 'none',
    padding: '0 15px',
    marginBottom: 30,
    [`&.Mui-expanded`]: {
      background: theme.palette.grey[500],
      marginBottom: 30
    }
  },
  accordSummary: {
    background: theme.palette.grey[500],
    padding: 0
  }
}));

export const transformArtifactCapabilities = (capabilities = {}) => {
  return Object.entries(capabilities).reduce((accu, [key, value]) => {
    if (!Array.isArray(value)) {
      accu.push({ key, primary: key, secondary: value });
    } else if (!key.startsWith('device_type')) {
      // we can expect this to be an array of artifacts or artifact groups this artifact depends on
      const dependencies = value.reduce((dependencies, dependency, index) => {
        const dependencyKey = value.length > 1 ? `${key}-${index + 1}` : key;
        dependencies.push({ key: dependencyKey, primary: dependencyKey, secondary: dependency });
        return dependencies;
      }, []);
      accu.push(...dependencies);
    }
    return accu;
  }, []);
};

export const transformArtifactMetadata = (metadata = {}) => {
  return Object.entries(metadata).reduce((accu, [key, value]) => {
    const commonProps = { key, primary: key, secondaryTypographyProps: { component: 'div' } };
    if (Array.isArray(value)) {
      accu.push({ ...commonProps, secondary: value.length ? value.join(',') : '-' });
    } else if (value instanceof Object) {
      accu.push({ ...commonProps, secondary: JSON.stringify(value) || '-' });
    } else {
      accu.push({ ...commonProps, secondary: value || '-' });
    }
    return accu;
  }, []);
};

const DevicesLink = ({ artifact: { installCount }, softwareItem: { key, name, version } }) => {
  const { classes } = useStyles();
  const text = `${installCount} ${pluralize('device', installCount)}`;
  if (!installCount) {
    return <div className={classes.link}>{text}</div>;
  }
  const attribute = `${key}${name ? `.${name}` : ''}.version`;
  return (
    <a
      className={`flexbox center-aligned ${classes.link}`}
      href={`${window.location.origin}/ui/devices/accepted?inventory=${attribute}:eq:${version}`}
      target="_blank"
      rel="noreferrer"
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

  const onDescriptionChanged = useCallback(description => dispatch(editArtifact(artifact.id, { description })), [artifact.id, dispatch]);

  const softwareItem = extractSoftwareItem(artifact.artifact_provides);
  const softwareInformation = softwareItem
    ? {
        title: 'Software versioning information',
        content: [
          { primary: 'Software filesystem', secondary: softwareItem.key },
          { primary: 'Software name', secondary: softwareItem.name },
          { primary: 'Software version', secondary: softwareItem.version }
        ]
      }
    : { content: [] };

  const artifactMetaInfo = [
    { title: 'Depends', content: transformArtifactCapabilities(artifact.artifact_depends) },
    { title: 'Clears', content: transformArtifactCapabilities(artifact.artifact_clears) },
    { title: 'Provides', content: transformArtifactCapabilities(artifact.artifact_provides) },
    { title: 'Artifact metadata', content: transformArtifactMetadata(artifact.metaData) }
  ];
  const hasMetaInfo = artifactMetaInfo.some(item => !!item.content.length);
  const { installCount } = artifact;
  const itemProps = { classes: { root: 'attributes', disabled: 'opaque' }, className: classes.listItemStyle };
  return (
    <div className={artifact.name == null ? 'muted' : null}>
      <List className="list-horizontal-flex">
        <ListItem {...itemProps}>
          <ListItemText
            primary="Description"
            style={{ marginBottom: -3, minWidth: 600 }}
            primaryTypographyProps={{ style: { marginBottom: 3 } }}
            secondary={<EditableLongText fullWidth original={artifact.description} onChange={onDescriptionChanged} />}
            secondaryTypographyProps={{ component: 'div' }}
          />
        </ListItem>
        <ListItem {...itemProps} className={`${classes.listItemStyle} ${classes.listItemStyle.bordered}`}>
          <ListItemText primary="Signed" secondary={artifact.signed ? <CheckCircleOutlineIcon className="green" /> : <CancelOutlinedIcon className="red" />} />
        </ListItem>
        {installCount !== undefined && softwareVersions.length === 1 && (
          <ExpandableAttribute
            classes={{ root: classes.paddingOverride }}
            disableGutters
            primary="Installed on"
            secondary={<DevicesLink artifact={artifact} softwareItem={softwareItem} />}
            secondaryTypographyProps={{ title: `installed on ${installCount} ${pluralize('device', installCount)}` }}
            style={{ padding: 0 }}
          />
        )}
      </List>
      <ArtifactMetadataList metaInfo={softwareInformation} />
      <Accordion square expanded={showPayloads} onChange={() => setShowPayloads(toggle)} className={classes.accordPanel1}>
        <AccordionSummary className={classes.accordSummary}>
          <p>Artifact contents</p>
          <div style={{ marginLeft: 'auto' }}>{showPayloads ? <RemoveIcon /> : <AddIcon />}</div>
        </AccordionSummary>
        <AccordionDetails className={classes.accordSummary}>
          {showPayloads &&
            !!artifact.updates.length &&
            artifact.updates.map((update, index) => <ArtifactPayload index={index} payload={update} key={`artifact-update-${index}`} />)}
        </AccordionDetails>
      </Accordion>
      {hasMetaInfo && (
        <Accordion square expanded={showProvidesDepends} onChange={() => setShowProvidesDepends(!showProvidesDepends)} className={classes.accordPanel1}>
          <AccordionSummary className={classes.accordSummary}>
            <p>Provides and Depends</p>
            <div style={{ marginLeft: 'auto' }}>{showProvidesDepends ? <RemoveIcon /> : <AddIcon />}</div>
          </AccordionSummary>
          <AccordionDetails className={classes.accordSummary}>
            {showProvidesDepends && artifactMetaInfo.map((info, index) => <ArtifactMetadataList metaInfo={info} key={`artifact-info-${index}`} />)}
          </AccordionDetails>
        </Accordion>
      )}
      <div className="two-columns margin-top-small" style={{ maxWidth: 'fit-content' }}>
        {canManageReleases && (
          <>
            <Button
              href={artifact.url}
              target="_blank"
              disabled={!artifact.url}
              download={artifact.name ? `${artifact.name}.mender` : true}
              startIcon={<ExitToAppIcon style={{ transform: 'rotateZ(90deg)' }} />}
            >
              Download Artifact
            </Button>
            <Button onClick={showRemoveArtifactDialog} startIcon={<CancelIcon className="red auth" />}>
              Remove this Artifact?
            </Button>
          </>
        )}
      </div>
    </div>
  );
};

export default ArtifactDetails;
