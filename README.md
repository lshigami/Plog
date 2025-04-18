# Plog: A Go Backend Blog API with AWS Deployment

Plog is a full-stack blogging application featuring a robust backend API built with Go and a frontend built with React (FE is generated by AI). 

## Overview

Plog provides a RESTful API for managing blog posts and user authentication. The backend is written in Go, leveraging modern practices and libraries. It is designed to be containerized using Docker and deployed onto AWS infrastructure.

## Features

* User registration and JWT-based authentication
* CRUD (Create, Read, Update, Delete) operations for blog posts
* Pagination for listing posts
* Association of posts with their authors
* Database migrations management
* API documentation via Swagger

## Technology Stack

* **Backend:**
  * Language: **Go** (v1.24+)
  * Framework: **Gin-Gonic** (`github.com/gin-gonic/gin`)
  * ORM/Query Builder: **sqlc** (`github.com/sqlc-dev/sqlc`) for type-safe SQL query generation
  * Migrations: **golang-migrate** (`github.com/golang-migrate/migrate`)
  * Authentication: **JWT** (`github.com/golang-jwt/jwt/v5`)
* **Database:**
  * PostgreSQL
* **Containerization:**
  * Docker, Docker Compose (for local dev)
* **AWS Services:**
  * **EC2:** Application hosting
  * **RDS:** Managed PostgreSQL database
  * **ECR:** Docker container registry (Public)
  * **Secrets Manager:** Secure storage for credentials and secrets
  * **IAM:** Permissions management for AWS services
* **CI/CD:**
  * GitHub Actions
## AWS Architecture

![Architecture](aws_images/architecture.png)

## Setup and Installation

### Prerequisites

* Go 1.24+
* Docker & Docker Compose
* Make (Optional, for using Makefile commands)
* AWS Account (If deploying to AWS)
* AWS CLI (Optional, for manual AWS setup/interaction)
* `migrate` CLI (Install via instructions [here](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate))

### Local Development

1. **Clone the repository:**
   ```bash
   git clone https://github.com/lshigami/Plog.git
   cd lshigami-plog
   ```

2. **Create `.env` file:**
   Create a file named `.env` in the project root with the following content (adjust values as needed):
   ```dotenv
   DATABASE_URL=postgres://admin:secret@localhost:5432/personal_blog_db?sslmode=disable
   JWT_SECRET=a_very_secret_key_should_be_longer_and_random_for_dev
   SERVER_PORT=8080
   ACCESS_TOKEN_DURATION=15m
   ```
   *Note: `docker-compose.yaml` also sets `DATABASE_URL` for the `api` service, overriding the `.env` file value for the container if both are present and docker-compose reads the env file.*

3. **Using Docker Compose (Recommended):**
   This starts the PostgreSQL database and the Go API service together.
   ```bash
   docker-compose up --build
   ```
   The API will be available at `http://localhost:8080`. The `entrypoint.sh` script within the `api` service's container will automatically wait for Postgres and run migrations.

4. **Using Makefile (Alternative):**
   If you have PostgreSQL running separately and `migrate` CLI installed locally:
   * Start your local PostgreSQL server ensuring the DB `personal_blog_db` exists with user `admin` and password `secret`
   * Update `DB_URL` in the `Makefile` if your local connection details differ
   * Run migrations: `make migrate_up`
   * Run the server: `make server` (This runs `go run cmd/server/main.go`)

### AWS Deployment

(This is a manual deployment guide)

1. **AWS Prerequisites:**
   * An RDS PostgreSQL instance running and accessible
   * An ECR Public repository (`public.ecr.aws/r8o3t2l0/go/plog`) where the image has been pushed (e.g., by the GitHub Action)
   * A secret created in AWS Secrets Manager (e.g., `goPlog` in `ap-southeast-1`) containing keys like `DATABASE_URL` (pointing to RDS), `JWT_SECRET`, `SERVER_PORT`, `ACCESS_TOKEN_DURATION`
   * An IAM Role for EC2 with Secrets Manager ReadWrite permission for the specific secret

2. **Launch EC2 Instance:**
   * Choose an AMI (e.g., Amazon Linux 2023)
   * Select an instance type (e.g., `t2.micro`)
   * Assign the **IAM Role** created above
   * Configure a **Security Group**:
     * Allow inbound SSH (Port 22) *from your IP only*
     * Allow inbound TCP (Port 8080) from `Anywhere (0.0.0.0/0)` or your specific IP/Load Balancer
     * Ensure outbound traffic to your RDS instance on port 5432 is allowed (often allowed by default, but verify)
   * Use a key pair you have access to

3. **Connect to EC2 via SSH:**
   ```bash
   ssh -i /path/to/your/key.pem ec2-user@YOUR_EC2_PUBLIC_IP
   ```

4. **Install Tools on EC2:**
   ```bash
   sudo dnf update -y
   sudo dnf install docker jq -y
   sudo systemctl start docker
   sudo systemctl enable docker
   sudo usermod -a -G docker ec2-user
   # Log out and log back in for group changes to take effect, or use 'newgrp docker'
   ```

5. **Deploy the Application:**
   ```bash
   # (Log back in if you logged out)
   mkdir ~/plog-app && cd ~/plog-app

   # Fetch secrets and create .env file (Ensure EC2 Role has permissions)
   aws secretsmanager get-secret-value --secret-id goPlog --region ap-southeast-1 --query 'SecretString' --output text | jq -r 'to_entries|map("\(.key)=\(.value|tostring)")|.[]' > .env

   # Verify .env content, especially DATABASE_URL pointing to RDS
   cat .env

   # Pull the latest image (or specific tag)
   export IMAGE_TAG=latest # Or your specific commit SHA tag
   export ECR_PUBLIC_REPO=public.ecr.aws/r8o3t2l0/go/plog
   docker pull $ECR_PUBLIC_REPO:$IMAGE_TAG

   # Run the container
   docker run -d --name plog-container \
     --restart unless-stopped \
     -p 8080:8080 \
     --env-file .env \
     $ECR_PUBLIC_REPO:$IMAGE_TAG
   ```

6. **Verify:**
   * Check logs: `docker logs plog-container`
   * Check status: `docker ps`
   * Access the API: `curl http://localhost:8080/health` (from within EC2) or `http://YOUR_EC2_PUBLIC_IP:8080/health` (from your browser)

## Usage

### API Endpoints

The main API endpoints include:

* `POST /register`: Register a new user
* `POST /login`: Login a user, returns JWT
* `GET /posts`: List posts with pagination (`limit`, `offset` query params)
* `POST /posts`: Create a new post (Requires Authentication)
* `GET /posts/{id}`: Get a specific post by ID
* `PUT /posts/{id}`: Update a specific post (Requires Authentication, user must own post)
* `GET /health`: Health check endpoint

## CI/CD

This project uses GitHub Actions for basic CI/CD:

* **`test.yml`**: Triggered on push/pull_request to `main`. Checks out code, sets up Go, builds the project, and runs unit tests (`go test -v ./...`)
* **`deploy.yml`**: Triggered on push to `main`
  1. Checks out the code
  2. Configures AWS credentials using an IAM Role assumed via OIDC 
  3. Logs into AWS ECR Public
  4. Builds the multi-stage Docker image
  5. Tags the image with the Git commit SHA
  6. Pushes the tagged image to the specified ECR Public repository (`public.ecr.aws/r8o3t2l0/go/plog`)

*Note: The `deploy.yml` workflow **only pushes the image**. It does *not* automatically deploy the new image to the EC2 instance. Manual steps or further automation (e.g., AWS CodeDeploy, SSM Run Command) are required to update the running container on EC2.*

![EC2](aws_images/EC2.png)
![ECR](aws_images/ECR.png)
![RDS](aws_images/RDS.png)

