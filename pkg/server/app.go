package server

import (
	"database/sql"
	"github.com/go-chi/chi"
	"log"
	"net/http"
	"strconv"
	"user_service/pkg/handler"
	"user_service/pkg/storage"

	_ "github.com/mattn/go-sqlite3"
)

func WebService(port int) {

	router := chi.NewRouter()

	//коннект с подежкой внешних ключей
	connect, err := sql.Open("sqlite3", "data-base.db?_foreign_keys=on")
	if err != nil {
		log.Fatalln(err)
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(connect)

	currentStorage := storage.MakeStorage(connect)

	log.Println("Run app - port: ", port)

	// хендлер создания пользователя
	router.Post("/create", handler.CreateUser(currentStorage))

	// хендлер делает друзей из двух пользователей
	router.Post("/make_friends", handler.MakeFriends(currentStorage))

	// хендлер удаляет пользователя
	router.Delete("/user", handler.DeleteUser(currentStorage))

	router.Get("/friends/{user_id}", handler.GetAllFriends(currentStorage))

	// хендлер обновляет возраст пользователя§
	router.Put("/user/{user_id}", handler.UpdateAge(currentStorage))

	// получает пользователя
	router.Get("/user/{user_id}", handler.GetUser(currentStorage))

	// хендлер вывводит всех пользователей
	router.Get("/get_users", handler.GetAllUsers(currentStorage))

	// заглушка
	router.Get("/", handler.GetRoot())

	// Очистка всех данных
	router.Delete("/flush", handler.Flush(currentStorage))

	log.Println(http.ListenAndServe("localhost:"+strconv.Itoa(port), router))
}
