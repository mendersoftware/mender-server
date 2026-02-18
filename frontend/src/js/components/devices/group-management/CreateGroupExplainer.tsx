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
import { Lock, SyncOutlined } from '@mui/icons-material';
import { Button, DialogActions, DialogContent } from '@mui/material';
import { Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import DocsLink, { DOCSTIPS } from '@northern.tech/common-ui/DocsLink';
import EnterpriseNotification from '@northern.tech/common-ui/EnterpriseNotification';
import { InfoHintContainer } from '@northern.tech/common-ui/InfoHint';
import { BaseDialog } from '@northern.tech/common-ui/dialogs/BaseDialog';
import { BENEFITS } from '@northern.tech/store/constants';

import dynamicImage from '../../../../assets/img/dynamic-group-creation.gif';
import staticImage from '../../../../assets/img/static-group-creation.gif';

const useStyles = makeStyles()(theme => ({
  groupType: {
    paddingRight: theme.spacing(2)
  },
  icon: { marginLeft: theme.spacing() },
  image: { maxWidth: '100%' }
}));

export const CreateGroupExplainer = ({ onClose }) => {
  const { classes } = useStyles();
  return (
    <BaseDialog title="Creating a group" open fullWidth maxWidth="md" onClose={onClose}>
      <DialogContent>
        <div className="flexbox column">
          <div className={`two-columns margin-bottom ${classes.groupType}`}>
            <div className="margin-right-large">
              <div className="flexbox align-items-center margin-bottom">
                <Typography variant="subtitle1">Static group</Typography>
                <Lock className={classes.icon} fontSize="small" />
              </div>
              <Typography>
                Select specific devices to add them to a static group. Filter to find devices in the &apos;All devices&apos; UI, check the devices you want to
                add, and click the action button to assign them to a group.
              </Typography>
            </div>
            <img className={classes.image} src={staticImage} />
          </div>
          <div className={`two-columns ${classes.groupType}`}>
            <div className="margin-right-large">
              <div className="flexbox align-items-center space-between margin-bottom">
                <Typography variant="subtitle1" className="flexbox align-items-center">
                  Dynamic group
                  <SyncOutlined className={`${classes.icon} margin-right`} fontSize="small" />
                </Typography>
                <InfoHintContainer>
                  <EnterpriseNotification id={BENEFITS.dynamicGroups.id} />
                </InfoHintContainer>
              </div>
              <Typography>
                You can set filters based on device attributes, and save them as a group. At any point in time, all devices that match the filters will be part
                of this group. This means that new devices will automatically join the group if they match the filters.{' '}
                <DocsLink path={DOCSTIPS.dynamicGroups.path} title="Learn more" /> about dynamic groups.
              </Typography>
            </div>
            <img className={classes.image} src={dynamicImage} />
          </div>
        </div>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
      </DialogActions>
    </BaseDialog>
  );
};

export default CreateGroupExplainer;
