import { OAuth2Client } from 'google-auth-library';
import http from 'http';
import open from 'open';
import destroyer from 'server-destroy';
import url from 'url';

/**
 * Start by acquiring a pre-authenticated oAuth2 client.
 */
async function main() {
  await getAuthenticatedClient();
}

/**
 * Create a new OAuth2Client, and go through the OAuth2 content
 * workflow.  Return the full client to the callback.
 */
function getAuthenticatedClient() {
  return new Promise((resolve, reject) => {
    if (!process.env.GOOGLE_CLIENT_ID || !process.env.GOOGLE_CLIENT_SECRET) {
      reject('Missing required environment variables GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET');
    }
    // create an oAuth client to authorize the API call.  Secrets are kept in a `keys.json` file,
    // which should be downloaded from the Google Developers Console.
    const oAuth2Client = new OAuth2Client({
      clientId: process.env.GOOGLE_CLIENT_ID,
      clientSecret: process.env.GOOGLE_CLIENT_SECRET,
      redirectUri: 'http://localhost:3000/oauth2callback'
    });

    // Generate the url that will be used for the consent dialog.
    const authorizeUrl = oAuth2Client.generateAuthUrl({
      access_type: 'offline',
      scope: [
        'https://www.googleapis.com/auth/gmail.readonly' // To read messages
      ]
    });

    // Open an http server to accept the oauth callback. In this simple example, the
    // only request to our webserver is to /oauth2callback?code=<code>
    const server = http
      .createServer(async (req, res) => {
        try {
          if (req.url.indexOf('/oauth2callback') > -1) {
            // acquire the code from the querystring, and close the web server.
            const qs = new url.URL(req.url, 'http://localhost:3000').searchParams;
            const code = qs.get('code');
            res.end('Authentication successful! Please return to the console.');
            server.destroy();

            // Now that we have the code, use that to acquire tokens.
            const r = await oAuth2Client.getToken(code);
            // Make sure to set the credentials on the OAuth2 client.
            oAuth2Client.setCredentials(r.tokens);
            console.info(`Tokens acquired: ${JSON.stringify(r.tokens, null, 2)}`);
            console.info(`To use gmail with the end-to-end tests, set the following environment variable:`);
            console.info('```sh');
            console.info(`export GOOGLE_REFRESH_TOKEN=${r.tokens.refresh_token}`);
            console.info('```');
            resolve(oAuth2Client);
          }
        } catch (e) {
          reject(e);
        }
      })
      .listen(3000, () => {
        // open the browser to the authorize url to start the workflow
        open(authorizeUrl, { wait: false, app: 'firefox' }).then(cp => cp.unref());
      });
    destroyer(server);
  });
}

main().catch(console.error);
