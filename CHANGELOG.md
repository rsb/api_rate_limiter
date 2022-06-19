# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

### Types of Changes
- `Added`: for new features.
- `Changed`:  for changes in existing functionality.
- `Deprecated`: for soon-to-be removed features.
- `Removed`:  for now removed features.
- `Fixed`: or any bug fixes.
- `Security`: in case of vulnerabilities.

## [Unreleased]


## [0.2.0] - 2022-06-19
# Added
- dockerfile for limiter api
- wired up serve command `limter api serve` will start up web server
- no routes or limiter yet

## [0.1.0] - 2022-06-17
### Added
- app, conf, construct, cli and api packages
- added health handlers for kubernetes health checks
- Makefile to help with kubernetes and basic development