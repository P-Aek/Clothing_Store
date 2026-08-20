## Phase 0 — Project foundation — Complete

- Go module and backend project structure
- Layered architecture under `internal/`
- Fiber application entry point
- Environment configuration with `.env` and `.env.example`
- `.gitignore` protection for local environment files
- Air hot reload configuration
- Global Fiber request logger
- `GET /health` endpoint

## Phase 1 — Database and runtime setup — Complete

- MongoDB Atlas connection from `MONGO_URI`
- Database selection from `MONGO_DATABASE`
- Startup MongoDB connectivity check with a timeout
- MongoDB connectivity check inside `GET /health`
- HTTP `503 Service Unavailable` when the database is disconnected
- Graceful MongoDB disconnect during application shutdown

## Phase 2 — User — Complete

- User model
- Register
- Login
- Password hash
- JWT

## Phase 3A — Category — Complete

- Category model
- GET categories
- Admin create category
- Admin update category
- Admin delete category

- Public category list and detail endpoints
- Admin mutations require JWT authentication and the `admin` role
- Category names and slugs are validated and normalized
- Unique category slug index
- Deletion is implemented as a soft delete to preserve product references
- Tests for validation, duplicate slugs, not-found behavior, authentication, and authorization

## Phase 3B — Product — Complete

- Product model
- Product variant
- `categoryId`
- GET products
- GET product/:id
- Filter by category
- Admin CRUD product

- Product variants are embedded in product documents
- Public product list and detail endpoints
- Filter products by active category with `categoryId`
- Admin mutations require JWT authentication and the `admin` role
- Product, image, price, stock, and variant validation
- Product category existence validation
- Product category/active index
- Product deletion is implemented as a soft delete
- Tests for validation, category errors, not-found behavior, authentication, and authorization

## Phase 4 — Cart

- Add item
- Update quantity
- Remove item
- Get cart

## Phase 5 — Order

- Create order from cart
- Get user's orders
- Admin get all orders
- Update order status
