# Changelog
## Unreleased
### Changed
- Gracefully handle the case where `elasticsearch_index_template` objects exist in the terraform state but not in the ES domain (e.g. because they were manually deleted.)
- Create index `aliases` and `mappings` even if no settings are set.
- Bump aws client to v1.35.33.

### Added
-

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
