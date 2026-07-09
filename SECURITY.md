# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| Latest release | ✅ |

## Reporting a Vulnerability

If you discover a security vulnerability in CovScan, please report it by opening a [GitHub Issue](https://github.com/EdgarOrtegaRamirez/covscan/issues) with the label "security".

Do **not** open public issues for critical vulnerabilities — instead, email the maintainer directly.

## Security Features

- **No network access** — CovScan operates entirely offline. It never sends data to external services.
- **Local file processing only** — Only parses files you explicitly provide as arguments.
- **Safe file I/O** — All file operations use standard library functions with proper error handling. No shell injection vectors.
- **No dependencies with network functionality** — All dependencies are CLI/language standard libraries.

## Best Practices

1. Only run `covscan` on coverage reports you trust
2. Use `--threshold` in CI/CD to enforce minimum coverage standards
3. Review coverage reports for sensitive information before sharing