# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.0.2] - unreleased

### Added

* Added support for FC17 (0x11) Read Server ID.
* Added `packet.LooksLikeModbusTCP()` to check if given bytes are possibly TCP packet or start of packet.
* Added `Parse*Request*` for every function type to help implement Modbus servers.
* Added `Server` package to implement your own modbus server

## [0.0.1] - 2021-04-11

### Added

* First implementations
