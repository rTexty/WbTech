# Order Service

A high-performance, production-ready Go microservice for handling orders. It consumes order data from Kafka, stores it in PostgreSQL with transactional integrity, caches it in-memory for fast retrieval, and serves it via a REST API and a modern Web UI.

## 🚀 Features

- **Event-Driven Architecture**: Consumes orders from **Kafka** asynchronously.
- **Robust Storage**: Uses **PostgreSQL** with **GORM**.
  - **Transactional Integrity**: Ensures atomicity when saving orders and items.
  - **Connection Retries**: Resilient startup logic for database connections.
- **High Performance**:
  - **In-Memory Caching**: Implements `go-cache` with TTL and automatic cleanup to prevent memory leaks.
- **Reliability**:
  - **Graceful Shutdown**: Handles `SIGTERM`/`SIGINT` to ensure in-flight requests and database operations complete safely.
  - **Input Validation**: Uses `validator/v10` to ensure data integrity before processing.
- **Quality Assurance**:
  - **Unit & Integration Tests**: Comprehensive test coverage.
  - **Linting**: strictly follows Go standards.
- **User Interface**: Modern, responsive Web UI to view order details.

## 🛠 Tech Stack

- **Language**: Go 1.25+
- **Database**: PostgreSQL 15
- **Message Broker**: Apache Kafka + Zookeeper
- **Libraries**:
  - `gorilla/mux`: HTTP Routing
  - `gorm`: ORM & Database Management
  - `sarama`: Kafka Client
  - `go-cache`: In-memory Caching
  - `validator`: Struct Validation
  - `gofakeit`: Realistic Data Generation (for testing)

## 📂 Project Structure

```
├── cmd/
│   ├── server/       # Main application entry point
│   └── producer/     # Data generator for Kafka
├── internal/
│   ├── cache/        # In-memory caching layer
│   ├── config/       # Configuration management
│   ├── handlers/     # HTTP handlers
│   ├── kafka/        # Kafka consumer logic
│   └── repository/   # Database access layer
├── migrations/       # SQL migration files
├── web/              # Static frontend assets
├── tests/            # Integration tests
├── docker-compose.yml # Infrastructure (DB, Kafka, Zookeeper)
└── Makefile          # Build and run commands
```

## 🏁 Getting Started

### Prerequisites

- Go 1.25+
- Docker & Docker Compose
- Make (optional, for convenience)

### 1. Start Infrastructure

Start PostgreSQL, Kafka, and Zookeeper using Docker Compose:

```bash
docker-compose up -d
```

### 2. Configure Environment

The service uses default configuration suitable for local development. You can customize it via `.env` file if needed (see `.env.example`).

### 3. Run the Service

Start the main API server and Kafka consumer:

```bash
make run
```
*The server will start on port `8080`.*

### 4. Generate Data

Simulate incoming orders by running the producer script:

```bash
make producer
```
*This sends random JSON order data to the Kafka topic.*

### 5. Access Web UI

Open your browser and navigate to:

[http://localhost:8080](http://localhost:8080)

Enter an Order ID (e.g., from the producer output) to view its details.

## 🧪 Testing

Run all unit and integration tests:

```bash
make test
```

## 🧹 Linting

Run linters to ensure code quality:

```bash
make lint
```

## 📜 API Endpoints

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/` | Serve Web UI |
| `GET` | `/order/{id}` | Get Order JSON by ID |

## 🤝 Contribution

1. Fork the repository
2. Create your feature branch (`git checkout -b feat/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feat/amazing-feature`)
5. Open a Pull Request

## 📄 License

Distributed under the MIT License.
