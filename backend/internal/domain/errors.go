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
	ErrDocumentRequired    = errors.New("required documents are missing")
	ErrFileTooLarge        = errors.New("file exceeds size limit")
	ErrInvalidFileType     = errors.New("invalid file type")
	ErrMissionExpired      = errors.New("mission deadline has passed")
	ErrCertificateNotReady = errors.New("certificate cannot be issued yet")
	ErrBusy                = errors.New("resource is busy, retry shortly")
)
