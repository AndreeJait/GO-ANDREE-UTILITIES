package andreerror

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	perrors "github.com/pkg/errors"
)

type (
	ErrorStandard interface {
		Error() string
		Wrap(string)
		AppendError(string)
		GetCode() string
		GetMessage() string
		GetErrors() []string
		GetHTTPStatus() int
		IsErrorOf(string) bool
	}

	tiketError struct {
		Errors     []string
		Code       string
		Message    string
		HTTPStatus int
	}

	printer interface {
		Printf(message string, args ...interface{})
		Errorf(message string, args ...interface{})
	}

	tracer interface {
		StackTrace() perrors.StackTrace
	}

	ResponseCode struct {
		code       string
		message    string
		httpStatus int
	}
)

const (
	SUCCESS                    = "SUCCESS"
	SYSTEM_ERROR               = "SYSTEM_ERROR"
	DUPLICATE_DATA             = "DUPLICATE_DATA"
	DATA_NOT_EXIST             = "DATA_NOT_EXIST"
	BIND_ERROR                 = "BIND_ERROR"
	RUNTIME_ERROR              = "RUNTIME_ERROR"
	DATE_NOT_VALID             = "DATE_NOT_VALID"
	VENDOR_SHUTDOWN            = "VENDOR_SHUTDOWN"
	METHOD_ARGUMENTS_NOT_VALID = "METHOD_ARGUMENTS_NOT_VALID"
	TOO_MANY_REQUEST           = "TOO_MANY_REQUEST"
	BAD_REQUEST                = "BAD_REQUEST"
	UNAUTHORIZE                = "UNAUTHORIZE"
	SOLDOUT                    = "SOLDOUT"
)

var (
	responseCodes = map[string]ResponseCode{
		SUCCESS: ResponseCode{
			code:       SUCCESS,
			message:    "Success",
			httpStatus: http.StatusOK,
		},
		SYSTEM_ERROR: ResponseCode{
			code:       SYSTEM_ERROR,
			message:    "Contact our team",
			httpStatus: http.StatusInternalServerError,
		},
		DUPLICATE_DATA: ResponseCode{
			code:       DUPLICATE_DATA,
			message:    "Duplicate data",
			httpStatus: http.StatusOK,
		},
		DATA_NOT_EXIST: ResponseCode{
			code:       DATA_NOT_EXIST,
			message:    "No data exist",
			httpStatus: http.StatusOK,
		},
		BIND_ERROR: ResponseCode{
			code:       BIND_ERROR,
			message:    "Please fill in mandatory parameter",
			httpStatus: http.StatusOK,
		},
		RUNTIME_ERROR: ResponseCode{
			code:       RUNTIME_ERROR,
			message:    "Runtime Error",
			httpStatus: http.StatusInternalServerError,
		},
		DATE_NOT_VALID: ResponseCode{
			code:       DATE_NOT_VALID,
			message:    "Date not valid",
			httpStatus: http.StatusOK,
		},
		VENDOR_SHUTDOWN: ResponseCode{
			code:       VENDOR_SHUTDOWN,
			message:    "Vendor is Shutdown",
			httpStatus: http.StatusOK,
		},
		METHOD_ARGUMENTS_NOT_VALID: ResponseCode{
			code:       METHOD_ARGUMENTS_NOT_VALID,
			message:    "Method argument is not valid",
			httpStatus: http.StatusOK,
		},
		TOO_MANY_REQUEST: ResponseCode{
			code:       TOO_MANY_REQUEST,
			message:    "Invalid data",
			httpStatus: http.StatusOK,
		},
		BAD_REQUEST: ResponseCode{
			code:       BAD_REQUEST,
			message:    "Bad request",
			httpStatus: http.StatusBadRequest,
		},
		UNAUTHORIZE: ResponseCode{
			code:       UNAUTHORIZE,
			message:    "Unauthorize",
			httpStatus: http.StatusUnauthorized,
		},
		SOLDOUT: ResponseCode{
			code:       SOLDOUT,
			message:    "Soldout",
			httpStatus: http.StatusOK,
		},
	}
)

func (e tiketError) Error() string {
	err := e.Errors
	if len(err) > 0 {
		return err[0]
	} else {
		return ""
	}
}

func (e tiketError) Wrap(errMessage string) {
	e.Errors[0] = fmt.Sprintf("%s: %s", errMessage, e.Errors[0])
}

func (e *tiketError) AppendError(errMessage string) {
	e.Errors = append(e.Errors, errMessage)
}

func (e tiketError) GetCode() string {
	return e.Code
}

func (e tiketError) GetMessage() string {
	return e.Message
}

func (e tiketError) GetErrors() []string {
	return e.Errors
}

func (e tiketError) GetHTTPStatus() int {
	return e.HTTPStatus
}

func (e tiketError) IsErrorOf(code string) bool {
	if strings.ToLower(e.Code) == strings.ToLower(code) {
		return true
	}
	return false
}

func New(code string, err error) ErrorStandard {
	if code == SUCCESS {
		errCode := responseCodes[SUCCESS].code
		errMessage := responseCodes[SUCCESS].message
		errHTTPStatus := responseCodes[SUCCESS].httpStatus

		return &tiketError{
			Errors:     []string{},
			Code:       errCode,
			Message:    errMessage,
			HTTPStatus: errHTTPStatus,
		}
	}

	errCode := responseCodes[SYSTEM_ERROR].code
	errMessage := responseCodes[SYSTEM_ERROR].message
	errHTTPStatus := responseCodes[SYSTEM_ERROR].httpStatus
	errorList := make([]string, 0)

	if tiketError, ok := responseCodes[code]; ok {
		errCode = tiketError.code
		errMessage = tiketError.message
		errHTTPStatus = tiketError.httpStatus

		if err != nil {
			errorList = append(errorList, err.Error())
		}
	}

	return &tiketError{
		Errors:     errorList,
		Code:       errCode,
		Message:    errMessage,
		HTTPStatus: errHTTPStatus,
	}
}

func NewError(code string, err error, status int) ErrorStandard {
	return &tiketError{
		Errors:     []string{err.Error()},
		Code:       code,
		Message:    err.Error(),
		HTTPStatus: status,
	}
}

func Wrap(err error, errMessage string) error {
	if err == nil {
		err = errors.New(errMessage)
		return err
	}

	if s, ok := err.(ErrorStandard); ok {
		s.Wrap(errMessage)
		return s
	} else {
		errTemp := errors.New(fmt.Sprintf("%s: %s", errMessage, err.Error()))
		return errTemp
	}
}

func AppendError(err error, errMessage string) error {
	if s, ok := err.(ErrorStandard); ok {
		s.AppendError(errMessage)
		return s
	}
	return err
}

func IsErrorOf(err error, code string) bool {
	if s, ok := err.(ErrorStandard); ok {
		return s.IsErrorOf(code)
	}
	return false
}

func PrintStackTrace(err error, p printer) {
	var (
		sb strings.Builder
	)

	if err, ok := err.(tracer); ok {
		for _, f := range err.StackTrace() {
			sb.WriteString(fmt.Sprintf("%+s:%d\n", f, f))
		}
	} else {
		sb.WriteString(fmt.Sprintf("%s", err))
	}

	p.Printf("%s", sb.String())
}
