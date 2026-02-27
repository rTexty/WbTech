package repository

import (
"regexp"
"testing"
"time"

"wildberries-tech/internal/models"

"github.com/DATA-DOG/go-sqlmock"
"github.com/stretchr/testify/require"
"gorm.io/driver/postgres"
"gorm.io/gorm"
)

func TestSaveOrder(t *testing.T) {
db, mock, err := sqlmock.New()
require.NoError(t, err)
defer func() { _ = db.Close() }()

dialector := postgres.New(postgres.Config{
Conn:       db,
DriverName: "postgres",
})
gdb, err := gorm.Open(dialector, &gorm.Config{})
require.NoError(t, err)

repo := &Repository{db: gdb}

order := models.Order{
OrderUID:    "test-uid",
TrackNumber: "test-track",
DateCreated: time.Now(),
Items: []models.Item{
{ChrtID: 123, TrackNumber: "item-track", Price: 100},
},
}

mock.ExpectBegin()
mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "orders"`)).
WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
WillReturnResult(sqlmock.NewResult(1, 1))

mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "items"`)).
WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
sqlmock.AnyArg(), sqlmock.AnyArg()).
WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

mock.ExpectCommit()

err = repo.SaveOrder(order)
require.NoError(t, err)

require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrder(t *testing.T) {
db, mock, err := sqlmock.New()
require.NoError(t, err)
defer func() { _ = db.Close() }()

dialector := postgres.New(postgres.Config{
Conn:       db,
DriverName: "postgres",
})
gdb, err := gorm.Open(dialector, &gorm.Config{})
require.NoError(t, err)

repo := &Repository{db: gdb}

mock.ExpectQuery(`SELECT \* FROM "orders" WHERE order_uid = \$1 ORDER BY "orders"\."order_uid" LIMIT \$2`).
WithArgs("test-uid", 1).
WillReturnRows(sqlmock.NewRows([]string{"order_uid", "track_number"}).AddRow("test-uid", "test-track"))

mock.ExpectQuery(`SELECT \* FROM "items" WHERE "items"\."order_uid" = \$1`).
WithArgs("test-uid").
WillReturnRows(sqlmock.NewRows([]string{"id", "order_uid", "track_number"}).AddRow(1, "test-uid", "item-track"))

order, err := repo.GetOrder("test-uid")
require.NoError(t, err)
require.NotNil(t, order)
require.Equal(t, "test-uid", order.OrderUID)

require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllOrders(t *testing.T) {
db, mock, err := sqlmock.New()
require.NoError(t, err)
defer func() { _ = db.Close() }()

dialector := postgres.New(postgres.Config{
Conn:       db,
DriverName: "postgres",
})
gdb, err := gorm.Open(dialector, &gorm.Config{})
require.NoError(t, err)

repo := &Repository{db: gdb}

mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orders"`)).
WillReturnRows(sqlmock.NewRows([]string{"order_uid", "track_number"}).
AddRow("test-uid-1", "track-1").
AddRow("test-uid-2", "track-2"))

mock.ExpectQuery(`SELECT \* FROM "items" WHERE "items"\."order_uid" IN \(\$1,\$2\)`).
WithArgs("test-uid-1", "test-uid-2").
WillReturnRows(sqlmock.NewRows([]string{"id", "order_uid", "track_number"}).
AddRow(1, "test-uid-1", "item-1").
AddRow(2, "test-uid-2", "item-2"))

orders, err := repo.GetAllOrders()
require.NoError(t, err)
require.Len(t, orders, 2)

require.NoError(t, mock.ExpectationsWereMet())
}
