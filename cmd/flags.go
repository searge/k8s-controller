// Package cmd contains shared flags and utilities for CLI commands.
// This file defines common Kubernetes-related flags used across multiple commands.
package cmd

// Shared flags for Kubernetes operations
var (
	// kubeconfigPath specifies the path to kubeconfig file.
	// Used by commands that interact with Kubernetes API.
	kubeconfigPath string

	// contextName specifies the Kubernetes context to use.
	// Used by commands that interact with Kubernetes API.
	contextName string

	// timeoutSeconds specifies the timeout for Kubernetes operations.
	// Used by commands that interact with Kubernetes API.
	timeoutSeconds int
)
