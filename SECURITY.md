# Security policy

## Reporting a vulnerability

Please report suspected vulnerabilities privately through the GitHub Security
tab for `NodeOps-app/createos-go-sdk`. Include the affected SDK version,
reproduction steps, and impact. Do not open a public issue or include live API
keys, customer data, or response bodies containing secrets.

## Supported versions

Until the first stable release, security fixes target the latest published
`0.x` version. Upgrade older versions before reporting an issue.

## Security boundary

The SDK protects its API key from request-header overrides and cross-origin
redirects. The service enforces authorization and sandbox isolation.
Applications remain responsible for storing keys securely, authorizing
commands, handling output, and cleaning up resources.
