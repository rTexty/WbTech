package repository

import (
	"wildberries-tech/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&models.Order{}, &models.Item{})

	return &Repository{db: db}, nil
}

func (r *Repository) SaveOrder(order models.Order) error {
	result := r.db.Create(&order)
	return result.Error
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