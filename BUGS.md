# Support

## Reporting Issues
We use GitHub Issues for tracking bugs, feature requests, and general support questions.

### Before Reporting
- Check existing issues to avoid duplicates
- Ensure you're using the latest version
- Test with the MCP Inspector to isolate the issue

### Bug Reports
When reporting bugs, please include:
- **Environment**: Operating system, Python version, MCP client
- **Configuration**: Relevant parts of your MCP configuration
- **Steps to Reproduce**: Clear, minimal steps to reproduce the issue
- **Expected vs Actual**: What you expected to happen vs what actually happened
- **Logs**: Relevant log entries (see Troubleshooting section below)
- **Repository Context**: Whether using single repo, multi-repo, or discovery mode

### Feature Requests
For feature requests, please describe:
- **Use Case**: The problem you're trying to solve
- **Proposed Solution**: Your idea for addressing it
- **Alternatives Considered**: Other approaches you've evaluated
- **Impact**: Who would benefit and how

## Troubleshooting

### Common Issues
**"Repository not found" errors:**
- Verify the repository path is correct and accessible
- Check that the directory contains a `.git` folder
- Ensure MCP client has proper permissions to access the path

**Discovery not finding repositories:**
- Confirm `--enable-discovery` flag is set
- Check that repositories are within MCP root paths
- Verify repositories aren't excluded by discovery patterns
- Try `--force-refresh` to clear discovery cache

**Performance issues with multi-repository operations:**
- Reduce `--max-discovery-depth` if scanning too deep
- Add exclusion patterns for large directories like `node_modules`
- Use `show_clean=false` in `git_multi_status` to reduce output

### Enable Detailed Logging
For debugging, enable verbose logging by setting environment variables:

```bash
# Enable debug logging
export MCP_LOG_LEVEL=DEBUG

# Enable git command tracing
export MCP_GIT_TRACE=1

# Log to file
export MCP_LOG_FILE=/path/to/debug.log
```

**Windows (PowerShell):**
```powershell
$env:MCP_LOG_LEVEL="DEBUG"
$env:MCP_GIT_TRACE="1"
$env:MCP_LOG_FILE="C:\path\to\debug.log"
```

**Claude Desktop Logs:**
- **macOS**: `~/Library/Logs/Claude/mcp*.log`
- **Windows**: `%APPDATA%\Claude\logs\mcp*.log`
- **Linux**: `~/.local/share/Claude/logs/mcp*.log`

### Testing with MCP Inspector
The MCP Inspector is invaluable for debugging:

```bash
# Test with Inspector
npx @modelcontextprotocol/inspector uvx mcp-server-git --repository /path/to/repo

# Test with discovery enabled
npx @modelcontextprotocol/inspector uvx mcp-server-git --enable-discovery
```

Access the Inspector at `http://localhost:6274` to:
- Test individual git operations interactively
- Validate configuration and arguments
- Examine server responses in real-time
- Debug parameter passing issues

## Security Issues
**Security vulnerabilities should be reported privately.** Please see our `SECURITY.md` file for detailed instructions on responsible disclosure.

Do **not** report security issues through public GitHub issues.

## Getting Help
- **Documentation**: Check `USING.md` for installation and usage instructions
- **Examples**: See configuration examples in the `rel/` directory
- **Community**: Join discussions in GitHub Discussions
- **Professional Support**: Contact information available in project documentation

## Contributing
We welcome contributions! Please see `CONTRIBUTING.md` for guidelines on:
- Code style and standards
- Testing requirements
- Pull request process
- Development setup
