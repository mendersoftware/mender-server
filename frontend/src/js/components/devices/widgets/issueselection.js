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
import React, { useCallback, useMemo, useState } from 'react';

// material ui
import { Checkbox, MenuItem, Select } from '@mui/material';

import { DEVICE_ISSUE_OPTIONS } from '../../../constants/deviceConstants';

const menuProps = {
  anchorOrigin: {
    vertical: 'bottom',
    horizontal: 'left'
  },
  transformOrigin: {
    vertical: 'top',
    horizontal: 'left'
  }
};
const groupOptions = (options = [], selection = []) => {
  const things = options.reduce((accu, value) => {
    const { issueCategory, key } = DEVICE_ISSUE_OPTIONS[value.key];
    const nestedValue = { ...value, level: 0, checked: selection.includes(key) };
    if (issueCategory) {
      nestedValue.level = 1;
      let categoryItem = { ...DEVICE_ISSUE_OPTIONS[issueCategory], count: nestedValue.count, checked: nestedValue.checked };
      let existingItems = [];
      if (Array.isArray(accu[issueCategory])) {
        categoryItem = {
          ...accu[issueCategory][0],
          count: accu[issueCategory][0].count + nestedValue.count,
          checked: accu[issueCategory][0].checked && nestedValue.checked
        };
        existingItems = accu[issueCategory].slice(1);
      }
      accu[issueCategory] = [categoryItem, ...existingItems, nestedValue];
    } else {
      accu[key] = nestedValue;
    }
    return accu;
  }, {});
  return Object.values(things).flat();
};

const getSelectionDisplayValue = ({ selected = [], options = [] }) => {
  let content = 'all';
  if (selected.length) {
    const { titles, sum } = selected.reduce(
      (accu, issue) => {
        accu.titles.push(DEVICE_ISSUE_OPTIONS[issue].title);
        accu.sum += options.find(option => option.key === issue)?.count || 0;
        return accu;
      },
      { titles: [], sum: 0 }
    );
    content = `${titles.join(', ')} (${sum})`;
  }
  return content;
};

const DeviceIssuesSelection = ({ classes, onChange, options, selection }) => {
  const [open, setOpen] = useState(false);

  const handleClose = () => setOpen(false);

  const handleOpen = e => {
    if (e && e.target.closest('input')?.hasOwnProperty('checked')) {
      return;
    }
    setOpen(true);
  };

  // eslint-disable-next-line react-hooks/exhaustive-deps
  const groupedOptions = useMemo(() => groupOptions(options, selection), [options.join(''), selection.join('')]);

  const onSelectionChange = useCallback(
    ({ target: { value: newSelection } }, { props: { value: clickedItem } }) => {
      const issue = DEVICE_ISSUE_OPTIONS[clickedItem];
      let categoryItems = [];
      if (issue.isCategory) {
        categoryItems = options.reduce(
          (collector, option) => (DEVICE_ISSUE_OPTIONS[option.key].issueCategory === clickedItem ? [...collector, option.key] : collector),
          categoryItems
        );
      }

      let selectedOptions = newSelection;
      if (categoryItems.length && categoryItems.every(item => selection.includes(item))) {
        selectedOptions = selectedOptions.filter(option => !(categoryItems.includes(option) || option === clickedItem));
      } else {
        selectedOptions = [...newSelection, ...categoryItems].filter(option => (issue.isCategory ? option !== clickedItem : true));
      }
      onChange({ target: { value: [...new Set(selectedOptions)] } });
    },
    [onChange, options, selection]
  );

  return (
    <div className="flexbox center-aligned margin-left">
      <div>Show:</div>
      <Select
        className={classes.selection}
        disableUnderline
        displayEmpty
        MenuProps={menuProps}
        multiple
        open={open}
        onClose={handleClose}
        onOpen={handleOpen}
        onChange={onSelectionChange}
        renderValue={selected => getSelectionDisplayValue({ selected, options: groupedOptions })}
        value={selection}
      >
        {groupedOptions.map(({ checked, count, key, title, level = 0 }) => (
          <MenuItem key={key} value={key} size="small">
            <Checkbox checked={checked} style={{ marginLeft: 8 * (level + 1) }} />
            {title} ({count})
          </MenuItem>
        ))}
      </Select>
    </div>
  );
};

export default DeviceIssuesSelection;
