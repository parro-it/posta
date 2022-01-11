package plex

type addOut struct {
	out output
}

type removeOut struct {
	out any
}

type outOf[T any] chan T

func (l outOf[T]) Post(a any) {
	if aa, isReqType := a.(T); isReqType {
		l <- aa
	}
}
func (l outOf[T]) Equal(a any) bool {
	if aa, isReqType := a.(chan T); isReqType {
		return aa == l
	}
	return false
}
func (l outOf[T]) Close() {
	close(l)
}

type outOf2[T1 any, T2 any] chan any

func (l outOf2[T1, T2]) Post(a any) {
	if aa, isReqType := a.(T1); isReqType {
		l <- aa
		return
	}
	if aa, isReqType := a.(T2); isReqType {
		l <- aa
		return
	}
}

func (l outOf2[T1, T2]) Close() {
	close(l)
}
func (l outOf2[T1, T2]) Equal(a any) bool {
	if aa, isReqType := a.(chan any); isReqType {
		return aa == l
	}
	return false
}

type outOf3[T1 any, T2 any, T3 any] chan any

func (l outOf3[T1, T2, T3]) Equal(a any) bool {
	if aa, isReqType := a.(chan any); isReqType {
		return aa == l
	}
	return false
}

func (l outOf3[T1, T2, T3]) Post(a any) {
	if aa, isReqType := a.(T1); isReqType {
		l <- aa
		return
	}
	if aa, isReqType := a.(T2); isReqType {
		l <- aa
		return
	}
	if aa, isReqType := a.(T3); isReqType {
		l <- aa
		return
	}
}

func (l outOf3[T1, T2, T3]) Close() {
	close(l)
}
