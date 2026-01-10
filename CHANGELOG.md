# Changelog

All notable changes to this project will be documented in this file.

## [0.6.0] - 2026-01-10

### 🚀 Features

- *(k8s)* Add client-go integration with deployment listing
- *(docs)* Init best practices
- *(tasks)* Add markdownlint
- *(conf)* Add SonarQube integration
- *(k8s)* Fix KUBECONFIG multi-path handling and add YAML output

### 🐛 Bug/Lint Fixes

- *(docs)* Update to more strict rules
- *(k8s)* Add Updated field and fix test constants
- Address CodeRabbit review issues
- *(test)* Add defer cleanup and improve test coverage

### 🚜 Refactor

- *(tests)* Extract string literals to constants
- *(tests)* Replace string literals with constants in client_test.go
- Reduce function complexity to meet Codacy requirements
- *(tests)* Simplify test files for tutorial project
- *(test)* Reduce cyclomatic complexity in TestFormatDeploymentJSON

### ⚙️ Miscellaneous Tasks

- Add AI specific rules
- *(ai)* Add claude folder
- *(docs)* Mark list deployments feature as completed

### Deps

- *(deps)* Bump k8s.io/client-go from 0.33.2 to 0.33.3 (#9)
- *(deps)* Bump github.com/valyala/fasthttp in the updates group (#8)
- *(deps)* Bump github.com/valyala/fasthttp in the updates group (#10)
- *(deps)* Bump k8s.io/client-go from 0.33.3 to 0.33.4 (#11)
- *(deps)* Bump the updates group with 3 updates (#13)
- *(deps)* Bump github.com/valyala/fasthttp in the updates group (#14)
- *(deps)* Bump k8s.io/client-go from 0.34.0 to 0.34.1 (#15)
- *(deps)* Bump the updates group across 1 directory with 3 updates (#18)

## [0.5.0] - 2025-07-07

### 🚀 Features

- *(k8s)* Add client-go integration with connection testing
- *(cmd)* Add list deployments command with namespace filtering and output formats
- *(ci)* Add CodeRabbit.ai config
- *(cmd)* Enhance namespace validation with DNS label compliance
- *(devcontainer)* Setup Kubernetes with Ansible playbooks and automation scripts (#5)

### 🐛 Bug/Lint Fixes

- *(client)* Add a constant instead RY
- *(client)* Update func name to match expression
- *(connection)* Add proper name
- *(k8s)* Improve TestConnection timeout handling and server version logging
- *(k8s)* Fixed the name of the example function for TestConnection
- *(task)* Update dev, ci commands

### 🚜 Refactor

- *(cmd)* Reduce cognitive complexity and eliminate string duplication in tests
- *(cmd)* Extract test case creation to reduce function length

### ⚙️ Miscellaneous Tasks

- *(config)* Update code owners file

### Dep

- *(go)* Integrate k8s libraries and updates go version

### Deps

- *(deps)* Bump github.com/valyala/fasthttp in the updates group

## [0.4.3] - 2025-06-24

### 🐛 Bug/Lint Fixes

- *(ci)* Fix Docker tag generation in workflow_run context

## [0.4.2] - 2025-06-24

### 🐛 Bug/Lint Fixes

- *(ci)* Simplify release workflow logic

## [0.4.1] - 2025-06-24

### 🐛 Bug/Lint Fixes

- *(ci)* Correct release workflow security gate and binary packaging

## [0.4.0] - 2025-06-24

### 🚀 Features

- *(ci)* Add pre-release security scanning with Trivy
- *(ci)* Add security gate and matrix builds to release workflow

## [0.3.0] - 2025-06-24

### 🚀 Features

- *(.github)* Add GitHub community health files and project templates

## [0.2.0] - 2025-06-24

### 🚀 Features

- *(scripts)* Simplify release script to use git-cliff --bump

### 🐛 Bug/Lint Fixes

- *(docker)* Remove binary verification to support multi-arch builds

## [0.1.0] - 2025-06-23

### 🚀 Features

- *(docs)* Add initial REAME.md
- *(docs)* Update architecture diagram
- *(docs)* Add badges
- *(docs)* Add K8s Control Plane Setup
- *(docs)* Add installation guide for Fedora
- Initialize k8s with podman
- Create a taskfile
- Add dotfiles
- *(ansible)* Initial provision playbook
- *(ansible)* Add certs generation & split code to blocks
- *(ansible)* Enhance authentication and API server setup
- Add control plane services
- *(ansible)* Configure and run Scheduler and Controller Manager
- *(ansible)* Add kubelet to provision
- *(ansible)* Add Containerd config
- *(ansible)* Implements containerd runtime with systemd service and configuration
- Initalize Go Notebook
- *(ansible)* Configure kubeconfig for core user
- *(docs)* Add Ansible documentation
- *(notebook)* Update basic golang knowledge
- *(docker)* Create a Dockerfile
- *(ci)* Add ci from @den-vasyliev repo
- *(tasks)* Add Golang tasks
- *(docs)* Add Podman, Ansible & Taskfile info
- *(tasks)* Add tidy, fmt, lint
- Add zerolog-based logger with log-level flag
- *(ci)* Add CODEOWNERS for automated review assignments
- *(tasks)* Enable podman support and adds golangci-lint
- *(logger)* Add tests for packages
- *(serve)* Implement serve command with fasthttp server
- *(server)* Implement HTTP server with health check and default endpoints
- *(tasks)* Add `revive` linter
- *(tasks)* Add task for all linters
- *(docs)* Add API documentation
- *(docs)* Update README with prerequisites and progress
- *(cmd)* Enhance version implementatin
- *(cmd)* Add standard CLI version flags support
- *(lint)* Add extra rules
- *(cmd)* Add port validation to serve command
- *(server)* Set content type for responses
- *(ci)* Add Git Cliff
- *(cicd)* Implement release workflow, go test enhancements, and removes redundant CI workflow file
- [**breaking**] Rename binary to `kc` and modernize Dockerfile for production

### 🐛 Bug/Lint Fixes

- Remove extra space
- *(docs)* Update diagram
- *(docs)* Convert errors to inline code
- *(docs)* Update documentation
- Add proper reboot commands
- Add proper reboot commands
- Update ownership for home
- Update Containerd installation
- Add cli args
- Update services
- *(ansible)* Add kubeconfig
- *(ci)* Add path of go app
- *(ansible)* Update missing variables
- *(ci)* Disable envtest for now
- *(ci)* Disable Docker build
- *(ci)* Update to check all go files
- *(ci)* Add ignore pattern to covarage
- Offer more tests for god of tests
- *(server)* Increase test coverage
- *(go)* Update code regarding Go Style Guide
- *(tasks)* Update the build flags to correct the version injection
- *(cmd)* Remove toggle
- *(cmd)* Add silence
- *(script)* Improve script with shellcheck  suggestions
- *(script)* Update flags for read

### 🚜 Refactor

- *(init)* Update script & playbook
- Configure Podman setup with Ansible provisioning
- *(ansible)* PKI setup with CSR for CA and API server
- *(ansible)* Change path
- *(licenses)* Move to the docs
- *(tasks)* Update build flags to include local bin path
- *(server)* Improves server test reliability and cmd args
- *(server)* Reduce cognitive complexity and eliminate string duplication in tests
- *(cmd)* Consolidate version handling to single source of truth
- *(cmd)* Improve version command tests with helper function
- *(tests)* Setup root command test for better isolation

### ⚙️ Miscellaneous Tasks

- Add configs an Ci
- *(ansible)* Add the rest
- *(licenses)* Move to .github
- *(licenses)* Move back to the root
- *(notebook)* Fast overview of all functions
- *(config)* Remove cache & telemetry from git
- Add VSCode Workspace to the git
- *(dep)* Add fasthttp
- *(tasks)* Add coverage task
- *(workspace)* Enhance dubug methods
- Add tmp
- *(script)* Update logic
- *(script)* Update logic II

<!-- generated by git-cliff -->
