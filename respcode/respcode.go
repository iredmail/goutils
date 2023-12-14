package respcode

import "errors"

const (
	InternalServerError      = "INTERNAL_SERVER_ERROR"
	MissingCSRFToken         = "MISSING_CSRF_TOKEN"
	PermissionDenied         = "PERMISSION_DENIED"
	LoginRequired            = "LOGIN_REQUIRED"
	LoginOrAPIKeyRequired    = "LOGIN_OR_API_KEY_REQUIRED"
	InvalidCredentials       = "INVALID_CREDENTIALS"
	InvalidLicenseKey        = "INVALID_LICENSE_KEY"
	InvalidProduct           = "INVALID_PRODUCT"
	InvalidDomain            = "INVALID_DOMAIN"
	LoggedOut                = "LOGGED_OUT"
	Added                    = "ADDED"
	Updated                  = "UPDATED"
	Deleted                  = "DELETED"
	Enabled                  = "ENABLED"
	Disabled                 = "DISABLED"
	SignedUp                 = "SIGNED_UP"
	PasswordMismatch         = "PASSWORD_MISMATCH"
	PasswordTooShort         = "PASSWORD_TOO_SHORT"
	InvalidPlatform          = "INVALID_PLATFORM"
	InvalidBackend           = "INVALID_BACKEND"
	InvalidComponent         = "INVALID_COMPONENT"
	InvalidParam             = "INVALID_PARAM"
	InvalidValue             = "INVALID_VALUE"
	InvalidUpgrade           = "INVALID_UPGRADE"
	InvalidUpdate            = "INVALID_UPDATE"
	InvalidFormData          = "INVALID_FORM_DATA"
	InvalidOtpCode           = "INVALID_OTP_CODE"
	InvalidDeploymentID      = "INVALID_DEPLOYMENT_ID"
	InvalidEmailAddress      = "INVALID_EMAIL_ADDRESS"
	ImportedSettings         = "IMPORTED_SETTINGS"
	ApplyingSavedChanges     = "APPLYING_SAVED_CHANGES"
	ApplyingUpgrade          = "APPLYING_UPGRADE"
	NotDeploying             = "NOT_DEPLOYING"
	AccountExists            = "ACCOUNT_EXISTS"
	NotAllowed               = "NOT_ALLOWED"
	ExceededDomainMLLimit    = "EXCEEDED_DOMAIN_ML_LIMIT"
	ExceededDomainAliasLimit = "EXCEEDED_DOMAIN_ALIAS_LIMIT"
	EmailAlreadyExists       = "EMAIL_ALREADY_EXISTS"
)

var (
	ErrInvalidCredentials         = errors.New(InvalidCredentials)
	ErrInternalServerError        = errors.New(InternalServerError)
	ErrInvalidLicenseKey          = errors.New(InvalidLicenseKey)
	ErrInvalidEmailAddress        = errors.New(InvalidEmailAddress)
	ErrInvalidProduct             = errors.New(InvalidProduct)
	ErrInvalidDomain              = errors.New(InvalidDomain)
	ErrInvalidParam               = errors.New(InvalidParam)
	ErrInvalidSignupToken         = errors.New("INVALID_SIGNUP_TOKEN")
	ErrInvalidVersionNumber       = errors.New("INVALID_VERSION_NUMBER")
	ErrInvalidReleaseDay          = errors.New("INVALID_RELEASE_DAY")
	ErrInvalidFormData            = errors.New("INVALID_FORM_DATA")
	ErrInvalidColumn              = errors.New("INVALID_COLUMN")
	ErrInvalidCustomer            = errors.New("INVALID_CUSTOMER")
	ErrEmptyPassword              = errors.New("EMPTY_PASSWORD")
	ErrInvalidAccount             = errors.New("INVALID_ACCOUNT")
	ErrInvalidAPIKey              = errors.New("INVALID_API_KEY")
	ErrPasswordMismatch           = errors.New(PasswordMismatch)
	ErrPasswordTooShort           = errors.New(PasswordTooShort)
	ErrPasswordTooLong            = errors.New("PASSWORD_TOO_LONG")
	ErrPasswordNoLetter           = errors.New("PASSWORD_NO_LETTER")
	ErrPasswordNoUpperLetter      = errors.New("PASSWORD_NO_UPPER_LETTER")
	ErrPasswordNoNumber           = errors.New("PASSWORD_NO_NUMBER")
	ErrPasswordNoSpecialChar      = errors.New("PASSWORD_NO_SPECIAL_CHAR")
	ErrPermissionDenied           = errors.New(PermissionDenied)
	ErrDomainExists               = errors.New("DOMAIN_EXISTS")
	ErrAccountExists              = errors.New(AccountExists)
	ErrNotAllowed                 = errors.New(NotAllowed)
	ErrInvalidPasswordScheme      = errors.New("INVALID_PASSWORD_SCHEME")
	ErrExceededDomainAccountLimit = errors.New("EXCEEDED_DOMAIN_ACCOUNT_LIMIT")
	ErrExceededDomainQuotaSize    = errors.New("EXCEEDED_DOMAIN_QUOTA_SIZE")
	ErrUnsupportedPasswordScheme  = errors.New("UNSUPPORTED_PASSWORD_SCHEME")
	ErrDisallowToCreateUser       = errors.New("DISALLOW_TO_CREATE_USER")
	ErrEmailAlreadyExists         = errors.New(EmailAlreadyExists)
)
