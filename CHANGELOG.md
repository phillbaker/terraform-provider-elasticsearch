# Changelog

## [Unreleased]
### Deprecated
- Datasource: elasticsearch_{destination} in favor of elasticsearch_opendistro_{destination}

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
