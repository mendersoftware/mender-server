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
//    limitations under the License.
import { FunctionComponent, ReactNode } from 'react';

import { Close as CloseIcon, Link as LinkIcon } from '@mui/icons-material';
import { IconButton } from '@mui/material';

interface DrawerTitleProps {
  onClose: () => void;
  title: string | ReactNode;
  onLinkCopy?: () => void;
  postTitle?: ReactNode;
  preCloser?: ReactNode;
}

export const DrawerTitle: FunctionComponent<DrawerTitleProps> = ({ onClose, onLinkCopy, postTitle, preCloser, title }) => (
  <div className="flexbox margin-bottom-small space-between">
    <div className="flexbox center-aligned">
      <h3 className="capitalized-start flexbox center-aligned">{title}</h3>
      {onLinkCopy && (
        <IconButton onClick={onLinkCopy} size="large">
          <LinkIcon />
        </IconButton>
      )}
      {postTitle}
    </div>
    <div className="flexbox center-aligned">
      {preCloser}
      <IconButton onClick={onClose} aria-label="close" size="large">
        <CloseIcon />
      </IconButton>
    </div>
  </div>
);
