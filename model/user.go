package model

// Users db table
type Users struct {
	ID       int32
	Name     string
	Password string
}

// GetUserInfoByUserName Get user info by user name
func GetUserInfoByUserName(userName string) (id int32, name string, password string, err error) {
	row := db.QueryRow("SELECT `id`, `name`, `password` FROM `users` WHERE `name` = ?", userName)
	err = row.Scan(&id, &name, &password)

	return id, name, password, err
}

// RegsiterUser RegsiterUser
func RegsiterUser(userName, password string) (int32, error) {
	stmt, err := db.Prepare("INSERT INTO `users` (`name`,`password`) VALUES (?,?)")

	if err != nil {
		return 0, err
	}

	val, err := stmt.Exec(userName, password)

	if err != nil {
		return 0, err
	}

	id, _ := val.LastInsertId()

	return int32(id), nil
}
