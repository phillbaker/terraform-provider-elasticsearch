# Changelog

## [Unreleased]

## [1.0.0] - 2020-03-18

### Added
- Ability to import xpack users, role mapping and roles (#58, #59)
  This includes changes to the role `field_security` field. In order to upgrade to this version, a  do a `terraform state rm` and then a `terraform import` of the resource may be necessary.

### Changed
- Move source files into specific package.
