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
import { Alert, Card, CardActions, CardContent, Checkbox, FormControlLabel, Typography } from '@mui/material';

import { Addon, AvailableAddon, Plan } from '@northern.tech/store/appSlice/constants';

interface AddonProps {
  addon: Addon;
  checked: boolean;
  disabled: boolean;
  onChange: (addon: AvailableAddon, checked: boolean) => void;
  selectedPlan: Plan;
}
export const SubscriptionAddon = (props: AddonProps) => {
  const { addon, disabled, checked = false, onChange, selectedPlan } = props;
  const disabledDueToPlan = !addon.eligible.includes(selectedPlan.id);
  return (
    <Card variant="outlined" className="margin-bottom-small">
      <CardContent className="padding-bottom-none">
        <FormControlLabel
          className="margin-none"
          disabled={disabled}
          value={checked}
          onChange={(event, checked) => onChange(addon.id, checked)}
          control={<Checkbox name={addon.id} className="padding-none margin-x-small" checked={checked} />}
          label={addon.title}
        />
        <Typography className="margin-top-x-small margin-bottom-x-small" variant="body2">
          {addon.description}
        </Typography>
        {disabledDueToPlan && (
          <Alert severity="info" className="margin-bottom-x-small">
            {addon.title} is not available on the {selectedPlan.name} plan.
          </Alert>
        )}
      </CardContent>
      <CardActions className="padding-small padding-top-none">
        <a href="http://mender.io" target="_blank" rel="noreferrer">
          Learn More
        </a>
      </CardActions>
    </Card>
  );
};
