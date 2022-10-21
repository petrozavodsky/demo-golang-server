package storage

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
	"strings"
	user "user_service/pkg/user"

	_ "github.com/mattn/go-sqlite3"
)

// Storage - хранилище пользователй
type Storage struct {
	Db *sql.DB
}

// MakeStorage - создает экземпляр хранилище пользователей
func MakeStorage(connect *sql.DB) *Storage {

	return &Storage{
		connect,
	}
}

func (s *Storage) SaveUser(user user.User) int {

	uid := user.GetId()

	if 1 > uid {
		q := "INSERT INTO us_users(name, age) VALUES (?, ?);"
		stmt, err := s.Db.Prepare(q)
		if err != nil {
			log.Fatal(err)
		}
		ret, err := stmt.Exec(user.GetName(), user.GetAge())

		if err != nil {
			log.Fatal(err)
		}
		returnId, err := ret.LastInsertId()

		user.SetId(int(returnId))

		s.SaveFriends(&user, user.GetFriends())

		return uid
	}

	chRow := s.Db.QueryRow("SELECT id FROM us_users WHERE id=?", user.GetId())
	var chId int
	chRow.Scan(&chId)

	if chId < 1 {
		return chId
	}

	q := "UPDATE us_users SET name =?, age=? WHERE id = ?;"
	stmt, err := s.Db.Prepare(q)
	if err != nil {
		log.Fatal(err)
	}
	_, err = stmt.Exec(user.GetName(), user.GetAge(), user.GetId())

	if err != nil {
		log.Fatal(err)
	}

	s.SaveFriends(&user, user.GetFriends())

	return user.GetId()
}

func (s *Storage) SaveFriends(user *user.User, friends []int) {
	uid := user.GetId()
	existsFriends := s.existsFriends(uid)
	var addFriends []int
	var delFriends []int

	if len(existsFriends) < 1 {
		addFriends = friends
	} else {
		for i := range friends {
			if !inArray(friends[i], existsFriends) {
				addFriends = appendIfMissing(addFriends, friends[i])
			}
		}

	}

	if len(addFriends) > 0 {
		s.addFriends(uid, addFriends)
	}

	if len(existsFriends) > 0 {

		if len(friends) >= 1 {

			for i := range existsFriends {

				if !inArray(existsFriends[i], friends) {
					delFriends = appendIfMissing(delFriends, existsFriends[i])
				}
			}

		} else {
			delFriends = existsFriends
		}

	}

	if len(delFriends) > 0 {
		s.deleteFriends(uid, delFriends)
	}
}

// appendIfMissing хелпер добавления только уникальных эллементов
func appendIfMissing(slice []int, i int) []int {
	for _, ele := range slice {
		if ele == i {
			return slice
		}
	}

	slice = append(slice, i)
	return slice
}

// inArray хелпер проверки вхождения
func inArray(what interface{}, where []int) bool {
	for _, v := range where {
		if v == what {
			return true
		}
	}
	return false
}

// existsFriends хелпер получения друзей
func (s *Storage) existsFriends(uid int) []int {
	var existsFriends []int
	rows, _ := s.Db.Query("SELECT friend_id FROM us_friends WHERE user_id=?", uid)

	for rows.Next() {
		var fid int

		if err := rows.Scan(&fid); err != nil {
			log.Fatalf("could not scan row: %v", err)
		}

		existsFriends = appendIfMissing(existsFriends, fid)
	}
	return existsFriends
}

// deleteFriends хелпер удаления друзей
func (s *Storage) deleteFriends(uid int, delFriends []int) sql.Result {
	var delFriendsStr []string

	for i := range delFriends {
		delFriendsStr = append(delFriendsStr, strconv.Itoa(delFriends[i]))
	}

	q := "DELETE FROM us_friends WHERE " +
		"( user_id =" + strconv.Itoa(uid) + " AND friend_id IN(" + strings.Join(delFriendsStr, ", ") + ") ) " +
		"OR ( user_id IN(" + strings.Join(delFriendsStr, ", ") + ") AND friend_id =" + strconv.Itoa(uid) + " )"

	output, err := s.Db.Exec(q)

	if err != nil {
		log.Fatal(err)
	}

	return output
}

// addFriends хелпер жоюавления друзей
func (s *Storage) addFriends(uid int, addFriends []int) {

	var addFriendsValuesStr string

	for i := range addFriends {
		s := " (" + strconv.Itoa(uid) + ", " + strconv.Itoa(addFriends[i]) + ") ,"
		s += " (" + strconv.Itoa(addFriends[i]) + ", " + strconv.Itoa(uid) + " )"

		if i < len(addFriends)-1 {
			s += ", "
		}

		addFriendsValuesStr += s
	}

	q := "INSERT INTO us_friends ( user_id, friend_id) VALUES " + addFriendsValuesStr

	_, _ = s.Db.Exec(q)
}

// AddUser - добавление пользователя
func (s *Storage) AddUser(name string, age int, friends []int) (userId int) {

	if name != "" {
		newUser := user.MakeUser()
		newUser.SetName(name)
		newUser.SetAge(age)
		newUser.SetFriends(friends)

		return s.SaveUser(*newUser)
	}

	return 0
}

// UpdateAge - обновление возраста
func (s *Storage) UpdateAge(id, age int) int {

	stmt, err := s.Db.Prepare("UPDATE us_users SET age=? WHERE id = ?;")
	if err != nil {
		log.Fatal(err)
	}

	_, err = stmt.Exec(
		age,
		id,
	)

	if err != nil {
		log.Fatalln(err)
	}

	return id
}

// DeleteUser - удаление пользователя
func (s *Storage) DeleteUser(id int) int {

	stmt, err := s.Db.Prepare("DELETE FROM us_users WHERE id = ?;")
	if err != nil {
		log.Fatal(err)
	}

	_, err = stmt.Exec(id)

	return id
}

// GetAllUsers - вывод всех пользователей
func (s *Storage) GetAllUsers() []*user.User {

	rows, err := s.Db.Query("SELECT id FROM us_users")
	all := make([]*user.User, 0)
	log.Println(all)

	if err == nil {
		for rows.Next() {
			var id int
			rows.Scan(&id)
			singleUser := s.GetUser(id)
			all = append(all, singleUser)
		}
	}

	return all
}

// GetUser - вывод пользователя
func (s *Storage) GetUser(id int) *user.User {

	var userId int
	var name string
	var age int
	var friends []int

	out := s.Db.QueryRow("SELECT `id`, `name`, `age` FROM us_users WHERE id =?", id)

	err := out.Scan(&userId, &name, &age)

	if err != nil {
		return nil
	}

	rows, err := s.Db.Query("SELECT friend_id FROM us_friends WHERE user_id=?", id)

	if err == nil {
		for rows.Next() {
			var friendId int
			rows.Scan(&friendId)
			friends = append(friends, friendId)
		}
	}

	newUser := user.MakeUser()
	newUser.SetId(userId)
	newUser.SetName(name)
	newUser.SetAge(age)
	newUser.SetFriends(friends)
	return newUser
}

// GetFriendsID - получение id друзей
func (s *Storage) GetFriends(id int) ([]*user.User, error) {

	friendsUsers := make([]*user.User, 0)

	targetUser := s.GetUser(id)

	if targetUser == nil {
		return friendsUsers, errors.New("пользователь не существует")
	}

	friends := targetUser.GetFriends()

	if len(friends) > 0 {

		for i := range friends {
			friend := s.GetUser(friends[i])
			log.Println(friend, friends, friends[i])
			friendsUsers = append(friendsUsers, friend)
		}

	}

	return friendsUsers, nil
}

// MakeFriends - создает дружеские связи
func (s *Storage) MakeFriends(id int, fid int) (int, error) {

	count := 0

	row := s.Db.QueryRow("SELECT COUNT(id) FROM us_users WHERE id =? OR id=?", id, fid)

	row.Scan(&count)

	log.Println(count)

	if count < 2 {
		return id, errors.New("пользователь не существует")
	}

	targetUser := s.GetUser(id)
	friends := []int{fid}

	s.SaveFriends(targetUser, friends)

	return id, nil
}

func (s *Storage) Flush() {
	q := "DELETE FROM us_friends;\n"
	q += "DELETE FROM us_users;\n"
	q += "UPDATE `sqlite_sequence` SET `seq` = 0 WHERE `name` = 'us_friends';\n"
	q += "UPDATE `sqlite_sequence` SET `seq` = 0 WHERE `name` = 'us_users';\n"

	_, _ = s.Db.Exec(q)
}
