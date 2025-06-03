
# Rescounts-Task Backend API Documentation

This documentation provides a concise overview of the available endpoints, their request/response formats, and authentication requirements.

---

## Authentication

- **Signup** and **Login** do not require authentication.
- All other endpoints require a valid JWT in the `Authorization` header:
  ```
  Authorization: Bearer <token>
  ```
- Admin endpoints additionally require the user to be in the `admins` table.

---

## 1. Public Endpoints

### 1.1 POST `/signup`

Create a new user.

- **Request Header**:
  - `Content-Type: application/json`
- **Request Body**:
  ```json
  {
    "username": "johndoe",
    "email": "john@example.com",
    "password": "MyStr0ngP@ss"
  }
  ```
- **Success Response** (201 Created):
  ```json
  {
    "id": 42,
    "username": "johndoe",
    "email": "john@example.com"
  }
  ```
- **Errors**:
  - 400 Bad Request: Invalid JSON or missing fields.
  - 409 Conflict: Email or username already exists.

---

### 1.2 POST `/login`

Authenticate a user and receive a JWT.

- **Request Header**:
  - `Content-Type: application/json`
- **Request Body**:
  ```json
  {
    "email": "john@example.com",
    "password": "MyStr0ngP@ss"
  }
  ```
- **Success Response** (200 OK):
  ```json
  {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
  ```
- **Errors**:
  - 400 Bad Request: Invalid JSON or missing fields.
  - 401 Unauthorized: Invalid credentials.

---

## 2. User (Authenticated) Endpoints

All endpoints below require:
```
Authorization: Bearer <jwt_token>
```

### 2.1 POST `/users/creditcards`

Add a new credit card (Stripe PaymentMethod) for the logged-in user.

- **Request Header**:
  - `Content-Type: application/json`
  - `Authorization: Bearer <jwt_token>`
- **Request Body**:
  ```json
  {
    "payment_method_id": "pm_XXXXXXXXXXXX"
  }
  ```
- **Success Response** (201 Created):
  ```json
  {
    "id": 1,
    "stripe_pm_id": "pm_XXXXXXXXXXXX",
    "brand": "visa",
    "last4": "4242",
    "exp_month": 12,
    "exp_year": 2025
  }
  ```
- **Errors**:
  - 400 Bad Request: Missing `payment_method_id`, invalid PM, or user has no Stripe customer.
  - 401 Unauthorized: Missing or invalid token.
  - 500 Internal Server Error: DB or Stripe API failure (Should not accure).

---

### 2.2 DELETE `/users/creditcards/{card_id}`

Delete a credit card for the logged-in user.

- **Request Header**:
  - `Authorization: Bearer <jwt_token>`
- **Path Parameter**:
  - `card_id` (integer)
- **Success Response**:
  - 204 No Content
- **Errors**:
  - 400 Bad Request: Invalid `card_id`.
  - 401 Unauthorized: Missing or invalid token.
  - 404 Not Found: Card not found or does not belong to user.
  - 500 Internal Server Error: Stripe or DB failure (Should not accure).

---

### 2.3 GET `/products`

List all available products.

- **Request Header**:
  - `Authorization: Bearer <jwt_token>`
- **Success Response** (200 OK):
  ```json
  [
    {
      "id": 1,
      "name": "Widget A",
      "description": "A basic widget",
      "price_cents": 500
    },
    {
      "id": 2,
      "name": "Gadget B",
      "description": "A fancy gadget",
      "price_cents": 1299
    }
  ]
  ```
- **Errors**:
  - 401 Unauthorized: Missing or invalid token.
  - 500 Internal Server Error: DB query failed (Should not accure).

---

### 2.4 POST `/users/buy`

Purchase multiple products in one transaction.

- **Request Header**:
  - `Content-Type: application/json`
  - `Authorization: Bearer <jwt_token>`
- **Request Body**:
  ```json
  {
    "items": [
      { "product_id": 1, "quantity": 2 },
      { "product_id": 3, "quantity": 1 }
    ],
    "payment_method_id": "pm_XXXXXXXXXXXXXXXXX"
  }
  ```
- **Success Response** (200 OK):
  ```json
  {
    "success": true,
    "stripe_payment_intent_id": "pi_1JGxxxxx"
  }
  ```
- **Errors**:
  - 400 Bad Request: Invalid JSON, missing fields, invalid `product_id`, or Stripe payment failure.
  - 401 Unauthorized: Missing or invalid token.
  - 500 Internal Server Error: DB transaction failure (Should not accure).

---

### 2.5 GET `/users/history`

Retrieve purchase history for the logged-in user.

- **Request Header**:
  - `Authorization: Bearer <jwt_token>`
- **Success Response** (200 OK):
  ```json
  [
    {
      "purchase_id": 17,
      "product_id": 3,
      "product_name": "Gadget B",
      "quantity": 1,
      "total_price_cents": 1299,
      "purchased_at": "2025-05-28T14:23:45Z"
    },
    {
      "purchase_id": 12,
      "product_id": 1,
      "product_name": "Widget A",
      "quantity": 2,
      "total_price_cents": 1000,
      "purchased_at": "2025-05-20T09:12:33Z"
    }
  ]
  ```
  - Returns `[]` if no purchases exist.
- **Errors**:
  - 401 Unauthorized: Missing or invalid token.
  - 500 Internal Server Error: DB query failed (Should not accure).

---

## 3. Admin (Authenticated + Admin) Endpoints

All endpoints below require:
```
Authorization: Bearer <jwt_token>
```
and the JWT's `user_id` must be in the `admins` table.

### 3.1 POST `/admin/products`

Create a new product.

- **Request Header**:
  - `Content-Type: application/json`
  - `Authorization: Bearer <jwt_token>`
- **Request Body**:
  ```json
  {
    "name": "SuperWidget",
    "description": "An awesome widget",
    "price_cents": 2499
  }
  ```
- **Success Response** (201 Created):
  ```json
  {
    "id": 1,
    "name": "SuperWidget",
    "description": "An awesome widget",
    "price_cents": 2499
  }
  ```
- **Errors**:
  - 400 Bad Request: Missing/invalid fields.
  - 401 Unauthorized: Missing or invalid token.
  - 403 Forbidden: User is not an admin.
  - 500 Internal Server Error: DB insertion failed (Should not accure).

---

### 3.2 PUT `/admin/products/{id}`

Update an existing product.

- **Request Header**:
  - `Content-Type: application/json`
  - `Authorization: Bearer <jwt_token>`
- **Path Parameter**:
  - `id` (integer)
- **Request Body**:
  ```json
  {
    "name": "SuperWidget V2",
    "description": "Improved widget",
    "price_cents": 2799
  }
  ```
- **Success Response** (200 OK):
  ```json
  {
    "id": 1,
    "name": "SuperWidget V2",
    "description": "Improved widget",
    "price_cents": 2799
  }
  ```
- **Errors**:
  - 400 Bad Request: Invalid `id` or JSON payload.
  - 401 Unauthorized: Missing or invalid token.
  - 403 Forbidden: User is not an admin.
  - 404 Not Found: Product ID does not exist.
  - 500 Internal Server Error: DB update failed (Should not accure).

---

### 3.3 DELETE `/admin/products/{id}`

Delete a product.

- **Request Header**:
  - `Authorization: Bearer <jwt_token>`
- **Path Parameter**:
  - `id` (integer)
- **Success Response**:
  - 204 No Content
- **Errors**:
  - 400 Bad Request: Invalid `id`.
  - 401 Unauthorized: Missing or invalid token.
  - 403 Forbidden: User is not an admin.
  - 404 Not Found: Product ID does not exist.
  - 500 Internal Server Error: DB deletion failed (Should not accure).

---

### 3.4 GET `/admin/sales`

Retrieve product sales, with optional filters.

- **Request Header**:
  - `Authorization: Bearer <jwt_token>`
- **Query Parameters** (all optional):
  - `from` (YYYY-MM-DD) — include sales on/after this date.
  - `to` (YYYY-MM-DD) — include sales on/before this date (end of day).
  - `username` (string) — only include sales by this exact username.
- **Example**:
  ```
  GET /admin/sales?from=2025-01-01&to=2025-06-01&username=johndoe
  ```
- **Success Response** (200 OK):
  ```json
  [
    {
      "purchase_id": 25,
      "product_id": 3,
      "product_name": "Gadget B",
      "user_id": 7,
      "username": "johndoe",
      "quantity": 1,
      "total_price_cents": 1299,
      "purchased_at": "2025-05-28T14:23:45Z"
    },
    {
      "purchase_id": 23,
      "product_id": 1,
      "product_name": "Widget A",
      "user_id": 7,
      "username": "johndoe",
      "quantity": 2,
      "total_price_cents": 1000,
      "purchased_at": "2025-01-15T09:12:33Z"
    }
  ]
  ```
  - Returns `[]` if no matching sales.
- **Errors**:
  - 400 Bad Request: Invalid date formats.
  - 401 Unauthorized: Missing or invalid token.
  - 403 Forbidden: User is not an admin.
  - 500 Internal Server Error: DB query failed (Should not accure).

---

## 4. Error Response Format

Most errors return a plain text message with the appropriate HTTP status code. Example:
```
400 Bad Request
Invalid JSON payload
```
For consistency, always check the HTTP status in your client and display the response body as-is.

---

## 5. Example Usage

1. **Signup / Login**
   - Create a new user, then log in to receive a JWT.
2. **Add a credit card**
   - `POST /users/creditcards` with a Stripe `payment_method_id`.
3. **List products**
   - `GET /products` with JWT.
4. **Buy products**
   - `POST /users/buy` with items and `payment_method_id`.
5. **Get history**
   - `GET /users/history` with JWT.
6. **Admin: manage products**
   - Create, update, delete via `/admin/products` endpoints.
7. **Admin: view sales**
   - `GET /admin/sales?from=...&to=...&username=...`.

---

_End of Documentation_
