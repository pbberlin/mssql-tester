package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

// The user needs to be setup in SQL Server as an SQL Server user.
// See create login and the create user SQL commands as well as the
// SQL Server Management Studio documentation to turn on Hybrid Authentication
// which allows both Windows Authentication and SQL Server Authentication.
// Also need to grant to the user the proper access permissions.
// Also need to enable TCP protocol in SQL Server Configuration Manager.

type tcn struct {
	Server   string `json:"server,omitempty"`
	Port     int    `json:"port,omitempty"`
	UserId   string `json:"user_id,omitempty"`
	Password string `json:"password,omitempty"`
	AppName  string `json:"app_name,omitempty"`

	Encrypt     bool   `json:"encrypt,omitempty"`
	Certificate string `json:"certificate,omitempty"` // file to public key

	Statement     string `json:"statement,omitempty"`
	EveryNMinutes int    `json:"every_n_minutes,omitempty"`
}

func newCn() tcn {
	t := tcn{
		Server:   "localhost",
		Port:     1433,
		UserId:   "gouser",
		Password: "g0us3r",
		AppName:  "pandb-mssql-heartbeat",

		Encrypt:     true,
		Certificate: "example.pem",

		Statement:     "SELECT SYSDATETIME() ",
		EveryNMinutes: 10,
	}
	return t
}

func shortTime() string {
	return time.Now().Format("15:04:05") + "\t"
}

func (t *tcn) CnString() string {
	str := fmt.Sprintf("server=%v;port=%d;user id=%v;password=%v;", t.Server, t.Port, t.UserId, t.Password)
	str += fmt.Sprintf("app name=%s;", t.AppName)
	str += fmt.Sprintf("encrypt=%t;", t.Encrypt)
	if t.Encrypt {
		str += fmt.Sprintf("certificate=%s;", t.Certificate)
		// this is still mysterious - if omitted we get
		// x509: certificate is not valid for any names, but wanted to match 'pandb'"
		str += fmt.Sprintf("TrustServerCertificate=true")
	}
	return str
}

func (t tcn) String() string {
	str := t.CnString()
	str = strings.Replace(str, t.Password, "xxxxxx", -1)
	return str
}

func main() {

	log.SetFlags(log.Lshortfile)
	f, err := os.Create("mssql-tester.log")
	if err != nil {
		log.Fatalf("Could not open log file: %v\n", err)
	}
	log.SetOutput(f)

	def := newCn()
	bts, err := json.MarshalIndent(def, "", "  ")
	if err != nil {
		log.Fatalf("Unmarshal default cn: %v\n", err)
	}
	err = ioutil.WriteFile("example.json", bts, 0644)
	if err != nil {
		log.Fatalf("Could not write example.json: %v\n", err)
	}

	//
	jsonstring, err := ioutil.ReadFile("connect.json")
	if err != nil {
		log.Fatalf("Could not read connect.json; use example.json as template: %v\n", err)
	}
	cn := def // using defaults
	json.Unmarshal([]byte(jsonstring), &cn)

	for {
		func() {
			testCn(&cn)
			time.Sleep(time.Duration(cn.EveryNMinutes) * time.Minute)
		}()
	}

}

func testCn(cn *tcn) {

	log.Printf("About to connect to %v", cn)

	conn, err := sql.Open("mssql", cn.CnString())
	if err != nil {
		log.Printf("%v Open connection failed: %v\n", shortTime(), err)
		return
	}
	defer conn.Close()

	err = conn.Ping()
	if err != nil {
		log.Printf("%v Ping failed: %v\n", shortTime(), err)
		return
	}

	makeQuery(cn, conn)
	time.Sleep(2 * time.Minute)
	makeQuery(cn, conn)
	// var ctx context.Context
	// conn.QueryContext(ctx, `select * from t where ID = @ID and Name = @p2;`, sql.Named("ID", 6), "Bob")
}

func makeQuery(cn *tcn, conn *sql.DB) {
	stmt, err := conn.Prepare(cn.Statement)
	if err != nil {
		log.Printf("%v Prepare failed: %v\n", shortTime(), err)
		return
	}
	defer stmt.Close()

	row := stmt.QueryRow()
	var resultCol string
	err = row.Scan(&resultCol)
	if err != nil {
		log.Printf("%v Scan failed: %v\n", shortTime(), err)
		return
	}
	log.Printf("%v Successful query: %v  =>  %s\n", shortTime(), cn.Statement, resultCol)

}
