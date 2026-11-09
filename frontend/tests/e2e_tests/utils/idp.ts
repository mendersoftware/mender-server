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
import escapeHtml from 'escape-html';
import express from 'express';
import * as fs from 'fs';
import type { Server } from 'http';
import * as path from 'path';
import samlp from 'samlp';

// A minimal SAML identity provider for the SSO tests: it serves the IdP metadata and a sign in page
// that lets the test pick the identity, then posts a signed assertion back to the service provider.

const port = 7000;
const ssoPath = '/saml/sso';
const signInPath = '/saml/signin';
const defaultIdentity = 'saml.jackson@example.com';

const idpOptions = {
  issuer: 'urn:example:idp',
  cert: fs.readFileSync(path.join(import.meta.dirname, '..', 'fixtures', 'idp-public-cert.pem'), 'utf8'),
  key: fs.readFileSync(path.join(import.meta.dirname, '..', 'fixtures', 'idp-private-key.pem'), 'utf8'),
  digestAlgorithm: 'sha256',
  signatureAlgorithm: 'rsa-sha256'
};

const profileMapper = ({ nameId, email }) => ({
  getClaims: () => ({ email }),
  getNameIdentifier: () => ({ nameIdentifier: nameId, nameIdentifierFormat: 'urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress' })
});

const signInPage = ({ acsUrl = '', audience = '', inResponseTo = '', relayState = '' }) => `<!DOCTYPE html>
<html>
  <head><title>Test Identity Provider</title></head>
  <body>
    <h1>Test Identity Provider</h1>
    <form method="post" action="${signInPath}">
      <input type="hidden" name="acsUrl" value="${escapeHtml(acsUrl)}" />
      <input type="hidden" name="audience" value="${escapeHtml(audience)}" />
      <input type="hidden" name="inResponseTo" value="${escapeHtml(inResponseTo)}" />
      <input type="hidden" name="RelayState" value="${escapeHtml(relayState)}" />
      <p>
        <label for="nameId">Subject NameID</label>
        <input id="nameId" name="nameId" value="${defaultIdentity}" />
      </p>
      <p>
        <label for="email">E-Mail Address</label>
        <input id="email" name="email" value="${defaultIdentity}" />
      </p>
      <button type="submit">Sign In</button>
    </form>
  </body>
</html>`;

const app = express();
app.use(express.urlencoded({ extended: false }));

app.get('/metadata', samlp.metadata({ ...idpOptions, redirectEndpointPath: ssoPath, postEndpointPath: ssoPath }));

app.all(ssoPath, (req, res) =>
  samlp.parseRequest(req, (error, authnRequest) => {
    if (error || !authnRequest?.assertionConsumerServiceURL) {
      console.error('test IdP received an invalid AuthnRequest', error);
      return res.status(400).send(`invalid AuthnRequest: ${error?.message ?? 'missing AssertionConsumerServiceURL'}`);
    }
    res.send(
      signInPage({
        acsUrl: authnRequest.assertionConsumerServiceURL,
        audience: authnRequest.issuer,
        inResponseTo: authnRequest.id,
        relayState: req.query.RelayState ?? req.body?.RelayState
      })
    );
  })
);

app.post(signInPath, (req, res, next) => {
  const { acsUrl, audience, email, inResponseTo, nameId } = req.body;
  samlp.auth({
    ...idpOptions,
    audience,
    destination: acsUrl,
    recipient: acsUrl,
    inResponseTo,
    profileMapper,
    getUserFromRequest: () => ({ nameId, email }),
    getPostURL: (_audience, _samlRequestDom, _req, callback) => callback(null, acsUrl)
  })(req, res, next);
});

export const startIdpServer = (): Promise<Server> =>
  new Promise((resolve, reject) => {
    const server = app.listen(port, error => (error ? reject(error) : resolve(server)));
  });
