package repositories

import "errors"

var (
	ErrCategoryNotFound          = errors.New("category not found")
	ErrCategorySlugAlreadyExists = errors.New("category slug already exists")
)
