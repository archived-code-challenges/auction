# Auction bid tracker

This project consists of an authentication and authorisation service, as well as a bidding system based on users and items. 

## Implementation decisions

I have decided to run this task without doing all possible CRUD methods, in order not to obfuscate the code.
I have considered this on the assumption that the structure is consistent enough to get an idea of how they would have been implemented. I hope you like it.

### Data chart

    ┌─────────────┐                       ┌─────────────┐
    │    USERS    ├───────────────────────┤     BIDS    │
    │             │1                  0..N│             │
    └─────────────┘                       └──────┬──────┘
                                            0..N │
    ┌─────────────┐                              │
    │             │ 1                            │
    │    ITEMS    ├──────────────────────────────┘
    └─────────────┘

### Data structure

| User  | Item   | Bid     |
| ----- | ------ | ------- |
| id    | id     | id      |
| name  | name   | item id |
|       | value* | user id |
|       |        | amount  |
---

*value is used as starting price of an object in the auction service.

### Chosen data structures and concurrency approach

I have used:

- A set of Golang [maps](https://blog.golang.org/maps) along with
- [RWMutexes](https://golang.org/pkg/sync/#RWMutex) and
- incremental integers used to simulate the sequential addition of elements into the map

Everything can be found on the [memdatabase.go](/internal/models/memdatabase.go) file

As far as the entities **user** and **item** are concerned, the mutex used will not return any error if two concurrent users try to add one of these items to the storage but wait until the first one is done to process the second one.
On the other hand, the code in charge of the **bids** will return an error if two users try to access the same resource, excluding the last one that consumes the service. I have decided to do this to ensure consistency of bids as well as to guarantee that the amount a user bids does not increase if the bid comes in second place.

Some tests can be found at [memdatabase_test.go](/internal/models/memdatabase_test.go)

## Instructions to run the project

The project is vendored so nothing else than `go run` might be needed. You can use the makefile that will ensure everything is correctly in place and point to the main file for you.

    make go-run

There is also a docker version of the execution. But depending of your SO you will need to [configure how to access to the container](https://stackoverflow.com/a/24326540) through your host.

    make docker-build
    make docker-run

## API usage

The project has a [Postman collection](/docs/auction-bid-tracker.postman_collection.json) attached, which can be used to interact with the auction service.

### Packaging

- entrypoint in `cmd/sales-api`
- HTTP layer in `cmd/sales-api/internal/handlers`
- business logic in `internal/models`
    * in-memory database in `internal/models/memdatabase.go`
- framework for common HTTP related tasks in `internal/web`
- helper functions to process data before response/request in `internal/views`
- documentation, images and helpful files in `docs/`

## Benchmarking

I did some testing to check if the performance was good enough and IMO the processing times are fast.

![bench-create-bids](/docs/images/bench-create-bids.png)
> Benchmarking of 500 sequential creations of a bid using the model and storage.

![bench-list-by-user-id](/docs/images/bench-listing-by-user-id.png)
> Benchmarking of 500 sequential listings of items given a user ID. For this one I created 1000 bids for 5 items with the same user and then queried to find the items per user.

I used [ali](https://github.com/nakabonne/ali) for the load testing.
