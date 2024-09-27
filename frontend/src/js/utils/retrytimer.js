// Copyright 2017 Northern.tech AS
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
import { TIMEOUTS } from '@northern.tech/store/constants';
import { extractErrorMessage, preformatWithRequestID } from '@northern.tech/store/utils';

let timers = {};

export function setRetryTimer(err, service, errorContext, timeLeft, setSnackbar) {
  // check if service not already retrying
  if (timers[service]) {
    return;
  }
  let remaining = timeLeft - TIMEOUTS.oneSecond;
  timers[service] = setInterval(() => {
    remaining -= TIMEOUTS.oneSecond;
    const errMsg = extractErrorMessage(err, 'Please check your connection.');
    remaining > 0
      ? setSnackbar(preformatWithRequestID(err.response, `${errorContext} ${errMsg} Retrying in ${remaining / 1000} seconds`))
      : clearRetryTimer(service, setSnackbar);
  }, TIMEOUTS.oneSecond);
}

export function clearRetryTimer(service, setSnackbar) {
  if (timers[service]) {
    clearInterval(timers[service]);
    delete timers[service];
    setSnackbar('');
  }
}

export function clearAllRetryTimers(setSnackbar) {
  Object.keys(timers).map(service => clearRetryTimer(service, setSnackbar));
  setSnackbar('');
}
