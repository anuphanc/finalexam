package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type Customer struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

func getCustomersHandler(c *gin.Context) {
	db, _ := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	stmt, err := db.Prepare("SELECT  id, name, email , status FROM customers")
	if err != nil {
		log.Fatal("stmt error ", err.Error())
	}
	defer db.Close()

	rows, _ := stmt.Query()
	customers := []Customer{}

	for rows.Next() {
		ct := Customer{}
		err := rows.Scan(&ct.ID, &ct.Name, &ct.Email, &ct.Status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error ": err.Error()})
			return
		}
		customers = append(customers, ct)
	}
	fmt.Println("YEHHHHHHH !!! connect to database already !!!!")
	fmt.Println(customers)
	//c.JSON(200, customers)
	c.JSON(http.StatusOK, customers)

	return
}

func postCustomersHandler(c *gin.Context) {
	url := os.Getenv("DATABASE_URL")

	db, err := sql.Open("postgres", url)
	if err != nil {
		log.Fatal("faltal", err.Error())
	}
	defer db.Close()

	ct := Customer{}
	if err := c.ShouldBindJSON(&ct); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	fmt.Println(ct)

	name := ct.Name
	status := ct.Status
	email := ct.Name

	query := `
	INSERT INTO customers (name , email ,  status) VALUES ($1, $2, $3 ) RETURNING id
	`
	var id int

	row := db.QueryRow(query, name, email, status)
	err = row.Scan(&id)

	if err != nil {
		log.Fatal("Can't scan id", err.Error())
	}
	fmt.Println("insert sucess id : ", id)
	ct.ID = id
	c.JSON(http.StatusCreated, ct)

	return
}

func getCustomersByIdHandler(c *gin.Context) {
	idInput := c.Param("id")

	db, _ := sql.Open("postgres", os.Getenv("DATABASE_URL"))

	stmt, _ := db.Prepare("SELECT id, name, email, status FROM customers WHERE id=$1")

	row := stmt.QueryRow(idInput)
	ct := Customer{}

	err := row.Scan(&ct.ID, &ct.Name, &ct.Email, &ct.Status)
	if err != nil {
		log.Fatal("error", err.Error())
	}

	name := ct.Name
	email := ct.Email
	status := ct.Status

	fmt.Println("one row ", idInput, name, email, status)

	fmt.Println("Select by ID !!!!")
	fmt.Println(ct)
	//c.JSON(200, ct)
	c.JSON(http.StatusOK, ct)
	return
}

func putCustomersByIdHandler(c *gin.Context) {
	idInput := c.Param("id")

	ct := Customer{}
	if err := c.ShouldBindJSON(&ct); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	db, _ := sql.Open("postgres", os.Getenv("DATABASE_URL"))

	stmt, err := db.Prepare("UPDATE customers SET name=$2, email=$3, status=$4 WHERE id=$1;")

	if err != nil {
		log.Fatal("Can't scan id", err.Error())
	}
	if _, err3 := stmt.Exec(idInput, ct.Name, ct.Email, ct.Status); err3 != nil {
		log.Fatal("ex error", err3.Error())
	}

	fmt.Println("ct inpu ", ct)

	stmt2, _ := db.Prepare("SELECT id, name, email, status FROM customers WHERE id=$1")

	row := stmt2.QueryRow(idInput)

	err2 := row.Scan(&ct.ID, &ct.Name, &ct.Email, &ct.Status)
	if err2 != nil {
		log.Fatal("error", err2.Error())
	}

	name := ct.Name
	email := ct.Email
	status := ct.Status

	fmt.Println("one row ", idInput, name, email, status)

	fmt.Println("update sucess id : ", idInput)
	c.JSON(http.StatusOK, ct)

	return
}

func deleteCustomersByIdHandler(c *gin.Context) {
	idInput := c.Param("id")

	db, _ := sql.Open("postgres", os.Getenv("DATABASE_URL"))

	ct := Customer{}

	fmt.Println("row for delete ", idInput)

	stmt, err := db.Prepare("DELETE FROM customers  WHERE id=$1 RETURNING Id, name, email, status;")

	if err != nil {
		log.Fatal("Can't scan id==> ", err.Error())
	}

	fmt.Println("one row ", idInput)
	_, err = stmt.Query(idInput)

	fmt.Println("Select by ID !!!!", idInput, stmt, ct)
	//c.JSON(200, gin.H{"status": "customer deleted"})
	c.JSON(http.StatusOK, gin.H{"message": "customer deleted"})
	return
}

func authMiddleware(c *gin.Context) {

	fmt.Println("Hello from middlewre")
	token := c.GetHeader("Authorization")
	fmt.Println("token:", token)
	if token != "token2019" {
		c.JSON(http.StatusUnauthorized, gin.H{"error token ": http.StatusText(http.StatusUnauthorized)})
		c.Abort()
		return
	}
	c.Next()
	fmt.Println("Goodbye from middleware")
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(authMiddleware)

	r.GET("customers", getCustomersHandler)
	r.POST("customers", postCustomersHandler)
	r.GET("customers/:id", getCustomersByIdHandler)
	r.PUT("customers/:id", putCustomersByIdHandler)
	r.DELETE("customers/:id", deleteCustomersByIdHandler)

	return r
}

func createTable() {
	url := os.Getenv("DATABASE_URL")

	db, err := sql.Open("postgres", url)
	if err != nil {
		log.Fatal("faltal", err.Error())
	}
	defer db.Close()

	createTB := `CREATE TABLE IF NOT EXISTS customers(id SERIAL PRIMARY KEY, 
		name TEXT,
		email TEXT,
		status TEXT);`
	_, err = db.Exec(createTB)
	if err != nil {
		log.Fatal("faltal", err.Error())
	}
	return
}

func main() {

	createTable()

	r := setupRouter()
	r.Run(":2019")
}
