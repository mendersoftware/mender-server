// Copyright 2021 Northern.tech AS
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
import { deepmerge } from '@mui/utils';
import { makeStyles } from 'tss-react/mui';

import { TwoColumnData } from '@northern.tech/common-ui/ConfigurationObject';
import { rootfsImageVersion, softwareTitleMap } from '@northern.tech/store/constants';
import { extractSoftware, isEmpty } from '@northern.tech/utils/helpers';

import DeviceDataCollapse from './DeviceDataCollapse';
import DeviceInventoryLoader from './DeviceInventoryLoader';

const borderStyle = theme => ({ borderLeft: 'solid 1px', borderLeftColor: theme.palette.grey[500] });

const useStyles = makeStyles()(theme => ({
  paddingOnly: { paddingLeft: theme.spacing(2) },
  nestingBorders: { ...borderStyle(theme), paddingLeft: theme.spacing(2), paddingBottom: theme.spacing(2) },
  topLevelBorder: { ...borderStyle(theme), paddingBottom: theme.spacing(2), marginBottom: theme.spacing(-2) }
}));

const mapLayerInformation = (key, value, i, path) => {
  const infoItems = key.split('.');
  let priority = i;
  const title = infoItems[0];
  if (softwareTitleMap[path]) {
    priority = softwareTitleMap[path].priority;
  }
  const itemKey = infoItems[infoItems.length - 1];
  let contents = {};
  if (infoItems.length > 2) {
    contents = { content: {}, children: { [infoItems[1]]: mapLayerInformation(infoItems.slice(1).join('.'), value, i + 1, key) } };
  } else {
    contents = { content: { [itemKey]: value }, children: {} };
  }
  return {
    priority,
    key: infoItems.slice(0, -1).join('.'),
    title,
    ...contents
  };
};

const sortAndHoist = thing =>
  Object.entries(thing)
    .sort((a, b) => a[1].priority - b[1].priority)
    .reduce((accu, entry) => {
      let { children, content, title } = entry[1];
      title = Object.keys(content).reduce((layerTitle, key) => {
        if (softwareTitleMap[`${title}.${key}`]) {
          return softwareTitleMap[`${title}.${key}`].title;
        }
        return layerTitle;
      }, title);
      children = sortAndHoist(children);
      if (isEmpty(content) && Object.keys(children).length === 1) {
        const child = Object.entries(children).reduce((result, [key, value]) => ({ title: key, ...value }), {});
        title = `${title}.${child.title}`;
        content = child.content;
        children = child.children;
      }
      accu = { ...accu, [title]: { content, children, title } };
      return accu;
    }, {});

/**
 * to get information about the software installed on the device we first need to:
 * - parse the inventory attributes that are likely software references (those with a key ending on '.version')
 * - recursively descend the attribute to create a tree and group software based on shared prefixes
 * - for a shallower tree rendering the resulting tree is descended recursively once more and all
 *    software with only a single "sublayer" is hoisted up & listed under the shared title
 */
export const extractSoftwareInformation = (attributes = {}, sort = true) => {
  const { software } = extractSoftware(attributes);

  const softwareLayers = software.reduce((accu, item, index) => {
    const layer = mapLayerInformation(item[0], item[1], index, item[0]);
    if (!accu[layer.title]) {
      accu[layer.title] = { content: {}, children: {}, title: '' };
    }
    accu[layer.title] = {
      ...accu[layer.title],
      ...layer,
      children: deepmerge(accu[layer.title].children, layer.children),
      content: deepmerge(accu[layer.title].content, layer.content),
      priority: accu[layer.title].priority < layer.priority ? accu[layer.title].priority : layer.priority
    };
    return accu;
  }, {});
  if (sort) {
    return sortAndHoist(softwareLayers);
  }
  return softwareLayers;
};

const SoftwareLayer = ({ classes, layer, isNested, overviewOnly, setSnackbar }) => (
  <div className={`margin-top-small ${overviewOnly ? classes.paddingOnly : ''}`}>
    <div className="muted">{layer.title}</div>
    {!isEmpty(layer.content) && (
      <div className={isNested || overviewOnly ? '' : classes.topLevelBorder}>
        <TwoColumnData
          className={`${isNested || overviewOnly ? 'margin-bottom-small' : ''} margin-left-small margin-top-small`}
          config={layer.content}
          compact
          setSnackbar={setSnackbar}
        />
      </div>
    )}
    {!overviewOnly && !isEmpty(layer.children) && (
      <div className={classes.nestingBorders}>
        {Object.entries(layer.children).map(([key, child]) => (
          <SoftwareLayer classes={classes} key={key} layer={child} isNested setSnackbar={setSnackbar} />
        ))}
      </div>
    )}
  </div>
);

export const InstalledSoftware = ({ device, setSnackbar }) => {
  const { classes } = useStyles();

  const { attributes = {} } = device;

  let softwareInformation = extractSoftwareInformation(attributes);

  if (isEmpty(softwareInformation)) {
    softwareInformation = [
      {
        children: [],
        content: { [softwareTitleMap[rootfsImageVersion].title]: attributes.artifact_name },
        title: softwareTitleMap[rootfsImageVersion].title
      }
    ];
  }

  const waiting = !Object.values(attributes).some(i => i);
  return (
    <DeviceDataCollapse header={waiting && <DeviceInventoryLoader />} title="Installed software">
      <div className={classes.nestingBorders}>
        {Object.entries(softwareInformation).map(([key, layer]) => (
          <SoftwareLayer classes={classes} key={key} layer={layer} setSnackbar={setSnackbar} />
        ))}
      </div>
    </DeviceDataCollapse>
  );
};

export default InstalledSoftware;
