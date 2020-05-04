package config

// JWT constants
const HmacSampleSecret = "(;.W2g]z[eN@Eck["
const JWTExpirationTimeInMinutes = 15
const JWTRefreshDeadlineInHours = 7 * 24

// Application status
const CurrentAPIVersion = 0
const DevStatus = true                       // dont forget to leave false for production
const LimitElementsReturnedFromDatabase = 10 // used for some getAll methods
