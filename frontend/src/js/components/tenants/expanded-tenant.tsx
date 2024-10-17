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
import { Drawer } from '@mui/material';

interface ExpandedTenantProps {
  tenantId: string;
  onCloseClick: () => void;
}

export const ExpandedTenant = (props: ExpandedTenantProps) => {
  const { tenantId, onCloseClick } = props;
  return (
    <Drawer onClose={onCloseClick} open={true} PaperProps={{ style: { minWidth: '67vw' } }} anchor="right">
      Information for tenant with id {tenantId}
    </Drawer>
  );
};
