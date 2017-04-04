package sqlite

import "github.com/jmoiron/sqlx"

// Link is used to insert and update in mysql
type Link struct{}

// MigrateUp creates the needed tables
func (link *Link) MigrateUp(exec sqlx.Execer) (errExec error) {
	_, errExec = exec.Exec(`
        CREATE TABLE IF NOT EXISTS entityone (
            entityone_id INTEGER PRIMARY KEY ASC,
            time_created DATETIME DEFAULT (datetime('now','localtime'))
        )
    `)
	if errExec != nil {
		return errExec
	}

	_, errExec = exec.Exec(`
        CREATE TABLE IF NOT EXISTS entityone_status (
            entityone_id INT NOT NULL,
            action_id INT NOT NULL DEFAULT 1,
            status_id INT NOT NULL DEFAULT 1,
            time_created DATETIME DEFAULT (datetime('now','localtime')),
            is_latest INT(1) NULL DEFAULT 1,
            UNIQUE (is_latest, entityone_id),
            CONSTRAINT es_fk_e
            FOREIGN KEY (entityone_id)
            REFERENCES entityone (entityone_id)
        )
    `)
	if errExec != nil {
		return errExec
	}

	_, errExec = exec.Exec(
		`CREATE INDEX es_idx1 ON entityone_status(status_id, is_latest)`,
	)
	if errExec != nil {
		return errExec
	}

	_, errExec = exec.Exec(
		`CREATE INDEX es_idx2 ON entityone_status(entityone_id)`,
	)
	return errExec
}

// MigrateDown destroys the needed tables
func (link *Link) MigrateDown(exec sqlx.Execer) (errExec error) {
	_, errExec = exec.Exec(`DROP TABLE IF EXISTS entityone_status`)
	if errExec != nil {
		return errExec
	}

	_, errExec = exec.Exec(`DROP TABLE IF EXISTS entityone`)
	return errExec
}
