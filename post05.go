package post05

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

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

// MSDSCourse is for holding full course data
// Coursedata table + Coursename
type MSDSCourse struct {
	ID          int
	Coursecode  string
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

// The function returns the Course ID of the coursecode
// -1 if the course does not exist
func exists(coursecode string) int {
	coursecode = strings.ToLower(coursecode)

	db, err := openConnection()
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer db.Close()

	courseID := -1
	statement := fmt.Sprintf(`SELECT "id" FROM "courses" where coursecode = '%s'`, coursecode)
	rows, err := db.Query(statement)

	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			fmt.Println("Scan", err)
			return -1
		}
		courseID = id
	}
	defer rows.Close()
	return courseID
}

// AddCourse adds a new course to the database
// Returns new Course ID
// -1 if there was an error
func AddCourse(d MSDSCourse) int {
	d.Coursecode = strings.ToLower(d.Coursecode)

	db, err := openConnection()
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer db.Close()

	courseID := exists(d.Coursecode)
	if courseID != -1 {
		fmt.Println("Course already exists:", d.Coursecode)
		return -1
	}

	insertStatement := `insert into "courses" ("coursecode") values ($1)`
	_, err = db.Exec(insertStatement, d.Coursecode)
	if err != nil {
		fmt.Println(err)
		return -1
	}

	courseID = exists(d.Coursecode)
	if courseID == -1 {
		return courseID
	}

	insertStatement = `insert into "coursedata" ("id", "cid", "cname", "cprereq")
	values ($1, $2, $3, $4)`
	_, err = db.Exec(insertStatement, courseID, d.CID, d.CNAME, d.CPREREQ)
	if err != nil {
		fmt.Println("db.Exec()", err)
		return -1
	}

	return courseID
}

// DeleteCourse deletes an existing course
func DeleteCourse(id int) error {
	db, err := openConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	// Does the ID exist?
	statement := fmt.Sprintf(`SELECT "coursecode" FROM "courses" where id = %d`, id)
	rows, err := db.Query(statement)

	var coursecode string
	for rows.Next() {
		err = rows.Scan(&coursecode)
		if err != nil {
			return err
		}
	}
	defer rows.Close()

	if exists(coursecode) != id {
		return fmt.Errorf("Course with ID %d does not exist", id)
	}

	// Delete from MSDSCourse
	deleteStatement := `delete from "coursedata" where courseid=$1`
	_, err = db.Exec(deleteStatement, id)
	if err != nil {
		return err
	}

	// Delete from Courses
	deleteStatement = `delete from "courses" where id=$1`
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

	rows, err := db.Query(`SELECT "id","coursecode","cid","cname","cprereq"
		FROM "courses","coursedata"
		WHERE courses.id = coursedata.courseid`)
	if err != nil {
		return Data, err
	}

	for rows.Next() {
		var id int
		var coursecode string
		var courseid string
		var coursename string
		var prerequisite string
		err = rows.Scan(&id, &coursecode, &courseid, &coursename, &prerequisite)
		temp := MSDSCourse{ID: id, Coursecode: coursecode, CID: courseid, CNAME: coursename, CPREREQ: prerequisite}
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

	courseID := exists(d.CID)
	if courseID == -1 {
		return errors.New("Course does not exist")
	}
	d.ID = courseID
	updateStatement := `update "coursedata" set "courseid"=$1, "coursename"=$2, "prerequisite"=$3 where "courseid"=$4`
	_, err = db.Exec(updateStatement, d.CID, d.CNAME, d.CPREREQ, d.ID)
	if err != nil {
		return err
	}

	return nil
}
