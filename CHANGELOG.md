# Changelog
How to release a new version:
- Update this file with the latest version.
- Manually release new version.

## [Unreleased]

## [0.7.0] - 2024-03-11
### Changed
- Logging interface changes to `log/slog`.

  NOTE: This is version presents a BREAKING CHANGE in the server's logger interface. The server now accepts a `slog.Logger` instead of a custom `ServerLogger`.
- Change in the logging middleware. The middleware now nests request specific data under the "request" group.
- Updated from Go 1.20 to Go 1.22.
- Updated packages:
  ```diff
  - github.com/go-chi/chi/v5 v5.0.8
  - github.com/google/uuid v1.3.0
  - github.com/stretchr/testify v1.8.0
  + github.com/go-chi/chi/v5 v5.0.12
  + github.com/google/uuid v1.6.0
  + github.com/stretchr/testify v1.9.0
  ```

## [0.6.2] - 2023-06-27
### Fixed
- Error logging when terminating HTTP server.

## [0.6.1] - 2023-03-28
### Changed
- package `http/param` does not zero the field if not tagged with any relevant tags

## [0.6.0] - 2023-03-03
### Added
- package `http/signature` to simplify defining http handler functions
- package `http/param` to simplify parsing http path and query parameters

## [0.5.0] - 2023-01-20
### Added
- `ErrorResponseOptions` contains public error message.
- `ErrorResponseOptions` contains request ID.
- Error response options:
  - `WithErrorMessage`
  - `WithRequestID`

## [0.4.0] - 2023-01-12
### Changed
- JSON tags in `ErrorResponseOptions`.

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

[Unreleased]: https://github.com/strvcom/strv-backend-go-net/compare/v0.7.0...HEAD
[0.7.0]: https://github.com/strvcom/strv-backend-go-net/compare/v0.6.2...v0.7.0
[0.6.2]: https://github.com/strvcom/strv-backend-go-net/compare/v0.6.1...v0.6.2
[0.6.1]: https://github.com/strvcom/strv-backend-go-net/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/strvcom/strv-backend-go-net/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/strvcom/strv-backend-go-net/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/strvcom/strv-backend-go-net/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/strvcom/strv-backend-go-net/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/strvcom/strv-backend-go-net/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/strvcom/strv-backend-go-net/releases/tag/v0.1.0
