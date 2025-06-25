// Пакет database предоставляет функционал для работы с PostgreSQL базой данных.
// Использует драйвер pgx для подключения и выполнения запросов.
package database

import (
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// DBStore представляет хранилище данных с подключением к БД.
type DBStore struct {
	DatabaseConf string  // Строка подключения к БД
	DB           *sql.DB // Подключение к базе данных
}

// InitDB инициализирует и возвращает новое подключение к БД.
//
// Параметры:
//   - conn: строка подключения к БД в формате "postgres://user:password@host:port/dbname"
//
// Возвращает:
//   - *DBStore: объект подключения к БД
//   - nil: если строка подключения пустая или не удалось подключиться
//
// Функция выполняет:
//  1. Проверку строки подключения
//  2. Установку соединения
//  3. Проверку доступности через ping
//
// Пример использования:
//
//	db := InitDB("postgres://user:pass@localhost:5432/mydb")
//	defer db.Close()
func InitDB(conn string) *DBStore {
	if conn == "" {
		return nil
	}
	db := NewDB(conn)
	if err := db.Open(); err != nil {
		log.Println("Don't connect DataBase")
		log.Fatal(err)
		return nil
	}
	if err := db.PingDB(); err != nil {
		log.Println("Don't ping DataBase")
		log.Fatal(err)
		return nil
	}
	return db
}

// NewDB создает новый объект DBStore без подключения к БД.
//
// Параметры:
//   - conf: строка подключения к БД
//
// Возвращает:
//   - *DBStore: объект с сохраненной конфигурацией подключения
//
// Примечание:
// Для установки соединения необходимо вызвать метод Open()
func NewDB(conf string) *DBStore {
	return &DBStore{
		DatabaseConf: conf,
	}
}

// Open устанавливает соединение с базой данных.
//
// Возвращает:
//   - error: ошибка подключения или ping проверки
//
// При успешном подключении сохраняет соединение в поле DB структуры DBStore.
func (store *DBStore) Open() error {

	db, err := sql.Open("pgx", store.DatabaseConf)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	store.DB = db
	return nil
}

// Close закрывает соединение с базой данных.
// Рекомендуется вызывать при завершении работы с БД.
func (store *DBStore) Close() {
	store.DB.Close()
}

// PingDB проверяет доступность базы данных.
//
// Возвращает:
//   - error: ошибка если соединение не активно
//
// Используется для проверки работоспособности подключения.
func (store *DBStore) PingDB() error {
	if err := store.DB.Ping(); err != nil {
		log.Println("don't ping Database")
		return err
	}
	return nil
}
