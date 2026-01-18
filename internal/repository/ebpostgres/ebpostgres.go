// Package ebpostgres provides methods for communicating with DB PostgresQL to service-layer
package ebpostgres

import (
	"context"
	"database/sql"

	"github.com/UnendingLoop/EventBooker/internal/model"
)

type PostgresRepo struct{}

// Executor provides a way to use both transactions and sql.DB for running queries from Service-layer
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (pr PostgresRepo) CreateEvent(ctx context.Context, exec Executor, newEvent *model.Event) error {
	return nil
} // только для админа

func (pr PostgresRepo) CreateBook(ctx context.Context, exec Executor, newBook *model.Book) error {
	return nil
}

func (pr PostgresRepo) CreateUser(ctx context.Context, exec Executor, newUser *model.User) error {
	return nil
}

func (pr PostgresRepo) DeleteEvent(ctx context.Context, exec Executor, eventID int) error {
	return nil
} // только для админа

func (pr PostgresRepo) DeleteBook(ctx context.Context, exec Executor, bookID int) error {
	return nil
} // эксклюзивно для воркера BookCleaner

func (pr PostgresRepo) UpdateBookStatus(ctx context.Context, exec Executor, bookID int, newStatus string) error {
	return nil
}

func (pr PostgresRepo) GetEventByID(ctx context.Context, exec Executor, id int) (*model.Event, error) { // select FOR UPDATE
	return nil, nil
}

func (pr PostgresRepo) GetEventsList(ctx context.Context, exec Executor) ([]model.Event, error) {
	return nil, nil
}

func (pr PostgresRepo) GetBookByID(ctx context.Context, exec Executor, id int) (*model.Book, error) {
	return nil, nil
}

func (pr PostgresRepo) GetBooksListByUser(ctx context.Context, exec Executor, id int) ([]*model.Book, error) {
	return nil, nil
}

func (pr PostgresRepo) GetExpiredBooksList(ctx context.Context, exec Executor) ([]*model.Book, error) {
	return nil, nil
}

func (pr PostgresRepo) GetUser(ctx context.Context, exec Executor, id int) (*model.User, error) {
	return nil, nil
}

func (pr PostgresRepo) IncrementAvailSeatsByEventID(ctx context.Context, exec Executor, eventID int) error {
	return nil
}

func (pr PostgresRepo) DecrementAvailSeatsByEventID(ctx context.Context, exec Executor, eventID int) error {
	return nil
}
