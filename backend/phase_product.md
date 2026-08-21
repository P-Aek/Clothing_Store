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

## Phase 4 — Cart — Complete

- Cart model and MongoDB repository
- One cart per user with a unique user index
- JWT-protected get cart endpoint
- Add item with product, variant, quantity, and stock validation
- Update item quantity with stock validation
- Remove item
- Empty carts are returned for users without a persisted cart
- Unit and route tests for authentication, validation, stock, and not-found behavior

Endpoints:

- `GET /api/cart/`
- `POST /api/cart/items`
- `PUT /api/cart/items/:productId/:variantId`
- `DELETE /api/cart/items/:productId/:variantId`

## Phase 5 — Order — Complete

- Create order from cart through a MongoDB transaction
- Snapshot product and variant details into order items
- Atomically decrement variant stock during checkout
- Clear the cart only after the order and stock updates succeed
- Get a user's orders and individual orders with ownership checks
- Admin get all orders
- Admin update order status
- Order indexes for user history and status-based administration
- Unit and route tests for checkout, empty carts, stock failures, ownership, authentication, and authorization

Endpoints:

- `POST /api/orders/`
- `GET /api/orders/`
- `GET /api/orders/:id`
- `GET /api/admin/orders/`
- `PUT /api/admin/orders/:id/status`

## Phase 6 — API hardening and production readiness

- Consistent global API error response format
- Request IDs and structured application logging
- Pagination and validated filtering for product, category, and order lists
- Request body size limits and stricter request validation
- Rate limiting for registration, login, and other public endpoints
- Restrictive CORS and security response headers
- Checkout idempotency keys to prevent duplicate orders on retries
- Valid order-status transitions and cancellation stock restoration
- MongoDB integration tests for transactions, stock, and concurrent checkout
- Authentication and authorization test matrix
- OpenAPI/Swagger API documentation
- Readiness and liveness endpoints
- Metrics for request latency, error rate, and checkout failures
- Docker and CI configuration
- Automated verification with `go test ./...`, `go vet ./...`, and `go test -race ./...`
