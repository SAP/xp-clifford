# Parsing and Sanitization

The `parsan` package transforms strings to conform with Kubernetes object naming rules.

## ParseAndSanitize

```go
func ParseAndSanitize(input string, rule Rule) []string
```

Returns all valid sanitized variants of the input for a given rule.

## Available Rules

### RFC1035Subdomain

Standard Kubernetes subdomain naming:

- Labels start and end with a letter (or end with a digit)
- Contains only letters, digits, and `-`
- Max 63 characters per label, 253 per subdomain

| Input                  | Sanitized              |
|------------------------|------------------------|
| `www.example.com`      | `www.example.com`      |
| `Can you sanitize me?` | `Can-you-sanitize-mex` |
| `99Luftballons`        | `x99Luftballons`       |
| `admin@example.com`    | `admin-at-example.com` |

### RFC1035LowerSubdomain

Same as `RFC1035Subdomain` but lowercase only.

| Input                  | Sanitized              |
|------------------------|------------------------|
| `Can you sanitize me?` | `can-you-sanitize-mex` |
| `99Luftballons`        | `x99luftballons`       |

### RFC1035SubdomainRelaxed

Same as `RFC1035Subdomain` but labels may start with digits.

| Input           | Sanitized       |
|-----------------|-----------------|
| `99Luftballons` | `99Luftballons` |

### RFC1035LowerSubdomainRelaxed

Combines lowercase and relaxed rules — labels may start with digits, all lowercase.

| Input               | Sanitized              |
|---------------------|------------------------|
| `99Luftballons`     | `99luftballons`        |
| `admin@example.com` | `admin-at-example.com` |

## Sanitization Behavior

- Invalid characters are replaced with `-` or `x`
- `@` is replaced with `-at-`
- Labels exceeding 63 characters are trimmed
- Subdomains exceeding 253 characters are trimmed
