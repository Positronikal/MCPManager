# Bug Reporting

This document explains how to report bugs for MCP Manager. Please use GitHub Issues for bug reports, following the guidelines below.

## How to Report a Bug

### Before Reporting
- Check existing bugs to avoid duplicates
- Ensure you're using the latest version
- Test in development mode (`wails dev`) for detailed logs

### Bug Report Template
When reporting bugs, include:
- **Environment**: OS, Go version, Wails version
- **Configuration**: Relevant server configs
- **Steps to Reproduce**: Clear, minimal reproduction steps
- **Expected vs Actual**: What should happen vs what does happen
- **Logs**: Terminal output from `wails dev`
- **Spec References**: Which FR/requirements are violated

### Severity Levels
- **Critical**: System unusable, data loss risk
- **High**: Major feature broken, workaround exists
- **Medium**: Feature impaired, minor impact
- **Low**: Cosmetic issue, no functional impact

### Priority
- **P0**: Blocks release, fix immediately
- **P1**: High impact, fix in current sprint
- **P2**: Medium impact, schedule for next sprint
- **P3**: Low impact, fix when convenient

---

## Known Limitations

These are intentional design decisions per specifications, not bugs:

1. **No MCP Client Config Modification** (FR-019)
   - MCP Manager displays but never modifies Claude Desktop/Cursor configs
   - Users must edit client configs manually

2. **No Remote Server Installation**
   - MCP Manager only manages already-installed servers
   - Users must install servers via npm/pip/etc. first

3. **Single Instance Only** (FR-051)
   - Only one MCP Manager instance per machine
   - Second launch shows existing window

4. **UI Real-Time Updates** (FR-005, FR-047)
   - Currently, UI table requires manual refresh after lifecycle operations
   - Backend cache updates correctly, but UI doesn't subscribe to state change events
   - This is a missing feature implementation, not a bug
   - Should be implemented as part of normal spec-kit development workflow

---

## Contact

- **Security Issues**: See SECURITY.md for private disclosure
- **General Bugs**: GitHub Issues (this file guides reporting)
- **Feature Requests**: specs/ directory (add new spec-kit feature)
