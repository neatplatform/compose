# Alertmanager Fixtures

Use these fixtures to test how a notification template renders against representative Alertmanager payloads.

## Quick Start

1. Edit `test.tmpl` with the template you want to validate.
2. Render the template using the `volt template` command.

```sh
# Render with a single fixture
volt template -json-file <fixture>.json -tmpl-file test.tmpl

# Render with all fixtures
find . -name "*.json" | sort | xargs -I {} volt template -json-file {}  -tmpl-file test.tmpl  
```
