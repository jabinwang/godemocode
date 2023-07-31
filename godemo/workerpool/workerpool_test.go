package workerpool

import "testing"

func TestExample(t *testing.T) {
	wp := newPool(2)
	reqs := []string{"alpha", "beta", "gamma", "delta", "epsilon"}
	rspChan := make(chan string, len(reqs))
	for _, r := range reqs {
		r := r
		wp.Submit(func() {
			rspChan <- r
		})
	}
	wp.StopWait()
	close(rspChan)
	rspSet := map[string]struct{}{}
	for rsp := range rspChan {
		t.Logf("rsp %v", rsp)
		rspSet[rsp] = struct{}{}
	}
	if len(rspSet) < len(reqs) {
		t.Fatalf("Did not handle all request")
	}
	for _, req := range reqs {
		if _, ok := rspSet[req]; !ok {
			t.Fatal("missing expected values: ", req)
		}
	}
}