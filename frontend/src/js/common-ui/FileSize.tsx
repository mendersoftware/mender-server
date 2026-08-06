// Copyright 2024 Northern.tech AS
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
import type { TypographyProps } from '@mui/material';
import { Typography } from '@mui/material';

import { getFormattedSize } from '@northern.tech/utils/helpers';

interface FileSizeProps extends TypographyProps {
  fileSize: number;
}

const FileSize = ({ fileSize, ...props }: FileSizeProps) => (
  <Typography variant="body2" {...props}>
    {getFormattedSize(fileSize)}
  </Typography>
);

export default FileSize;
