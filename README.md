# Gamedev-backend
Backend for university gamedev project
### CLI flags

| Parameter (short)    | Default       | Usage                      |
|----------------------|---------------|----------------------------|
| --port (-p)          | 8080          | TCP port for API server    |
| --redis (-r)         | redis:6379    | Address of redis server    |
| --redispass (-b)     |               | Address of redis server    |
| --postrgresAddr (-a) | postgres:5432 | Address of Postgres server |
| --db (-d)            | fict          | Postgres database name     |
| --user (-u)          | fict          | Postgres user name         |
| --pass (-c)          | fict          | Postgres user password     |
| --verbose (-v)       | false         | Show debug level logs      |

### API

API accepts JSON format

**/api/v1/register** - POST - register player

JSON body params:

| Parameter | type   | Usage                                |
|-----------|--------|--------------------------------------|
| login     | string | Player's login                       |
| password  | string | Player's password                    |

Responses:

| Status code | Body                                          | Case                                             |
|-------------|-----------------------------------------------|--------------------------------------------------|
| 200         | {"err":null}                                  | Player successfully registered                   |
| 400         | {"err": "Bad json"}                           | Wrong or malformed request body                  |
| 400         | {"err": "Empty login or password"}            | No login or password provided                    |
| 403         | {"err": "User already exists"}                | Player with provided login is already registered |
| 500         |                                               | Internal error                                   |

**/api/v1/login** - POST - Get user's authorization token

JSON body params:

| Parameter     | type   | Usage                                |
|---------------|--------|--------------------------------------|
| login         | string | Player's login                       |
| password      | string | Player's pasword                     |

Responses:

| Status code | Body                                        | Case                                             |
|-------------|---------------------------------------------|--------------------------------------------------|
| 200         | {"token": "<token>"}                        | Successfully authorized                          |
| 400         | {"err": "Bad json"}                         | Wrong or malformed request body                  |
| 400         | {"err": "Empty login or password"}          | No login or password provided                    |
| 403         | {"err": "Failed to login"}                  | Wrong credentials or user not found              |
| 500         |                                             | Internal error                                   |

**/api/v1/game** - WS - Game session websocket

### Websocket API

**Authorization** - To autorize, send message in format of
```
{"channel": "auth", "authToken" : "<token>"}
```

Response:

```
{"channel": "auth", "response": <bool>}
```

response is true if user is authorized, false otherwise

**Proceed forward** - To go to next page of the story
```
{"channel": "story_move", "answerId": <answerId, optional>}
```
answerId must be provided if current page has a question

Response:

```
{"channel": "story_move", "response": <bool>}
```

response is true if story went on next page, false otherwise

**Save game** - when user wants to save game manually
```
{"channel": "story_save"}
```

Response:

```
{"channel": "story_save", "response": <bool>}
```

response is true if successfully saved, false otherwise

**Story text** - ws server sends story text in message of format 

```
{"channel": "story_text", "text": "<text>", "answers": [{"answerId": <answerId>, "text": "<answer text>"}...]}
```

answers array is optional and provided only if current text page has a question

**User stats** - ws server sends user's stats each time they change and on authorization, in message of format
```
{"channel": "stats", "stats": {"knowledge": int, "soberness": int, "performance": int,
 "prestige": int, "connections": int}}
```