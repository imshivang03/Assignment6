package main

//cd "C:\Users\hp word\Desktop\go-workspace\src\Mux_postgre"
// go run main.go
import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"

	"github.com/gorilla/mux"
)

type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Location string `json:"Location"`
	Age      int64  `json:"age"`
}

func main() {
	r := Router()
	fmt.Println("Starting server on the port 8080...")

	log.Fatal(http.ListenAndServe(":8080", r))
}

func Router() *mux.Router {

	router := mux.NewRouter()

	router.HandleFunc("/user/{id}", GetUser).Methods("GET", "OPTIONS")
	router.HandleFunc("/user", GetAllUser).Methods("GET", "OPTIONS")
	router.HandleFunc("/newuser", CreateUser).Methods("POST", "OPTIONS")
	router.HandleFunc("/user/{id}", UpdateUser).Methods("PUT", "OPTIONS")
	router.HandleFunc("/deleteuser/{id}", DeleteUser).Methods("DELETE", "OPTIONS")

	return router
}

// response format
type response struct {
	ID      int64  `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}

// create connection with postgres db
func createConnection() *sql.DB {
	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", "postgres", "Verma_123", "information")

	//var err error
	db, err := sql.Open("postgres", connectionString)
	//db, err := sql.Open("postgres", "postgres://postgres:7046365527@localhost/postgres?sslmode=disable")

	if err != nil {
		panic(err)
	}
	// check the connection
	err = db.Ping()

	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected!")
	return db
}

// CreateUser create a user in the postgres db
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Fatalf("Unable to decode the request body.  %v", err)
	}
	insertID := insertUser(user)

	res := response{
		ID:      insertID,
		Message: "User created successfully",
	}
	json.NewEncoder(w).Encode(res)
}

// GetUser will return a single user by its id
func GetUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])

	if err != nil {
		log.Fatalf("Unable to convert the string into int.  %v", err)
	}
	user, err := getUser(int64(id))

	if err != nil {
		log.Fatalf("Unable to get user. %v", err)
	}

	json.NewEncoder(w).Encode(user)
}

// GetAllUser will return all the users
func GetAllUser(w http.ResponseWriter, r *http.Request) {
	users, err := getAllUsers()

	if err != nil {
		log.Fatalf("Unable to get all user. %v", err)
	}

	json.NewEncoder(w).Encode(users)
}

// UpdateUser update user's detail in the postgres db
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])

	if err != nil {
		log.Fatalf("Unable to convert the string into int.  %v", err)
	}
	var user User
	err = json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		log.Fatalf("Unable to decode the request body.  %v", err)
	}
	updatedRows := updateUser(int64(id), user)
	msg := fmt.Sprintf("User updated successfully. Total rows/record affected %v", updatedRows)
	res := response{
		ID:      int64(id),
		Message: msg,
	}
	json.NewEncoder(w).Encode(res)
}

// DeleteUser delete user's detail in the postgres db
func DeleteUser(w http.ResponseWriter, r *http.Request) {

	// get the userid from the request params, key is "id"
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])

	if err != nil {
		log.Fatalf("Unable to convert the string into int.  %v", err)
	}
	deletedRows := deleteUser(int64(id))
	msg := fmt.Sprintf("User updated successfully. Total rows/record affected %v", deletedRows)

	// format the reponse message
	res := response{
		ID:      int64(id),
		Message: msg,
	}

	// send the response
	json.NewEncoder(w).Encode(res)
}

//------------------------- handler functions ----------------
// insert one user in the DB
func insertUser(user User) int64 {

	// create the postgres db connection
	db := createConnection()

	// close the db connection
	defer db.Close()

	// create the insert sql query
	// returning userid will return the id of the inserted user
	sqlStatement := `INSERT INTO customers (id,name, Location, age) VALUES ($1, $2, $3,$4) RETURNING id`

	// the inserted id will store in this id
	var id int64

	// execute the sql statement
	// Scan function will save the insert id in the id
	err := db.QueryRow(sqlStatement, user.ID, user.Name, user.Location, user.Age).Scan(&id)

	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}

	fmt.Printf("Inserted a single record %v", id)

	// return the inserted id
	return id
}

// get one user from the DB by its userid
func getUser(id int64) (User, error) {
	db := createConnection()
	defer db.Close()
	var user User
	sqlStatement := `SELECT * FROM customers WHERE id=$1`
	row := db.QueryRow(sqlStatement, id)
	err := row.Scan(&user.ID, &user.Name, &user.Location, &user.Age)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
		return user, nil
	case nil:
		return user, nil
	default:
		log.Fatalf("Unable to scan the row. %v", err)
	}

	// return empty user on error
	return user, err
}

// get one user from the DB by its userid
func getAllUsers() ([]User, error) {
	db := createConnection()
	defer db.Close()
	var users []User
	sqlStatement := `SELECT * FROM customers`
	rows, err := db.Query(sqlStatement)
	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var user User
		err = rows.Scan(&user.ID, &user.Name, &user.Location, &user.Age)

		if err != nil {
			log.Fatalf("Unable to scan the row. %v", err)
		}

		// append the user in the users slice
		users = append(users, user)

	}
	return users, err
}

// update user in the DB
func updateUser(id int64, user User) int64 {
	db := createConnection()
	defer db.Close()
	sqlStatement := `UPDATE customers SET name=$2, Location=$3, age=$4 WHERE id=$1`

	res, err := db.Exec(sqlStatement, id, user.Name, user.Location, user.Age)

	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}
	rowsAffected, err := res.RowsAffected()

	if err != nil {
		log.Fatalf("Error while checking the affected rows. %v", err)
	}

	fmt.Printf("Total rows/record affected %v", rowsAffected)

	return rowsAffected
}

// delete user in the DB
func deleteUser(id int64) int64 {
	db := createConnection()
	defer db.Close()
	sqlStatement := `DELETE FROM customers WHERE id=$1`

	res, err := db.Exec(sqlStatement, id)

	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}

	rowsAffected, err := res.RowsAffected()

	if err != nil {
		log.Fatalf("Error while checking the affected rows. %v", err)
	}

	fmt.Printf("Total rows/record affected %v", rowsAffected)

	return rowsAffected
}
