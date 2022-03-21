# Changelog
## Unreleased
### Changed

### Added

### Fixed

## [2.0.0] - 2022-03-13
### Changed
- [index] Remove primary shards default so templates can specify.

### Added
- [provider] Allow configurable service name in AWS signatures (#254)
- [kibana alert] A `params_json` attribute to allow all alert types

### Fixed
- [provider] Fix interface panic conversions in defaultHttpClient and tokenHttpClient (#255)

## [2.0.0-beta.4] - 2022-01-29
### Changed
- Fix renamed opensearch resources to start with provider name

### Added
- Recognize Opensearch version without compatibility mode.
- [index] Support number_of_routing_shards setting

## [2.0.0-beta.3] - 2022-01-07
### Changed
- [build] Clarify that a minimum of Go 1.13 is required to build the provider
- [opendistro *] Rename elasticsearch_opendistro_* resources to opensearch_*. Test OpenSearch against 1.x.

### Added
- [provider] Output ES Client logs in JSON format, depending on TF_LOG_PROVIDER
- [index] Add support for rollover alias in opensearch.

### Fixed
- [composable_index_template, kibana_alert] Use provider's version, instead of pinging the server.


## [2.0.0-beta.2] - 2021-09-25
### Changed
- [provider] Change default for sniffing to false, see https://github.com/phillbaker/terraform-provider-elasticsearch/pull/161.
- [aws] Reuse session options, ensure synchronization before using credentials, see https://github.com/phillbaker/terraform-provider-elasticsearch/issues/124.

### Added
- [index] Add include_type_name for compatibility between ESv6/7
- [xpack license] Handle ackowledged only reponse.
- [kibana alert] Fix storing actions, missing descriptions.

### Fixed
- [provider] Add a timeout for pinging ES in case of no network access.


## [2.0.0-beta.1] - 2021-08-30
### Changed
- Upgraded to Terraform plugin SDK v2. This removes support for Terraform 0.11 and earlier.
- Remove deprecated resources and attributes.
- Drop elasticsearch v5 support.


## [1.6.3] - 2021-08-29
### Added
- [opendistro tenant] Retry to avoid conflicts


## [1.6.2] - 2021-08-08
### Added
- [xpack user] Add support for elasticsearch v6 (#205)


## [1.6.1] - 2021-07-20
### Changed
- [kibana object] Diffs will now be detected and the stringified version of kibana objects saved to terraform state (#182).

### Added
- [index] Add normalizer and filter attributes
- [provider] Add host_override parameter to allow connections via SSH tunnel when using sign_aws_requests = true. Add default http client and set ServerName for other clients (#203)
- [aws] Correctly pass through insecure setting in awsHttpClient (#200)
- [aws] Also load shared config for a profile (#196)

### Fixed
- [host] Fix url attribute being empty
- [opendistro tenant] Fix casing of opendistro tenant
- [index] Fix updates on index settings with '.' in name (#198)


## [1.6.0] - 2021-07-03
### Changed
- [aws] Always enable AWS shared configuration file support

### Added
- [opendistro user,role,ism policy] Retry handling for conflicts
- [kibana alert] Add kibana alert resource
- [opendistro tenant] Add computed attribute for index


## [1.5.8] - 2021-06-25
### Added
- [component template] Resource to manage component templates.
- [index] Add attributes for analysis settings.

### Fixed
- [aws] Fix concurrent map writes crash when using AWS assume role.


## [1.5.7] - 2021-06-06
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


## [1.5.6] - 2021-05-10
### Changed
- [opendistro destination] Use API to get in odfe>=1.11.0 (#158)

### Added
- [aws auth] Set pass profile on assume role

### Fixed
- [opendistro ism policy] Fix perpetual diff in error_notification, only delete the attribute if it's null. (#165)
- [opendistro monitor] Normalize IDs


## [1.5.5] - 2021-04-06
### Changed
- Updated AWS client to v1.37.0 for AWS SSO auth using `aws_provider_profile` (#162)

### Added
- Support for specifying Authorization header (Bearer or ApiKey) to authenticate to elasticsearch.

### Fixed
- [opendistro ism] Retry on 409.


## [1.5.4] - 2021-03-17
### Fixed
- [opendistro destination] normalize destination for nested "id" key in newer versions of ES (#148)
- [index] Handle not found on resource read
- [opendistro role] Fix crash on import (#150)
- [opendistro/xpack user] Fix user update leading to the password being set to the hashed value
#157


## [1.5.3] - 2021-02-18
### Changed
- [xpack_user,opendistro_user] Hash user passwords in state to detect change (#132)

### Added
-  Open Distro Kibana tenant resource (#144)


## [1.5.2] - 2021-02-05
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
