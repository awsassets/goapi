# GoAPI

GoAPI is a simple, secure and fast API writen in Golang, user authentication ready.

This template is forked from one of my personal projects, every feature is not implemented yet.

## 👨‍🍳 What is it made of

Useby is based on the [Fiber](https://github.com/gofiber/fiber) Go framework, which is built on top of [Fasthttp](https://github.com/valyala/fasthttp), the fastest http engine for go.

[JWT](https://github.com/dgrijalva/jwt-go) handles the authentication.

Useby is conceived to work with [MongoDB](https://github.com/mongodb/mongo-go-driver).
It does not use any ODM, first to avoid having an  extra abstraction level, second because its useless with mongoDB thanks to the "omitempty" tag which validate data when mapping json to objects.

And it is built on an MVC + Service-Repository architecture for best flexibility.

## ⚡️ Quick start

```
go get -u github.com/natnatf/goapi
go build main.go
go run main.go
```
In [Postman](https://www.postman.com)
import the requests collection and localhost environment
[available here](../doc).

First **Post** a user to get a **JWT**, and use it as bearer token for the other requests.

## ⚙️ Project Architecture

#### 🤖 Models
They set what the objects are.

#### 📬 Controllers
They receive the queries' data.
They call services to get the data asked by the query.
Finally, they send back what the services return them to the client.

#### 🧠 Services
They handle the business logic.
Their role is to receive data from controllers, process it and call repositories to get data from the database.
They return data or errors to controllers.

#### 📚 Repositories
Their role is to interact with the database.
All the operations with the database are made here, only services should call them.

#### 🔧 Config
Basic server config info.

#### ❌ Errors
Some errors text and codes.
Codes are useful for the client application, it allows it to react depending on it and tell the user what's wrong.
Text provide more details on what wrong.

## 👨‍💻 Customisation

You can easily adapt this template and implement your own objects and modifying the config file.