package initq

import "testing"

func TestRunQItem(t *testing.T) {

	var rqi *runQItem

	rqi = newRunQItem("test", func() ReqResult { return Satisfied })

	if rqi.state != UnRun {
		t.Errorf("Expected UnRun on initialization")
	}

	rqi.run()

	if rqi.state != Satisfied {
		t.Errorf("Expected Satisfied after run")
	}

}
