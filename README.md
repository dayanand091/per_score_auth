# per score auth
This service provide Validation for user authentication and return auth token to perscore server service .It implements both GRPC calls in order to interacts between other services and its internal services

### Dependencies

* [Postgresql](https://wiki.westfieldlabs.com/display/WL/PostgreSQL)
* [GORM](https://github.com/jinzhu/gorm)
* [Go](https://wiki.westfieldlabs.com/display/WL/Go)

### Build and run this project

1. To give privilege to ur .sh file
    ```
  chmod +x setupPostgres.sh
    ```
2. Run .sh file to create role and database
    ```
    ./setupPostgres.sh
    ```
3. Run command to migrate database
    ```
    go run main.go createDB
    ```
4. Run command to start server
    ```
    go run main.go serve
    ```
