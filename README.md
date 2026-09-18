# Tili-backend

## Coding Style

**To run the coding style execute this script**

```bash
./scripts/go_coding_style_checker.sh
```

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

## Developpement

### Stripe-CLI

You can use Stripe CLI to test webhook payment, without exposing the app

Download stripe CLI here [Stripe-cli]([stripe](https://github.com/stripe/stripe-cli))

```bash
stripe login && stripe listen --forward-to localhost:8000/api/webhooks/stripe
```




GET /categorie/catalog/:catalog_id - Get all categories for a catalog
GET /categorie/catalog/:catalog_id/:id - Get a specific category by ID
GET /categorie/catalog/:catalog_id/type/:type - Get a category by type
POST /categorie/catalog/:catalog_id - Create a new category
PUT /categorie/catalog/:catalog_id/:id - Update a category
DELETE /categorie/catalog/:catalog_id/:id - Delete a category by ID
DELETE /categorie/catalog/:catalog_id/type/:type - Delete a category by type


GET /catalog/store/:store_id - Get all catalogs for a store
GET /catalog/:id - Get a specific catalog by ID
POST /catalog/store/:store_id - Create a new catalog
PUT /catalog/:id - Update a catalog
DELETE /catalog/:id - Delete a catalog
