# Clothing Store Backend API

Backend REST API for a clothing store, built with Go, Fiber, and MongoDB. It provides authentication, role-based catalog management, cart operations, transactional checkout, order history, admin order management, health checks, and interactive Swagger documentation.

## Features

- Customer registration and login with JWT authentication
- Password hashing with bcrypt
- Customer and admin role authorization
- Category and product management with soft deletion
- Product variants with color, size, and stock
- One cart per customer with stock validation
- Transactional checkout with atomic stock updates
- Customer order ownership checks
- Admin order listing and status management
- MongoDB indexes created during application startup
- Graceful HTTP server and MongoDB shutdown
- OpenAPI 3.0 specification and Swagger UI
- Unit and route tests for important business behavior

## Technology

- Go 1.26.4
- Fiber v2
- MongoDB Go Driver
- MongoDB Atlas or a replica set capable of transactions
- JWT using `github.com/golang-jwt/jwt/v5`
- Swagger UI using Fiber contrib Swagger middleware

## Project Structure

```text
cmd/api/                 Application entry point and dependency wiring
internal/apidocs/        Embedded OpenAPI specification and Swagger handler
internal/config/         Environment configuration
internal/controllers/    HTTP request and response handling
internal/database/       MongoDB connection setup
internal/middleware/     JWT authentication and role authorization
internal/models/         Domain and persistence models
internal/repositories/   MongoDB queries, indexes, and transactions
internal/routes/         Route registration
internal/services/       Validation and business logic
internal/utils/          Password and JWT helpers
```

The application follows a layered flow:

```text
HTTP route -> middleware -> controller -> service -> repository -> MongoDB
```

Controllers remain thin, services own validation and business rules, and repositories contain database operations.

## Prerequisites

- Go 1.26 or compatible project toolchain
- MongoDB Atlas or MongoDB configured as a replica set
- Git

MongoDB transactions are required for checkout so that order creation, stock reduction, and cart clearing succeed or fail together.

## Configuration

Copy the example environment file:

```powershell
Copy-Item .env.example .env
```

Configure the following values:

```dotenv
PORT=8081
MONGO_URI=mongodb+srv://username:password@cluster.example.mongodb.net/
MONGO_DATABASE=clothing_store
JWT_SECRET=replace-with-a-long-random-secret
```

| Variable | Required | Description |
| --- | --- | --- |
| `PORT` | No | HTTP port. The application default is `8080`; this project can run on `8081`. |
| `MONGO_URI` | Yes | MongoDB connection URI. |
| `MONGO_DATABASE` | Yes | MongoDB database name. |
| `JWT_SECRET` | Yes | Secret used to sign and verify JWTs. The value `change-me` is rejected. |

Do not commit `.env`. Use a strong, randomly generated JWT secret and store production secrets in a secret manager.

## Installation

Download the Go modules:

```powershell
go mod download
```

Run the API:

```powershell
go run ./cmd/api
```

With `PORT=8081`, the API listens at `http://localhost:8081`.

For development with Air:

```powershell
air
```

The repository includes an `.air.toml` configuration.

## Swagger API Documentation

After starting the API on port `8081`, open:

- Swagger UI: `http://localhost:8081/docs`
- OpenAPI YAML: `http://localhost:8081/docs/openapi.yaml`

The OpenAPI document is embedded into the Go binary, so the specification does not depend on the server's working directory. Swagger UI assets are loaded from a pinned CDN version, so the browser needs internet access to render the interactive UI. The raw OpenAPI YAML remains available directly from the API.

To call a protected endpoint from Swagger:

1. Register or log in.
2. Copy the `token` returned by `POST /api/auth/login`.
3. Select **Authorize** in Swagger UI.
4. Enter the JWT token.

## Authentication and Authorization

Protected endpoints expect this header:

```http
Authorization: Bearer <jwt-token>
```

Registration always creates a user with the `customer` role. Admin users must be provisioned separately through a trusted administrative process. Do not accept a role from the public registration request.

The JWT contains the user ID as its subject and includes the user's role. Authentication verifies the token and user ID format; authorization separately checks whether the role is permitted.

## API Endpoints

### Health

| Method | Endpoint | Access | Description |
| --- | --- | --- | --- |
| `GET` | `/health` | Public | Check API and MongoDB health. |

### Authentication

| Method | Endpoint | Access | Description |
| --- | --- | --- | --- |
| `POST` | `/api/auth/register` | Public | Register a customer account. |
| `POST` | `/api/auth/login` | Public | Log in and receive a JWT. |
| `GET` | `/api/auth/me` | Authenticated | Return the authenticated user ID. |

Example registration request:

```json
{
  "name": "Jane Doe",
  "email": "jane@example.com",
  "password": "strong-password"
}
```

Example login response:

```json
{
  "token": "<jwt-token>",
  "user": {
    "id": "507f1f77bcf86cd799439011",
    "name": "Jane Doe",
    "email": "jane@example.com",
    "role": "customer"
  }
}
```

### Categories

| Method | Endpoint | Access | Description |
| --- | --- | --- | --- |
| `GET` | `/api/categories/` | Public | List active categories. |
| `GET` | `/api/categories/:id` | Public | Get an active category. |
| `POST` | `/api/categories/` | Admin | Create a category. |
| `PUT` | `/api/categories/:id` | Admin | Update a category. |
| `DELETE` | `/api/categories/:id` | Admin | Soft-delete a category. |

### Products

| Method | Endpoint | Access | Description |
| --- | --- | --- | --- |
| `GET` | `/api/products/` | Public | List active products. |
| `GET` | `/api/products/?categoryId=:id` | Public | Filter products by category. |
| `GET` | `/api/products/:id` | Public | Get an active product. |
| `POST` | `/api/products/` | Admin | Create a product and variants. |
| `PUT` | `/api/products/:id` | Admin | Update a product and variants. |
| `DELETE` | `/api/products/:id` | Admin | Soft-delete a product. |

Example product request:

```json
{
  "categoryId": "507f1f77bcf86cd799439011",
  "name": "Oxford Shirt",
  "description": "Classic cotton shirt",
  "price": 49.99,
  "images": ["https://example.com/products/oxford-shirt.jpg"],
  "variants": [
    {
      "color": "Blue",
      "size": "M",
      "stock": 10
    }
  ]
}
```

### Cart

All cart endpoints require authentication.

| Method | Endpoint | Description |
| --- | --- | --- |
| `GET` | `/api/cart/` | Get the authenticated user's cart. |
| `POST` | `/api/cart/items` | Add a product variant to the cart. |
| `PUT` | `/api/cart/items/:productId/:variantId` | Update an item's quantity. |
| `DELETE` | `/api/cart/items/:productId/:variantId` | Remove an item. |

Example add-item request:

```json
{
  "productId": "507f1f77bcf86cd799439011",
  "variantId": "507f191e810c19729de860ea",
  "quantity": 1
}
```

### Customer Orders

All customer order endpoints require authentication and enforce order ownership.

| Method | Endpoint | Description |
| --- | --- | --- |
| `POST` | `/api/orders/` | Create an order from the current cart. |
| `GET` | `/api/orders/` | List the authenticated user's orders. |
| `GET` | `/api/orders/:id` | Get one owned order. |

Checkout snapshots product details into the order, atomically decrements stock, and clears the cart within a MongoDB transaction.

### Admin Orders

| Method | Endpoint | Access | Description |
| --- | --- | --- | --- |
| `GET` | `/api/admin/orders/` | Admin | List all orders. |
| `PUT` | `/api/admin/orders/:id/status` | Admin | Update an order status. |

Supported order status values:

- `pending`
- `processing`
- `shipped`
- `delivered`
- `cancelled`

Example status request:

```json
{
  "status": "processing"
}
```

## HTTP Status Codes

The API commonly returns:

| Status | Meaning |
| --- | --- |
| `200 OK` | Successful read or update. |
| `201 Created` | Resource or order created. |
| `204 No Content` | Resource or cart item removed. |
| `400 Bad Request` | Invalid body, identifier, query, or domain input. |
| `401 Unauthorized` | Missing or invalid JWT. |
| `403 Forbidden` | Authenticated user lacks the admin role. |
| `404 Not Found` | Resource does not exist or is not accessible. |
| `409 Conflict` | Duplicate data, empty cart, changed cart, or insufficient stock. |
| `500 Internal Server Error` | Unexpected internal failure. |
| `503 Service Unavailable` | MongoDB health check failed. |

Fiber's current default error handler returns controller and middleware errors as plain text. Health responses and successful API resources use JSON.

## Database

The application uses these MongoDB collections:

- `users`
- `categories`
- `products`
- `carts`
- `orders`

Indexes are created at startup:

- Unique user email
- Unique category slug
- Product category and active status
- Unique cart user
- Orders by user and creation time
- Orders by status and creation time

See [erd.md](./erd.md) for the entity relationship diagram.

## Testing and Verification

Run all tests:

```powershell
go test ./...
```

Run static analysis:

```powershell
go vet ./...
```

Run the race detector before a production release:

```powershell
go test -race ./...
```

The test suite covers services, authentication middleware, health checks, route protection, validation, stock behavior, cart operations, order ownership, and Swagger routes.

## Security Notes

- Passwords are hashed and are never returned in API responses.
- Registration does not allow clients to select an admin role.
- JWT authentication and role authorization are separate checks.
- Customer-specific cart and order operations use the authenticated user ID.
- Category and product deletion is soft deletion.
- Checkout uses a transaction and conditional stock updates.
- Secrets are loaded from environment variables and must not be committed.
- Public authentication endpoints should be rate-limited before internet-facing production deployment.
- Production deployments should add restrictive CORS, security headers, request size limits, structured logging, and a consistent JSON error format.

## Deployment Notes

- Provide all required environment variables through the deployment platform.
- Use a MongoDB deployment that supports transactions.
- Ensure shutdown signals reach the application for graceful shutdown.
- Swagger documentation is public at `/docs`; restrict or disable it at the gateway if API surface disclosure is not desired in production.
- No schema migration command is required. MongoDB indexes are created automatically when the API starts.
