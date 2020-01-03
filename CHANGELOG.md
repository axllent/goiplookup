# Changelog

## [0.2.2]

- Add support for MaxMind (free) license key requirement to update GeoLite2 databases
- Switch to [ghru](https://github.com/axllent/ghru) for binary release updates


## [0.2.1]

- Fix dataDir flag parsing


## [0.2.0]

- Switch to go modules (go >= 1.11 required)
- Switch to [pflag](github.com/spf13/pflag) for more flexibility


## [0.1.0]

- Add corruption check for downloaded databases before overwriting
- Code cleanup


## [0.0.5]

- Better version feedback and update information
- Strip binaries with `-ldflags "-s -w"`


## [0.0.4]

- Fix bug whereby executable path wasn't detected on `self-update`


## [0.0.3]

- Split project into multiple files
- Add self-updater, `-v` will return current version plus latest if there is an update
- Support for darwin (Mac)
- Release multiple OS binaries


## [0.0.2]

- Add README.md
- Use `db-update` for update flag


## [0.0.1]

- Rename to goiplookup
- Add database update functionality
- Tidy code and basic versioning
- Return IP Address not found
- Update build script to use tags
- Usage to include hostname
- Fix debug info
