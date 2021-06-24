package tunnel

type Hook struct {
	fns func(int)
}

func (h *Hook) fire(status int) {
	if h.fns != nil {
		h.fns(status)
	}
}

func (h *Hook) Subscribe(openFn func(int)) {
	h.fns = openFn
}
func (h *Hook) UnSubscribe() {
	h.fns = nil
}
