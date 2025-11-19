// Copyright 2018 Northern.tech AS
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
import { ReactNode } from 'react';

import { FormControl, FormControlLabel, Switch, Typography } from '@mui/material';

export const ToggleSetting = ({
  description,
  disabled = false,
  title,
  onClick,
  value
}: {
  description?: string;
  disabled?: boolean;
  onClick: () => void;
  title: string | ReactNode;
  value: boolean;
}) => (
  <div className="flexbox column">
    <FormControl variant="standard">
      <FormControlLabel
        disabled={disabled}
        classes={{ label: 'capitalized-start' }}
        className="align-self-start margin-left-none margin-top-none"
        control={<Switch className="margin-left-small" checked={value} onClick={onClick} />}
        label={title}
        labelPlacement="start"
      />
    </FormControl>
    {!!description && (
      <Typography className="margin-top-x-small" variant="body2">
        {description}
      </Typography>
    )}
  </div>
);
