package repository

import "github.com/jmoiron/sqlx"

type statementsItem struct {
	name      string
	query     string
	statement *sqlx.Stmt
}

type statements struct {
	createTransaction     statementsItem
	getTransactionsByUser statementsItem
	getTransactionsByType statementsItem
	getBalance            statementsItem
	setBalance            statementsItem
}

var statementsList = []statementsItem{
	{
		name: "createTransaction",
		query: `
    INSERT INTO transactions (user_id, type, amount)
    VALUES ($1, $2, $3)
    RETURNING id, user_id, type, amount
    `,
	},
	{
		name: "getTransactionsByUser",
		query: `
    SELECT * FROM transactions WHERE user_id = $1 ORDER BY created_at DESC
    `,
	},
	{
		name: "getTransactionsByType",
		query: `
    SELECT * FROM transactions
    WHERE user_id = $1 AND type = $2
    ORDER BY created_at DESC
    `,
	},
	{
		name: "getBalance",
		query: `
    SELECT balance
    FROM accounts
    WHERE user_id = $1
    `,
	},
	{
		name: "setBalance",
		query: `
    UPDATE accounts
    SET balance = balance + $1
    WHERE user_id = $2 AND balance + $1 >= 0
    `,
	},
}
