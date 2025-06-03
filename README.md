# rescounts-task

HTTP Web Server for Users and Products Management

---

## 🛠 Setup and Run Instructions

### Prerequisites
- Go  
- Docker & Docker Compose  
- Git  
- `psql` client (for local DB access)  

### Environment Variables
Set the following environment variables (e.g. in a `.env` file):

```env
DB_HOST=db
DB_PORT=5432
DB_USER=rescounts_user
DB_PASSWORD=rescounts_pass
DB_NAME=rescounts_db
JWT_SECRET=<your_jwt_secret>
STRIPE_SECRET_KEY=<your_stripe_test_key>
```

---

## Running with Docker Compose

1. **Clone the repository:**
   ```bash
   git clone https://github.com/Brossef/rescounts-task.git
   cd rescounts-task
   ```

2. **Create a `.env` file in the root directory** using the example above.

3. **Build and start the services:**
   ```bash
   docker-compose up --build
   ```

4. **Access the Go server at:**
   ```
   http://localhost:8080
   ```

5. **To stop the services:**
   ```bash
   docker-compose down
   ```

---

## 🗃 Database Schema (Postgres)

- `users` (id, username, email, password_hash, stripe_customer_id, created_at)  
- `admins` (user_id)  
- `products` (id, name, description, price_cents, created_at)  
- `credit_cards` (id, user_id, stripe_pm_id, brand, last4, exp_month, exp_year, created_at)  
- `purchases` (id, user_id, product_id, quantity, total_price_cents, stripe_payment_intent_id, purchased_at)  

---

## Testing with Postman

1. **Import the provided Postman collection:**  
   File: `postman/Rescounts.postman_collection.json`

2. **Key Endpoints:**
   ```
   POST    /signup
   POST    /login
   POST    /users/creditcards
   DELETE  /users/creditcards/{card_id}
   GET     /products
   POST    /users/buy
   GET     /users/history
   Admin Only:
     POST    /admin/products
     PUT     /admin/products/{id}
     DELETE  /admin/products/{id}
     GET     /admin/sales
   ```

3. **Notes:**
   - Pass JWT tokens in the `Authorization` header:
     ```
     Authorization: Bearer <token>
     ```
   - Use Stripe's test card:
     ```
     pm_card_visa
     ```

---

## 🔗 Postman Collection Link

[Open Postman Collection](https://postman.co/workspace/My-Workspace~61498fee-56fb-46b3-8468-f0f8e28a7135/collection/45502323-6c201d3f-c6a9-4555-9183-502f4e4e90b6?action=share&creator=45502323)
