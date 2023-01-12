# Changelog
How to release a new version:
- Update this file with the latest version.
- Manually release new version.

## [Unreleased]

### Fixed
- JSON tag for `MaxHeaderBytes` field in `Limits` configuration.

## [0.3.0] - 2023-01-09
### Added
- HTTP response writer contains error field.

### Changed
- `LoggingMiddleware` logs `err` field with message if error is present.
- Updated packages:
  ```diff
  - go.strv.io/time: v0.1.0
  + go.strv.io/time: v0.2.0
  ```

## [0.2.0] - 2022-08-22
### Added
- HTTP response writer implements hijacking to support web sockets.

## [0.1.0] - 2022-08-01
### Added
- Added Changelog.

[Unreleased]: https://github.com/strvcom/strv-backend-go-net/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/strvcom/strv-backend-go-net/releases/tag/v0.3.0
[0.2.0]: https://github.com/strvcom/strv-backend-go-net/releases/tag/v0.2.0
[0.1.0]: https://github.com/strvcom/strv-backend-go-net/releases/tag/v0.1.0
