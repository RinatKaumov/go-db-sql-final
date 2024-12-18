package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered, // используем константу "registered"
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // подключаемся к файлу tracker.db
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db) // создаем хранилище для работы с базой данных
	parcel := getTestParcel()   // создаем тестовую посылку

	// add
	id, err := store.Add(parcel) // добавляем новую посылку
	require.NoError(t, err)
	require.NotZero(t, id)
	parcel.Number = id // сохраняем идентификатор посылки

	// get
	retrievedParcel, err := store.Get(parcel.Number) // получаем посылку по идентификатору
	require.NoError(t, err)
	require.Equal(t, parcel.Client, retrievedParcel.Client)
	require.Equal(t, ParcelStatusRegistered, retrievedParcel.Status) // проверяем статус "registered"
	require.Equal(t, parcel.Address, retrievedParcel.Address)
	require.Equal(t, parcel.CreatedAt, retrievedParcel.CreatedAt)

	// delete
	err = store.Delete(parcel.Number) // удаляем посылку
	require.NoError(t, err)

	_, err = store.Get(parcel.Number) // проверяем, что посылка удалена
	require.Error(t, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // подключаемся к файлу tracker.db
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)
	parcel.Number = id

	// set address
	newAddress := "new test address"
	err = store.SetAddress(parcel.Number, newAddress)
	require.NoError(t, err)

	// check
	updatedParcel, err := store.Get(parcel.Number)
	require.NoError(t, err)
	require.Equal(t, newAddress, updatedParcel.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // подключаемся к файлу tracker.db
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)
	parcel.Number = id

	// set status
	newStatus := ParcelStatusSent // используем статус "sent"
	err = store.SetStatus(parcel.Number, newStatus)
	require.NoError(t, err)

	// check
	updatedParcel, err := store.Get(parcel.Number)
	require.NoError(t, err)
	require.Equal(t, ParcelStatusSent, updatedParcel.Status) // проверяем статус "sent"
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // подключаемся к файлу tracker.db
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	for i := range parcels {
		parcels[i].Client = client
	}

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotZero(t, id)
		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcels))

	// check
	for _, parcel := range storedParcels {
		original, exists := parcelMap[parcel.Number]
		require.True(t, exists)
		require.Equal(t, original.Client, parcel.Client)
		require.Equal(t, ParcelStatusRegistered, parcel.Status) // проверяем статус "registered"
		require.Equal(t, original.Address, parcel.Address)
		require.Equal(t, original.CreatedAt, parcel.CreatedAt)
	}
}
