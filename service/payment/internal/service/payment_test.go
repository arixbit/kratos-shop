package service

import (
	"strings"
	"testing"
)

func TestGeneratePaymentNo(t *testing.T) {
	no := generatePaymentNo()
	if !strings.HasPrefix(no, "PAY-") {
		t.Fatalf("payment no should start with PAY-: %s", no)
	}
	if len(no) != len("PAY-")+14+6 {
		t.Fatalf("unexpected payment no length: %s", no)
	}
	if no == generatePaymentNo() {
		t.Fatal("payment no should be unique")
	}
}
