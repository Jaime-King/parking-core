package data

import (
	"database/sql"
	"os"
	"time"

	i "github.com/Jaime-King/parking-core/logger"
	"github.com/Jaime-King/parking-core/models"
	_ "github.com/go-sql-driver/mysql"
)

// Table Name Constants
const TABLE_SCHEDULE string = "schedule"
const TABLE_USER string = "user"
const DATE_FORMAT = "2006-01-02 15:04:05"

func getMysqlClient() *sql.DB {
	var user = os.Getenv("DB_USER")
	var pass = os.Getenv("MYSQL_ROOT_PASSWORD")
	var host = os.Getenv("DB_HOST")
	var port = os.Getenv("DB_PORT")

	db, err := sql.Open("mysql", "" + user + ":" + pass + "@tcp(" + host + ":" + port + ")/")
    // if there is an error opening the connection, handle it
    if err != nil {
        panic(err.Error())
    }

	exists, err := checkDatabaseExists(db);
	if (err != nil) {
		panic(err.Error())
	}

	// Initialise the schema if needed
	if (!exists) {
		initialise(db)
	}

	_, err = db.Exec("USE peter_parker;")
	if err != nil {
		panic(err)
	}

	return db
}

func closeDbConnection(db *sql.DB) {
	i.Log.Info("Closing Mysql Connection")
	db.Close()
}

func initialise(db *sql.DB) {
	i.Log.Info("Creating database")
	_, err := db.Exec("CREATE DATABASE IF NOT EXISTS peter_parker;")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("USE peter_parker;")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user (
			username VARCHAR(50) NOT NULL,
			name VARCHAR(50) NOT NULL,
			atEmail VARCHAR(50) NOT NULL,
			atPassword VARCHAR(50) NOT NULL,
			cycleLength INT unsigned DEFAULT 60,
			plate VARCHAR(50) NOT NULL,
			PRIMARY KEY (username)
		);
	`)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS schedules (
			username VARCHAR(50) NOT NULL,
			startTime datetime NOT NULL,
			endTime datetime NOT NULL,
			area INT unsigned NOT NULL,
			nextParkTime datetime NOT NULL,
			progress VARCHAR(50) NOT NULL DEFAULT 'pending',
			message VARCHAR(500) NOT NULL DEFAULT '',
			sessions INT unsigned DEFAULT 0,
			PRIMARY KEY (username, startTime),
			KEY schedule_user_time (username) USING BTREE,
			FOREIGN KEY (username) REFERENCES user(username)
		);
	`)
	if err != nil {
		panic(err)
	}
}

func GetSchedules() []models.Schedule {
	db := getMysqlClient()
	defer closeDbConnection(db)

	var schedules []models.Schedule

	rows, err := db.Query(`
		SELECT username, startTime, endTime, area, nextParkTime, progress, message, sessions
		FROM schedules
		WHERE progress <> 'complete' AND startTime < NOW();
	`)

	if err != nil {
		i.Log.Error("Error while retrieving schedules", err)
		return schedules
	}

	defer rows.Close()

	for rows.Next() {
		var s models.Schedule
		var startTime, endTime, nextParkTime string
		err := rows.Scan(&s.Username, &startTime, &endTime, &s.Area, &nextParkTime, &s.Progress, &s.Message, &s.Sessions)

		if err != nil {
			i.Log.Error("Error while mapping schedule from database", err)
			continue
		}

		s.StartTime = parseDateTime(startTime)
		s.EndTime = parseDateTime(endTime)
		s.NextParkTime = parseDateTime(nextParkTime)

		schedules = append(schedules, s)
	}

	return schedules
}

func parseDateTime(dtStr string) time.Time {
	loc, err := time.LoadLocation("Local")
	if err != nil {
		i.Log.Error("Error while parsing datetime", err)
		return time.Time{}
	}
	t, err := time.ParseInLocation(DATE_FORMAT, dtStr, loc)
	if err != nil {
		i.Log.Error("Error while parsing datetime", err)
		return time.Time{}
	}
	return t
}


func checkDatabaseExists(db *sql.DB) (bool, error) {
	// Prepare a query to check if the database exists
	query := "SELECT COUNT(*) FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = 'peter_parker';"

	// Execute the query and check the result
	var count int
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		// An error occurred while checking the database existence
		return false, err
	}

	// Check if the count is greater than zero
	return count > 0, nil
}

func SaveSchedule(s models.Schedule) error {
	db := getMysqlClient()
	defer closeDbConnection(db)

	nextParkTimeStr := s.NextParkTime.Format(DATE_FORMAT)
	startTimeStr := s.StartTime.Format(DATE_FORMAT)
	endTimeStr := s.EndTime.Format(DATE_FORMAT)
	
	// Prepare the update statement
	stmt, err := db.Prepare("UPDATE schedules SET progress=?, nextParkTime=?, endTime=?, message=?, sessions=? WHERE username=? AND startTime=?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	
	// Execute the update statement
	_, err = stmt.Exec(s.Progress, nextParkTimeStr, endTimeStr, s.Message, s.Sessions, s.Username, startTimeStr)
	if err != nil {
		return err
	}
	return nil
}

func GetUser(username string) (models.User, error) {
	db := getMysqlClient()
	defer closeDbConnection(db)

	var user models.User

	// Prepare the select statement
	stmt, err := db.Prepare("SELECT username, name, atEmail, atPassword, plate, cycleLength FROM user WHERE username = ?")
	if err != nil {
		return user, err
	}
	defer stmt.Close()

	var cycleLenghtInt int

	// Execute the select statement
	row := stmt.QueryRow(username)

	// Scan the row into the User struct
	err = row.Scan(&user.Username, &user.Name, &user.AtEmail, &user.AtPassword, &user.Plate, &cycleLenghtInt)
	if err != nil {
		return user, err
	}

	// Convert cycleLength to time.Duration
	user.CycleLength = time.Duration(cycleLenghtInt) * time.Minute

	return user, nil
}