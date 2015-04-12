package user

import (
	_ "fmt"
	. "github.com/smartystreets/goconvey/convey"
	_ "github.com/gopistolet/gopistolet/helpers"
	"testing"
)

func TestUserDB(t *testing.T) {
	Convey("Testing UserDB.Add()", t, func() {

		db := UserDB{}

		err := db.Add(User{Name: "Mathias"})
		So(err, ShouldEqual, nil)

		user, err := db.Get("Mathias")
		So(err, ShouldEqual, nil)
		So(user.Name, ShouldEqual, "Mathias")

		err = db.Add(User{Name: "Mathias"})
		So(err, ShouldNotEqual, nil)

	})

	Convey("Testing LoadDB() UserDB", t, func() {

		db, err := LoadDB("./users.json")

		if err != nil {
			panic(err.Error())
		}

		user, err := db.Get("Mathias")
		So(err, ShouldEqual, nil)
		So(user.Name, ShouldEqual, "Mathias")

		//So(true, ShouldEqual, true)
		if err != nil {
			panic(err.Error())
		}

	})

}
