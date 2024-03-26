# creatorDB

<div>
  <img src="https://github.com/devicons/devicon/blob/master/icons/go/go-original.svg" title="go" alt="go" width="40" height="40"/>&nbsp;
  <img src="https://github.com/devicons/devicon/blob/master/icons/docker/docker-original.svg" title="docker" alt="docker" width="40" height="40"/>&nbsp;
  <img src="https://github.com/devicons/devicon/blob/master/icons/grpc/grpc-original.svg" title="grpc" alt=grpc" width="40" height="40"/>&nbsp;
</div>

# 🐣 GRPC service on Go for creating a database of your choice(PostgreSQL, MySQL) in a docker container

# 📞 Description

The user makes a gRPC request passing his nickname, login for database, password, Database name and database type (PostgreSQL, MySQL)

# 💻 Example

```json
{
  "user": "username",
  "dbName": "databaseName",
  "dbType": "postgresql",
  "login": "login",
  "password": "password"
}
```
# 🏴‍☠️ Response 

```json
{
    "connectionString": "postgres://a:a@localhost:63858/a?sslmode=disable"
}
```
