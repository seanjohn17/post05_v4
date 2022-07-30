package post05

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"encoding/json"

	_ "github.com/lib/pq"
)

// Connection details
var (
	Hostname = ""
	Port     = 2345
	Username = ""
	Password = ""
	Database = ""
)

// Userdata is for holding full user data
// Userdata table + Username
type MSDSCourse struct {
	CID         string `json:"course_ID"`
	CNAME       string `json:"course_name"`
	CPREREQ     string `json:"prerequisite"`
}

func openConnection() (*sql.DB, error) {
	// connection string
	conn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		Hostname, Port, Username, Password, Database)

	// open database
	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// The function returns the Course ID of the course
// -1 if the user does not exist
func exists(CNAME string) int {
	CNAME = strings.ToLower(CNAME)

	db, err := openConnection()
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer db.Close()

	CID := -1
	statement := fmt.Sprintf(`SELECT "CID" FROM "Courses" where CNAME = '%s'`, CNAME)
	rows, err := db.Query(statement)

	for rows.Next() {
		var CID int
		err = rows.Scan(&CID)
		if err != nil {
			fmt.Println("Scan", err)
			return -1
		}
		CID = CID
	}
	defer rows.Close()
	return CID
}

// AddUser adds a new user to the database
// Returns new User ID
// -1 if there was an error
func AddCourse(d MSDSCourse) int {
	d.CNAME = strings.ToLower(d.CNAME)

	db, err := openConnection()
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer db.Close()

	CID := exists(d.CNAME)
	if CID != -1 {
		fmt.Println("Course already exists:", CID)
		return -1
	}

	insertStatement := `insert into "Courses" ("CID") values ($1)`
	_, err = db.Exec(insertStatement, d.CNAME)
	if err != nil {
		fmt.Println(err)
		return -1
	}

	CID = exists(d.CNAME)
	if CID == -1 {
		return CID
	}

	insertStatement = `insert into "Coursedata" ("CID", "CNAME", "CPREREQ")
	values ($1, $2, $3)`
	_, err = db.Exec(insertStatement, CID, d.CNAME, d.CPREREQ)
	if err != nil {
		fmt.Println("db.Exec()", err)
		return -1
	}

	return CID
}

// DeleteUser deletes an existing user
func DeleteCourse(id int) error {
	db, err := openConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	// Does the ID exist?
	statement := fmt.Sprintf(`SELECT "CNAME" FROM "Courses" where id = %d`, id)
	rows, err := db.Query(statement)

	var CNAME string
	for rows.Next() {
		err = rows.Scan(&CNAME)
		if err != nil {
			return err
		}
	}
	defer rows.Close()

	if exists(CNAME) != id {
		return fmt.Errorf("Course with ID %d does not exist", id)
	}

	// Delete from MSDSCourse
	deleteStatement := `delete from "Coursedata" where CID=$1`
	_, err = db.Exec(deleteStatement, id)
	if err != nil {
		return err
	}

	// Delete from Courses
	deleteStatement = `delete from "Courses" where id=$1`
	_, err = db.Exec(deleteStatement, id)
	if err != nil {
		return err
	}

	return nil
}

// ListUsers lists all users in the database
func ListCourses() ([]MSDSCourse, error) {
	Data := []MSDSCourse{}
	db, err := openConnection()
	if err != nil {
		return Data, err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT "CID","CNAME","CPREREQ"
		FROM "Courses","Coursedata"
		WHERE Courses.CID = Coursedata.CID`)
	if err != nil {
		return Data, err
	}

	for rows.Next() {
		var CID string
		var CNAME string
		var CPREREQ string
		err = rows.Scan(&CID, &CNAME, &CPREREQ)
		temp := MSDSCourse{CID: CID, CNAME: CNAME, CPREREQ: CPREREQ}
		Data = append(Data, temp)
		if err != nil {
			return Data, err
		}
	}
	defer rows.Close()
	return Data, nil
}

// UpdateCourse is for updating an existing course
func UpdateCourse(d MSDSCourse) error {
	db, err := openConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	CID := exists(d.CNAME)
	if CID == -1 {
		return errors.New("Course does not exist")
	}
	d.CID = CID
	updateStatement := `update "msdscourse" set "coursename"=$1, "prerequisite"=$2 where "courseid"=$4`
	_, err = db.Exec(updateStatement, d.CNAME, d.CPREREQ, d.CID)
	if err != nil {
		return err
	}

	return nil
}
