# Order Service

Service for processing orders from Kafka, saving them to PostgreSQL, and serving them via HTTP API with in-memory caching.

## Tech Stack
- **Go** 1.25+
- **PostgreSQL** (Storage)
- **Kafka** (Message Broker)
- **Gorilla Mux** (Router)
- **GORM** (ORM)
- **Go-Cache** (In-memory cache with TTL)
- **Validator/v10** (Data validation)

## Architecture
- **Producer**: Generates fake order data and sends it to Kafka topic `orders`.
- **Consumer**: Reads from Kafka, validates data, saves to DB, and updates Cache.
- **API**: Serves order data by ID from Cache (fallback to DB).
- **Cache**: In-memory storage with TTL to reduce DB load.

## Configuration
Configuration is managed via environment variables or `.env` file. See `.env.example`.

## Usage

### Prerequisites
- Docker & Docker Compose
- Go 1.25+

### Running
1. Start infrastructure:
   ```bash
   docker-compose up -d
   ```
2. Run migrations (using golang-migrate or manually applying `migrations/` SQL files).
3. Start the service:
   ```bash
   make run
   ```
4. Start the producer (to generate data):
   ```bash
   make producer
   ```

### API
- `GET /order/{id}` - Get order by ID
- `GET /` - Web Interface

## Development
- `make test` - Run tests
- `make lint` - Run linters
