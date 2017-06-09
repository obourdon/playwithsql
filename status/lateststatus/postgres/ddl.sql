
CREATE TABLE entityone (
    entityone_id BIGSERIAL NOT NULL,
    time_created DATE NOT NULL DEFAULT CURRENT_DATE,
    PRIMARY KEY (entityone_id)
);

CREATE TABLE IF NOT EXISTS entityone_status (
    entityone_status_id BIGSERIAL NOT NULL,
    entityone_id BIGINT NOT NULL,
    action_id INT NOT NULL DEFAULT 1,
    status_id INT NOT NULL DEFAULT 1,
    time_created DATE NOT NULL DEFAULT CURRENT_DATE,
    PRIMARY KEY(entityone_status_id),
    CONSTRAINT es_fk_ei_e
        FOREIGN KEY (entityone_id)
        REFERENCES entityone (entityone_id)
);

CREATE TABLE IF NOT EXISTS entityone_lateststatus (
    entityone_id BIGINT NOT NULL,
    entityone_status_id BIGINT NOT NULL,
    UNIQUE (entityone_status_id),
    UNIQUE (entityone_id),
    PRIMARY KEY (entityone_id, entityone_status_id),
    CONSTRAINT el_fk_e_ei
        FOREIGN KEY (entityone_id)
        REFERENCES entityone (entityone_id),
    CONSTRAINT el_fk_es_esi
        FOREIGN KEY (entityone_status_id)
        REFERENCES entityone_status (entityone_status_id)
);

CREATE INDEX es_idx_sid ON entityone_status(status_id);

CREATE INDEX es_fk_ei_e_idx ON entityone_status(entityone_id);


