package internal

import "errors"

var (
	ErrNotFound = errors.New("not found")
)

var (
	InternalErrorText  = "На сервере произошла ошибка, попробуйте позже"
	UserNotFoundText   = "Вы не найдены среди пользователей. Сначала зарегистрируйтесь при помощи команды /reg"
	DomainNotFoundText = "Домена нет среди отслеживаемых"
)
