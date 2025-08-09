# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2025-08-09

### Added
- Initial release of gin-pprof
- Dynamic profiling middleware for Gin web framework
- Support for CPU, heap, and goroutine profiling
- File-based configuration with YAML format
- Nacos configuration center integration
- Smart sampling with configurable rates
- Concurrency control and resource management
- Built-in monitoring and statistics endpoints
- File storage backend with automatic cleanup
- Memory storage backend for testing
- Comprehensive documentation and examples
- Production-safe design with non-blocking operations

### Features
- Route-aware profiling with parameter support (e.g., `/users/:id`)
- Real-time configuration updates via Nacos
- Automatic profile file cleanup
- Detailed logging and error handling
- Builder pattern API for easy configuration
- Multiple profiler types with extensible architecture

### Examples
- Basic file-based configuration example
- Nacos integration example
- Advanced configuration with custom options
- Complete API application example

### Documentation
- Comprehensive README with quick start guide
- Getting started guide
- Configuration reference
- API documentation
- Production deployment best practices