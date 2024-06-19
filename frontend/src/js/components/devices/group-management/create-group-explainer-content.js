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
import React from 'react';

import { Autorenew, LockOutlined } from '@mui/icons-material';
import { makeStyles } from 'tss-react/mui';

import dynamicImage from '../../../../assets/img/dynamic-group-creation.gif';
import staticImage from '../../../../assets/img/static-group-creation.gif';
import { BENEFITS } from '../../../constants/appConstants';
import { DOCSTIPS, DocsTooltip } from '../../common/docslink';
import EnterpriseNotification from '../../common/enterpriseNotification';
import { InfoHintContainer } from '../../common/info-hint';

const useStyles = makeStyles()(theme => ({
  groupType: {
    flexGrow: 1,
    padding: 25,
    ['&.non-enterprise']: {
      backgroundColor: theme.palette.background.lightgrey
    }
  },

  icon: { color: '#7b7b7b' },
  image: { maxWidth: '100%' }
}));

export const CreateGroupExplainerContent = ({ isEnterprise }) => {
  const { classes } = useStyles();
  return (
    <div className="flexbox column">
      <div className={`two-columns margin-bottom ${classes.groupType}`}>
        <div className="margin-right-large">
          <div className="flexbox center-aligned margin-bottom">
            <LockOutlined className={classes.icon} fontSize="small" />
            <div className="bold margin-left-small">Static group</div>
          </div>
          <p className="help-content">
            Select specific devices to add them to a static group. Filter to find devices in the &apos;All devices&apos; UI, check the devices you want to add,
            and click the action button to assign them to a group.
          </p>
        </div>
        <img className={classes.image} src={staticImage} />
      </div>
      <div className={`two-columns ${classes.groupType} ${isEnterprise ? '' : 'non-enterprise'}`}>
        <div className="margin-right-large">
          <div className="flexbox center-aligned margin-bottom">
            <Autorenew className={classes.icon} fontSize="small" />
            <div className="bold margin-left-small margin-right">Dynamic group</div>
            <InfoHintContainer>
              <EnterpriseNotification id={BENEFITS.dynamicGroups.id} />
              <DocsTooltip id={DOCSTIPS.dynamicGroups.id} />
            </InfoHintContainer>
          </div>
          <p className="help-content">
            You can set filters based on device attributes, and save them as a group. At any point in time, all devices that match the filters will be part of
            this group. This means that new devices will automatically join the group if they match the filters.
          </p>
        </div>
        <img className={classes.image} src={dynamicImage} />
      </div>
    </div>
  );
};

export default CreateGroupExplainerContent;
