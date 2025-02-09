# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2025.xx.xx

Alot of breaking changes. Biggest things:

* Introduced `Field` struct to be used instead of `modbus.BField`. `BField` and fluent interface is deprecated. When
  adding fields in bulk and dynamically the struct (more idiomatic) approach is more ergonomic than fluent API.
* Added implementation of Poller in `poller` package. For example usage
  see [cmd/modbus-poller/main.go](cmd/modbus-poller/main.go)

Breaking changes to following structs/methods/functions

* struct field `modbus.Field.RegisterAddress` was renamed to `Address`
* struct `modbus.RegisterRequest` was renamed to `BuilderRequest`
* method `BuilderRequest.ExtractFields()` signature changed
* removed type `packet.LooksLikeType` and related consts
    * const `packet.DataTooShort`
    * const `packet.IsNotTPCPacket`
    * const `packet.LooksLikeTCPPacket`
    * const `packet.UnsupportedFunctionCode`
* error `modbus.ErrClientNotConnected`: changed from `ClientError` to `*ClientError`
* error `modbus.ErrPacketTooLong`: changed from `ClientError` to `*ClientError`

### Added

* Added FC1/FC2 support to builder. You can register coils with `b.Coild(address)` to be requested and extracted.
  Builder has now following methods for splitting:
    * `ReadCoilsTCP` combines fields into TCP Read Coils (FC1) requests
    * `ReadCoilsRTU` combines fields into RTU Read Coils (FC1) requests
    * `ReadDiscreteInputsTCP` combines fields into TCP Read Discrete Inputs (FC2) requests
    * `ReadDiscreteInputsRTU` combines fields into RTU Read Discrete Inputs (FC2) requests

## [0.2.0] - 2024.02.04

### Added

* Added support for FC17 (0x11) Read Server ID.
* Added `packet.LooksLikeModbusTCP()` to check if given bytes are possibly TCP packet or start of packet.
* Added `Parse*Request*` for every function type to help implement Modbus servers.
* Added `Server` package to implement your own modbus server

## [0.0.1] - 2021-04-11

### Added

* First implementations
