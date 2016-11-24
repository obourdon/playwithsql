package islatest

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
)

// Entityone represents an event
type Entityone struct {
	ID          int64     `db:"entityone_id"`
	TimeCreated time.Time `db:"time_created"`
	Status
}

// Status of the entity
type Status struct {
	EntityID    int64     `db:"status_entityone_id"`
	ActionID    ActionID  `db:"action_id"`
	StatusID    StatusID  `db:"status_id"`
	TimeCreated time.Time `db:"status_time_created"`
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
	InitDB(exec sqlx.Execer, dbName string) (errExec error)
	DestroyDB(exec sqlx.Execer, dbName string) (errExec error)
	MigrateUp(exec sqlx.Execer) (errExec error)
	MigrateDown(exec sqlx.Execer) (errExec error)
	InsertOne(sqlx.Ext) (int64, error)
	SaveStatus(exec sqlx.Execer, entityID int64, actionID int, statusID int) error
	GetFilterSelectEntityOneQuery(idFilter []int64, statusFilter []int) ([]interface{}, string)
}

// Create will create an entityone
func (e *Entityone) Create(db *sqlx.DB, link SQLLink) (err error) {
	tx := db.MustBegin()
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	e.ID, err = link.InsertOne(db)
	if err != nil {
		return fmt.Errorf("Entityone createEntityone: %v", err)
	}

	err = link.SaveStatus(tx, e.ID, int(ActionCreate), int(StatusCreated))
	if err != nil {
		return fmt.Errorf("Entityone createEntityone: %v", err)
	}

	e.Status = Status{
		EntityID:    e.ID,
		ActionID:    ActionCreate,
		StatusID:    StatusCreated,
		TimeCreated: time.Now(),
	}

	return nil
}

// UpdateStatus will update the status of an Entityone into db
func (e *Entityone) UpdateStatus(exec sqlx.Ext, link SQLLink, actionID ActionID, statusID StatusID) error {
	err := link.SaveStatus(exec, e.ID, int(actionID), int(statusID))

	if err != nil {
		return fmt.Errorf("entityone UpdateStatus(): %v", err)
	}

	// Update status
	e.ActionID = actionID
	e.StatusID = statusID

	return nil
}

// SelectEntityoneOneByStatus will retrieve one entityone from a selected status
func SelectEntityoneOneByStatus(
	q sqlx.Queryer,
	link SQLLink,
	statusID StatusID,
) (selectedEntity *Entityone, err error) {
	entityOnes, err := selectEntity(q, link, 0, int(statusID), 1)
	if err != nil {
		return nil, err
	}
	if len(entityOnes) == 0 {
		return nil, fmt.Errorf("no entity found for status %d", statusID)
	}
	selectedEntity = entityOnes[0]
	return selectedEntity, err
}

// SelectEntityoneOneByPK will retrieve one entityone from a selected status
func SelectEntityoneOneByPK(
	q sqlx.Queryer,
	link SQLLink,
	entityID int64,
) (selectedEntity *Entityone, err error) {
	entityOnes, err := selectEntity(q, link, entityID, 0, 0)
	if err != nil {
		return nil, err
	}
	if len(entityOnes) == 0 {
		return nil, fmt.Errorf("no entity found for %d", entityID)
	}

	selectedEntity = entityOnes[0]
	return selectedEntity, err
}

// selectEntity will retrieve a slice of entityones that are in status created
func selectEntity(
	q sqlx.Queryer,
	link SQLLink,
	entityID int64,
	statusID int,
	limit int,
) (entityOnes []*Entityone, err error) {

	query := `
        SELECT
            e.entityone_id, e.time_created,
            es.entityone_id as status_entityone_id, es.action_id, es.status_id, es.time_created as status_time_created
        FROM entityone e
        INNER JOIN entityone_status es ON es.entityone_id = e.entityone_id
            AND es.is_latest = 1
        WHERE 0 = 0
    `

	var filterEntities []int64
	var filterStatuses []int

	if entityID != 0 {
		filterEntities = append(filterEntities, entityID)
	}

	if statusID != 0 {
		filterStatuses = append(filterStatuses, statusID)
	}

	params, queryFilter := link.GetFilterSelectEntityOneQuery(
		filterEntities,
		filterStatuses,
	)
	query += queryFilter

	if limit > 0 {
		limit := ` LIMIT 1`
		query += limit
	}

	rows, err := q.Queryx(query, params...)
	if err != nil {
		return entityOnes, fmt.Errorf("entityone Select: %v", err)
	}

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
