// Package service provides all methods for app
package service

import (
	"context"
	"log"
	"time"

	"github.com/UnendingLoop/EventBooker/internal/model"
	"github.com/UnendingLoop/EventBooker/internal/repository"
	"github.com/wb-go/wbf/dbpg"
)

type EBService struct {
	repo repository.EBRepo
	db   *dbpg.DB
}

func NewEBService(ebrepo repository.EBRepo, ebdb *dbpg.DB) *EBService {
	return &EBService{repo: ebrepo, db: ebdb}
}

func (eb EBService) CreateUser(ctx context.Context, user *model.User) error {
	if err := validateNormalizeUser(user); err != nil {
		return err
	}

	if err := eb.repo.CreateUser(ctx, eb.db, user); err != nil { // кейс имейл уже существует: unique_violation -> ErrUserAlreadyExists
		log.Printf("Failed to create new user in DB: %v", err)
		return ErrCommon500
	}

	return nil
}

func (eb EBService) CreateEvent(ctx context.Context, event *model.Event) error {
	if err := validateNormalizeEvent(event); err != nil {
		return err // 400
	}

	if err := eb.repo.CreateEvent(ctx, eb.db, event); err != nil {
		log.Printf("Failed to create new event in DB: %v", err)
		return ErrCommon500
	}

	return nil
}

func (eb EBService) BookEvent(ctx context.Context, book *model.Book) error {
	if book.EventID <= 0 || book.UserID <= 0 {
		return ErrEmptyBookInfo // 400
	}
	book.Status = model.BookStatusCreated

	// транзакция - бегин
	tx, err := eb.db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Failed to begin transaction:", err)
		return ErrCommon500 // 500
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			log.Println("Failed to rollback transaction:", err)
		}
	}()
	// получаем ивент и проверяем доступность мест
	event, err := eb.repo.GetEventByID(ctx, tx, book.EventID) // нужна проверка на 404
	if err != nil {
		log.Println("Failed to get eventID from DB:", err)
		return ErrCommon500 // 500
	}
	if event.Status != model.EventStatusActual {
		return ErrExpiredEvent // 409
	}
	if event.EventDate.Before(time.Now().UTC()) {
		return ErrExpiredEvent // 409
	}
	if event.AvailSeats == 0 {
		return ErrNoSeatsAvailable // 409
	}

	book.ConfirmDeadline = time.Now().UTC().Add(time.Duration(event.BookWindow) * time.Second)

	// создание записи
	if err := eb.repo.CreateBook(ctx, tx, book); err != nil {
		log.Println("Failed to create new book in DB:", err)
		return ErrCommon500 // 500
	}

	// декремент event.availSeats
	if err := eb.repo.DecrementAvailSeatsByEventID(ctx, tx, book.EventID); err != nil {
		log.Println("Failed to decrement event avail.seats:", err)
		return ErrCommon500
	}
	// коммит транзакции
	if err := tx.Commit(); err != nil {
		log.Println("Failed to commit transaction:", err)
		return ErrCommon500
	}
	return nil
}

func (eb EBService) ConfirmBook(ctx context.Context, bid int) error {
	// бегин транзакции
	tx, err := eb.db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Failed to begin transaction:", err)
		return ErrCommon500
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			log.Println("Failed to rollback transaction:", err)
		}
	}()
	// проверяем бронь
	book, err := eb.repo.GetBookByID(ctx, tx, bid)
	if err != nil {
		log.Printf("Failed to get book info from DB: %v", err)
		return ErrCommon500
	}
	if book.Status == model.BookStatusCancelled {
		return ErrBookIsCancelled
	}

	// апдейтим статус
	if err := eb.repo.UpdateBookStatus(ctx, tx, bid, model.BookStatusConfirmed); err != nil { // добавить обработку 404
		log.Printf("Failed to confirm book in DB: %q", err)
		return ErrCommon500
	}

	// коммит транзакции
	if err := tx.Commit(); err != nil {
		log.Println("Failed to commit transaction:", err)
		return ErrCommon500
	}

	return nil
}

func (eb EBService) CancelBook(ctx context.Context, bid int) error { // не удаляет бронь, а помечает как cancelled и инкрементит availseats
	// бегин транзакции
	tx, err := eb.db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Failed to begin transaction:", err)
		return ErrCommon500
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			log.Println("Failed to rollback transaction:", err)
		}
	}()

	// проверяем бронь
	book, err := eb.repo.GetBookByID(ctx, tx, bid)
	if err != nil {
		log.Printf("Failed to get book info from DB: %v", err)
		return ErrCommon500
	}
	if book.Status == model.BookStatusCancelled {
		return nil
	}

	// отменяем бронь
	if err := eb.repo.UpdateBookStatus(ctx, tx, bid, model.BookStatusCancelled); err != nil {
		log.Println("Failed to update book status in DB:", err)
		return ErrCommon500
	}

	// инкрементим event.availSeats
	if err := eb.repo.IncrementAvailSeatsByEventID(ctx, tx, book.EventID); err != nil { // добавить обработку 404
		log.Println("Failed to increment event avail.seats:", err)
		return ErrCommon500
	}

	// коммит транзакции
	if err := tx.Commit(); err != nil {
		log.Println("Failed to commit transaction:", err)
		return ErrCommon500
	}

	return nil
}

func (eb EBService) DeleteEvent(ctx context.Context, eid int) error { // добавить проверку роли пользователя
	if err := eb.repo.DeleteEvent(ctx, eb.db, eid); err != nil {
		log.Printf("Failed to delete event in DB: %v", err)
		return ErrCommon500
	}

	return nil
}

func (eb EBService) CleanExpiredBooks(ctx context.Context) error {
	// транзакция - бегин
	tx, err := eb.db.BeginTx(ctx, nil)
	if err != nil {
		log.Println("Failed to begin transaction:", err)
		return ErrCommon500
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			log.Println("Failed to rollback transaction:", err)
		}
	}()

	// запрос всех подходящих броней
	books, err := eb.repo.GetExpiredBooksList(ctx, tx)
	if err != nil {
		log.Println("Failed to fetch expired books:", err)
		return ErrCommon500
	}
	if len(books) == 0 {
		return nil
	}

	// в цикле проделать декремент ивентов и удаление броней
	for _, b := range books {
		// если статус брони cancelled - availSeats уже инкрементирован
		if b.Status != model.BookStatusCancelled {
			if err := eb.repo.IncrementAvailSeatsByEventID(ctx, tx, b.EventID); err != nil {
				log.Println("Failed to decrement event avail.seats:", err)
				return ErrCommon500
			}
		}

		if err = eb.repo.DeleteBook(ctx, tx, b.ID); err != nil {
			log.Println("Failed to delete expired book:", err)
			return ErrCommon500
		}
	}

	// закоммитить транзакцию
	if err := tx.Commit(); err != nil {
		log.Println("Failed to commit transaction:", err)
		return ErrCommon500
	}

	log.Printf("Cleaned %d expired bookings\n", len(books))
	return nil
}

func (eb EBService) GetBooksListByUserID(ctx context.Context, uid int) ([]*model.Book, error) {
	res, err := eb.repo.GetBooksListByUser(ctx, eb.db, uid)
	if err != nil { // добавить обработку 404
		log.Printf("Failed to confirm book in DB: %q", err)
		return nil, ErrCommon500
	}

	return res, nil
}
