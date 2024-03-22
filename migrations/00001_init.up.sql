CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       login VARCHAR(255),
                       password text,
                       dbName VARCHAR(255),
                       dbType VARCHAR(255),
                       connectionString VARCHAR(255)
);



