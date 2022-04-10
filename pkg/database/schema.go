package database

var schema = []string{
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
}
