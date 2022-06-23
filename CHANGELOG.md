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
### Remaining 
- document README.md to instruct users how to run the example
- add ARCHITECTURE.md to explain code layout


## [0.6.0] - 2022-06-22
### Added
- `tests/api_test.go` integration tests to show the system is working

## [0.5.0] - 2022-06-22
### Added 
- `limits` package to foundation it handles the main logic
- `app/api/middle/limiter` middleware package that integrates rate limiter into fiber
- `conf` added configuration to control middleware
- `construct` added logic in constructors to configure and wire up middleware

## [0.4.0] - 2022-06-19
### Added
- kubernetes support in `infra/k8s`
- currently only supporting `kind`
- `env_sample` to document env configurations for the system
- Readme section for installation

### Fixed
- Makefile commands supporting kubernetes management


## [0.3.0] - 2022-06-19
### Added
- ping package to handle route 'GET /ping' the route our example will be using

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