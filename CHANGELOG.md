# Changelog
## Unreleased
### Changed

### Added

### Fixed


## [1.5.7] - 2020-06-06
### Changed
- [index] Refactor to use flattened settings, prepare for adding more settings.
- [opendistro ISM policy mapping] Poll for updates to mapped indices.
- [opendistro ISM policy mapping] Mark as deprecated based on ODFE marking `opendistro.index_state_management.policy_id` as deprecated starting from version 1.13.0.

### Added
- [xpack watcher] Ability to deactivate a watcher. (#173)
- [index] Add explicit attributes for many index settings
- [auth] Support configuring provider token from environment (#180)
- [build] Bump go build version to 1.16 for apple silicon
- [opendistro ism] Support opendistro ISM policy for elasticsearch 6.8 (#152)

### Fixed
- [opendistro ism policy] Fix perpetual diff in ism_template if not specified
- [opendistro ism mapping] Fix provider crash when using ODFE >= 1.13.0
- [aws] Fix regression in 1.5.1 that broke the use of `aws_assume_role_arn` (#124)


## [1.5.6] - 2020-05-10
### Changed
- [opendistro destination] Use API to get in odfe>=1.11.0 (#158)

### Added
- [aws auth] Set pass profile on assume role

### Fixed
- [opendistro ism policy] Fix perpetual diff in error_notification, only delete the attribute if it's null. (#165)
- [opendistro monitor] Normalize IDs


## [1.5.5] - 2020-04-06
### Changed
- Updated AWS client to v1.37.0 for AWS SSO auth using `aws_provider_profile` (#162)

### Added
- Support for specifying Authorization header (Bearer or ApiKey) to authenticate to elasticsearch.

### Fixed
- [opendistro ism] Retry on 409.


## [1.5.4] - 2020-03-17
### Fixed
- [opendistro destination] normalize destination for nested "id" key in newer versions of ES (#148)
- [index] Handle not found on resource read
- [opendistro role] Fix crash on import (#150)
- [opendistro/xpack user] Fix user update leading to the password being set to the hashed value
#157


## [1.5.3] - 2020-02-18
### Changed
- [xpack_user,opendistro_user] Hash user passwords in state to detect change (#132)

### Added
-  Open Distro Kibana tenant resource (#144)


## [1.5.2] - 2020-02-05
### Changed
- [open distro role] Rename fls in favor of field_level_security, deprecate fls.
- [aws auth] Revert bump aws client to v1.35.33, downgrade to v1.35.20
- [aws auth] Revert pass profile on assume role.

### Added
- [open distro role] Add support for OpenDistro document-level-security in role.

### Fixed
- [composable template] Fix elasticsearch_composable_index_template with Elasticsearch 7.10 (#134)


## [1.5.1] - 2020-12-23
### Changed
- Gracefully handle the case where `elasticsearch_index_template` objects exist in the terraform state but not in the ES domain (e.g. because they were manually deleted.)
- Create index `aliases` and `mappings` even if no settings are set.
- Bump aws client to v1.35.33.
- Allow provider variable interpolation by deferring client instantiation, `providerConfigure` only returns a configuration struct.
- Fix XPack license resource having perpetual diff if using basic license.
- [aws auth] Pass profile on assume role.
- [aws auth] Pass down the `insecure` provider parameter to allow skipping TLS verification.
- [index] Return errors in the case where index definitions are invalid.

### Added
- Composable Index Template resource, available in ESv7.8+

## [1.5.0] - 2020-10-26
### Changed
- Deprecated elasticsearch_index_lifecycle_policy in favor of elasticsearch_xpack_index_lifecycle_policy.
- Don't recreate indices which are managed by ILM or ISM for the resource `elasticsearch_index` (#75).
- Fix opendistro ism policy version conflict

### Added
- Validation for kibana objects to prevent provider crashes.
- Add XPack license resource, compatible with ES version >= 6.2.
- Add `assume_role` support for authenticating to AWS


## [1.4.4] - 2020-09-23
### Added
- Add OpenDistro user resource (#83)


## [1.4.3] - 2020-09-10
### Changed
- Fixed issue where diffs may not be detected for the resources `elasticsearch_kibana_object`, `elasticsearch_opendistro_destination`, `elasticsearch_opendistro_ism_policy_mapping`, `elasticsearch_opendistro_monitor`, `elasticsearch_xpack_role`, `elasticsearch_xpack_role_mapping`, and data source `elasticsearch_opendistro_destination` (#65).


## [1.4.2] - 2020-09-03
### Changed
- Fixed import for resources: `elasticsearch_opendistro_ism_policy`, `elasticsearch_opendistro_role`, `elasticsearch_opendistro_roles_mapping`.
- Fixed diffs for sets in `elasticsearch_opendistro_role`.
- Allow omitting tenant permissions
- Fix `elasticsearch_xpack_watch` resource not detecting diffs outside of terraform (#65). The watch API may return default values that were not passed in the original request, e.g. for log actions, `"level": "info"`, which will result in a perpetual diff unless it's pulled into the definition.

### Added
- Allow specifying AWS region for URL signing explicitly


## [1.4.1] - 2020-07-27
### Changed
- Fixed build process.


## [1.4.0] - 2020-07-27
### Changed
- Releases are now built with goreleaser tooling and packaged as zip files. Binaries are built with `CGO_ENABLED=0` as per the recommendations of goreleaser/terraform, these correspond to the previous `_static` binaries.


## [1.3.0] - 2020-06-20
### Added
- Support Snapshot lifecycle management

### Deprecated
- Datasource: elasticsearch_{destination} in favor of elasticsearch_opendistro_{destination}

### Fixed
- Using date math in index names - an index resource is tied to the resolved index it is created with.


## [1.2.0] - 2020-05-31
### Added
- Add aws profile authentication option

### Fixed
- Fix error on only updating index.force_destroy.

### Changed
- Rename (backward compatible) elasticsearch_{monitor,destination} to be under opendistro, elasticsearch_opendistro_{monitor,destination}
- Bump terraform plugin sdk to 1.12.0.

### Deprecated
- Resource: elasticsearch_{monitor,destination} in favor of elasticsearch_opendistro_{monitor,destination}


## [1.1.1] - 2020-05-09
### Added
- Make ping to elasticsearch during provider config optional


## [1.1.0] - 2020-05-04

### Added
- Add OpenDistro Index State Management (https://opendistro.github.io/for-elasticsearch-docs/docs/ism/api/).
- Add OpenDistro Roles and Role Mappings.

### Changed
- Clarify naming, watch resources are from xpack.

### Deprecated
- elasticsearch_{watch} in favor of elasticsearch_xpack_watch


## [1.0.0] - 2020-03-18

### Added
- Ability to import xpack users, role mapping and roles (#58, #59)
  This includes changes to the role `field_security` field. In order to upgrade to this version:

  1. Run the following to generate a list of state remove and import commands:
        ```sh
        terraform state list | grep role\\.  > roles.txt
        for i in $(cat roles.txt); do
          id=$(terraform state show $i | egrep "id" | tr -d '"' | cut -d'=' -f 2)
          echo "terraform state rm $i"
          echo "terraform import $i $id"
        done
        ```
  2. Upgrade to the new provider.
  3. Copy and paste the state commands that were printed above.



### Changed
- Move source files into specific package.
