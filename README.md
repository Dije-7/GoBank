# GoBank - A Simple Banking System

GoBank is a feature-rich banking API developed using the Go programming language (Golang). The project harnesses the power of PostgreSQL for data management and employs JSON Web Tokens (JWT) for secure authentication and authorization. With a containerized PostgreSQL database managed by Docker, GoBank ensures a seamless and reliable deployment process.

## Features

- **User Authentication**: Utilize the power of JWT for secure user authentication and authorization, protecting your API endpoints.
- **Account Management**: Create and manage user accounts with support for administrative privileges.
- **Funds Transfer**: Seamlessly transfer funds between accounts while maintaining a comprehensive transaction history.
- **Dockerized PostgreSQL**: Employ Docker to manage the PostgreSQL database, simplifying deployment and ensuring consistency across environments.

## Prerequisites

- Go (Golang) - [Installation Guide](https://golang.org/doc/install)
- Docker - [Installation Guide](https://docs.docker.com/)
- PostgreSQL - [Installation Guide](https://www.postgresql.org/download/)

## Getting Started

1. Clone the repository:

```
git clone https://github.com/yourusername/gobank.git
cd gobank
```

2. Docker Setup: If you don't have PostgreSQL installed locally, you can use Docker to run a PostgreSQL container. Run the following command to start a PostgreSQL container:
```
docker run --name gobank-postgres -e POSTGRES_PASSWORD=mysecretpassword -d postgres
```
Replace mysecretpassword with your desired password.

3. Install dependencies and run the application:
```
make run
make run-seed (using this will seed an account)
```
4. The API server will start running on http://localhost:3000.

##API Endpoints
- POST /login: Log in to the system and obtain a JWT token.
- POST /account: Create a new user account (admin only).
- GET /account: View all user accounts (admin only).
- GET /account/{id}: View account details by ID (admin only).
- DELETE /account/{id}: Delete an account by ID (admin only).
- POST /set-admin/{accountNumber}: Set admin status for an account (admin only).
- POST /transfer: Perform a fund transfer between accounts (authenticated users only).

##Technologies Used
- Go (Golang): The core language for building the robust API.
- Dockerized PostgreSQL: Employ Docker to run a PostgreSQL database instance, simplifying deployment and setup.
- JSON Web Tokens (JWT): Ensure secure authentication and authorization of API requests.
- Gorilla Mux: A powerful router and dispatcher for the HTTP server in Go.



