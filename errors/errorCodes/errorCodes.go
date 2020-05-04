package errorCodes

const Unknown = "unknown"
const InternalServerError = "internalServerError"

const BadRequest = "badRequest"
const ResourceNotFound = "resourceNotFound"
const RequiredFieldEmpty = "requiredFieldEmpty"
const EmailAddressAlreadyExists = "emailAddressAlreadyExists"
const EmailAddressDomainForbidden = "emailAddressDomainForbidden"

const JWTExpiredCanBeRefreshed = "jwtExpiredCanBeRefreshed"
const JWTExpiredCannotBeRefreshed = "jwtExpiredCannotBeRefreshed"
const JWTIsStillValid = "jwtIsStillValid"

const CredentialDoesNotMatch = "credentialDoesNotMatch"