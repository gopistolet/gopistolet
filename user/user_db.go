package user

import "github.com/gopistolet/gopistolet/helpers"
import "errors"
import "io/ioutil"
import "encoding/json"

type UserDB struct {
	Users map[string]User
}


// UserExists checks if a user exists in the DB
func (db *UserDB) UserExists(name string) bool {
	_, found := db.Users[name]
	return found
}

// Get user from the database
func (db *UserDB) Get(name string) (*User, error) {
	helpers.Assert(true, "Test")
	if db.UserExists(name) {
		user := db.Users[name]
		return &user, nil
	} else {
		return nil, errors.New("User not found")
	}
}

// Add user to the database
func (db *UserDB) Add(user User) error {
	if db.Users == nil {
		db.Users = make(map[string]User)
	}
	if db.UserExists(user.Name) {
		return errors.New("User already exists")
	}
	db.Users[user.Name] = user
	return nil
}

// Save database to file
func (db *UserDB) SaveDB(file string) error {
	output, err := json.MarshalIndent(db, "", "\t")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(file, output, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Load database from file
func LoadDB(file string) (*UserDB, error) {
	input, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	db := UserDB{}
	err = json.Unmarshal(input, &db)

	if err != nil {
		return nil, err
	}

	return &db, nil
}
