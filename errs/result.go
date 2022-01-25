package errs

type Result[T any] struct {
	Res chan T
	Err error
}

func NewResult[T any]() Result[T] {
	return Result[T]{
		Res: make(chan T),
	}
}

func (r *Result[_]) End(err error) {

}
