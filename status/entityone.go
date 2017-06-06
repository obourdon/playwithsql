package status

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	hmssql "github.com/vincentserpoul/playwithsql/status/history/mssql"
	hmysql "github.com/vincentserpoul/playwithsql/status/history/mysql"
	hpostgres "github.com/vincentserpoul/playwithsql/status/history/postgres"
	hsqlite "github.com/vincentserpoul/playwithsql/status/history/sqlite"
	ilmssql "github.com/vincentserpoul/playwithsql/status/islatest/mssql"
	ilmysql "github.com/vincentserpoul/playwithsql/status/islatest/mysql"
	ilpostgres "github.com/vincentserpoul/playwithsql/status/islatest/postgres"
	ilsqlite "github.com/vincentserpoul/playwithsql/status/islatest/sqlite"
	lsmssql "github.com/vincentserpoul/playwithsql/status/lateststatus/mssql"
	lsmysql "github.com/vincentserpoul/playwithsql/status/lateststatus/mysql"
	lspostgres "github.com/vincentserpoul/playwithsql/status/lateststatus/postgres"
	lssqlite "github.com/vincentserpoul/playwithsql/status/lateststatus/sqlite"
)

// Entityone represents an event
type Entityone struct {
	ID          int64     `db:"entityone_id" json:"entityone_id"`
	TimeCreated time.Time `db:"time_created" json:"time_created"`
	Status      `json:"status"`
}

// Status of the entity
type Status struct {
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
	MigrateUp(ctx context.Context, exec sqlx.ExecerContext) (errExec error)
	MigrateDown(ctx context.Context, exec sqlx.ExecerContext) (errExec error)
	Create(ctx context.Context, exec *sqlx.Tx, actionID, statusID int) (int64, error)
	SaveStatus(
		ctx context.Context,
		exec *sqlx.Tx,
		entityID int64,
		actionID int,
		statusID int,
	) error
	SelectEntityone(
		ctx context.Context,
		q *sqlx.DB,
		entityIDs []int64,
		isStatusIDs []int,
		notStatusIDs []int,
		neverStatusIDs []int,
		hasStatusIDs []int,
		limit int,
	) (*sqlx.Rows, error)
}

// Create will create an entityone
func (e *Entityone) Create(ctx context.Context, db *sqlx.DB, link SQLLink) (err error) {

	tx := db.MustBegin()
	defer func() {
		if err != nil {
			errRoll := tx.Rollback()
			err = fmt.Errorf("%v (rollback errors: %v)", err, errRoll)
		} else {
			err = tx.Commit()
		}
	}()

	e.ID, err = link.Create(ctx, tx, int(ActionCreate), int(StatusCreated))
	if err != nil {
		return fmt.Errorf("entityone Create(): %v", err)
	}

	e.TimeCreated = time.Now()
	// Update status
	e.Status.TimeCreated = time.Now()
	e.ActionID = ActionCreate
	e.StatusID = StatusCreated

	return nil
}

// UpdateStatus will update the status of an Entityone into db
func (e *Entityone) UpdateStatus(
	ctx context.Context,
	db *sqlx.DB,
	link SQLLink,
	actionID ActionID,
	statusID StatusID,
) (err error) {
	tx := db.MustBegin()
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
	q *sqlx.DB,
	link SQLLink,
	statusID StatusID,
) (selectedEntity []*Entityone, err error) {
	rows, err := link.SelectEntityone(ctx, q, []int64{}, []int{int(statusID)}, []int{}, []int{}, []int{}, 3)
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
	q *sqlx.DB,
	link SQLLink,
	entityID int64,
) (selectedEntity *Entityone, err error) {
	rows, err := link.SelectEntityone(ctx, q, []int64{entityID}, []int{}, []int{}, []int{}, []int{}, 0)
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

func extractEntityonesFromRows(rows *sqlx.Rows) (entityOnes []*Entityone, err error) {
	for rows.Next() {

		eo := Entityone{}
		err := rows.StructScan(&eo)
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
func GetSQLIntImpl(dbType string, schemaType string) *SQLIntImpl {
	switch schemaType {
	case "lateststatus":
		switch dbType {
		case "mysql", "percona", "mariadb", "gcpmysql":
			return &SQLIntImpl{&lsmysql.Link{}}
		case "sqlite":
			return &SQLIntImpl{&lssqlite.Link{}}
		case "postgres", "gcppostgres", "cockroachdb":
			return &SQLIntImpl{&lspostgres.Link{}}
		case "mssql":
			return &SQLIntImpl{&lsmssql.Link{}}
		}
	case "islatest":
		switch dbType {
		case "mysql", "percona", "mariadb", "gcpmysql":
			return &SQLIntImpl{&ilmysql.Link{}}
		case "sqlite":
			return &SQLIntImpl{&ilsqlite.Link{}}
		case "postgres", "gcppostgres", "cockroachdb":
			return &SQLIntImpl{&ilpostgres.Link{}}
		case "mssql":
			return &SQLIntImpl{&ilmssql.Link{}}
		}
	case "history":
		switch dbType {
		case "mysql", "percona", "mariadb", "gcpmysql":
			return &SQLIntImpl{&hmysql.Link{}}
		case "sqlite":
			return &SQLIntImpl{&hsqlite.Link{}}
		case "postgres", "gcppostgres", "cockroachdb":
			return &SQLIntImpl{&hpostgres.Link{}}
		case "mssql":
			return &SQLIntImpl{&hmssql.Link{}}
		}
	}
	return nil
}
