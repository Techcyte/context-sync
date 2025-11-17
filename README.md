# Local Interconnection Service Protocol

This is a demo repository for the Local Interconnection Service Protocol (LIS Protocol).

The `host` folder contains the code for the example host. This is a `go` project. See the `README.md` file in the
`host` folder for instructions on running it.

The `client` folder contains the code for a sample client. This is an `npm` project. See the `README.md` file in the
`client` folder for instructions on running it.

It's important to note that the demo client and the demo host are meant to be starting points to show the flow of the
protocol and do not handle edge cases adequetely. 

## Overview

The Local Interconnection Service Protocol (LIS Protocol or LISP) is a protocol to allow Fusion to synchronize context
with LIS solutions running locally. When the case is switched in one application, the change should be reflected in
the other application. The LIS Protocol uses a WebSocket connection to send JSON based messages to allow Fusion and the
LIS to communicate and sync the current context. The protocol will allow either party to request a new context and
accept or reject context change requests. The protocol will also allow the client and host to handle de-synchronization
errors.

The LIS Protocol requires the LIS application to be the host and Fusion will be the client. Only one client can be
subscribed to the host at a time. Multiple clients can be connected to the host and if the subscribed client disconnects
the host can choose to make one of the remaining connected clients the subscribed client. The subscribed client is the
client the LIS application will synchronize context with.

## Message Kinds

The following message types are used to maintain context synchronization:
* `sub-request`
* `sub-accept`
* `sub-reject`
* `ctx-change-request`
* `ctx-change-accept`
* `ctx-change-reject`
* `ctx-null`
* `sync-error`

## Message Structure
* `kind: string` - The kind of message being sent or received. Required for all message types.
* `info` - Required for `sub-request`, `sub-accept`, `sub-reject messages`.
	* `version: number` - The minimum protocol version supported by the application.
	* `application: string` - The name of the application sending the message.
	* `replace_exiting_client: boolean` - Optional. If true then the subscribed client will be unsubscribed and the requesting client will become the subscribed client.
* `context` - An array of context objects. Required for `sub-accept`, and `ctx-change-request` messages.
	* `key: string` - The context kind.
	* `value: string` - The context value.
* `rejection` - Explains why the request was rejected. Required for `sub-reject`, and `ctx-change-reject` messages.
	* `reason: string` - Why the request was rejected.
	* `status: number` - The HTTP status code for the request rejection. 
* `current_context` - The current context. Required for `ctx-change-reject` messages. Can only be used with `ctx-change-reject` messages.
	* `key: string` - The context kind.
	* `value: string` - The context value.
* `error` - For errors that result in de-synchronization. Required for `sync-error` messages.
	* `message: string` - The error message.
	* `status: number` - The HTTP status code for the error.

## Status Codes

| Name              | Value |
|-------------------|-------|
| OK                | 200   |
| BadRequest        | 400   |
| Conflict          | 409   |
| ConflictWithRetry | 419   |
| UpgradeRequired   | 426   |
| TooManyRequests   | 429   |
| ServerError.      | 500   |
| Unknown.          | 520   |

## Sample Messages


### Subscription Request
```JSON
{
  "kind": "sub-request",
  "info": {
    "version": 1,
    "application": "Fusion"
  }
}
```

Sent By the client to the host. The host should never send a `sub-request` message.

### Subscription Request with Initial Context
```JSON
{
  "kind": "sub-request",
  "info": {
    "version": 1,
    "application": "Fusion"
  },
  "context": [
    {
      "key": "case",
      "value": "N123456"
    }
  ]
}
```

If the user is already in a case when they initiate context sync send a requet with the case as the initial context.
The host should switch to the context sent in this request unless it is unable to do so.

### Subscription Accept
```JSON
{
  "kind": "sub-accept",
  "info": {
    "version": 1,
    "application": "LIS"
  }
}
```

Sent by the host to the requesting client if no client is subscribed. Also sent to a client if the subscribed client's
connection is closed, indicating the the client that previously had its `sub-request` rejected is now the subscribed
client. The client should never send a `sub-accept` message.

### Subscription Accept with Initial Context
```JSON
{
  "kind": "sub-accept",
  "info": {
    "version": 1,
    "application": "LIS"
  },
  "context": [
    {
      "key": "case",
      "value": "N123456"
    }
  ]
}
```

If the host has an active context when the client makes a subscription request the host should include that context in
the `sub-accept` message.

### Subscription Reject
```JSON
{
  "kind": "sub-reject",
  "info": {
    "version": 1,
    "application": "LIS"
  },
  "rejection": {
    "reason": "Client already connected.",
    "status": 409
  }
}
```

Sent by the host to the requesting client if another client is already subscribed. The host can send a status of
`409 (Conflict)` to indicate the client will not later be sent a `sub-accept` and should close the connection. The host
can send a status of `419 (ConflictWithRetry)` to indicate the client may later become the subscribed client and should
keep the connection open. Other status codes can be used to indicate other errors, for instance a `400` status if the
message is malformed or a status in the `500s` for internal errors preventing subscription. The client should never
send a `sub-reject` message.

### Context Change Request
```JSON
{
  "kind": "ctx-change-request",
  "context": [
    {
      "key": "case",
      "value": "N123456"
    }
  ]
}
```

Sent by either the host or the client to request changing the context.

### Context Change Accepted
```JSON
{
  "kind": "ctx-change-accept",
  "context": [
    {
      "key": "case",
      "value": "N123456"
    }
  ]
}
```

### Context Change Rejected
```JSON
{
  "kind": "ctx-change-reject",
  "current_context": [
    {
      "key": "case",
      "value": "N4567890"
    }
  ],
  "context": [
    {
      "key": "case",
      "value": "N123456"
    }
  ],
  "rejection": {
    "reason": "The doodad field has not been saved."
  }
}
```

Where `current_context` is the current context of the rejector and `context` is the context being rejected.

### Context Change Rejected (Outstanding Request)

If one side has an outsanding request and it recieves a new request it should send a `ctx-change-reject` message with
the `rejection.status` set to 409.

```JSON
{
  "kind": "ctx-change-reject",
  "context": [
    {
      "key": "case",
      "value": "N4567890"
    }
  ]
  "context": [
    {
      "key": "case",
      "value": "N123456"
    }
  ],
  "rejection": {
    "reason": "Rejected because of outstanding request.",
	"status": 409
  }
}
```

### Empty Context
```JSON
{
  "kind": "ctx-null"
}
```

Sent when the sender has no current context. As an example the user could navigate from a case to the worklist and no
longer have a context. This would be sent to let the other know the sender is not in sync but there isn't an out of
sync error either.

### Sync Error
```JSON
{
  "kind": "sync-error",
  "context": [
    {
      "key": "case",
      "value": "N123456"
    }
  ],
  "error": {
    "message": "Failed to navigate to N123456",
    "status": 500
  }
}
```

Can be sent by either the host or the client if there is some error that causes a desynchronization before or after the
context has been agreed upon by the host and the client. As an example lets say the both the host and the client have
agreed to swwitch to a new context and after sending the `ctx-change-accept` message the host fails to switch to the
new context due to an unexpected internal error. The host would send a `sync-error` error and the client could decide
to show an error message to the user and potentially rollback to the previous context to restore syncronization.

## Scenarios

The following scenarios are from the point of view of the client.

### Client connects to host and accepts new context:
1. Connect to websocket
2. Send `sub-request` message with connection info
3. Receive `sub-accept` message with connection info
6. Receive `ctx-change-request` message with new context
7. Navigate to new context
8. Send `ctx-change-accept` message

### Client with initial context connects to host and accepts new context:
1. Connect to websocket
2. Send `sub-request` message with connection info and initial context
3. Receive `sub-accept` message with connection info and same initial context
6. Receive `ctx-change-request` message with new context
7. Navigate to new context
8. Send `ctx-change-accept` message

### Client connects to host with initial context and accepts new context:
1. Connect to websocket
2. Send `sub-request` message with connection info
3. Receive `sub-accept` message with connection info and initial context
4. Navigate to initial context
5. Send `ctx-change-accept` message
6. Receive `ctx-change-request` message with new context
7. Navigate to new context
8. Send `ctx-change-accept` message

### Client connects to host with initial context and fails to navigate to new context:
1. Connect to websocket
2. Send `sub-request` message with connection info
3. Receive `sub-accept` message with connection info and initial context
4. Navigate to initial context
5. Send `ctx-change-accept` message
6. Receive `ctx-change-request` message with new context
7. Fail to navigate to new context, show the user an error
8. Send `sync-error` message with error message and status code

### Client attempts to connect while host has active connection:
1. Connect to websocket
2. Send `sub-request` message with connection info
3. Receive `sub-reject` message with connection info, rejection reason, and status code 409
4. Close connection

### Client attempts to connect while host has active connection later becomes active connection:
1. Connect to websocket
2. Send `sub-request` message with connection info
3. Receive `sub-reject` message with connection info, rejection reason, and status code 419
4. Keep connection open and wait
5. Receive `sub-accept` message with connection info and initial context
6. Navigate to initial context
7. Send `ctx-change-accept` message

### Client successfully requests context change:
1. Connect to websocket
2. Send `sub-request` message with connection info
3. Receive `sub-accept` message with connection info and initial context
4. Navigate to initial context
5. Send `ctx-change-accept` message
6. User clicks on new case
7. Send `ctx-change-request` message with new context
8. Optimistically navigate to new context
9. Receive `ctx-change-accept` message

### Host rejects context change request:
1. Connect to websocket
2. Send `sub-request` message with connection info
3. Receive `sub-accept` message with connection info and initial context
4. Navigate to initial context
5. Send `ctx-change-accept` message
6. User clicks on new case
7. Send `ctx-change-request` message with new context
8. Optimistically navigate to new context
9. Receive `ctx-change-reject` message with reason
10. Show the user an error message so they know the context is out of sync

### User navigates from case view to worklist view:
1. Connect to websocket
2. Send `sub-request` message with connection info
3. Receive `sub-accept` message with connection info and initial context
4. Navigate to initial context
5. Send `ctx-change-accept` message
6. User navigates to worklist from the case view
7. Send `ctx-null` message

### Client connects to host and fails to navigate to initial context:
1. Connect to websocket
2. Send `sub-request` message with connection info
3. Receive `sub-accept` message with connection info and initial context
4. Fail to navigate to initial context, indicate to user the connection is still open but out of sync
5. Send `sync-error` message with error message and status code

### Client connects to host and fails to navigate to initial context, succeeds with later context:
1. Connect to websocket
2. Send `sub-request` message with connection info
3. Receive `sub-accept` message with connection info and initial context
4. Fail to navigate to initial context, indicate to user the connection is still open but out of sync
5. Send `sync-error` message
6. Receive `ctx-change-request` message with new context
7. Navigate to new context
8. Send `ctx-change-accept` message

### Client connects to host and fails to navigate to initial context, client requests new context:
1. Connect to websocket
2. Send `sub-request` message with connection info
3. Receive `sub-accept` message with connection info and initial context
4. Fail to navigate to initial context, indicate to user the connection is still open but out of sync
5. Send `ctx-change-reject` message with reason
6. User clicks on new case
7. Send `ctx-change-request` message with new context
8. Optimistically navigate to new context
9. Receive `ctx-change-accept` message

### Client successfully requests context change and fails to navigate:
1. Connect to websocket
2. Send `sub-request` message with connection info
3. Receive `sub-accept` message with connection info and initial context
4. Navigate to initial context
5. Send `ctx-change-accept` message
6. User clicks on new case
7. Send `ctx-change-request` message with new context
8. Fail to optimistically navigate to new context
8. Receive `ctx-change-accept` message
10. Send `sync-error` message with error message and status code

### Host successfully requests context change and fails to switch:
1. Connect to websocket
2. Send `sub-request` message with connection info
3. Receive `sub-accept` message with connection info and initial context
4. Navigate to initial context
5. Send `ctx-change-accept` message
6. User clicks on new case in LIS
7. Receive `ctx-change-request` message with new context
8. Navigate to new context.
9. Send `ctx-change-accept` message
10. LIS fails to switch to new context
11. Receive `sync-error` message with error message and status code
12. Indicate to user the connection is still open but out of sync

### Reccieve context change request with outstanding context change request:
1. Connect to websocket
2. Send `sub-request` message with connection info
3. Receive `sub-accept` message with connection info and initial context
4. Navigate to initial context
5. Send `ctx-change-accept` message
6. User clicks on new case
7. Send `ctx-change-request` message with new context
8. Optimistically navigate to new context
9. Receive `ctx-change-request` message with new context
10. Send `ctx-change-reject` message with reason and status code indicating there is an outstanding request (409)
11. Receive `ctx-change-accept` or `ctx-change-reject` message.
    * If `ctx-change-reject` show the user a desync error and send a `sync-error` message
	* If `ctx-change-accept` stay at currect case.
