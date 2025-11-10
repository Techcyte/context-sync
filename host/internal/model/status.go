package model

type StatusCode int

const (
	OK                StatusCode = 200
	BadRequest        StatusCode = 400
	MethodNotAllowed  StatusCode = 405
	RequestTimeout    StatusCode = 408
	Conflict          StatusCode = 409
	ConflictWithRetry StatusCode = 419
	TooManyRequests   StatusCode = 429
	ServerError       StatusCode = 500
)

func (ack StatusCode) String() string {
	switch ack {
	case OK:
		return "OK"
	case BadRequest:
		return "BadRequest"
	case MethodNotAllowed:
		return "MethodNotAllowed"
	case RequestTimeout:
		return "RequestTimeout"
	case Conflict:
		return "Conflict"
	case ConflictWithRetry:
		return "ConflictWithRetry"
	case TooManyRequests:
		return "TooManyRequests"
	case ServerError:
		return "ServerError"
	}

	return "UnknwonwStatusCode"
}
