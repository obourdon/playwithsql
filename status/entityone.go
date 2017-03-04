package status

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	ilcockroachdb "github.com/vincentserpoul/playwithsql/status/islatest/cockroachdb"
	ilmssql "github.com/vincentserpoul/playwithsql/status/islatest/mssql"
	ilmysql "github.com/vincentserpoul/playwithsql/status/islatest/mysql"
	iloracle "github.com/vincentserpoul/playwithsql/status/islatest/oracle"
	ilpostgres "github.com/vincentserpoul/playwithsql/status/islatest/postgres"
	ilsqlite "github.com/vincentserpoul/playwithsql/status/islatest/sqlite"
)

// Entityone represents an event
type Entityone struct {
	ID          int64     `db:"entityone_id" json:"entityone_id"`
	TimeCreated time.Time `db:"time_created" json:"time_created"`
	Status      `json:"status"`
}

// Status of the entity
type Status struct {
	EntityID    int64     `db:"status_entityone_id" json:"entityone_id"`
	ActionID    ActionID  `db:"action_id" json:"action_id"`
	StatusID    StatusID  `db:"status_id" json:"status_id"`
	TimeCreated time.Time `db:"status_time_created" json:"time_created"`
}

// ActionID represents the action performed on the tradeoffer request
type ActionID int

const (
	// ActionCreate is triggered when the Entityone is created
	ActionCreate ActionID = 1
	// ActionCancel  is triggered when the Entityone is cancelled
	ActionCancel ActionID = 999
	// ActionClose is triggered when the Entityone is closed
	ActionClose ActionID = 500
)

func (s ActionID) String() string {
	return strconv.Itoa(int(s))
}

// StatusID represents the state of the tradeoffer, see constants
type StatusID int

const (
	// StatusCreated is when a Entityone is just created
	StatusCreated StatusID = 1
	// StatusCancelled when a Entityone is cancelled
	StatusCancelled StatusID = 999
	// StatusClosed is not changeable anymore, final status
	StatusClosed StatusID = 1000
)

func (s StatusID) String() string {
	return strconv.Itoa(int(s))
}

// SQLLink is used to define SQL interactions
type SQLLink interface {
	InitDB(ctx context.Context, db *sql.DB, dbName string) (errExec error)
	DestroyDB(ctx context.Context, db *sql.DB, dbName string) (errExec error)
	MigrateUp(ctx context.Context, db *sql.DB) (errExec error)
	MigrateDown(ctx context.Context, db *sql.DB) (errExec error)
	InsertOne(ctx context.Context, db *sql.DB) (int64, error)
	SaveStatus(
		ctx context.Context,
		tx *sql.Tx,
		entityID int64,
		actionID int,
		statusID int,
	) error
	SelectEntity(
		ctx context.Context,
		db *sql.DB,
		entityIDs []int64,
		isStatusIDs []int,
		notStatusIDs []int,
		neverStatusIDs []int,
		hasStatusIDs []int,
		limit int,
	) (*sql.Rows, error)
}

// Create will create an entityone
func (e *Entityone) Create(ctx context.Context, db *sql.DB, link SQLLink) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("Entityone Create: %v", err)
	}
	defer func() {
		if err != nil {
			errRoll := tx.Rollback()
			err = fmt.Errorf("%v (rollback errors: %v)", err, errRoll)
		} else {
			err = tx.Commit()
		}
	}()

	e.ID, err = link.InsertOne(ctx, db)
	if err != nil {
		return fmt.Errorf("Entityone Create: %v", err)
	}

	e.TimeCreated = time.Now()

	err = link.SaveStatus(ctx, tx, e.ID, int(ActionCreate), int(StatusCreated))
	if err != nil {
		return fmt.Errorf("Entityone Create: %v", err)
	}

	e.Status = Status{
		EntityID:    e.ID,
		ActionID:    ActionCreate,
		StatusID:    StatusCreated,
		TimeCreated: time.Now(),
	}

	return err
}

// UpdateStatus will update the status of an Entityone into db
func (e *Entityone) UpdateStatus(
	ctx context.Context,
	db *sql.DB,
	link SQLLink,
	actionID ActionID,
	statusID StatusID,
) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("Entityone UpdateStatus: %v", err)
	}
	defer func() {
		if err != nil {
			errRoll := tx.Rollback()
			err = fmt.Errorf("%v (rollback errors: %v)", err, errRoll)
		} else {
			err = tx.Commit()
		}
	}()

	err = link.SaveStatus(ctx, tx, e.ID, int(actionID), int(statusID))
	if err != nil {
		return fmt.Errorf("entityone UpdateStatus(): %v", err)
	}

	// Update status
	e.ActionID = actionID
	e.StatusID = statusID

	return nil
}

// SelectEntityoneByStatus will retrieve one entityone from a selected status
func SelectEntityoneByStatus(
	ctx context.Context,
	db *sql.DB,
	link SQLLink,
	statusID StatusID,
) (selectedEntity []*Entityone, err error) {
	rows, err := link.SelectEntity(
		ctx, db, []int64{}, []int{int(statusID)}, []int{}, []int{}, []int{}, 3)
	if err != nil {
		return nil, err
	}

	entityOnes, err := extractEntityonesFromRows(rows)
	if err != nil {
		return nil, err
	}
	if len(entityOnes) == 0 {
		return nil, fmt.Errorf("no entity found for status %d", int(statusID))
	}

	return entityOnes, err
}

// SelectEntityoneOneByPK will retrieve one entityone from a selected status
func SelectEntityoneOneByPK(
	ctx context.Context,
	db *sql.DB,
	link SQLLink,
	entityID int64,
) (selectedEntity *Entityone, err error) {
	rows, err := link.SelectEntity(ctx, db, []int64{entityID}, []int{}, []int{}, []int{}, []int{}, 0)
	if err != nil {
		return nil, err
	}

	entityOnes, err := extractEntityonesFromRows(rows)
	if err != nil {
		return nil, err
	}
	if len(entityOnes) == 0 {
		return nil, fmt.Errorf("no entity found for %d", entityID)
	}

	selectedEntity = entityOnes[0]
	return selectedEntity, err
}

func extractEntityonesFromRows(rows *sql.Rows) (entityOnes []*Entityone, err error) {
	for rows.Next() {

		eo := Entityone{}
		err := rows.Scan(&eo)
		if err != nil {
			return entityOnes, fmt.Errorf("entityone Select: %v", err)
		}
		entityOnes = append(entityOnes, &eo)

	}

	return entityOnes, nil
}

// SQLIntImpl allows to contains an interface
type SQLIntImpl struct {
	SQLLink
}

// GetSQLIntImpl returns the type of link according to the dbtype
func GetSQLIntImpl(dbType string) *SQLIntImpl {
	switch dbType {
	case "mysql", "percona", "mariadb":
		return &SQLIntImpl{&ilmysql.Link{}}
	case "sqlite":
		return &SQLIntImpl{&ilsqlite.Link{}}
	case "postgres":
		return &SQLIntImpl{&ilpostgres.Link{}}
	case "cockroachdb":
		return &SQLIntImpl{&ilcockroachdb.Link{}}
	case "mssql":
		return &SQLIntImpl{&ilmssql.Link{}}
	case "oracle":
		return &SQLIntImpl{&iloracle.Link{}}
	}

	return nil
}
