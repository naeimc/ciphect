DROP TABLE IF EXISTS keys;
DROP TABLE IF EXISTS users;

CREATE TABLE users (
    id          VARCHAR(32)     PRIMARY KEY,
    created     TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified    TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    username    VARCHAR(255)    NOT NULL UNIQUE,
    hash_type   VARCHAR(16)     NOT NULL,
    hash        BYTEA           NOT NULL
);

CREATE TABLE keys (
    id          VARCHAR(64)     PRIMARY KEY,
    user_id     VARCHAR(32)     NOT NULL,
    created     TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name        VARCHAR(255)    NOT NULL,
    session     BOOLEAN         NOT NULL,
    expires     BOOLEAN         NOT NULL,
    expiration  TIMESTAMPTZ     NOT NULL
);