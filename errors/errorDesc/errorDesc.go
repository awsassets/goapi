package errorDesc

// These error description are useful for developers to debug

const Unknown = "unknown error, operation has not been finished"
const InternalServerError = "internal server error"

const ResourceNotFound = "resource not found"
const RequiredFieldEmpty = "at least of the required fields is empty, maybe you mistyped it, or left it empty but it must be filled with something to be inserted in database"

const EmailAddressAlreadyExists = "email address already exists"
const EmailAddressDomainForbidden = "email address domain is forbidden"

const NoTokenWereProvided = "no token were provided"
const AuthorizationHeaderMustBeBearerToken = "authorization header format must be Bearer {token}"
const TokenSignatureIsNotValid = "token signature is not valid"
const JWTExpiredCanBeRefreshed = "this jwt is expired but can be refreshed"
const JWTExpiredCannotBeRefreshed = "this jwt is expired and cannot be refreshed, you must login"
const JWTIsStillValid = "this token is still valid and cannot be refreshed yet"

const CredentialDoesNotMatch = "invalid email address or password"
