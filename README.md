# WIP: Auth Service

This is a repository for the authentication service. It's a microservice that handles user authentication and authorization. It's built with Go and uses gRPC for communication with an [API Gateway](https://github.com/kakitomeru/gateway).

## How to run

```bash
make start
# or
(set -a; source .env; set +a; go run cmd/main.go) # appends all env variables to the running command
```

## Endpoints

Here are the endpoints that are available in the service, represented via http transcodings.

### Register: `Register` | `POST /api/v1/auth/register`

<!-- #### Request body
```json
{
  "username": "string",
  "email": "string",
  "password": "string"
}
```

#### Response body
```json
{
  "userId": "string"
}
``` -->

### Login: `Login` | `POST /api/v1/auth/login`

<!-- #### Request body
```json
{
  "email": "string",
  "password": "string"
}
```

#### Response body
```json
{
  "userId": "string"
}
``` -->

### Logout: `Logout` | `POST /api/v1/auth/logout`
### Refresh Token: `RefreshToken` | `POST /api/v1/auth/refresh`
### Get User: `GetUser` | `GET /api/v1/me`