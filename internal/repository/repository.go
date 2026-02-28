// Package repository handles database operations for orders.
package repository

import (
	"database/sql"
	"fmt"
	"time"
	"wildberries-tech/internal/config"
	"wildberries-tech/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// OrderRepository defines the interface for database interactions.
type OrderRepository interface {
	SaveOrder(order models.Order) error
	GetOrder(orderUID string) (*models.Order, error)
	GetAllOrders() ([]models.Order, error)
	Close() error
	DB() (*sql.DB, error)
}

// Repository implements OrderRepository using GORM.
type Repository struct {
	db *gorm.DB
}

// New creates a new Repository with retry logic for database connection.
func New(cfg config.DatabaseConfig) (*Repository, error) {
	var db *gorm.DB
	var err error

	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		db, err = gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
		if err == nil {
			// Connection successful, break the retry loop
			break
		}

		if attempt < cfg.MaxRetries {
			fmt.Printf("Failed to connect to database (attempt %d/%d): %v. Retrying in %v...\n",
				attempt, cfg.MaxRetries, err, cfg.RetryDelay)
			time.Sleep(cfg.RetryDelay)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", cfg.MaxRetries, err)
	}

	if err := db.AutoMigrate(&models.Order{}, &models.Item{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database schema: %w", err)
	}

	return &Repository{db: db}, nil
}

// SaveOrder persists an order and its nested items to the database within an explicit transaction.
func (r *Repository) SaveOrder(order models.Order) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&order).Error; err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}
		return nil
	})
}

// GetOrder retrieves a single order by its UID.
func (r *Repository) GetOrder(orderUID string) (*models.Order, error) {
	var order models.Order

	result := r.db.Preload("Items").Where("order_uid = ?", orderUID).First(&order)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get order %s: %w", orderUID, result.Error)
	}

	return &order, nil
}

// GetAllOrders retrieves all orders from the database.
func (r *Repository) GetAllOrders() ([]models.Order, error) {
	var orders []models.Order

	result := r.db.Preload("Items").Find(&orders)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get all orders: %w", result.Error)
	}

	return orders, nil
}

// Close closes the underlying database connection.
func (r *Repository) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql db: %w", err)
	}
	return sqlDB.Close()
}

// DB returns the underlying sql.DB for health checks.
func (r *Repository) DB() (*sql.DB, error) {
	return r.db.DB()
}
