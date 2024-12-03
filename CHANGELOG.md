---
## v4.0.0-rc.3 - 2024-12-03


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




### Features


- *(deployments)* New endpoint for getting release by name
([MEN-7575](https://northerntech.atlassian.net/browse/MEN-7575)) ([3a18e88](https://github.com/mendersoftware/mender-server/commit/3a18e880ec5cddedc19ed08949777caedda4350d))  by @kjaskiewiczz

- *(gui)* Added the possibility to create service provider administering roles
([MEN-7570](https://northerntech.atlassian.net/browse/MEN-7570)) ([92d7e50](https://github.com/mendersoftware/mender-server/commit/92d7e50e311d8c88f9847a83cec7b797ef9cebcc))  by @mzedel

- *(gui)* Aligned role removal dialog with other parts of the UI
 ([8661704](https://github.com/mendersoftware/mender-server/commit/866170425bef1f01f3a4a25f0d4e19fe5da94a6e))  by @mzedel

- *(gui)* Added support for Personal Access Token auditlog entries
([MEN-7622](https://northerntech.atlassian.net/browse/MEN-7622)) ([9a9a6c3](https://github.com/mendersoftware/mender-server/commit/9a9a6c3829611c35622e3812db7bbedd9bc9f9e5))  by @mzedel






---
