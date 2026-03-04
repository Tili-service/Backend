# Tili-backend

## Commit Message Guidelines

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

| Type | Description | Example |
|------|-------------|---------|
| `feat` | New feature | `feat(client): add chat system` |
| `fix` | Bug fix | `fix(server): resolve disconnect crash` |
| `docs` | Documentation | `docs(readme): update installation steps` |
| `style` | Code style changes | `style(client): format according to guidelines` |
| `refactor` | Code refactoring | `refactor(ecs): optimize component storage` |
| `test` | Test additions | `test(network): add packet parsing tests` |
| `perf` | Performance | `perf(render): optimize sprite batching` |
| `chore` | Maintenance | `chore(deps): update SFML to 2.6` |

### Scope

Indicates the affected module:
- `client`: Client code
- `server`: Server code
- `ecs`: ECS framework
- `network`: Network code
- `ui`: User interface
- `docs`: Documentation
- `build`: Build system

### Subject

- Use imperative mood: "add" not "added"
- Don't capitalize first letter
- No period at the end
- Maximum 50 characters

### Body

- Explain **what** and **why**, not how
- Wrap at 72 characters
- Separate from subject with blank line