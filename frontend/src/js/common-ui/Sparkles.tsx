// Copyright 2025 Northern.tech AS
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
import { keyframes } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { mdiStarFourPointsOutline as StarOutlined } from '@mdi/js';
import MaterialDesignIcon from '@northern.tech/common-ui/MaterialDesignIcon';

const sparkle = keyframes`
  0%, 100% {
    transform: scale(0.6);
  }
  50% {
    transform: scale(1);
  }
`;

const useStyles = makeStyles()(theme => ({
  star: {
    animation: `${sparkle} 2s ease-in-out infinite`,
    color: theme.palette.primary.main,
    fontSize: 80,
    position: 'absolute'
  }
}));

export const SparkleAnimation = () => {
  const { classes } = useStyles();
  return (
    <div className="relative" style={{ width: 200, height: 200 }}>
      <MaterialDesignIcon
        className={classes.star}
        path={StarOutlined}
        style={{
          animationDelay: '0s',
          top: 40,
          left: 0
        }}
      />
      <MaterialDesignIcon
        className={classes.star}
        path={StarOutlined}
        style={{
          top: 0,
          left: 55,
          animationDelay: '0.5s'
        }}
      />
      <MaterialDesignIcon
        className={classes.star}
        path={StarOutlined}
        style={{
          top: 75,
          left: 60,
          animationDelay: '1s'
        }}
      />
    </div>
  );
};
