package users

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestUserInsert(t *testing.T) {
	cases := []struct {
		name       string
		u          *User
		expectedID int64
	}{
		{
			"test case 1",
			&User{
				Email:    "abc@uw.edu",
				PassHash: []byte("goodpwd"),
				UserName: "goodboy",
				PhotoURL: "some-url",
			},
			1,
		},
		{
			"test case 2",
			&User{
				Email:    "abcd@uw.edu",
				PassHash: []byte("goodpwd"),
				UserName: "goodboy2",
				PhotoURL: "some-new-url",
			},
			2,
		},
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating sqlmock : %v", err)
	}
	defer db.Close()

	SqlStore := NewSqlStore(db)

	for _, c := range cases {
		expectedSQL := regexp.QuoteMeta("insert into users(" + userAttr + ") values (?,?,?,?,?,?)")
		mock.ExpectExec(expectedSQL).WithArgs(c.u.Email, c.u.PassHash, c.u.UserName, c.u.FirstName, c.u.LastName, c.u.PhotoURL).
			WillReturnResult(sqlmock.NewResult(c.expectedID, 1))

		insertedUser, err := SqlStore.Insert(c.u)
		if err != nil {
			t.Errorf("unexpected error during successful insert: %v", err)
		}
		if insertedUser == nil {
			t.Errorf("nil user returned from insert")
		} else if insertedUser.ID != c.expectedID {
			t.Errorf("case %s: incorrect new ID: expected %d but got %d", c.name, c.expectedID, insertedUser.ID)
		}
	}
}

func TestGetByID(t *testing.T) {
	cases := []struct {
		u            *User
		newID        int64
		expectedRows *sqlmock.Rows
	}{
		{
			&User{
				Email:    "abc@uw.edu",
				PassHash: []byte("goodpwd"),
				UserName: "goodboy",
				PhotoURL: "some-url",
			},
			1,
			sqlmock.NewRows(append([]string{"id"}, strings.Split(userAttr, ", ")...)).
				AddRow(1, "abc@uw.edu", []byte("goodpwd"), "goodboy", "", "", "some-url"),
		},
		{
			&User{
				Email:    "abcd@uw.edu",
				PassHash: []byte("goodpwd"),
				UserName: "goodboy2",
				PhotoURL: "some-new-url",
			},
			2,
			sqlmock.NewRows(append([]string{"id"}, strings.Split(userAttr, ", ")...)).
				AddRow(2, "abcd@uw.edu", []byte("goodpwd"), "goodboy2", "", "", "some-new-url"),
		},
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating sqlmock : %v", err)
	}
	defer db.Close()

	SqlStore := NewSqlStore(db)
	for _, c := range cases {
		expectedSQL := regexp.QuoteMeta("insert into users(" + userAttr + ") values (?,?,?,?,?,?)")
		mock.ExpectExec(expectedSQL).WithArgs(c.u.Email, c.u.PassHash, c.u.UserName, c.u.FirstName, c.u.LastName, c.u.PhotoURL).
			WillReturnResult(sqlmock.NewResult(c.newID, 1))

		insertedUser, err := SqlStore.Insert(c.u)
		if err != nil {
			t.Fatalf("unexpected error during successful insert: %v", err)
		}
		if insertedUser == nil {
			t.Fatalf("nil user returned from insert")
		}
	}

	for _, c := range cases {
		expectedSQL := regexp.QuoteMeta("select id, " + userAttr + " from users where id =?")
		mock.ExpectQuery(expectedSQL).WithArgs(c.newID).WillReturnRows(c.expectedRows)

		user, err := SqlStore.GetByID(c.newID)
		if err != nil {
			t.Errorf("unexpected error when retrieving data: %v", err)
		} else if user.Email != c.u.Email {
			t.Errorf("incorrect user retrieved: expected %s but got %s", c.u.Email, user.Email)
		}
	}

	expectedSQL := regexp.QuoteMeta("select id, " + userAttr + " from users where id =?")
	mock.ExpectQuery(expectedSQL).WithArgs(3).WillReturnError(sql.ErrNoRows)

	_, err = SqlStore.GetByID(3)
	if err != sql.ErrNoRows {
		t.Errorf("expected ErrNoRow but actually got: %v", err)
	}

	rowErr := fmt.Errorf("row error")
	expectedRows := sqlmock.NewRows(append([]string{"id"}, strings.Split(userAttr, ", ")...)).
		AddRow(1, "abcd@uw.edu", []byte("goodpwd"), "goodboy2", "", "", "some-new-url").RowError(0, rowErr)
	mock.ExpectQuery(expectedSQL).WithArgs(4).WillReturnRows(expectedRows)

	_, err = SqlStore.GetByID(4)
	if err != rowErr {
		t.Errorf("expected row error but actually got: %v", err)
	}
}

func TestGetByEmail(t *testing.T) {
	cases := []struct {
		u            *User
		newID        int64
		expectedRows *sqlmock.Rows
	}{
		{
			&User{
				Email:    "abc@uw.edu",
				PassHash: []byte("goodpwd"),
				UserName: "goodboy",
				PhotoURL: "some-url",
			},
			1,
			sqlmock.NewRows(append([]string{"id"}, strings.Split(userAttr, ", ")...)).
				AddRow(1, "abc@uw.edu", []byte("goodpwd"), "goodboy", "", "", "some-url"),
		},
		{
			&User{
				Email:    "abcd@uw.edu",
				PassHash: []byte("goodpwd"),
				UserName: "goodboy2",
				PhotoURL: "some-new-url",
			},
			2,
			sqlmock.NewRows(append([]string{"id"}, strings.Split(userAttr, ", ")...)).
				AddRow(2, "abcd@uw.edu", []byte("goodpwd"), "goodboy2", "", "", "some-new-url"),
		},
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating sqlmock : %v", err)
	}
	defer db.Close()

	SqlStore := NewSqlStore(db)
	for _, c := range cases {
		expectedSQL := regexp.QuoteMeta("insert into users(" + userAttr + ") values (?,?,?,?,?,?)")
		mock.ExpectExec(expectedSQL).WithArgs(c.u.Email, c.u.PassHash, c.u.UserName, c.u.FirstName, c.u.LastName, c.u.PhotoURL).
			WillReturnResult(sqlmock.NewResult(c.newID, 1))

		insertedUser, err := SqlStore.Insert(c.u)
		if err != nil {
			t.Fatalf("unexpected error during successful insert: %v", err)
		}
		if insertedUser == nil {
			t.Fatalf("nil user returned from insert")
		}
	}

	for _, c := range cases {
		expectedSQL := regexp.QuoteMeta("select id, " + userAttr + " from users where email =?")
		mock.ExpectQuery(expectedSQL).WithArgs(c.u.Email).WillReturnRows(c.expectedRows)

		user, err := SqlStore.GetByEmail(c.u.Email)
		if err != nil {
			t.Errorf("unexpected error when retrieving data: %v", err)
		} else if user.Email != c.u.Email {
			t.Errorf("incorrect user retrieved: expected %s but got %s", c.u.Email, user.Email)
		}
	}

	expectedSQL := regexp.QuoteMeta("select id, " + userAttr + " from users where email =?")
	mock.ExpectQuery(expectedSQL).WithArgs("ac").WillReturnError(sql.ErrNoRows)

	_, err = SqlStore.GetByEmail("ac")
	if err != sql.ErrNoRows {
		t.Errorf("expected ErrNoRow but actually got: %v", err)
	}

	rowErr := fmt.Errorf("row error")
	expectedRows := sqlmock.NewRows(append([]string{"id"}, strings.Split(userAttr, ", ")...)).
		AddRow(1, "abcd@uw.edu", []byte("goodpwd"), "goodboy2", "", "", "some-new-url").RowError(0, rowErr)
	mock.ExpectQuery(expectedSQL).WithArgs("ad").WillReturnRows(expectedRows)

	_, err = SqlStore.GetByEmail("ad")
	if err != rowErr {
		t.Errorf("expected row error but actually got: %v", err)
	}
}

func TestGetByUserName(t *testing.T) {
	cases := []struct {
		u            *User
		newID        int64
		expectedRows *sqlmock.Rows
	}{
		{
			&User{
				Email:    "abc@uw.edu",
				PassHash: []byte("goodpwd"),
				UserName: "goodboy",
				PhotoURL: "some-url",
			},
			1,
			sqlmock.NewRows(append([]string{"id"}, strings.Split(userAttr, ", ")...)).
				AddRow(1, "abc@uw.edu", []byte("goodpwd"), "goodboy", "", "", "some-url"),
		},
		{
			&User{
				Email:    "abcd@uw.edu",
				PassHash: []byte("goodpwd"),
				UserName: "goodboy2",
				PhotoURL: "some-new-url",
			},
			2,
			sqlmock.NewRows(append([]string{"id"}, strings.Split(userAttr, ", ")...)).
				AddRow(2, "abcd@uw.edu", []byte("goodpwd"), "goodboy2", "", "", "some-new-url"),
		},
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating sqlmock : %v", err)
	}
	defer db.Close()

	SqlStore := NewSqlStore(db)
	for _, c := range cases {
		expectedSQL := regexp.QuoteMeta("insert into users(" + userAttr + ") values (?,?,?,?,?,?)")
		mock.ExpectExec(expectedSQL).WithArgs(c.u.Email, c.u.PassHash, c.u.UserName, c.u.FirstName, c.u.LastName, c.u.PhotoURL).
			WillReturnResult(sqlmock.NewResult(c.newID, 1))

		insertedUser, err := SqlStore.Insert(c.u)
		if err != nil {
			t.Fatalf("unexpected error during successful insert: %v", err)
		}
		if insertedUser == nil {
			t.Fatalf("nil user returned from insert")
		}
	}

	for _, c := range cases {
		expectedSQL := regexp.QuoteMeta("select id, " + userAttr + " from users where user_name = ?")
		mock.ExpectQuery(expectedSQL).WithArgs(c.u.UserName).WillReturnRows(c.expectedRows)

		user, err := SqlStore.GetByUserName(c.u.UserName)
		if err != nil {
			t.Errorf("unexpected error when retrieving data: %v", err)
		} else if user.Email != c.u.Email {
			t.Errorf("incorrect user retrieved: expected %s but got %s", c.u.Email, user.Email)
		}
	}

	expectedSQL := regexp.QuoteMeta("select id, " + userAttr + " from users where user_name = ?")
	mock.ExpectQuery(expectedSQL).WithArgs("ac").WillReturnError(sql.ErrNoRows)

	_, err = SqlStore.GetByUserName("ac")
	if err != sql.ErrNoRows {
		t.Errorf("expected ErrNoRow but actually got: %v", err)
	}

	rowErr := fmt.Errorf("row error")
	expectedRows := sqlmock.NewRows(append([]string{"id"}, strings.Split(userAttr, ", ")...)).
		AddRow(1, "abcd@uw.edu", []byte("goodpwd"), "goodboy2", "", "", "some-new-url").RowError(0, rowErr)
	mock.ExpectQuery(expectedSQL).WithArgs("ad").WillReturnRows(expectedRows)

	_, err = SqlStore.GetByUserName("ad")
	if err != rowErr {
		t.Errorf("expected row error but actually got: %v", err)
	}
}

func TestDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating sqlmock : %v", err)
	}
	defer db.Close()
	SqlStore := NewSqlStore(db)

	u := &User{
		Email:    "abc@uw.edu",
		PassHash: []byte("goodpwd"),
		UserName: "goodboy",
		PhotoURL: "some-url",
	}

	newID := (int64)(1)
	expectedSQL := regexp.QuoteMeta("insert into users(" + userAttr + ") values (?,?,?,?,?,?)")
	mock.ExpectExec(expectedSQL).WithArgs(u.Email, u.PassHash, u.UserName, u.FirstName, u.LastName, u.PhotoURL).
		WillReturnResult(sqlmock.NewResult(newID, 1))

	insertedUser, err := SqlStore.Insert(u)
	if err != nil {
		t.Fatalf("unexpected error during successful insert: %v", err)
	}
	if insertedUser == nil {
		t.Fatalf("nil user returned from insert")
	}

	mock.ExpectExec("delete from users where id = ?").WithArgs(newID).
		WillReturnResult(sqlmock.NewResult(newID, 1))
	err = SqlStore.Delete(newID)
	if err != nil {
		t.Fatalf("unexpected error when deleting: %v", err)
	}

	expectedSQL = regexp.QuoteMeta("select id, " + userAttr + " from users where id =?")
	mock.ExpectQuery(expectedSQL).WithArgs(newID).WillReturnError(sql.ErrNoRows)
	_, err = SqlStore.GetByID(newID)
	if err != sql.ErrNoRows {
		t.Errorf("delete failed: %v", err)
	}

	mock.ExpectExec("delete from users where id = ?").WithArgs(newID).
		WillReturnResult(sqlmock.NewResult(newID, 1))
	err = SqlStore.Delete(newID)
	if err != nil {
		t.Fatalf("unexpected error when deleting: %v", err)
	}
}

// func TestSqlUpdate(t *testing.T) {
// 	db, mock, err := sqlmock.New()
// 	if err != nil {
// 		t.Fatalf("error creating sqlmock : %v", err)
// 	}
// 	defer db.Close()
// 	SqlStore := NewSqlStore(db)

// 	cases := []struct {
// 		u            *User
// 		newID        int64
// 		expectedRows *sqlmock.Rows
// 		up           *Updates
// 		updateTarget int64
// 		expectedErr  error
// 	}{
// 		{
// 			&User{
// 				Email:    "abc@uw.edu",
// 				PassHash: []byte("goodpwd"),
// 				UserName: "goodboy",
// 				PhotoURL: "some-url",
// 			},
// 			1,
// 			sqlmock.NewRows(append([]string{"id"}, strings.Split(userAttr, ", ")...)).
// 				AddRow(1, "abc@uw.edu", []byte("goodpwd"), "goodboy", "", "", "some-url"),
// 			&Updates{
// 				FirstName: "Mo",
// 				LastName:  "Ki",
// 			},
// 			1,
// 			nil,
// 		},
// 		{
// 			&User{
// 				Email:    "abc@uw.edu",
// 				PassHash: []byte("goodpwd"),
// 				UserName: "goodboy",
// 				PhotoURL: "some-url",
// 			},
// 			1,
// 			nil,
// 			&Updates{
// 				FirstName: "Mo",
// 				LastName:  "Ki",
// 			},
// 			2,
// 			sql.ErrNoRows,
// 		},
// 	}

// 	for _, c := range cases {
// 		expectedSQL := regexp.QuoteMeta("insert into users(" + userAttr + ") values (?,?,?,?,?,?)")
// 		mock.ExpectExec(expectedSQL).WithArgs(c.u.Email, c.u.PassHash, c.u.UserName, c.u.FirstName, c.u.LastName, c.u.PhotoURL).
// 			WillReturnResult(sqlmock.NewResult(c.newID, 1))

// 		insertedUser, err := SqlStore.Insert(c.u)
// 		if err != nil {
// 			t.Fatalf("unexpected error during successful insert: %v", err)
// 		}
// 		if insertedUser == nil {
// 			t.Fatalf("nil user returned from insert")
// 		}

// 		expectedSQL = regexp.QuoteMeta("select id, " + userAttr + " from users where id =?")
// 		if c.expectedRows == nil {
// 			mock.ExpectQuery(expectedSQL).WithArgs(c.updateTarget).WillReturnError(sql.ErrNoRows)
// 			c.expectedErr = sql.ErrNoRows
// 		} else {
// 			mock.ExpectQuery(expectedSQL).WithArgs(c.updateTarget).WillReturnRows(c.expectedRows)
// 			mock.ExpectExec("delete from users where id = ?").WithArgs(c.updateTarget).
// 				WillReturnResult(sqlmock.NewResult(c.updateTarget, 1))
// 			expectedSQL = regexp.QuoteMeta("insert into users(" + userAttr + ") values (?,?,?,?,?,?)")
// 			mock.ExpectExec(expectedSQL).WithArgs(c.u.Email, c.u.PassHash, c.u.UserName, c.up.FirstName, c.up.LastName, c.u.PhotoURL).
// 				WillReturnResult(sqlmock.NewResult(c.updateTarget, 1))
// 		}

// 		user, err := SqlStore.Update(c.updateTarget, c.up)
// 		if err != c.expectedErr {
// 			t.Fatalf("unexpected error during updating: expected %v, actually got %v", c.expectedErr, err)
// 		}

// 		if err == nil && (user.FirstName != c.up.FirstName || user.LastName != c.up.LastName) {
// 			t.Errorf("update didn't work")
// 		}
// 	}

// }
