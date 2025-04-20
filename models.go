package main

import (
	"database/sql"
	"errors"
)

// ErrNotFound is returned when a row isn't found.
var ErrNotFound = errors.New("not found")

// User
func CreateUser(db *sql.DB, username, pwHash string) error {
	_, err := db.Exec(`INSERT INTO users(username,password_hash) VALUES(?,?)`, username, pwHash)
	return err
}
func GetUserByUsername(db *sql.DB, username string) (id int, pwHash string, err error) {
	err = db.QueryRow(`SELECT id,password_hash FROM users WHERE username = ?`, username).Scan(&id, &pwHash)
	if err == sql.ErrNoRows {
		return 0, "", ErrNotFound
	}
	return
}

// Project
func CreateProject(db *sql.DB, userID int, name, token string) (int, error) {
	res, err := db.Exec(
		`INSERT INTO projects(user_id,name,token) VALUES(?,?,?)`,
		userID, name, token,
	)
	if err != nil {
		return 0, err
	}
	pid, _ := res.LastInsertId()
	return int(pid), nil
}

func ListProjects(db *sql.DB, userID int) ([]struct {
	ID   int
	Name string
}, error) {
	rows, err := db.Query(`SELECT id, name FROM projects WHERE user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	var projects []struct {
		ID   int
		Name string
	}
	for rows.Next() {
		var project struct {
			ID   int
			Name string
		}
		if err := rows.Scan(&project.ID, &project.Name); err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	return projects, nil
}

func GetProjectByToken(db *sql.DB, token string) (id, userID int, err error) {
	err = db.QueryRow(`SELECT id,user_id FROM projects WHERE token = ?`, token).Scan(&id, &userID)
	if err == sql.ErrNoRows {
		return 0, 0, ErrNotFound
	}
	return
}
func GetProject(db *sql.DB, projID, userID int) (map[string]any, error) {
	var name string
	err := db.QueryRow(
		`SELECT name FROM projects WHERE id = ? AND user_id = ?`,
		projID, userID,
	).Scan(&name)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	tables, err := ListTables(db, projID, userID)
	if err != nil {
		return nil, err
	}

	type TableWithVariables struct {
		ID        int
		Name      string
		Variables []struct {
			Name  string
			Type  string
			Value string
		}
	}
	var tablesWithVariables []TableWithVariables
	// For each table, get the variables
	for _, table := range tables {
		variables, _ := ListVariables(db, table.ID, userID)
		tablesWithVariables = append(tablesWithVariables, TableWithVariables{ID: table.ID, Name: table.Name, Variables: variables})
	}

	return map[string]any{"name": name, "tables": tablesWithVariables}, err
}

func RenameProject(db *sql.DB, projID int, name string, userID int) error {
	_, err := db.Exec(`UPDATE projects SET name = ? WHERE id = ? AND user_id = ?`, name, projID, userID)
	return err
}

func DeleteProject(db *sql.DB, projID int, userID int) error {
	_, err := db.Exec(`DELETE FROM projects WHERE id = ? AND user_id = ?`, projID, userID)
	return err
}

// Table
func CreateTable(db *sql.DB, projectID int, name string, userID int) (int, error) {
	res, err := db.Exec(
		`INSERT INTO tables(project_id,user_id,name)
         SELECT ?,?,?
         FROM projects WHERE id = ? AND user_id = ?`,
		projectID, userID, name,
		projectID, userID,
	)
	if err != nil {
		return 0, err
	}
	tid, _ := res.LastInsertId()
	return int(tid), nil
}
func GetTableID(db *sql.DB, projectID int, name string, userID int) (int, error) {
	var id int
	err := db.QueryRow(
		`SELECT id FROM tables WHERE project_id = ? AND name = ? AND user_id = ?`,
		projectID, name, userID,
	).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, ErrNotFound
	}
	return id, err
}

func DeleteTable(db *sql.DB, tableID int, userID int) error {
	_, err := db.Exec(`DELETE FROM tables WHERE id = ? AND user_id = ?`, tableID, userID)
	return err
}

func RenameTable(db *sql.DB, tableID int, name string, userID int) error {
	_, err := db.Exec(`UPDATE tables SET name = ? WHERE id = ? AND user_id = ?`, name, tableID, userID)
	return err
}

func ListTables(db *sql.DB, projectID int, userID int) ([]struct {
	ID   int
	Name string
}, error) {
	rows, err := db.Query(`SELECT id, name FROM tables WHERE project_id = ? AND user_id = ?`, projectID, userID)
	if err != nil {
		return nil, err
	}

	var tables []struct {
		ID   int
		Name string
	}
	for rows.Next() {
		var table struct {
			ID   int
			Name string
		}
		if err := rows.Scan(&table.ID, &table.Name); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
}

// Variable
func CreateVariable(db *sql.DB, tableID int, name, value, typ string, userID int) error {
	_, err := db.Exec(
		`INSERT INTO variables(table_id,user_id,name,value,type) VALUES(?,?,?,?,?)`,
		tableID, userID, name, value, typ,
	)
	return err
}

func ListVariables(db *sql.DB, tableID int, userID int) ([]struct {
	Name  string
	Type  string
	Value string
}, error) {
	rows, err := db.Query(`SELECT name, type, value FROM variables WHERE table_id = ? AND user_id = ?`, tableID, userID)
	if err != nil {
		return nil, err
	}
	var variables []struct {
		Name  string
		Type  string
		Value string
	}
	for rows.Next() {
		var variable struct {
			Name  string
			Type  string
			Value string
		}
		if err := rows.Scan(&variable.Name, &variable.Type, &variable.Value); err != nil {
			return nil, err
		}
		variables = append(variables, variable)
	}
	return variables, nil
}

func DeleteVariable(db *sql.DB, tableID int, name string, userID int) error {
	_, err := db.Exec(`DELETE FROM variables WHERE table_id = ? AND name = ? AND user_id = ?`, tableID, name, userID)
	return err
}

func UpdateVariable(db *sql.DB, tableID int, name, newName, typ string, userID int) error {
	_, err := db.Exec(
		`UPDATE variables SET name = ?, type = ? WHERE table_id = ? AND name = ? AND user_id = ?`,
		newName, typ, tableID, name, userID,
	)
	return err
}

func GetVariable(db *sql.DB, tableID int, name string) (value, typ string, err error) {
	err = db.QueryRow(
		`SELECT value, type FROM variables WHERE table_id = ? AND name = ?`,
		tableID, name,
	).Scan(&value, &typ)
	if err == sql.ErrNoRows {
		return "", "", ErrNotFound
	}
	return
}
func SetVariable(db *sql.DB, tableID int, name, value, typ string) error {
	_, err := db.Exec(
		`INSERT INTO variables(table_id,name,value,type)
         VALUES(?,?,?,?)
         ON CONFLICT(table_id,name) DO UPDATE SET value=excluded.value, type=excluded.type`,
		tableID, name, value, typ,
	)
	return err
}
