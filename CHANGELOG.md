---
## 4.0.0-rc.7 - 2025-01-09


### Bug Fixes


- *(gui)* Fixed an issue that would prevent showing deployment reports for phased deployments
 ([132d6b2](https://github.com/mendersoftware/mender-server/commit/132d6b2aa932924c6612a8ca39867f246b388a88))  by @mzedel

- *(gui)* Fixed an issue that would prevent upgrading a running session to a different plan
([MEN-7898](https://northerntech.atlassian.net/browse/MEN-7898)) ([7668b29](https://github.com/mendersoftware/mender-server/commit/7668b293bc71d4eaffda0c00823cac5026dfbf4c))  by @mzedel

- *(gui)* Fixed an issue that would crash the site when showing release details with multiple artifacts
 ([fd06f66](https://github.com/mendersoftware/mender-server/commit/fd06f66516e4f33e52c9f305fb058594eecf714f))  by @mzedel

- *(gui)* Made addon availability rely more on addons where possible to prevent erroneous device config retrieval
([MEN-7895](https://northerntech.atlassian.net/browse/MEN-7895)) ([62d6516](https://github.com/mendersoftware/mender-server/commit/62d6516b848b35e493db8be9908abaf7e573e008))  by @mzedel

- Implement signal handler for `server` commands
([QA-782](https://northerntech.atlassian.net/browse/QA-782)) ([6e17ada](https://github.com/mendersoftware/mender-server/commit/6e17adaaffa6778dc021353248d83b08cf645182))  by @bahaa-ghazal




### Features


- *(gui)* Aligned notions of "latest device activity" in listing & details
 ([40ee57d](https://github.com/mendersoftware/mender-server/commit/40ee57da173c1d5bc2a39b2a5534b62dc986f53c))  by @mzedel

- *(gui)* Limited onboarding to hosted Mender to ensure a streamlined experience
([MEN-7896](https://northerntech.atlassian.net/browse/MEN-7896)) ([cee60f8](https://github.com/mendersoftware/mender-server/commit/cee60f8de9eb5b6940828f5e09e11dbd4e2d4059))  by @mzedel








## 4.0.0-rc.5 - 2024-12-24


### Bug Fixes


- Stop user from having similar email and password
([MEN-6462](https://northerntech.atlassian.net/browse/MEN-6462)) ([3fa4a43](https://github.com/mendersoftware/mender-server/commit/3fa4a432780a40fb9b8c37633c7feca6ba3445c5))  by @bahaa-ghazal






## v4.0.0-rc.3 - 2024-12-17


### Bug Fixes


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

- Fixed an issue that prevented the UI from showing deeply nested software installations
([MEN-7640](https://northerntech.atlassian.net/browse/MEN-7640)) ([13496f3](https://github.com/mendersoftware/mender-server/commit/13496f3468fd08dcc9656ba07463eba682cfaff4))  by @mzedel




### Documentation


- *(README)* Add step to clone repository
 ([f9d3bbd](https://github.com/mendersoftware/mender-server/commit/f9d3bbde382bca4592f41e3d6be7e8292dcb221f))  by @alfrunes

- *(README)* Consistently add syntax highlighting to code blocks
 ([8583102](https://github.com/mendersoftware/mender-server/commit/8583102cbbf49882b9a9ab1b80257516ec13dc24))  by @alfrunes

- Update README.md
 ([f7a1b09](https://github.com/mendersoftware/mender-server/commit/f7a1b097726672dd40ed7df17551229c5cf6ce7f))  by @alfrunes


  Adjusted styling (note color, added 1st level indentation,  taxonomy i.e., Mender Server, Mender Enterprise) to make it easy to follow and read.
- Document how to bring up the Virtual Device for enterprise setup
 ([c674566](https://github.com/mendersoftware/mender-server/commit/c674566e6d834c64d6e64d321c6e09b5f2a36259))  by @lluiscampos




### Features


- *(deployments)* New endpoint for getting release by name
([MEN-7575](https://northerntech.atlassian.net/browse/MEN-7575)) ([3a18e88](https://github.com/mendersoftware/mender-server/commit/3a18e880ec5cddedc19ed08949777caedda4350d))  by @kjaskiewiczz

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





### Revert


- "fix(gui): fixed an issue that caused number comparisons in device filters to not work"
 ([787237e](https://github.com/mendersoftware/mender-server/commit/787237ec8689d96c73beefbc74bcea7b96b274ba))  by @mzedel


  This reverts commit 84e2398fece6b10fddcf6f60e3ff744af903c707.
  Signed-off-by: Manuel Zedel <manuel.zedel@northern.tech>
- Revert "docs(deviceauth): migration to OpenAPI3"
 ([93ab08a](https://github.com/mendersoftware/mender-server/commit/93ab08ab6051aec3508bb550a4455d30ba2a9b56))  by @kjaskiewiczz


  This reverts commit f7a33e9a71339522ee33f3808e7d6a8598144d2a.





---
