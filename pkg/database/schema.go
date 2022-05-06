package database

var schema = []string{
	`
    PRAGMA foreign_keys = ON;
    `,
	`
    CREATE TABLE IF NOT EXISTS credential(
        id TEXT NOT NULL PRIMARY KEY,
        name TEXT NOT NULL,
        value TEXT NOT NULL
    );
    `,
	`
    CREATE TABLE IF NOT EXISTS config(
        id TEXT NOT NULL PRIMARY KEY,
        host TEXT NOT NULL,
        name TEXT NOT NULL,
        user_c TEXT,
        pass_c TEXT,
        ca_c TEXT,
        ovpn_c TEXT,
        FOREIGN KEY(user_c) REFERENCES credential(id),
        FOREIGN KEY(pass_c) REFERENCES credential(id),
        FOREIGN KEY(ca_c) REFERENCES credential(id),
        FOREIGN KEY(ovpn_c) REFERENCES credential(id)
    );
    `,
	`
    CREATE TABLE IF NOT EXISTS network(
        id TEXT NOT NULL PRIMARY KEY,
        name TEXT NOT NULL,
        config TEXT NOT NULL,
        FOREIGN KEY(config) REFERENCES config(id)
    );
    `,
	`
    CREATE TABLE IF NOT EXISTS client(
        id TEXT NOT NULL PRIMARY KEY,
        name TEXT NOT NULL,
        address TEXT NOT NULL
    );
    `,
	`
    CREATE TABLE IF NOT EXISTS client_network(
        client_id TEXT NOT NULL PRIMARY KEY,
        network_id TEXT,
        FOREIGN KEY(client_id) REFERENCES client(id),
        FOREIGN KEY(network_id) REFERENCES network(id)
    );
    `,
}
