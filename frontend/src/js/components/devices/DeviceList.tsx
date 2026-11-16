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
import { useMemo, useState } from 'react';

import DeviceIdentityDisplay from '@northern.tech/common-ui/DeviceIdentity';
import { CommonList } from '@northern.tech/common-ui/List';
import { useColumnWidths } from '@northern.tech/common-ui/useColumnWidths';

import { DefaultAttributeRenderer } from './BaseDevices';

const IdentityCell = ({ item: device }) => {
  const [isHovering, setIsHovering] = useState(false);
  return (
    <div className="full-width" onMouseEnter={() => setIsHovering(true)} onMouseLeave={() => setIsHovering(false)}>
      <DeviceIdentityDisplay device={device} isHovered={isHovering} />
    </div>
  );
};

export const DeviceList = ({
  columnHeaders,
  devices,
  deviceListState,
  idAttribute,
  onChangeRowsPerPage,
  PaginationProps = {},
  onExpandClick,
  onPageChange,
  onSelect,
  onSort,
  pageLoading
}) => {
  const { columnWidths, onResizeColumns } = useColumnWidths('device-column-widths');
  const columns = useMemo(
    () =>
      columnHeaders.map((column, index) => {
        const { component: Component = DefaultAttributeRenderer } = column;
        // the device column renderers expect the listed item as `device`
        return { ...column, component: index ? ({ item, ...props }) => <Component {...props} device={item} idAttribute={idAttribute} /> : IdentityCell };
      }),
    [columnHeaders, idAttribute]
  );
  return (
    <CommonList
      columnHeaders={columns}
      columnWidths={columnWidths}
      listItems={devices}
      listState={deviceListState}
      onChangeRowsPerPage={onChangeRowsPerPage}
      onExpandClick={onExpandClick}
      onPageChange={onPageChange}
      onResizeColumns={onResizeColumns}
      onSelect={onSelect}
      onSort={onSort}
      pageLoading={pageLoading}
      PaginationProps={PaginationProps}
    />
  );
};

export default DeviceList;
