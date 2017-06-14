Bruschetta
---

UEK Hackathon entry backend codenamed **Bruschetta**. The project is named "Platforma UEK"

## Running

In order to start the **bruschetta** backend you need to have a `docker-compose` and run the following command:

```
cp ./credentials.env.sample ./credentials.env
vim ./credentials.env #change values in credentials.env to yours
docker-compose up --build
```

## Endpoints

### POST /accounts/register/

Sample request:

```json
{
    "name": "Your Name",
    "email": "email@example.com",
    "group": 8801,
    "password": "your-password"
}
```

Sample response:

```json
{
    "token": "JWT-token"
}
```

All future requests have to contain `Authorization: Bearer YOUR-JWT-TOKEN` header. Token is valid for 24h

### POST /accounts/login/

```json
{
    "email": "email@example.com",
    "password": "your-password"
}
```

Sample response:

```json
{
    "token": "JWT-token"
}
```

### GET /accounts/token/

Used to refresh expired token. Pass old token in `Authorization` header and a new one will be returned.

Sample response:

```json
{
    "token": "new-jwt-token"
}
```

### GET /events/

**Role:** User

Lists all events appropriate to a supplied token.

Sample response:

```json
[{
    "ID": 1,
    "CreatedAt": "2017-06-13T10:02:23.009069Z",
    "UpdatedAt": "2017-06-13T10:02:23.012834Z",
    "DeletedAt": null,
    "user_id": 5,
    "name": "Test!",
    "description": "Another test",
    "message": "subtitle test",
    "priority": 2
}]
```

### GET /events/:id/?channel=:campaign

**Role:** User with a specific group or admin

Get specific event.

Campaign parameter is not required and is only used for interactions.

Sample response:

```json
{
    "ID": 1,
    "CreatedAt": "2017-06-13T10:02:23.009069Z",
    "UpdatedAt": "2017-06-13T10:02:23.012834Z",
    "DeletedAt": null,
    "user_id": 5,
    "name": "Test!",
    "description": "Another test",
    "message": "subtitle test",
    "priority": 2,
    "group": 8801
}
```

### GET /events/:id/interactions/

Lists all interactions with channel from where the traffic originates.

**Role:** Admin

Sample response:

```json
[
    {
        "id": 1,
        "event_id": 1,
        "timestamp": "2017-06-13T19:29:12.106076Z",
        "user_id": 2,
        "channel": "messenger"
    },
    {
        "id": 2,
        "event_id": 1,
        "timestamp": "2017-06-13T19:29:33.841749Z",
        "user_id": 2
    }
]
```

### POST /events/

Posts an event and sends notifications to all matching students. Specifying `group` parameter limits the message to a specific group only.

**Role:** Admin

Sample request:

```json
{
  "name": "Test!",
  "description": "Another test",
  "image": "https://example.com/example.jpg",
  "message": "Group Target priority test - 1",
  "priority": 1,
  "group": 8801
}
```

### DELETE /events/:id/

Deletes the event speified by `:id`

**Role:** Admin

### PUT/PATCH /events/:id/

Updates the event by replacing (PUT) or changing parameters (PATCH)

**Role:** Admin

### GET /subscriptions/

Lists all user subscriptions.

**Role:** User

### POST /subscriptions/

Adds user's subscription.

**Role:** User

Sample request:

```json
{
    "channel": "messenger",
    "channel_id": "messenger-page-id",
    "priority": 2
}
```
### PATCH /subscriptions/:id/

Patches user's subscription

**Role:** User

### DELETE /subscriptions/:id/

Deletes user's subscription

**Role:** User

### GET /timetable/

Gets user's timetable.

**Role:** User

### GET /timetable/groups/

Gets all category->group->id associations.
