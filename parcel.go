package main

import (
	"database/sql"
	"errors"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	// SQL-запрос для добавления новой записи в таблицу parcel
	query := `INSERT INTO parcel (client, status, address, created_at) 
	VALUES (:client, :status, :address, :created_at)`

	// Выполнение запроса
	result, err := s.db.Exec(query, sql.Named("client", p.Client), sql.Named("status", p.Status),
		sql.Named("address", p.Address), sql.Named("created_at", p.CreatedAt))
	if err != nil {
		return 0, err
	}

	// Получение идентификатора добавленной записи
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	// SQL-запрос для получения одной записи по номеру посылки
	query := `SELECT number, client, status, address, created_at FROM parcel WHERE number = :number`

	// Выполнение запроса
	row := s.db.QueryRow(query, sql.Named("number", number))

	// Создание объекта для заполнения данными из БД
	var p Parcel
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return p, err // Возвращаем только оригинальную ошибку sql.ErrNoRows
		}
		return p, err // Возвращаем оригинальную ошибку
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// SQL-запрос для получения всех посылок по клиенту
	query := `SELECT number, client, status, address, created_at FROM parcel WHERE client = :client`

	// Выполнение запроса
	rows, err := s.db.Query(query, sql.Named("client", client))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Срез для хранения полученных посылок
	var parcels []Parcel

	// Чтение всех строк из результата запроса
	for rows.Next() {
		var p Parcel
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		parcels = append(parcels, p)
	}

	// Проверка на наличие ошибок после завершения итерации
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return parcels, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// SQL-запрос для обновления статуса посылки по её номеру
	query := `UPDATE parcel SET status = :status WHERE number = :number`

	// Выполнение запроса
	_, err := s.db.Exec(query, sql.Named("status", status), sql.Named("number", number))
	if err != nil {
		return err
	}

	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// SQL-запрос для обновления адреса, если статус посылки равен 'зарегистрирован'
	query := `
		UPDATE parcel 
		SET address = :address 
		WHERE number = :number AND status = :status`

	// Выполнение запроса
	_, err := s.db.Exec(query,
		sql.Named("address", address),
		sql.Named("number", number),
		sql.Named("status", ParcelStatusRegistered),
	)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении адреса: %w", err)
	}

	return nil
}

func (s ParcelStore) Delete(number int) error {
	// Выполняем запрос на удаление только если статус посылки 'registered'
	_, err := s.db.Exec(
		"DELETE FROM parcel WHERE number = :number AND status = :status",
		sql.Named("number", number),
		sql.Named("status", ParcelStatusRegistered),
	)

	if err != nil {
		return fmt.Errorf("ошибка выполнения запроса на удаление: %w", err)
	}

	// Если ошибка не возникла, но ни одна строка не была затронута, значит посылка не найдена или её статус не 'registered'
	return nil
}
