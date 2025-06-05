package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

var db *sql.DB

func Connect(connString string) {
	_db, err := sql.Open("libsql", connString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", connString, err)
		os.Exit(1)
	}

	db = _db
}

func Disconnect() {
	db.Close()
}

type UserResponse struct {
	Exist   bool
	IsAdmin bool
}

func CheckUser(mail string, ctx context.Context) (UserResponse, error) {
	var isAdminUser bool
	statement := `SELECT admin FROM users WHERE email=?`

	err := db.QueryRowContext(ctx, statement, mail).Scan(&isAdminUser)
	switch {
	case err == sql.ErrNoRows:
		return UserResponse{Exist: false, IsAdmin: false}, nil
	case err != nil:
		log.Printf("query error: %v\n", err)
		return UserResponse{}, fmt.Errorf("failed to check user: %w", err)
	}

	return UserResponse{
		Exist:   true,
		IsAdmin: isAdminUser,
	}, nil
}

func RegisterUser(mail, name, studentID string, ctx context.Context) error {
	statement := `INSERT INTO users (email, name, studentid) VALUES (?, ?, ?)`

	_, err := db.ExecContext(ctx, statement, mail, name, studentID)
	if err != nil {
		log.Printf("insert error: %v\n", err)
		return err
	}

	return nil
}
