CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       loginFromTables VARCHAR(255),
                       login VARCHAR(255),
                       password text,
                       dbName VARCHAR(255),
                       dbType VARCHAR(255),
                       connectionString VARCHAR(255)
);



