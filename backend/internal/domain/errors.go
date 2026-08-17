package domain

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("conflict")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrInvalidInput        = errors.New("invalid input")
	ErrInvalidTransition   = errors.New("invalid status transition")
	ErrCapacityFull        = errors.New("task capacity is full")
	ErrNotEligible         = errors.New("volunteer is not eligible for this task")
	ErrNotApproved         = errors.New("volunteer is not approved")
	ErrAlreadyAssigned     = errors.New("volunteer already assigned to this task")
	ErrDocumentRequired    = errors.New("تصویر کارت ملی الزامی است")
	ErrFileTooLarge        = errors.New("حجم فایل بیش از ۵ مگابایت است")
	ErrInvalidFileType     = errors.New("فقط JPG، PNG یا PDF مجاز است")
	ErrMissionExpired      = errors.New("مهلت ماموریت گذشته است")
	ErrCertificateNotReady = errors.New("گواهی هنوز قابل صدور نیست")
)

type ValidationError struct {
	Msg string
}

func (e ValidationError) Error() string { return e.Msg }

func (e ValidationError) Is(target error) bool { return target == ErrInvalidInput }

func Invalid(msg string) error { return ValidationError{Msg: msg} }
