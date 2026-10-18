package shopqua

// FieldType is a column type of a model. The quanar turns it into its database's type.
type FieldType string

const (
	FieldString   FieldType = "string"
	FieldText     FieldType = "text"
	FieldInteger  FieldType = "integer"
	FieldNumber   FieldType = "number"
	FieldBoolean  FieldType = "boolean"
	FieldDatetime FieldType = "datetime"
	FieldJSON     FieldType = "json"
	FieldEnum     FieldType = "enum"
)

// HttpMethod is a method a quan can ask the quanar to listen to.
type HttpMethod string

const (
	GET    HttpMethod = "GET"
	POST   HttpMethod = "POST"
	PUT    HttpMethod = "PUT"
	PATCH  HttpMethod = "PATCH"
	DELETE HttpMethod = "DELETE"
)

// QuanarMiddleware is a middleware the quanar provides. A quan names it in an api; it runs in
// the quanar, before the request reaches the quan.
type QuanarMiddleware string

const (
	// Authenticate answers 401 unless the request carries a valid token.
	Authenticate QuanarMiddleware = "authenticate"
	// RateLimit answers 429 when one client calls the api too often.
	RateLimit QuanarMiddleware = "rate-limit"
)

// RuleType is a value type a validation rule checks.
type RuleType string

const (
	RuleString  RuleType = "string"
	RuleEmail   RuleType = "email"
	RuleInteger RuleType = "integer"
	RuleNumber  RuleType = "number"
	RuleBoolean RuleType = "boolean"
	RuleEnum    RuleType = "enum"
	RuleArray   RuleType = "array"
	RuleObject  RuleType = "object"
)
