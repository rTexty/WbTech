package repository

import (
	"fmt"
	"time"
	"wildberries-tech/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type OrderRepository interface {
	SaveOrder(order models.Order) error
	GetOrder(orderUID string) (*models.Order, error)
	GetAllOrders() ([]models.Order, error)
	Close() error
}

type Repository struct {
	db *gorm.DB
}

// New creates a new Repository with retry logic for database connection.
// It attempts to connect to the database up to 5 times with a 2-second delay between attempts.
func New(dsn string) (*Repository, error) {
	const maxRetries = 5
	const retryDelay = 2 * time.Second

	var db *gorm.DB
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
	// Connection successful, break the retry loop
			break
		}

		if attempt < maxRetries {
			fmt.Printf("Failed to connect to database (attempt %d/%d): %v. Retrying in %v...\n",
				attempt, maxRetries, err, retryDelay)
			time.Sleep(retryDelay)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
	}

	if err := db.AutoMigrate(&models.Order{}, &models.Item{}); err != nil {
		return nil, err
	}

	return &Repository{db: db}, nil
}

// SaveOrder persists an order and its nested items to the database within an explicit transaction.
func (r *Repository) SaveOrder(order models.Order) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *Repository) GetOrder(orderUID string) (*models.Order, error) {
	var order models.Order

	result := r.db.Preload("Items").Where("order_uid = ?", orderUID).First(&order)
	if result.Error != nil {
		return nil, result.Error
	}

	return &order, nil
}

func (r *Repository) GetAllOrders() ([]models.Order, error) {
	var orders []models.Order

	result := r.db.Preload("Items").Find(&orders)
	if result.Error != nil {
		return nil, result.Error
	}

	return orders, nil
}

func (r *Repository) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
