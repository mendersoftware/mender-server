---
## 4.0.1 - 2025-05-19


### Bug fixes


- *(gui)* Fixed an issue that would prevent showing unassigned static group devices
([ME-519](https://northerntech.atlassian.net/browse/ME-519)) ([b595a57](https://github.com/mendersoftware/mender-server/commit/b595a574d2e3d2d4fb61757198007ff4bc308f56)) 

- Race when provisioning and connecting a device
([MEN-8164](https://northerntech.atlassian.net/browse/MEN-8164)) ([5535078](https://github.com/mendersoftware/mender-server/commit/5535078d027b213fb0ee6d37dd4c4934b075e51c)) 


  If a device gets provisioned and submits a connection request while the
  device is getting provisioned, the device might end up in an
  inconsistent state where the connection status gets overwritten to
  "unknown".
  The issue was discovered in a test where the system was under load
  and the device was running on the same network (artificially low RTT).
- Ensure email is always encoded in lowercase when stored
([MEN-8328](https://northerntech.atlassian.net/browse/MEN-8328)) ([0568d2e](https://github.com/mendersoftware/mender-server/commit/0568d2e09699961cb8aabc0faa16813a4928bc44)) 


  Added a bson codec for model.Email that will ensure that emails are
  always encoded in lowercase in the database to ensure case insensitive
  queries.






## 4.0.0 - 2025-02-10


### Bug fixes


- *(deployments)* Deprecate v1 endpoint for listing deployments
([MEN-7543](https://northerntech.atlassian.net/browse/MEN-7543)) ([879b589](https://github.com/mendersoftware/mender-server/commit/879b58986f7e34906cff649c687d47de9152455c))  by @kjaskiewiczz


  We deprecated GET v1 /deployments/deployments endpoint because of an
  issue with "search" query parameter behavior. As a replacement we
  introduce v2 version of the endpoint, where we replaced "search"
  parameter with "id" and "name" parameters.

- *(deviceconfig)* Enable multiplatform build
([QA-673](https://northerntech.atlassian.net/browse/QA-673)) ([fbbe646](https://github.com/mendersoftware/mender-server/commit/fbbe6466981015f47f250ad673f35f00004d1589))  by @oldgiova


  The required TARGETOS and TARGETARCH variables were missing from the
  Dockerfile.

- *(gui)* Fixed an issue that could prevent browsers from following programmatically triggered downloads
 ([f2b6189](https://github.com/mendersoftware/mender-server/commit/f2b61896fc476d08f65d11a604777bb7809d6de1))  by @mzedel


  - relative download addresses seem not to be followed, switched to absolute instead

- *(gui)* Fixed an issue that would sometimes prevent users from switching between tenants
([MEN-7774](https://northerntech.atlassian.net/browse/MEN-7774)) ([ce777fd](https://github.com/mendersoftware/mender-server/commit/ce777fdc9ae558a21a630384152152872c94b7a5))  by @mzedel


  can't rely on the user list data as it doesn't contain all the user details

- *(gui)* Fixed an issue that prevented deployment sizes from being shown
 ([d2bbb8d](https://github.com/mendersoftware/mender-server/commit/d2bbb8df54aea9288af6d77944a516a075816928))  by @mzedel

- *(gui)* Fixed an issue that caused number comparisons in device filters to not work
([MEN-7717](https://northerntech.atlassian.net/browse/MEN-7717)) ([84e2398](https://github.com/mendersoftware/mender-server/commit/84e2398fece6b10fddcf6f60e3ff744af903c707))  by @mzedel

- *(gui)* Added readable name for ltne device filter
([MEN-7717](https://northerntech.atlassian.net/browse/MEN-7717)) ([a741011](https://github.com/mendersoftware/mender-server/commit/a74101176c22df69455a9d0634494912e219daab))  by @mzedel

- *(gui)* Fixed an issue that could lead to unexpected locations in the UI when accessing unauthorized sections while authorized
([MEN-7842](https://northerntech.atlassian.net/browse/MEN-7842)) ([7938291](https://github.com/mendersoftware/mender-server/commit/7938291f8ac37c7ee3366c0cf2773e2c0053438f))  by @mzedel

- *(gui)* Enable device configuration for non enterprise users
 ([67170c5](https://github.com/mendersoftware/mender-server/commit/67170c5edb27a1061abf2826234fabab45e4dedf))  by @thall


  Currently it's not possible to see device configuration if you host
  Mender self and have environment variable `HAVE_DEVICECONFIG=true`.

  Changes the predicate to be the same as for `hasDeviceConnect`.

- *(gui)* Added missing link to rbac docs in the cooresponding section
([MEN-7826](https://northerntech.atlassian.net/browse/MEN-7826)) ([1d8c4ff](https://github.com/mendersoftware/mender-server/commit/1d8c4ff3f71f5918ea98ff277f96c31a85ebffe5))  by @mzedel

- *(gui)* Prevented disabled form inputs from showing validation errors
 ([2e7215a](https://github.com/mendersoftware/mender-server/commit/2e7215aa93a3d357cfad34ec24e852ca66faf7df))  by @mzedel

- *(gui)* Aligned quick actions in release details with actually possibile actions
 ([365f564](https://github.com/mendersoftware/mender-server/commit/365f5646f2c32956fa8c0cee22c20d8c3757948d))  by @mzedel

- *(gui)* Fixed an issue that would prevent showing deployment reports for phased deployments
 ([132d6b2](https://github.com/mendersoftware/mender-server/commit/132d6b2aa932924c6612a8ca39867f246b388a88))  by @mzedel

- *(gui)* Fixed an issue that would prevent upgrading a running session to a different plan
([MEN-7898](https://northerntech.atlassian.net/browse/MEN-7898)) ([7668b29](https://github.com/mendersoftware/mender-server/commit/7668b293bc71d4eaffda0c00823cac5026dfbf4c))  by @mzedel

- *(gui)* Fixed an issue that would crash the site when showing release details with multiple artifacts
 ([fd06f66](https://github.com/mendersoftware/mender-server/commit/fd06f66516e4f33e52c9f305fb058594eecf714f))  by @mzedel

- *(gui)* Made addon availability rely more on addons where possible to prevent erroneous device config retrieval
([MEN-7895](https://northerntech.atlassian.net/browse/MEN-7895)) ([62d6516](https://github.com/mendersoftware/mender-server/commit/62d6516b848b35e493db8be9908abaf7e573e008))  by @mzedel

- *(gui)* Aligned webhook details behaviour w/ rest of UI
([MEN-7955](https://northerntech.atlassian.net/browse/MEN-7955)) ([7860b5b](https://github.com/mendersoftware/mender-server/commit/7860b5b40c698b580f5a299ee8c9206490ea5710))  by @mzedel

- *(gui)* Prevented sso config retrieval on plans that don't support this
 ([fe6da5d](https://github.com/mendersoftware/mender-server/commit/fe6da5dbea68222226ab01f52df7e5975fedc09d))  by @mzedel

- *(gui)* Fixed an issue that would prevent deleting & tagging releases in the release overview
([MEN-7960](https://northerntech.atlassian.net/browse/MEN-7960)) ([16b2628](https://github.com/mendersoftware/mender-server/commit/16b2628feaf39eba631b5ab013bf3eeecfa95217))  by @mzedel

- *(gui)* Let on-prem installations refer to the docs to prevent server-url misconfiguration following monorepo transition
([MEN-7948](https://northerntech.atlassian.net/browse/MEN-7948)) ([e0dae51](https://github.com/mendersoftware/mender-server/commit/e0dae512f67f08f312a61bd3be8192b7bbb7d6db))  by @mzedel

- *(gui)* Fixed end date filters out today's entries in the Audit log, Deployments and Devices
 ([3ee84f2](https://github.com/mendersoftware/mender-server/commit/3ee84f2d743b51462f72e68efeae870a51c4d12c))  by @aleksandrychev

- *(gui)* Ensured target directory is nonempty on artifact generation
([MEN-8010](https://northerntech.atlassian.net/browse/MEN-8010)) ([5616722](https://github.com/mendersoftware/mender-server/commit/561672221f31d0be257c4e1da98c63eda40f792d))  by @mzedel

- *(gui)* Fixed an issue that could prevent listing devices with their custom identity in a deployment report
 ([f1fcf26](https://github.com/mendersoftware/mender-server/commit/f1fcf26f33a3ae458e60f994fed60231d75abd5c))  by @mzedel

- *(gui)* Fixed an issue that would prevent navigating to devices from a software distribution chart
([MEN-8038](https://northerntech.atlassian.net/browse/MEN-8038)) ([6516986](https://github.com/mendersoftware/mender-server/commit/6516986b82ce08ecdfd9ed3c790590796c60da2d))  by @mzedel

- *(gui)* Made user list tracking rely only on backend data instead of local store to prevent duplicate users listed
([MEN-8049](https://northerntech.atlassian.net/browse/MEN-8049)) ([7d1b060](https://github.com/mendersoftware/mender-server/commit/7d1b060d096ff2549305cc0f453d8a35b21257b0))  by @mzedel

- *(gui)* Fixed an issue that would cause the ui to crash when creating phased deployments
 ([9827ba9](https://github.com/mendersoftware/mender-server/commit/9827ba928889f52e7eb4216b68707386c082dc74))  by @mzedel

- *(gui)* Fixed remaining device percentage not being displayed correctly on phased deployment creation
 ([5600913](https://github.com/mendersoftware/mender-server/commit/560091349c778c288ad1bdc5a8bcd9b54c39f399))  by @mzedel

- *(iot-core)* Incosistent serialization format for device private key
([MEN-7478](https://northerntech.atlassian.net/browse/MEN-7478)) ([6deadef](https://github.com/mendersoftware/mender-server/commit/6deadef48a11f8f845fdb9c1e33ddf33a366d531))  by @alfrunes


  The generated key is serialized using SEC 1 (RFC5915) ASN.1 encoding,
  but encoded to PEM using PKCS8 (RFC5208) block header/trailer.

- Aligned identity attribute usage with scoped inventory data to prevent overriding custom naming attributes with name tags
([MEN-7218](https://northerntech.atlassian.net/browse/MEN-7218)) ([9d82ea1](https://github.com/mendersoftware/mender-server/commit/9d82ea13bbc705dae08f277d3cbd5905386a9452))  by @mzedel
- Document the 409 return for creating deployment to a group
([MEN-7414](https://northerntech.atlassian.net/browse/MEN-7414)) ([5327dac](https://github.com/mendersoftware/mender-server/commit/5327daca07acc2e5b955ed2a23ae6ead39182e0d))  by @kjaskiewiczz
- Prevented commercial client components are only selected when plan/ addon accessible
([MEN-7458](https://northerntech.atlassian.net/browse/MEN-7458)) ([81e0b73](https://github.com/mendersoftware/mender-server/commit/81e0b73047ba17c2301b080fa5a16f1200d7975a))  by @mzedel
- Fixed an issue that prevented retrieving group devices & related reports
([MEN-7461](https://northerntech.atlassian.net/browse/MEN-7461)) ([95ea85f](https://github.com/mendersoftware/mender-server/commit/95ea85f46e26cacd59ce3f8c2a10fc086c756405))  by @mzedel
- Fixed an issue that prevented showing PATs on page refresh
 ([bde80fe](https://github.com/mendersoftware/mender-server/commit/bde80fe781be0c07684d2d4689227708695a548e))  by @mzedel
- Fixed an issue that prevented onboarding tips from showing
 ([c2ecfcf](https://github.com/mendersoftware/mender-server/commit/c2ecfcffd0a21f17ea8b3485a5d87efa21ab233a))  by @mzedel
- Fixed an issue that prevented the UI from showing deeply nested software installations
([MEN-7640](https://northerntech.atlassian.net/browse/MEN-7640)) ([13496f3](https://github.com/mendersoftware/mender-server/commit/13496f3468fd08dcc9656ba07463eba682cfaff4))  by @mzedel
- Stop user from having similar email and password
([MEN-6462](https://northerntech.atlassian.net/browse/MEN-6462)) ([3fa4a43](https://github.com/mendersoftware/mender-server/commit/3fa4a432780a40fb9b8c37633c7feca6ba3445c5))  by @bahaa-ghazal
- Implement signal handler for `server` commands
([QA-782](https://northerntech.atlassian.net/browse/QA-782)) ([6e17ada](https://github.com/mendersoftware/mender-server/commit/6e17adaaffa6778dc021353248d83b08cf645182))  by @bahaa-ghazal
- Deviceauth healthcheck panics malformed inventory address
 ([70d493a](https://github.com/mendersoftware/mender-server/commit/70d493a6913827d893758cd481a535de67fbeff9))  by @alfrunes
- Use internal URLs for storage backend when generating artifacts
([MEN-7939](https://northerntech.atlassian.net/browse/MEN-7939)) ([3d72d5e](https://github.com/mendersoftware/mender-server/commit/3d72d5e5b0294a3dcf3faa4413104ef27f95ba19))  by @alfrunes
  - **BREAKING**: Generate artifacts API ignores `storage.proxy_uri` and
`aws.external_url` configuration values and instead access the API using
the same URL as deployments service.


  When generating artifacts, the backend will use the direct access URL
  instead of rewriting the URL using the configured `storage.proxy_uri` or
  `aws.external_url`.
- Deployment device count should not exceed max devices
([MEN-7847](https://northerntech.atlassian.net/browse/MEN-7847)) ([15e5fee](https://github.com/mendersoftware/mender-server/commit/15e5feec727e4257a1ee4345265146a194edb4ab))  by @alfrunes


  Added a condition to skip deployments when the device count reaches max
  devices.
- Update outdated api endpoints in the `inventory` service
([MEN-7017](https://northerntech.atlassian.net/browse/MEN-7017)) ([73c7149](https://github.com/mendersoftware/mender-server/commit/73c714951a61642b2fc100214c61e5f66c27ee0c))  by @bahaa-ghazal
- Limiting the size of metadata when uploading and generating artifacts
([MEN-7166](https://northerntech.atlassian.net/browse/MEN-7166)) ([9e80728](https://github.com/mendersoftware/mender-server/commit/9e8072874c94da9e7a3659207e08c6a05fc48cc4))  by @bahaa-ghazal

- *(create-artifact-worker)* do not install openssl1.1-compat

  Changes:

  - use mender-artifact which does not depend on openssl1.1-compat
  - do not install openssl1.1-compat
- *(deployments)* Accesslog catches panic traces and remove "dev" middleware ([MC-7155](https://northerntech.atlassian.net/browse/MC-7155))
- *(deployments)* Improve error message when uploading too large artifacts ([MEN-7175](https://northerntech.atlassian.net/browse/MEN-7175))
- *(deployments)* fix release filtering and RBAC for releases
-  (deployments)* With the old implementation, when using more than one tag in the filter, or when using role which grants access to releases with given tag (and more than one tag was specified), deployments will present only releases containg ALL the tags each. With the new behavior, deployments will retrun all the releases containg ANY of the tags. ([MEN-7272](https://northerntech.atlassian.net/browse/MEN-7272))
- *(deployments)* Compatibility with MongoDB \> 5.0 ([MEN-6956](https://northerntech.atlassian.net/browse/MEN-6956))
- *(deviceauth)* Handling preauthorized auth set when device is accepted
- *(deviceauth)* The previous behavior was putting the device in a conflicting state and returning 500 errors on auth requests. With this commit, the preauthorized auth set will take precedence and take over as the accepted auth set. ([ALV-213](https://northerntech.atlassian.net/browse/ALV-213))
- *(deviceauth)* Wrong Content-Type header on successful authentication
- *(deviceauth)* On success, the Content-Type header is set to `application/jwt` instead of invalid `application/json` on 200 responses to POST /api/devices/v1/authentication/auth\_requests ([MEN-6912](https://northerntech.atlassian.net/browse/MEN-6912))
- *(deviceauth)* Preauthorize force behavior applies to existing auth sets
- *(deviceauth)* Updates the behavior of the Preauthorize endpoint if "force" paremeter is set:
   * If an authset already exist, the status will be forced to "preauthorized".
   * If the auth set does not exist, a new one will be created. ([MEN-7241](https://northerntech.atlassian.net/browse/MEN-7241))
- *(deviceauth)* Inconsistent device `check_in_time` when listing devices

  The lookup for `check_in_time` from the cache does not work when running Redis in cluster mode because of the MGET command requires keys to hash to the same slot. This commit replaces MGET with multiple batched GET commands when running Redis in cluster mode. ([MEN-7337](https://northerntech.atlassian.net/browse/MEN-7337))
- *(gui)* fixed missing theme global variables ([MEN-7044](https://northerntech.atlassian.net/browse/MEN-7044))
- *(gui)* fixed terminal closure made by exit command ([MEN-7081](https://northerntech.atlassian.net/browse/MEN-7081))
- *(gui)* Devices tab not showing in the UI with Deployments manager ([MEN-7111](https://northerntech.atlassian.net/browse/MEN-7111))
- *(gui)* fixed Software distribution widget displayed wrong other devices count
- *(gui)* ensured release is retrieved on deployment recreation ([MEN-7228](https://northerntech.atlassian.net/browse/MEN-7228))
- *(gui)* ensured an attempt to show fresh device information is made in every device related auditlog entry ([MEN-7034](https://northerntech.atlassian.net/browse/MEN-7034))
- *(gui)* fixed an issue that sometimes prevented reopening paginated auditlog links
- *(gui)* fixed an issue that could prevent SSO logins depending on the type of SSO
- *(gui)* fixed SSO information not being adjusted depending on the type of SSO configured ([MEN-7277](https://northerntech.atlassian.net/browse/MEN-7277))
- *(gui)* fixed an issue that prevented accessing releases with routing relevant symbols in their name ([MEN-7209](https://northerntech.atlassian.net/browse/MEN-7209))
- *(gui)* ensured browser generated reports are refreshed on every full device data retrieval to prevent partly initialized report data to show misleading software distributions ([MEN-7123](https://northerntech.atlassian.net/browse/MEN-7123))
- *(gui)* fixed an issue that would prevent promoting a device to a gateway device ([MEN-7334](https://northerntech.atlassian.net/browse/MEN-7334))
- *(gui)* limited global settings saving for less privileged users ([MEN-6970](https://northerntech.atlassian.net/browse/MEN-6970))
- *(inventory)* Accesslog middleware log panic traces and remove "dev" middleware ([MC-7155](https://northerntech.atlassian.net/browse/MC-7155))
- *(inventory)* Bound the number of devices considered when aggregating filter attributes to maximum 5,000. ([MEN-6917](https://northerntech.atlassian.net/browse/MEN-6917))
- *(inventory)* do not return updated\_ts as zero time if updated\_ts is not set
- *(inventory)* store "check\_in\_time" attribute as ISODate instead of string ([MEN-7259](https://northerntech.atlassian.net/browse/MEN-7259))
- *(inventory)* attribute modification in the range loop

- *(iot-manager)* Event APIs return OK if event is saved to database ([MEN-6898](https://northerntech.atlassian.net/browse/MEN-6898))
- *(iot-manager)* Create TTL index for removing expired logs ([MEN-7101](https://northerntech.atlassian.net/browse/MEN-7101))
- *(iot-manager)* Incosistent serialization format for device private key

  The generated key is serialized using (RFC5915) ASN.1 encoding, but encoded to PEM using PKCS8 (RFC5208) block header/trailer. ([MEN-7478](https://northerntech.atlassian.net/browse/MEN-7478), \[SEC 1\]([https://northerntech.atlassian.net/browse/SEC](https://northerntech.atlassian.net/browse/SEC) 1))

- *(useradm)* Update accesslog middleware to catch panic traces and remove dev mode ([MC-7155](https://northerntech.atlassian.net/browse/MC-7155))

### Documentation


- *(README)* Add step to clone repository
 ([f9d3bbd](https://github.com/mendersoftware/mender-server/commit/f9d3bbde382bca4592f41e3d6be7e8292dcb221f))  by @alfrunes

- *(README)* Consistently add syntax highlighting to code blocks
 ([8583102](https://github.com/mendersoftware/mender-server/commit/8583102cbbf49882b9a9ab1b80257516ec13dc24))  by @alfrunes

- *(deployments)* Clarifications for the GET /deployments version 2 endpoint.
([MEN-8053](https://northerntech.atlassian.net/browse/MEN-8053)) ([ea9fda0](https://github.com/mendersoftware/mender-server/commit/ea9fda0593e982f15eaa6be0d5b0f240a7994878))

- Update README.md with instructions on using the docker composition
 ([c9aa7dc](https://github.com/mendersoftware/mender-server/commit/c9aa7dc73db9717f436def2b14dc8f7cce74903f))  by @alfrunes
- Add section about testing build artifacts
 ([5c7eaaa](https://github.com/mendersoftware/mender-server/commit/5c7eaaa22c9cdb0016b9f06afc3bccdfdd9b9578))  by @alfrunes
- Add snippet for starting a mender client to README
 ([a322b2d](https://github.com/mendersoftware/mender-server/commit/a322b2d8c22e5e9e409a9a5c96c3b8412cef555f))  by @alfrunes
- Update README.md
 ([f7a1b09](https://github.com/mendersoftware/mender-server/commit/f7a1b097726672dd40ed7df17551229c5cf6ce7f))  by @alfrunes


  Adjusted styling (note color, added 1st level indentation,  taxonomy i.e., Mender Server, Mender Enterprise) to make it easy to follow and read.
- Document how to bring up the Virtual Device for enterprise setup
 ([c674566](https://github.com/mendersoftware/mender-server/commit/c674566e6d834c64d6e64d321c6e09b5f2a36259))  by @lluiscampos
- Fix typo in snippet for creating tenant
 ([a346d33](https://github.com/mendersoftware/mender-server/commit/a346d33781086d157d831478cfb64bebeef6c3bd))  by @alfrunes
- Docmentation on backend integration tests running separately
([QA-683](https://northerntech.atlassian.net/browse/QA-683)) ([a8f8d54](https://github.com/mendersoftware/mender-server/commit/a8f8d545573100186fba953c7179592a23196b23))  by @merlin-northern




### Features


- *(deployments)* Add filter field to deployment object
([MEN-7416](https://northerntech.atlassian.net/browse/MEN-7416)) ([fec5b91](https://github.com/mendersoftware/mender-server/commit/fec5b91d59d07b1a0d85ccf077cd56aa5b192278))  by @kjaskiewiczz


  The filter field contains information about devices targeted by the
  deployment.

- *(deployments)* New endpoint for getting release by name
([MEN-7575](https://northerntech.atlassian.net/browse/MEN-7575)) ([3a18e88](https://github.com/mendersoftware/mender-server/commit/3a18e880ec5cddedc19ed08949777caedda4350d))  by @kjaskiewiczz

- *(gui)* Enabled webhook scope selection
([MEN-7455](https://northerntech.atlassian.net/browse/MEN-7455)) ([cec277d](https://github.com/mendersoftware/mender-server/commit/cec277d83adf930de47ca5bb03935aa350ea1af5))  by @mzedel

- *(gui)* Extended webhook event details
([MEN-7574](https://northerntech.atlassian.net/browse/MEN-7574)) ([0bfda40](https://github.com/mendersoftware/mender-server/commit/0bfda409122ed6837e13cf7f5418b093bf3ef97b))  by @mzedel

- *(gui)* Aligned webhook listing with updated design
([MEN-7573](https://northerntech.atlassian.net/browse/MEN-7573)) ([80e55d1](https://github.com/mendersoftware/mender-server/commit/80e55d15e361c21988e192bf715a219bb300f487))  by @mzedel

- *(gui)* Added the possibility to create service provider administering roles
([MEN-7570](https://northerntech.atlassian.net/browse/MEN-7570)) ([92d7e50](https://github.com/mendersoftware/mender-server/commit/92d7e50e311d8c88f9847a83cec7b797ef9cebcc))  by @mzedel

- *(gui)* Aligned role removal dialog with other parts of the UI
 ([8661704](https://github.com/mendersoftware/mender-server/commit/866170425bef1f01f3a4a25f0d4e19fe5da94a6e))  by @mzedel

- *(gui)* Added support for Personal Access Token auditlog entries
([MEN-7622](https://northerntech.atlassian.net/browse/MEN-7622)) ([9a9a6c3](https://github.com/mendersoftware/mender-server/commit/9a9a6c3829611c35622e3812db7bbedd9bc9f9e5))  by @mzedel

- *(gui)* Added possibility to trigger deployment & inventory data updates when troubleshooting
([MEN-7657](https://northerntech.atlassian.net/browse/MEN-7657)) ([11a9b7a](https://github.com/mendersoftware/mender-server/commit/11a9b7a57a179c3d9605779b41f6d10b6dbc72fb))  by @mzedel

- *(gui)* Made deployment targets rely on filter information in the deployment to more reliably display target devices etc.
([MEN-7647](https://northerntech.atlassian.net/browse/MEN-7647)) ([47c92d4](https://github.com/mendersoftware/mender-server/commit/47c92d4db494cfc77116258fc2ed7fdca8691400))  by @mzedel

- *(gui)* Aligned notions of "latest device activity" in listing & details
 ([40ee57d](https://github.com/mendersoftware/mender-server/commit/40ee57da173c1d5bc2a39b2a5534b62dc986f53c))  by @mzedel

- *(gui)* Limited onboarding to hosted Mender to ensure a streamlined experience
([MEN-7896](https://northerntech.atlassian.net/browse/MEN-7896)) ([cee60f8](https://github.com/mendersoftware/mender-server/commit/cee60f8de9eb5b6940828f5e09e11dbd4e2d4059))  by @mzedel

- *(gui)* Added feedback on file size limits to artifact upload dialog
([MEN-7858](https://northerntech.atlassian.net/browse/MEN-7858)) ([d612334](https://github.com/mendersoftware/mender-server/commit/d612334ebfae6b1a3d416016ee500b89daa70804))  by @mzedel

- *(gui)* Aligned text input appearance with MUI updated guidelines
([MEN-7838](https://northerntech.atlassian.net/browse/MEN-7838)) ([e5d5672](https://github.com/mendersoftware/mender-server/commit/e5d56720b901a451fa47a514424f710763b50291))  by @mzedel

- *(gui)* Added explanation about integration number limitation
([MEN-7899](https://northerntech.atlassian.net/browse/MEN-7899)) ([dbdfa67](https://github.com/mendersoftware/mender-server/commit/dbdfa672a62eb30745fb4b0e73bea7d345644932))  by @mzedel

- *(gui)* Clarified user creation capabilities for non-enterprise users
([MEN-7883](https://northerntech.atlassian.net/browse/MEN-7883)) ([d2fd192](https://github.com/mendersoftware/mender-server/commit/d2fd192e8c09443dabb694288d73636853b02d86))  by @mzedel

- *(gui)* Added automatic refresh to get webhook events
([MEN-8045](https://northerntech.atlassian.net/browse/MEN-8045)) ([502e06a](https://github.com/mendersoftware/mender-server/commit/502e06aab96f3ce980cd6fe197a2fd6d7e99233a))  by @mzedel

- *(inventory)* Add support for "$in" operator in the device search API
([MEN-7667](https://northerntech.atlassian.net/browse/MEN-7667)) ([fd4eaf0](https://github.com/mendersoftware/mender-server/commit/fd4eaf0ecc8b72ff7fa9cfe7b6f214bc4678a97f))  by @kjaskiewiczz

- Added option to limit deployments to a maximum number of devices in a dynamic group
([MEN-7403](https://northerntech.atlassian.net/browse/MEN-7403)) ([c04d736](https://github.com/mendersoftware/mender-server/commit/c04d736fd58adba549858b86b936d076f855eb7c))  by @mzedel
- Made search results reopen whenever the search field is clicked again and has a search term
([MEN-6894](https://northerntech.atlassian.net/browse/MEN-6894)) ([c36eb96](https://github.com/mendersoftware/mender-server/commit/c36eb96c9c1790a0b23f05507021378c59267690))  by @mzedel
- Added feedback dialog
([MEN-7355](https://northerntech.atlassian.net/browse/MEN-7355)) ([8c0a3ba](https://github.com/mendersoftware/mender-server/commit/8c0a3baa2fa4e4cf935d818235e651bd4c5ed85c))  by @mzedel
- New endpoint for listing deployments
([MEN-7541](https://northerntech.atlassian.net/browse/MEN-7541)) ([afb1566](https://github.com/mendersoftware/mender-server/commit/afb15665474440751e0463582e5d08d07b626da8))  by @kjaskiewiczz
- Add `version` command to all Go binaries
 ([ff439c9](https://github.com/mendersoftware/mender-server/commit/ff439c93552ae7e32d3a0cb932339902f45271ec))  by @alfrunes


  The `version` command will display the app version (linked at build
  time) as well as runtime version and commit SHA1.

- *(deployments)* Add configuration for max data size when generating artifacts
- *(deployments)* Adds a new configuration option for setting the max data section size when generating an image with a default of 512MiB. The configuraiton path is `storage.max_generate_data_size` or environment variable `DEPLOYMENTS_STORAGE_MAX_GENERATE_DATA_SIZE`. ([MEN-7134](https://northerntech.atlassian.net/browse/MEN-7134))
- *(deployments)* prevent the creation of deployments if there is already an active deployment with the same constructor parameters ([MEN-6622](https://northerntech.atlassian.net/browse/MEN-6622))
- *(deviceauth)* accept and support preauth at any time ([MEN-6961](https://northerntech.atlassian.net/browse/MEN-6961))
- *(deviceauth)* sync check\_in\_time with inventory if reporting is disabled ([MEN-7202](https://northerntech.atlassian.net/browse/MEN-7202))
- *(deviceconfig)* internal endpoint to delete all records related to a tenant ([MEN-7312](https://northerntech.atlassian.net/browse/MEN-7312))
- *(deviceconnect)* Forward filetransfer statuscode from client ([ALV-209](https://northerntech.atlassian.net/browse/ALV-209))
- *(deviceconnect)* internal endpoint to delete all records related to a tenant ([MEN-7317](https://northerntech.atlassian.net/browse/MEN-7317))
- *(devicemonitor)* internal endpoint to delete all records related to a tenant ([MEN-7318](https://northerntech.atlassian.net/browse/MEN-7318))

- *(gui)* treat devices which didn't contact server after being accepted as offline ([MEN-6880](https://northerntech.atlassian.net/browse/MEN-6880))
- *(gui)* treat devices without update\_ts as offline
- *(gui)  The new "$ltne" filter operator allows to get list of device where the update\_ts is lower than given value or update\_ts doesn't exist. ([MEN-6880](https://northerntech.atlassian.net/browse/MEN-6880))
- *(gui)* Added UI interface to save the Open ID connect Single sign-on ([MEN-6922](https://northerntech.atlassian.net/browse/MEN-6922))
- *(gui)* allow to save SAML Single Sign-On without config providing
- *(gui)* allow personal access tokens generation for the SSO users ([MEN-6824](https://northerntech.atlassian.net/browse/MEN-6824))
- *(gui)* Added releases quick actions support ([MEN-6859](https://northerntech.atlassian.net/browse/MEN-6859))
- *(gui)* added two-step login for enterprise users ([MEN-6823](https://northerntech.atlassian.net/browse/MEN-6823))
- *(gui)* enabled password reset during user creation ([MEN-7192](https://northerntech.atlassian.net/browse/MEN-7192))
- *(gui)* use inventory's check\_in\_time to extract and list offline devices ([MEN-7251](https://northerntech.atlassian.net/browse/MEN-7251))
- *(gui)* gave device deployment log files more descriptive file names ([MEN-7221](https://northerntech.atlassian.net/browse/MEN-7221))
- *(gui)* made log viewer wider to ease going through deployment logs ([MEN-7220](https://northerntech.atlassian.net/browse/MEN-7220))
- *(gui)* added copyable userid to user information ([MEN-7277](https://northerntech.atlassian.net/browse/MEN-7277))
- *(gui)* allowed adding users by user id in user creation dialog ([MEN-7277](https://northerntech.atlassian.net/browse/MEN-7277))
- *(gui)* restructured account menu & added option to switch tenant in supporting setups ([MEN-6906](https://northerntech.atlassian.net/browse/MEN-6906))
- *(gui)* let device details remain open when adding the device to a group ([MEN-7336](https://northerntech.atlassian.net/browse/MEN-7336))
- *(gui)* added notification about changes to the device offline threshold ([MEN-7288](https://northerntech.atlassian.net/browse/MEN-7288))

- *(inventory)* do not set updated\_ts field when inserting the device ([MEN-6878](https://northerntech.atlassian.net/browse/MEN-6878))

- *(iot-manager)* process webhook requests asynchronously, returing `202 Accepted` instead of `204 No Content` or `200 OK` ([MEN-7227](https://northerntech.atlassian.net/browse/MEN-7227))
- *(iot-manager)* add a timeout for webhook requests, defaults to 10 seconds; you can modify it using the `webhooks_timeout_seconds` configuration setting ([MEN-7227](https://northerntech.atlassian.net/browse/MEN-7227))
- *(iot-manager)* internal endpoint to delete all records related to a tenant ([MEN-7319](https://northerntech.atlassian.net/browse/MEN-7319))

- *(workflows)* Add encoding option "html" for html-escaping string parameters ([MEN-7003](https://northerntech.atlassian.net/browse/MEN-7003))



### Build

- (gui) **BREAKING**: Changed container image to unprivileged port 8090 and unprivileged user ([13b2268](https://github.com/mendersoftware/mender-server/commit/13b2268027f678e52ce69aa8bfa912c713d12093)) by @alfrunes

- (docker) **BREAKING**: Changed container image tag scheme from `mender-x.y.z` to `vX.Y.Z`

  The new versioning scheme uses the **Mender Server** version which is decoupled from the other components in the Mender ecosystem.

- *(docker)* Add build stage to Dockerfiles
 ([ba3692e](https://github.com/mendersoftware/mender-server/commit/ba3692eb52dd146081014de05ffae9b6331d6ff6))  by @alfrunes


  The Dockerfiles are now self-contained by moving the build stage into
  the Dockerfile.

- *(docker)* Build images on BUILDPLATFORM
 ([44e5b7f](https://github.com/mendersoftware/mender-server/commit/44e5b7f574f9437ccad0954dafbf75ea78f511f3))  by @alfrunes

- *(docker)* Use make(1) when building inside docker images
 ([153269e](https://github.com/mendersoftware/mender-server/commit/153269e6916ddfde68d4846621cd87a8b89c3dcf))  by @alfrunes


  For consistent builds.

- *(make)* Update `docker` targets to use updated Dockerfiles
 ([11f26d6](https://github.com/mendersoftware/mender-server/commit/11f26d6684f4732e3003cac69b4d03b11ae9924c))  by @alfrunes


  Refactored common parts to parent directory.

- *(make)* Change default target to `docker` and add variable TAGS
 ([92ac12a](https://github.com/mendersoftware/mender-server/commit/92ac12a512235c8b0013513935fe63bb712fea85))  by @alfrunes


  Containers are the primary build artifacts for this repo so it makes
  sense to build them by default.

- *(make)* Fix acceptance test targets after refactor
 ([70919bd](https://github.com/mendersoftware/mender-server/commit/70919bdda6ca3af15bdd61f3cbe03f01156fe71c))  by @alfrunes


  Put common acceptance test targets in Makefile.common and made
  exceptions for `create-artifact-worker` and `reporting`.

- *(make)* Do not expand `go` shell commands unconditionally
 ([1c68b83](https://github.com/mendersoftware/mender-server/commit/1c68b833d7823f6b0666e0f290ea0c073f482634))  by @alfrunes

- *(make)* Run acceptance tests without rebuilding the containers
 ([ce241cc](https://github.com/mendersoftware/mender-server/commit/ce241cc9cb5da92372787fbcd236a8b43c705242))  by @alfrunes


  Removed the dependency on `docker-acceptance` for the
  `test-acceptance-run`.

- *(make)* `test-unit` target runs in same environment as build
 ([141ea40](https://github.com/mendersoftware/mender-server/commit/141ea401f90b63d6667d403743b294908fb662b1))  by @alfrunes

- *(make)* Change TAGS behavior to always include required build tags
 ([5bae608](https://github.com/mendersoftware/mender-server/commit/5bae608e706384cd5ebb88dbd6869291c1fa33d8))  by @alfrunes

- *(make)* Rename DOCKER_ARGS to DOCKER_BUILDARGS, TAGS to BUILDTAGS
 ([1a97891](https://github.com/mendersoftware/mender-server/commit/1a978917316ce620ecc0fee01002ece409437765))  by @alfrunes


  It seems like Gitlab has a built in TAGS env variable which conflicts
  with the Make environment.

- *(make)* Add `docker-pull` target for pulling images
 ([37f4391](https://github.com/mendersoftware/mender-server/commit/37f4391dee7624d583c68aba6235f7f61deddd76))  by @alfrunes

- *(make)* Define DOCKER_PLATFORM template as multiline variable
 ([0db0c9a](https://github.com/mendersoftware/mender-server/commit/0db0c9ad67430ba3bbe109cb9d5e8613eb2af621))  by @alfrunes

- *(make)* Fix tag override for `docker-acceptance`
 ([7f0b260](https://github.com/mendersoftware/mender-server/commit/7f0b26011d422b539cdbdf6acbcb22156ac6a076))  by @alfrunes


  MENDER_IMAGE_TAG_TEST should set the tag when building the target.

- *(make)* Remove make 4.4 function `let`
 ([75f980e](https://github.com/mendersoftware/mender-server/commit/75f980eb118298e0fcaf3dc08dcda0eb46131541))  by @alfrunes

- *(make)* Added target `docker-publish` for publishing images
 ([c400b04](https://github.com/mendersoftware/mender-server/commit/c400b04c2c30a4a2e2ac5a2e2f53fddf44caeffb))  by @alfrunes

- *(make)* Split MENDER_PUBLISH_REGISTRY into registry and repository
 ([e27c770](https://github.com/mendersoftware/mender-server/commit/e27c77001704218b36442641598e3b3d1ef65fd1))  by @alfrunes

- *(test)* Force serialize unit tests for deviceauth
 ([a0ab55e](https://github.com/mendersoftware/mender-server/commit/a0ab55e6f84242b34b176520b7c2218e1c1d678e))  by @alfrunes





### Check


- Make `sed(1)` Linux compatible again
 ([1271396](https://github.com/mendersoftware/mender-server/commit/1271396faf49b2220a39cd5e4ef7a1cd1cdee443))  by @alfrunes


  Replace flag `-i=''` with `-i.bak` and removing the files.




### Refac


- *(compose)* Refactor SeaweedFS topology and optimize startup/shutdown time
 ([fe7ee2e](https://github.com/mendersoftware/mender-server/commit/fe7ee2eb37177521cb9ab060f551b2d441c86874))  by @alfrunes


  Instead of running SeaweedFS as a monolith using the `server` command,
  we explicitly launch all services in different containers. It appears
  that the `server` command has some issues when initializing the master
  and sometimes enter a deadlock the healthcheck interval is too low at
  startup. Moreover, running the services in different containers makes it
  easier to debug and interpret the logs.

- *(iotcore)* Break on errors instead of falling through
 ([733f8ab](https://github.com/mendersoftware/mender-server/commit/733f8ab45cc8a76566b79d86507306b78742f324))  by @alfrunes


  Using long chains of fallthrough error conditions makes it very
  difficult to read and error prone to extend. Refactoring to use common
  coding patterns instead.

- Use an overlay directory to create Makefiles and Dockerfiles
 ([85e93e0](https://github.com/mendersoftware/mender-server/commit/85e93e0081410b1fa295ddd6b639af969d3f47ac))  by @alfrunes


  Allows for easier individual customization required for accepatnce
  tests.
- Move compat tests to dedicated test suite
 ([059f437](https://github.com/mendersoftware/mender-server/commit/059f4375d3b33073e711fdbf81212a2cd5dacfbb))  by @alfrunes




### Revert


- *(docker)* Revert generate-delta-worker dockerfile
 ([d205b3e](https://github.com/mendersoftware/mender-server/commit/d205b3e4a279da0ee9660f3e14f828012ff0f5ab))  by @alfrunes


  Reverts the dockerfile to the upstream docker file with the two
  exceptions of copying the binaries from this repositories rather than
  relying on master docker images.

- Change docker entrypoint to launch workflows worker
 ([0d39c96](https://github.com/mendersoftware/mender-server/commit/0d39c964eeec8b833604b91be7f707c7ce1357a4))  by @alfrunes


  This was done by mistake when updating the Dockerfile for the monorepo.


