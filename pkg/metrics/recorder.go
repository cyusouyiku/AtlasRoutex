package metrics

import "sync/atomic"

type Recorder struct {
requests int64
errors   int64
}

func NewRecorder() *Recorder { return &Recorder{} }
func (r *Recorder) IncRequests() { atomic.AddInt64(&r.requests, 1) }
func (r *Recorder) IncErrors()   { atomic.AddInt64(&r.errors, 1) }
func (r *Recorder) Snapshot() (int64, int64) {
return atomic.LoadInt64(&r.requests), atomic.LoadInt64(&r.errors)
}
