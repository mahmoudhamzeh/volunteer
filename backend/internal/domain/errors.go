package domain

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("conflict")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrInvalidInput        = errors.New("invalid input")
	ErrInvalidTransition   = errors.New("این وضعیت قابل تغییر نیست")
	ErrCapacityFull        = errors.New("ظرفیت این فعالیت تکمیل است")
	ErrNotEligible         = errors.New("برای این فعالیت واجد شرایط نیستید")
	ErrNotApproved         = errors.New("پروفایل داوطلب هنوز تایید نشده است")
	ErrAlreadyAssigned     = errors.New("برای این فعالیت قبلاً درخواست داده‌اید")
	ErrDocumentRequired    = errors.New("تصویر کارت ملی الزامی است")
	ErrFileTooLarge        = errors.New("حجم فایل بیش از ۵ مگابایت است")
	ErrInvalidFileType     = errors.New("فقط JPG، PNG یا PDF مجاز است")
	ErrMissionExpired      = errors.New("مهلت ماموریت گذشته است")
	ErrMissionNotVerified  = errors.New("انجام این ماموریت هنوز تأیید نشده است")
	ErrCertificateNotReady = errors.New("گواهی هنوز قابل صدور نیست")
	ErrBusy                = errors.New("سیستم مشغول است؛ کمی بعد دوباره تلاش کنید")
)

type ValidationError struct {
	Msg string
}

func (e ValidationError) Error() string { return e.Msg }

func (e ValidationError) Is(target error) bool { return target == ErrInvalidInput }

func Invalid(msg string) error { return ValidationError{Msg: msg} }
