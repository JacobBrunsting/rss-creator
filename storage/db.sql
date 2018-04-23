CREATE TABLE Users (
    username VARCHAR(64) NOT NULL,
    password VARCHAR(128) NOT NULL,
    email VARCHAR(256) NOT NULL,
    invalidatedtokens BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (username)
);
