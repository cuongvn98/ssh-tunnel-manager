package tunnel

type Hook struct {
	fns func(int, string)
}

func (h *Hook) fire(status int, extra string) {
	if h.fns != nil {
		h.fns(status, extra)
	}
}

func (h *Hook) Subscribe(openFn func(int, string)) {
	h.fns = openFn
}
func (h *Hook) UnSubscribe() {
	h.fns = nil
}
