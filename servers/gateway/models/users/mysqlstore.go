package users

import (
	"database/sql"
	"strings"

	"github.com/UW-Info-441-Winter-Quarter-2020/homework-GarsonYang/servers/gateway/indexes"
)

type SqlStore struct {
	db *sql.DB
}

const userAttr = "email, pass_hash, user_name, first_name, last_name, photo_URL"

func NewSqlStore(db *sql.DB) *SqlStore {
	if db == nil {
		panic("nil database pointer passed to NewSqlStore")
	}
	return &SqlStore{
		db: db,
	}
}

func (s *SqlStore) Insert(user *User) (*User, error) {
	insq := "insert into users(" + userAttr + ") values (?,?,?,?,?,?)"
	res, err := s.db.Exec(insq, user.Email, user.PassHash, user.UserName, user.FirstName, user.LastName, user.PhotoURL)
	if err != nil {
		return user, err
	}

	newId, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	user.ID = newId
	return user, nil
}

func (s *SqlStore) GetByID(id int64) (*User, error) {
	rows, err := s.db.Query("select id, "+userAttr+" from users where id =?", id)
	if err != nil {
		return &User{}, err
	}
	defer rows.Close()

	u := &User{}
	rows.Next()
	if err := rows.Scan(&u.ID, &u.Email, &u.PassHash, &u.UserName,
		&u.FirstName, &u.LastName, &u.PhotoURL); err != nil {
		return u, err
	}

	return u, nil
}

func (s *SqlStore) GetByEmail(email string) (*User, error) {
	rows, err := s.db.Query("select id, "+userAttr+" from users where email =?", email)
	if err != nil {
		return &User{}, err
	}
	defer rows.Close()

	u := &User{}
	rows.Next()
	if err := rows.Scan(&u.ID, &u.Email, &u.PassHash, &u.UserName,
		&u.FirstName, &u.LastName, &u.PhotoURL); err != nil {
		return u, err
	}

	return u, nil
}

func (s *SqlStore) GetByUserName(username string) (*User, error) {
	rows, err := s.db.Query("select id, "+userAttr+" from users where user_name = ?", username)
	if err != nil {
		return &User{}, err
	}
	defer rows.Close()

	u := &User{}
	rows.Next()
	if err := rows.Scan(&u.ID, &u.Email, &u.PassHash, &u.UserName,
		&u.FirstName, &u.LastName, &u.PhotoURL); err != nil {
		return u, err
	}

	return u, nil
}

func (s *SqlStore) Update(id int64, updates *Updates) (*User, error) {
	u, err := s.GetByID(id)
	if err != nil {
		return u, err
	}

	err = u.ApplyUpdates(updates)
	if err != nil {
		return u, err
	}

	upq := "update users set email=?, pass_hash=?, user_name=?, first_name=?, last_name=?, photo_URL=? where id = ?"
	_, err = s.db.Exec(upq, u.Email, u.PassHash, u.UserName, u.FirstName, u.LastName, u.PhotoURL, u.ID)
	return u, err
}

func (s *SqlStore) Delete(id int64) error {
	_, err := s.db.Exec("delete from users where id = ?", id)
	return err
}

// laodToTrie loads existing user accounts in the db into the trie for searching
func LoadToTrie(db *sql.DB, root *indexes.TrieNode) error {
	rows, err := db.Query("select id, user_name, first_name, last_name from users")
	if err != nil {
		return err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		u := &User{}
		err = rows.Scan(&u.ID, &u.UserName, &u.FirstName, &u.LastName)
		if err != nil {
			return err
		}
		users = append(users, *u)
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	for _, u := range users {
		values := []string{}
		values = append(values, strings.ToLower(u.UserName))
		values = append(values, strings.Split(strings.ToLower(u.FirstName), " ")...)
		values = append(values, strings.Split(strings.ToLower(u.LastName), " ")...)
		for _, v := range values {
			if len(v) > 0 {
				root.Add(v, u.ID)
			}
		}
	}

	return nil
}
