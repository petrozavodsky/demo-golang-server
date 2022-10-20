package handler

import (
	"bytes"
	"database/sql"
	"github.com/kinbiko/jsonassert"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"user_service/pkg/storage"
)

var currentStorage *storage.Storage

func TestMain(m *testing.M) {

	connect, err := sql.Open("sqlite3", os.TempDir()+"test-base.db?_foreign_keys=on")

	if err != nil {
		log.Fatal(err)
	}

	currentStorage = storage.MakeStorage(connect)
	connect.Exec("create table us_users ( id INTEGER not null constraint us_users_pk primary key autoincrement, name TEXT not null, age INTEGER not null);")
	connect.Exec("create table us_users ( id INTEGER not null constraint us_users_pk primary key autoincrement, name TEXT not null, age INTEGER not null);")
	connect.Exec("create table us_friends ( id INTEGER not null constraint us_friends_pk primary key autoincrement, user_id INTEGER constraint us_friends_us_users_id_fk references us_users on update cascade on delete cascade, friend_id INTEGER not null constraint us_friends_us_users_id_fk_2 references us_users on update cascade on delete cascade );")
	connect.Exec("create index us_friends_friend_id_index on us_friends (friend_id);")
	connect.Exec("create index us_friends_user_id_index on us_friends (user_id);")
	exitVal := m.Run()

	defer connect.Close()

	err = os.Remove(os.TempDir() + "test-base.db")

	if err != nil {
		log.Fatal(err)
	}

	os.Exit(exitVal)
}

func TestGetRoot(t *testing.T) {

	// Формирование запроса
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Запуск кода
	r := httptest.NewRecorder()
	handler := GetRoot()
	handler.ServeHTTP(r, req)

	// Проверяем код
	if status := r.Code; status != http.StatusOK {
		t.Errorf("обработчик вернул неожиданный код состояния: %v ожидалось %v", status, http.StatusOK)
	}

	// Проверка тела ответа
	expected := `root`
	if r.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			r.Body.String(), expected)
	}
}

func TestGetAllUsers(t *testing.T) {

	// Формирование запроса
	req, err := http.NewRequest("GET", "/get_users", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Подготовка сотояния
	currentStorage.Db.Exec("INSERT INTO us_users (id,name, age) VALUES (5,'Иван', 23);")

	// Сброс сотояния
	defer currentStorage.Db.Exec("DELETE us_users;")

	// Запуск кода
	r := httptest.NewRecorder()
	handler := GetAllUsers(currentStorage)
	handler.ServeHTTP(r, req)

	// Проверяем код
	if status := r.Code; status != http.StatusOK {
		t.Errorf("обработчик вернул неожиданный код состояния: %v ожидалось %v", status, http.StatusOK)
	}

	// Проверка тела ответа )
	jsonassert.New(t).Assertf(r.Body.String(), "[{\"id\":5,\"name\":\"Иван\",\"age\":23}]")

}

func TestCreateUser(t *testing.T) {

	// Формирование запроса
	var jsonStr = []byte(`{"name": "test_user","age": 22}`)
	req, err := http.NewRequest("POST", "/create", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}

	// Запуск кода
	r := httptest.NewRecorder()
	handler := CreateUser(currentStorage)
	handler.ServeHTTP(r, req)

	// Проверяем код
	if status := r.Code; status != http.StatusCreated {
		t.Errorf("обработчик вернул неожиданный код состояния: %v ожидалось %v", status, http.StatusCreated)
	}

	// Проверка тела ответа
	jsonassert.New(t).Assertf(r.Body.String(), "{\"Message\":\"Пользователь test_user с ID 0 создан\\n\"}")

	// Проверяем состояни
	row := currentStorage.Db.QueryRow("SELECT name FROM us_users ORDER BY id DESC")
	var name string
	row.Scan(&name)

	if "test_user" != name {
		t.Fatal(err, name)
	}

}

func TestMakeFriends(t *testing.T) {

	// Формирование запроса
	var jsonStr = []byte(`{"source_id": "10","target_id": "20"}`)
	req, err := http.NewRequest("POST", "/make_friends", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}

	// Подготовка сотояния
	ins := "INSERT INTO us_users (`id`, `name`, `age`) VALUES (10, 'Алексей', 32), (20, 'Дмитрий', 45);\n"
	_, err = currentStorage.Db.Exec(ins)

	if err != nil {
		t.Fatal(err)
	}

	// Сброс сотояния
	defer currentStorage.Db.Exec("DELETE us_users;\n DELETE us_friends;")

	// Запуск кода
	r := httptest.NewRecorder()
	handler := MakeFriends(currentStorage)
	handler.ServeHTTP(r, req)

	// Проверяем код
	if status := r.Code; status != http.StatusOK {
		t.Errorf("обработчик вернул неожиданный код состояния: %v ожидалось %v", status, http.StatusOK)
	}

	// Проверка тела ответа
	jsonassert.New(t).Assertf(r.Body.String(), "{\"Message\":\"Алексей и Дмитрий теперь друзь\\n\"}")

	// Проверяем состояни
	rows, err := currentStorage.Db.Query("SELECT friend_id FROM us_friends")
	friends := make([]int, 0)

	if err == nil {
		for rows.Next() {
			var id int
			rows.Scan(&id)

			friends = append(friends, id)
		}
	}

	if friends[0] != 10 || friends[1] != 20 {
		t.Error("Связь не создана")
	}

}

func TestDeleteUser(t *testing.T) {
	// Формирование запроса
	var jsonStr = []byte(`{"target_id": "7"}`)
	req, err := http.NewRequest("DELETE", "/user", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}

	// Подготовка сотояния
	currentStorage.Db.Exec("INSERT INTO us_users (id,name, age) VALUES (7,'Леонид', 54);")

	//Сброс сотояния
	defer currentStorage.Db.Exec("DELETE us_users;")

	// Запуск кода
	r := httptest.NewRecorder()
	handler := DeleteUser(currentStorage)
	handler.ServeHTTP(r, req)

	// Проверяем код
	if status := r.Code; status != http.StatusOK {
		t.Errorf("обработчик вернул неожиданный код состояния: %v ожидалось %v", status, http.StatusOK)
	}
	// Проверка тела ответа
	jsonassert.New(t).Assertf(r.Body.String(), "{\"Message\":\"Пользователь Леонид удален\\n\"}")

	// Проверяем состояни
	row := currentStorage.Db.QueryRow("SELECT id FROM us_users WHERE id=7;")

	var id int

	row.Scan(&id)

	if id != 0 {
		t.Fatal("Пользователь не удалился")
	}
}
