package status

import (
	"context"
	"flag"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vincentserpoul/playwithsql/dbhandler"
)

func TestCreate(t *testing.T) {
	var e Entityone
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	err := e.Create(ctx, testDBConn, testSQLLink)
	if err != nil {
		t.Errorf("create entityone: %v", err)
	}
}

func BenchmarkCreate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var e Entityone
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		_ = e.Create(ctx, testDBConn, testSQLLink)

		// limit the number of tests
		if len(testEntityoneIDs) < 500 {
			testEntityoneIDs = append(testEntityoneIDs, e.ID)
		}
	}
}

func TestUpdateStatus(t *testing.T) {
	var e Entityone

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := e.Create(ctx, testDBConn, testSQLLink)
	if err != nil {
		t.Errorf("UpdateStatus entityone: %v", err)
	}
	err = e.UpdateStatus(ctx, testDBConn, testSQLLink, ActionCancel, StatusCancelled)
	if err != nil {
		t.Errorf("UpdateStatus entityone: %v", err)
	}

	if e.ActionID != ActionCancel && e.StatusID != StatusCancelled {
		t.Errorf("UpdateStatus entityone: status and action not updated")
	}

	err = e.UpdateStatus(ctx, testDBConn, testSQLLink, ActionCancel, StatusCreated)
	if err != nil {
		t.Errorf("UpdateStatus entityone: %v", err)
	}

	if e.ActionID != ActionCancel && e.StatusID != StatusCancelled {
		t.Errorf("UpdateStatus entityone: status and action not updated")
	}
}

func BenchmarkUpdateStatus(b *testing.B) {
	var e Entityone
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		e.ID = testEntityoneIDs[b.N%len(testEntityoneIDs)]
		_ = e.UpdateStatus(ctx, testDBConn, testSQLLink, ActionCancel, StatusCancelled)
	}
}

func TestSelectEntityoneByStatus(t *testing.T) {
	var e Entityone
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	err := e.Create(ctx, testDBConn, testSQLLink)
	if err != nil {
		t.Errorf("Select entityone by status: %v", err)
	}

	_, err = SelectEntityoneByStatus(ctx, testDBConn, testSQLLink, StatusCreated)
	if err != nil {
		t.Errorf("Select entityone by status: %v", err)
	}
}

func BenchmarkSelectEntityoneByStatus(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		_, err := SelectEntityoneByStatus(ctx, testDBConn, testSQLLink, StatusCreated)
		if err != nil {
			b.Errorf("Select entityone by status: %v", err)
			return
		}
	}
}

func TestSelectEntityoneOneByPK(t *testing.T) {
	var e Entityone
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	err := e.Create(ctx, testDBConn, testSQLLink)
	if err != nil {
		t.Errorf("Select entityone by pk: %v", err)
		return
	}

	entityOne, err := SelectEntityoneOneByPK(ctx, testDBConn, testSQLLink, e.ID)
	if err != nil {
		t.Errorf("Select entityone by pk: %v", err)
		return
	}

	if entityOne.ID != e.ID {
		t.Errorf("Select entityone by pk retrieved entity %d instead of %d", entityOne.ID, e.ID)
		return
	}
	var emptyTime time.Time
	if entityOne.TimeCreated == emptyTime {
		t.Errorf("Select entityone by pk retrieved but entity time created not correctly retrieved: %v", entityOne)
		return
	}
	if entityOne.Status.TimeCreated == emptyTime {
		t.Errorf("Select entityone by pk retrieved but entity status time created not correctly retrieved: %v", entityOne)
		return
	}
	if entityOne.Status.ActionID == 0 {
		t.Errorf("Select entityone by pk retrieved but entity status actionid created not correctly retrieved: %v", entityOne)
		return
	}
	if entityOne.Status.StatusID == 0 {
		t.Errorf("Select entityone by pk retrieved but entity status statusid created not correctly retrieved: %v", entityOne)
		return
	}
}

func BenchmarkSelectEntityoneOneByPK(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		_, err := SelectEntityoneOneByPK(ctx, testDBConn, testSQLLink, testEntityoneIDs[b.N%len(testEntityoneIDs)])
		if err != nil {
			b.Errorf("Select entityone by status: %v", err)
			return
		}
	}
}

var testDBConn *sqlx.DB
var testSQLLink *SQLIntImpl
var testEntityoneIDs []int64

func TestMain(m *testing.M) {

	var err error

	dbName := "entityone_test"
	dbType := flag.String("db", "mysql", "type of db to bench: mysql, cockroachdb, postgres...")
	schemaType := flag.String("sch", "islatest", "type of schema to use, is latest, latest status...")
	host := flag.String("host", "127.0.0.1", "host IP")
	flag.Parse()

	tempDBConn, err := dbhandler.Get(*dbType, *host, dbName)
	defer func() {
		errClose := testDBConn.Close()
		if errClose != nil {
			log.Fatalf("%v", errClose)
		}
	}()
	if err != nil {
		log.Fatalf("%v", err)
	}
	testDBConn = tempDBConn

	// create context to cancel working if 10 seconds passed
	ctx := context.Background()

	tempSQLLink := GetSQLIntImpl(*dbType, *schemaType)
	testSQLLink = tempSQLLink

	err = testSQLLink.MigrateDown(ctx, testDBConn)
	if err != nil {
		log.Fatalf("%v", err)
	}
	err = testSQLLink.MigrateUp(ctx, testDBConn)
	if err != nil {
		log.Fatalf("%v", err)
	}

	retCode := m.Run()

	err = testSQLLink.MigrateDown(ctx, testDBConn)
	if err != nil {
		log.Fatalf("%v", err)
	}

	os.Exit(retCode)
}
