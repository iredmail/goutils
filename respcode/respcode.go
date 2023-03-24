package respcode

import "errors"

const (
	InternalServerError   = "INTERNAL_SERVER_ERROR"
	PermissionDenied      = "PERMISSION_DENIED"
	LoginRequired         = "LOGIN_REQUIRED"
	LoginOrAPIKeyRequired = "LOGIN_OR_API_KEY_REQUIRED"
	InvalidCredentials    = "INVALID_CREDENTIALS"
	InvalidLicenseKey     = "INVALID_LICENSE_KEY"
	InvalidProduct        = "INVALID_PRODUCT"
	InvalidDomain         = "INVALID_DOMAIN"
	LoggedOut             = "LOGGED_OUT"
	Added                 = "ADDED"
	Updated               = "UPDATED"
	SignedUp              = "SIGNED_UP"
	Deleted               = "DELETED"
	PasswordMismatch      = "PASSWORD_MISMATCH"
	PasswordTooShort      = "PASSWORD_TOO_SHORT"
	InvalidBackend        = "INVALID_BACKEND"
	InvalidComponent      = "INVALID_COMPONENT"
	InvalidParam          = "INVALID_PARAM"
	InvalidValue          = "INVALID_VALUE"
	InvalidUpgrade        = "INVALID_UPGRADE"
	InvalidUpdate         = "INVALID_UPDATE"
	InvalidFormData       = "INVALID_FORM_DATA"
	InvalidOtpCode        = "INVALID_OTP_CODE"
	InvalidDeploymentID   = "INVALID_DEPLOYMENT_ID"
	ImportedSettings      = "IMPORTED_SETTINGS"
	ApplyingSavedChanges  = "APPLYING_SAVED_CHANGES"
	ApplyingUpgrade       = "APPLYING_UPGRADE"
	NotDeploying          = "NOT_DEPLOYING"
)

var (
	ErrInvalidCredentials   = errors.New(InvalidCredentials)
	ErrInternalServerError  = errors.New(InternalServerError)
	ErrInvalidLicenseKey    = errors.New(InvalidLicenseKey)
	ErrInvalidEmailAddress  = errors.New("INVALID_EMAIL_ADDRESS")
	ErrInvalidProduct       = errors.New(InvalidProduct)
	ErrInvalidDomain        = errors.New(InvalidDomain)
	ErrInvalidParam         = errors.New(InvalidParam)
	ErrInvalidSignupToken   = errors.New("INVALID_SIGNUP_TOKEN")
	ErrInvalidVersionNumber = errors.New("INVALID_VERSION_NUMBER")
	ErrInvalidReleaseDay    = errors.New("INVALID_RELEASE_DAY")
	ErrInvalidFormData      = errors.New("INVALID_FORM_DATA")
	ErrInvalidColumn        = errors.New("INVALID_COLUMN")
	ErrInvalidCustomer      = errors.New("INVALID_CUSTOMER")
	ErrEmptyPassword        = errors.New("EMPTY_PASSWORD")
	ErrInvalidAccount       = errors.New("INVALID_ACCOUNT")
	ErrPasswordMismatch     = errors.New(PasswordMismatch)
	ErrPasswordTooShort     = errors.New(PasswordTooShort)
	ErrPasswordTooLong      = errors.New("PASSWORD_TOO_LONG")
	ErrPermissionDenied     = errors.New(PermissionDenied)
	ErrDomainExists         = errors.New("DOMAIN_EXISTS")
)
