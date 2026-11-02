# Security policy

## Supported versions

Only the latest release receives security fixes. ginamite is not at v1 yet: a fix may come with incompatible changes,
described in the notes of the release.

## Reporting a vulnerability

Do not open a public issue. Report the vulnerability privately with GitHub:
[Report a vulnerability](https://github.com/cygy/ginamite/security/advisories/new).

Describe the version, the configuration and the steps that reproduce it, and its impact. You get an answer within
7 days. Once the fix is released, the advisory is published and credits you, unless you prefer otherwise.

## Scope

The vulnerabilities of ginamite and of its plugins (`plugins/mongodb`, `plugins/kafka`) are in scope. The ones of a
dependency belong to that dependency, unless ginamite uses it unsafely. The [security features](documentation/security.md)
describe what ginamite protects against, and what the application must configure.
