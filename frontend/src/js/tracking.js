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
import ReactGA4 from 'react-ga4';

const cookieConsentCSS = 'https://cdn.jsdelivr.net/npm/cookieconsent@3/build/cookieconsent.min.css';
const cookieConsentJS = 'https://cdn.jsdelivr.net/npm/cookieconsent@3/build/cookieconsent.min.js';

const ReactGA = ReactGA4.default;

class Tracker {
  constructor() {
    this.initialized = false;
    this.trackingEnabled = true;
    this.currentPageView = null;
    this.currentOrganizationUser = null;
  }
  cookieconsent() {
    return new Promise(resolve => {
      const style = document.createElement('link');
      style.href = cookieConsentCSS;
      style.rel = 'stylesheet';
      style.async = true;
      document.head.appendChild(style);
      //
      const script = document.createElement('script');
      script.src = cookieConsentJS;
      script.async = false;
      script.addEventListener('load', () => {
        window.cookieconsent.initialise({
          palette: {
            popup: {
              background: '#5d0f43',
              text: '#ffffff'
            },
            button: {
              background: '#73a4ad',
              text: '#ffffff'
            }
          },
          position: 'bottom-left',
          type: 'opt-out',
          content: {
            message: 'We use cookies to analyze our traffic so we can improve our website and give you a better experience.',
            link: 'View our cookie policy',
            href: 'https://northern.tech/legal/cookies'
          },
          autoOpen: true,
          revokable: false,
          law: {
            regionalLaw: false
          },
          onStatusChange: status => {
            let hasConsented = status == 'allow';
            resolve({ trackingConsentGiven: hasConsented });
          }
        });
      });
      document.body.appendChild(script);
    });
  }
  exception(error) {
    if (this.initialized && this.trackingEnabled) {
      ReactGA.event('error', error);
    }
  }
  event(data) {
    if (this.initialized && this.trackingEnabled) {
      ReactGA.event(data);
    }
  }
  initialize(trackingCode) {
    if (this.initialized && this.trackingEnabled) {
      return false;
    }
    ReactGA.initialize(trackingCode);
    this.initialized = true;
    return true;
  }
  pageview(data) {
    if (data) {
      this.currentPageView = data;
    }
  }
  set(value) {
    if (this.initialized && this.trackingEnabled) {
      ReactGA.set(value);
    }
  }
  setOrganizationUser(organization, user) {
    if (this.initialized && this.trackingEnabled && this.currentOrganizationUser != { organization, user }) {
      this.currentOrganizationUser = { organization, user };
      this.set({ dimension1: organization.plan });
      this.set({ dimension2: organization.id });
      this.set({ dimension3: user.id });
      this.set({ userId: user.id });
    }
  }
  setTrackingEnabled(trackingEnabled) {
    this.trackingEnabled = trackingEnabled;
  }
}

const Tracking = new Tracker();
export default Tracking;
