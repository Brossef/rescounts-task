# rescounts-task
HTTP Web Server for Users and Products Management


Setup and Run Instructions

Prerequisites:
- Go 
- Docker & Docker Compose
- Git
- psql client for local DB access

Environment Variables:
- DB_HOST: db
- DB_PORT: 5432
- DB_USER: rescounts_user
- DB_PASSWORD: rescounts_pass
- DB_NAME: rescounts_db
- JWT_SECRET: a random secret key for signing JWTs
- STRIPE_SECRET_KEY: your Stripe test secret key

Using Docker Compose:
1. Clone the repo:
   git clone https://github.com/Brossef/rescounts-task.git
   cd rescounts-task

2. Create a .env file in the project root with:
   DB_HOST=db
   DB_PORT=5432
   DB_USER=rescounts_user
   DB_PASSWORD=rescounts_pass
   DB_NAME=rescounts_db
   JWT_SECRET=<your_jwt_secret>
   STRIPE_SECRET_KEY=<your_stripe_test_key>

3. Build and start services:
   docker-compose up --build

4. The Go server will be available at:
   http://localhost:8080

5. To stop:
   docker-compose down


Database Schema (Postgres):
- users (id, username, email, password_hash, stripe_customer_id, created_at)
- admins (user_id)
- products (id, name, description, price_cents, created_at)
- credit_cards (id, user_id, stripe_pm_id, brand, last4, exp_month, exp_year, created_at)
- purchases (id, user_id, product_id, quantity, total_price_cents, stripe_payment_intent_id, purchased_at)

Testing with Postman:
- Import the provided Postman collection (postman/Rescounts.postman_collection.json).
- Use the following endpoints:
  POST /signup
  POST /login
  POST /users/creditcards
  DELETE /users/creditcards/{card_id}
  GET /products
  POST /users/buy
  GET /users/history
  (Admin only: POST/PUT/DELETE /admin/products /admin/sales)

Notes:
- Ensure JWT tokens are passed in Authorization headers: "Bearer <token>".
- Use Stripe test card "pm_card_visa" to create PaymentMethods.

Postman collection for testing the API endpoints:
https://.postman.co/workspace/My-Workspace~61498fee-56fb-46b3-8468-f0f8e28a7135/collection/45502323-6c201d3f-c6a9-4555-9183-502f4e4e90b6?action=share&creator=45502323