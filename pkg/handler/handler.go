package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"user_service/pkg/storage"
	"user_service/pkg/user"

	"github.com/go-chi/chi"
)

// Get - Заглушка
func Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("root")); err != nil {
			log.Fatalln(err)
		}
	}
}

// GetAllUsers  - http хендлер выводит всех пользователей
func GetAllUsers(s *storage.Storage) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		users := s.GetAllUsers()

		if err := json.NewEncoder(w).Encode(users); err != nil {
			log.Fatalln(err)
		}
	}
}

// CreateUser - http хендлер создания пользователя
func CreateUser(s *storage.Storage) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Fatalln(err)
			}
		}(r.Body)

		requestUser := user.User{}
		if err := json.NewDecoder(r.Body).Decode(&requestUser); err != nil {
			log.Fatalln(err)
		}

		userID := s.SaveUser(requestUser)
		response := fmt.Sprintf("Пользователь %s с ID %d создан\n", requestUser.GetName(), userID)

		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte(response)); err != nil {
			log.Fatalln(err)
		}
	}
}

// MakeFriends - http хендлер создания дружеских связей
func MakeFriends(s *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Fatalln(err)
			}
		}(r.Body)

		request := map[string]string{}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Fatalln(err)
		}

		id1, err := strconv.Atoi(request["source_id"])
		if err != nil {
			log.Fatalln(err)
		}
		id2, err := strconv.Atoi(request["target_id"])
		if err != nil {
			log.Fatalln(err)
		}

		_, err = s.MakeFriends(id1, id2)

		if err != nil {
			body := MakeBody()
			w.WriteHeader(http.StatusOK)
			body.SetMessage("Идентификатор пользователя не корректен")

			if err := json.NewEncoder(w).Encode(body); err != nil {
				log.Fatalln(err)
			}
			return
		}

		response := fmt.Sprintf("%s и %s теперь друзь\n", s.GetUser(id1).GetName(), s.GetUser(id2).GetName())
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte(response)); err != nil {
			log.Fatalln(err)
		}

		return
	}
}

// DeleteUser - http хендлер удаления пользователя
func DeleteUser(s *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Fatalln(err)
			}
		}(r.Body)

		request := map[string]string{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Fatalln(err)
		}

		id, err := strconv.Atoi(request["target_id"])
		if err != nil {
			log.Fatalln(err)
		}

		response := fmt.Sprintf("Пользователь %s удален\n", s.GetUser(id).GetName())

		s.DeleteUser(id)

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(response)); err != nil {
			log.Fatalln(err)
		}
	}
}

// GetAllFriends -  http хендлер получения друзей пользователя
func GetAllFriends(s *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		val := chi.URLParam(r, "user_id")

		userID, err := strconv.Atoi(val)

		if err != nil {
			body := MakeBody()
			w.WriteHeader(http.StatusNotFound)
			body.SetMessage("Идентификатор не корректен")

			if err := json.NewEncoder(w).Encode(body); err != nil {
				log.Fatalln(err)
			}
			return
		}

		friendsID := s.GetFriends(userID)

		if len(friendsID) < 1 {
			body := MakeBody()
			w.WriteHeader(http.StatusOK)
			body.SetMessage("Нет друзей")

			if err := json.NewEncoder(w).Encode(body); err != nil {
				log.Fatalln(err)
			}
			return
		}

		if err := json.NewEncoder(w).Encode(friendsID); err != nil {
			log.Fatalln(err)
		}

	}
}

// UpdateAge - http хендлер обновления возраста пользователя
func UpdateAge(s *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Fatalln(err)
			}
		}(r.Body)

		val := chi.URLParam(r, "user_id")

		userID, err := strconv.Atoi(val)
		if err != nil {
			body := MakeBody()
			w.WriteHeader(http.StatusNotFound)
			body.SetMessage("Идентификатор не корректен")

			if err := json.NewEncoder(w).Encode(body); err != nil {
				log.Fatalln(err)
			}
		}

		request := map[string]string{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Fatalln(err)
		}

		age, err := strconv.Atoi(request["new age"])
		if err != nil {
			body := MakeBody()
			w.WriteHeader(http.StatusOK)
			body.SetMessage("Возраст не корректен")

			if err := json.NewEncoder(w).Encode(body); err != nil {
				log.Fatalln(err)
			}
			return
		}

		s.UpdateAge(userID, age)
		response := fmt.Sprintf("Пользователь %d обновлен %d\n", userID, age)
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte(response)); err != nil {
			log.Fatalln(err)
		}

	}
}

// GetUser http хендлер получет пользователя
func GetUser(s *storage.Storage) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		val := chi.URLParam(r, "user_id")

		userID, err := strconv.Atoi(val)

		if err != nil {
			body := MakeBody()
			w.WriteHeader(http.StatusNotFound)
			body.SetMessage("Идентификатор не корректен")

			if err := json.NewEncoder(w).Encode(body); err != nil {
				log.Fatalln(err)
			}
			return
		}

		currentUser := s.GetUser(userID)

		if currentUser == nil {
			body := MakeBody()
			w.WriteHeader(http.StatusNotFound)
			body.SetMessage("Пользователь не найден")

			if err := json.NewEncoder(w).Encode(body); err != nil {
				log.Fatalln(err)
			}
			return
		}

		if err := json.NewEncoder(w).Encode(currentUser); err != nil {
			log.Fatalln(err)
		}

	}
}
