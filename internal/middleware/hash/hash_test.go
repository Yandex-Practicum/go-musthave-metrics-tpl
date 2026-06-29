package hash

import "testing"

func TestComputeHMAC_Deterministic(t *testing.T) {
	data := []byte("hello world")
	key := "secret"

	h1 := ComputeHMAC(data, key)
	h2 := ComputeHMAC(data, key)
	if h1 != h2 {
		t.Fatalf("HMAC должен быть детерминирован: %s != %s", h1, h2)
	}
	if h1 == "" {
		t.Fatal("HMAC не должен быть пустым")
	}
}

func TestComputeHMAC_DifferentKey(t *testing.T) {
	data := []byte("payload")
	if ComputeHMAC(data, "key1") == ComputeHMAC(data, "key2") {
		t.Fatal("разные ключи должны давать разный HMAC")
	}
}

func TestVerify(t *testing.T) {
	data := []byte("audit event")
	key := "topsecret"
	sum := ComputeHMAC(data, key)

	if !Verify(data, key, sum) {
		t.Fatal("Verify должен принимать корректную подпись")
	}
	if Verify(data, key, "deadbeef") {
		t.Fatal("Verify должен отклонять неверную подпись")
	}
	if Verify([]byte("tampered"), key, sum) {
		t.Fatal("Verify должен отклонять изменённые данные")
	}
}
