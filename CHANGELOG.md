---
## 4.0.0 - 2024-10-30


### Bug Fixes


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




### Documentation


- Update README.md with instructions on using the docker composition
 ([c9aa7dc](https://github.com/mendersoftware/mender-server/commit/c9aa7dc73db9717f436def2b14dc8f7cce74903f))  by @alfrunes
- Add section about testing build artifacts
 ([5c7eaaa](https://github.com/mendersoftware/mender-server/commit/5c7eaaa22c9cdb0016b9f06afc3bccdfdd9b9578))  by @alfrunes
- Add snippet for starting a mender client to README
 ([a322b2d](https://github.com/mendersoftware/mender-server/commit/a322b2d8c22e5e9e409a9a5c96c3b8412cef555f))  by @alfrunes




### Features


- *(deployments)* Add filter field to deployment object
([MEN-7416](https://northerntech.atlassian.net/browse/MEN-7416)) ([fec5b91](https://github.com/mendersoftware/mender-server/commit/fec5b91d59d07b1a0d85ccf077cd56aa5b192278))  by @kjaskiewiczz


  The filter field contains information about devices targeted by the
  deployment.

- *(gui)* Enabled webhook scope selection
([MEN-7455](https://northerntech.atlassian.net/browse/MEN-7455)) ([cec277d](https://github.com/mendersoftware/mender-server/commit/cec277d83adf930de47ca5bb03935aa350ea1af5))  by @mzedel

- *(gui)* Extended webhook event details
([MEN-7574](https://northerntech.atlassian.net/browse/MEN-7574)) ([0bfda40](https://github.com/mendersoftware/mender-server/commit/0bfda409122ed6837e13cf7f5418b093bf3ef97b))  by @mzedel

- *(gui)* Aligned webhook listing with updated design
([MEN-7573](https://northerntech.atlassian.net/browse/MEN-7573)) ([80e55d1](https://github.com/mendersoftware/mender-server/commit/80e55d15e361c21988e192bf715a219bb300f487))  by @mzedel

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




### Build


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




### Revert


- *(docker)* Revert generate-delta-worker dockerfile
 ([d205b3e](https://github.com/mendersoftware/mender-server/commit/d205b3e4a279da0ee9660f3e14f828012ff0f5ab))  by @alfrunes


  Reverts the dockerfile to the upstream docker file with the two
  exceptions of copying the binaries from this repositories rather than
  relying on master docker images.

- Change docker entrypoint to launch workflows worker
 ([0d39c96](https://github.com/mendersoftware/mender-server/commit/0d39c964eeec8b833604b91be7f707c7ce1357a4))  by @alfrunes


  This was done by mistake when updating the Dockerfile for the monorepo.




---
