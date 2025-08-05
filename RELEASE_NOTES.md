# LazyOC Release Notes

## v0.2.3 (2025-08-05)

### 🐛 Bug Fixes
- **README**: Fix failing status badges by updating CI badge to point to specific workflow file and removing non-functional badges (#3)
  - Updated CI badge to use correct workflow file path
  - Changed release badge to use proper v/release format
  - Removed Go Report Card badge (showing error)
  - Removed Codecov badge (not configured)
  - Added Go Version badge for better project information

### 🔧 Technical Improvements
- Remove unused `getLogViewHeight` function to resolve linter warnings
- Clean up codebase for better maintainability

---

## v0.2.2 (Previous Release)

### ✨ Features
- Re-enable Homebrew integration with proper token support
- Improve log streaming performance for high-frequency logs
- Implement real-time log streaming with line-anchored scrolling

### 🐛 Bug Fixes
- Temporarily disable Homebrew integration to resolve 403 error
- Update GoReleaser config to resolve deprecation warnings
- Resolve linting issues for clean CI builds

### 🧪 Testing
- Remove outdated test files that were failing