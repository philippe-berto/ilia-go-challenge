package repository

import "github.com/jmoiron/sqlx"

type statementsItem struct {
	name      string
	query     string
	statement *sqlx.Stmt
}

var statementsList = []statementsItem{
	{
		name: "createUser",
		query: `
    INSERT INTO users (first_name, last_name, email, password_hash)
    VALUES ($1, $2, $3, $4)
    RETURNING id, first_name, last_name, email
    `,
	},
	{
		name: "getUserByID",
		query: `
    SELECT id, first_name, last_name, email
    FROM users
    WHERE id = $1
    `,
	},
	{
		name: "getUserByEmail",
		query: `
    SELECT id, first_name, last_name, email, password_hash
    FROM users
    WHERE email = $1
    `,
	},
	{
		name: "getUsers",
		query: `
    SELECT id, first_name, last_name, email
    FROM users
    ORDER BY created_at DESC
    `,
	},
	{
		name: "updateUser",
		query: `
    UPDATE users
    SET first_name = $1, last_name = $2, email = $3
    WHERE id = $4
    RETURNING id, first_name, last_name, email
    `,
	},
	{
		name: "deleteUser",
		query: `
    DELETE FROM users WHERE id = $1
    `,
	},
}
