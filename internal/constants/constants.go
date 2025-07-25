// Package constants provides centralized constant definitions for the LazyOC application.
// This package helps maintain consistency and makes configuration changes easier by
// consolidating all magic numbers, strings, and configuration values in one place.
//
// The constants are organized into logical categories:
//   - time.go: Timeouts, intervals, and duration-related constants
//   - limits.go: Resource limits, buffer sizes, and retry configurations
//   - paths.go: File paths, directories, and configuration locations
//   - ui.go: UI themes, dimensions, and display-related constants
//   - status.go: Status strings for connections and resources
//   - errors.go: Error messages and error detection keywords
//   - http.go: HTTP status codes and related constants
//   - api.go: API endpoints and paths
//
// When adding new constants:
//  1. Choose the appropriate file based on the constant's category
//  2. Use clear, descriptive names following Go naming conventions
//  3. Add documentation explaining the purpose and any important notes
//  4. Include units in comments where applicable (e.g., seconds, pixels)
//  5. Consider relationships and dependencies with other constants
package constants
