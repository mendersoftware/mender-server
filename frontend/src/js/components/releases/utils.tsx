// Copyright 2026 Northern.tech AS
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
import { GppGoodOutlined as SignedIcon, GppBadOutlined as UnsignedIcon } from '@mui/icons-material';
import { Typography } from '@mui/material';

const variants = {
  signed: { Icon: SignedIcon, color: 'green', label: 'Signed' },
  unsigned: { Icon: UnsignedIcon, color: 'red', label: 'Unsigned' }
} as const;

export const SignatureSign = ({ isSigned }: { isSigned: boolean }) => {
  const { Icon, color, label } = variants[isSigned ? 'signed' : 'unsigned'];

  return (
    <div className="flexbox align-items-center">
      <Icon className={`${color} margin-right-x-small`} />
      <Typography variant="body2">{label}</Typography>
    </div>
  );
};
