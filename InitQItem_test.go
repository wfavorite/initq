package initq

import "testing"

func TestInitQItem(t *testing.T) {

	var rqi *initQItem

	rqi = newInitQItem("test", func() ReqResult { return Satisfied })

	if rqi.state != UnRun {
		t.Errorf("Expected UnRun on initialization")
	}

	rqi.run()

	if rqi.state != Satisfied {
		t.Errorf("Expected Satisfied after run")
	}

}
