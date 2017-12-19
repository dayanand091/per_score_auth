Per Score Auth!
===================

This service is responsible for managing **Validations**  and ** Secure Authentication** for user come from **perScoreServer** and calling any other internal services using **GRPC** as needed. It has implemented  **GRPC** calls in order to interacts between two services.

----------

Description
-------------

This application is implemented on Go lang **NET/HTTP** package for starting server and making **GRPC** call between  other internal services. following features  are below .

> **Note:**
> - Before Run the application ,You must load ```.sh```  script in the root directory to help you get everything setup to run it locally use any of the go cobra command  so that it can load all the environment variable and setup postgres to create role and database at the same time.
> - Sometimes the bash does not load the environment ```.env``` file form the bash ```.sh``` file, if there is any issue while running the code than you must fetch the environment file by the following command ```source.env``` before running any of the go cobra command.
> - Make sure you have all the required and correct environment variable available before running the service.
> - Make sure that you have postgres installed in your machine.

#### <i class="icon-file"></i> GetSession

This is one of the route that will be called by the  **perScoreServer** using GRPC calls.getSession will use to authentication as well as validation . if it is Authenticated the response will contains auth_token and success message otherwise a failure response will be return to perScoreServer.

#### <i class="icon-folder-open"></i> CreateUser

This route will be called by **perScoreServer** when user is trying to register himself as admin, questioner or responder using GRPC calls. **CreateUser** service will use to store the user data in database if successful created  then The response return Status,Token,Message.if response is failed then response contains Status,Token,Message, Fields.

> **Github URL:**  [<i class="icon-download"></i> perScoreAuth](#https://github.com/dayanand091/per_score_auth)

----------


Dependencies
-------------------
1> Packages

*  [Cobra](https://github.com/spf13/cobra)
* [Postgresql](https://wiki.westfieldlabs.com/display/WL/PostgreSQL)
* [GORM](https://github.com/jinzhu/gorm)
* [Go](https://wiki.westfieldlabs.com/display/WL/Go)
* [GRPC](https://google.golang.org/grpc)

2> Testing packages

* [GINKO](https://github.com/onsi/ginkgo)
* [GOMEGA](https://github.com/onsi/gomega)

----------


Build and run this project
-------------

>1. To give privilege to ur .sh file
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

----------


## PerScoreAuth Directories
--------------------
below is a basic layout for projects. It represents the most common directory structure with a number of small enhancements along with several supporting directories common to all applications.

### `/perScoreAuth`

Main applications or root directory for this project.

All the subdirectories are in the application directory.

### `/cmd`

This directories is created while initialising the cobra. It will contain all the file related to the cobra cli commands including the root file generated by cobra by default

### `/models`

This directory contains struct and data migration that represent data relation

### `/server`

It will start the server of this go service on the provided host and port.

### `/perScoreProto`

 It will contain all the proto and the compiled file used by the application.

`setupPostgres.sh` : - this file will automatically set up the postgres in your system including the ROLE AND DATABASE in one go.

### Tables

**it will have one table  which will contain following columns** :

Item           | Value
--------       | ---
ID		       | int
Token          |string
ExpirationTime | int
Email          | string
