# Changelog

## [Unreleased]

## [1.1.1] - 2020-05-09
### Added
- Make ping to elasticsearch during provider config optional

## [1.1.0] - 2020-05-04

### Added
- Add OpenDistro Index State Management (https://opendistro.github.io/for-elasticsearch-docs/docs/ism/api/).
- Add OpenDistro Roles and Role Mappings.

### Changed
- Clarify naming, watch resources are from xpack.


## [1.0.0] - 2020-03-18

### Added
- Ability to import xpack users, role mapping and roles (#58, #59)
  This includes changes to the role `field_security` field. In order to upgrade to this version, a  do a `terraform state rm` and then a `terraform import` of the resource may be necessary.

### Changed
- Move source files into specific package.
