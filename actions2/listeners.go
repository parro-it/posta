package actions2

type addListener struct {
	l listener
}

type removeListener struct {
	l listener
}

type listenerOf[T any] chan T

func (l listenerOf[T]) Post(a Action) {
	if aa, isReqType := a.(T); isReqType {
		l <- aa
	}
}

func (l listenerOf[T]) Close() {
	close(l)
}

type listenerOf2[T1 any, T2 any] chan Action

func (l listenerOf2[T1, T2]) Post(a Action) {
	if aa, isReqType := a.(T1); isReqType {
		l <- aa
	}
	if aa, isReqType := a.(T2); isReqType {
		l <- aa
	}
}

func (l listenerOf2[T1, T2]) Close() {
	close(l)
}

type listenerOf3[T1 any, T2 any, T3 any] chan Action

func (l listenerOf3[T1, T2, T3]) Post(a Action) {
	if aa, isReqType := a.(T1); isReqType {
		l <- aa
	}
	if aa, isReqType := a.(T2); isReqType {
		l <- aa
	}

	if aa, isReqType := a.(T3); isReqType {
		l <- aa
	}
}

func (l listenerOf3[T1, T2, T3]) Close() {
	close(l)
}
