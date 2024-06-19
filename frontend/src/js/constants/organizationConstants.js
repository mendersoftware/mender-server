// Copyright 2020 Northern.tech AS
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
import { EXTERNAL_PROVIDER } from './deviceConstants';
import { useradmApiUrl } from './userConstants';

export const XML_METADATA_FORMAT = 'xml';
export const JSON_METADATA_FORMAT = 'json';

export const getSamlStartUrl = id => `${window.location.origin}${useradmApiUrl}/auth/sso/${id}/login`;
export const getOidcStartUrl = id => `${window.location.origin}${useradmApiUrl}/oidc/${id}/start`;

export const SSO_TYPES = {
  saml: {
    id: 'saml',
    type: 'saml',
    title: 'SAML',
    metadataFormat: XML_METADATA_FORMAT,
    editorLanguage: XML_METADATA_FORMAT,
    contentType: 'application/samlmetadata+xml',
    getStartUrl: getSamlStartUrl,
    configDetails: [
      { key: 'entityID', label: 'Entity ID', getValue: id => `${window.location.origin}${useradmApiUrl}/sso/sp/metadata/${id}` },
      { key: 'acs', label: 'ACS URL', getValue: id => `${window.location.origin}${useradmApiUrl}/auth/sso/${id}/acs` },
      { key: 'startURL', label: 'Start URL', getValue: getSamlStartUrl }
    ]
  },
  oidc: {
    id: 'oidc',
    type: 'oidc',
    title: 'OpenID Connect',
    metadataFormat: JSON_METADATA_FORMAT,
    editorLanguage: JSON_METADATA_FORMAT,
    contentType: 'application/json',
    getStartUrl: getOidcStartUrl,
    configDetails: [{ key: 'startURL', label: 'Start Url', getValue: getOidcStartUrl }]
  }
};

export const AUDIT_LOGS_TYPES = [
  { title: 'Artifact', queryParameter: 'object_type', value: 'artifact' },
  { title: 'Deployment', queryParameter: 'object_deployment_name', value: 'deployment' },
  { title: 'Device', queryParameter: 'object_id', value: 'device' },
  { title: 'User', queryParameter: 'object_id', value: 'user' }
];
export const RECEIVE_AUDIT_LOGS = 'RECEIVE_AUDIT_LOGS';
export const RECEIVE_CURRENT_CARD = 'RECEIVE_CURRENT_CARD';
export const RECEIVE_SETUP_INTENT = 'RECEIVE_SETUP_INTENT';
export const SET_AUDITLOG_STATE = 'SET_AUDITLOG_STATE';
export const SET_ORGANIZATION = 'SET_ORGANIZATION';
export const RECEIVE_EXTERNAL_DEVICE_INTEGRATIONS = 'RECEIVE_EXTERNAL_DEVICE_INTEGRATIONS';
export const RECEIVE_SSO_CONFIGS = 'RECEIVE_SSO_CONFIGS';
export const RECEIVE_WEBHOOK_EVENTS = 'RECEIVE_WEBHOOK_EVENTS';
export const emptyWebhook = {
  description: '',
  enabled: true,
  credentials: {
    type: EXTERNAL_PROVIDER.webhook.credentialsType,
    [EXTERNAL_PROVIDER.webhook.credentialsType]: {
      secret: '',
      url: ''
    }
  }
};
