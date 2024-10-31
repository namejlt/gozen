package gozen

import (
	"encoding/base64"
	"testing"
)

func TestAesCbcDecrypt(t *testing.T) {
	r := `ce8JIVoyveyYlGRm/+I04/cRZgYVZFiJXwVmjzuJ6VQSfATy+tLCC5WXzcnB8Pw9XopzROwp9V2UCVak+h3nbTYgXG0CoF+ObzAcU/ocrpaarH+/rW8UShF/4yXYRDGh+VQ7YUkftvpbow+2fOgqYqn//4FoTmB24gdNJC0D+TSmxQ4hMl1UGSygVcGvco13Jds2vyIt4tXslALj7/UqmVeqmnh4+DOt0JBWpRLzlKs03XzgMa/RCVOpgWvEXAeRpYgEeO+cF9EH7GMqo8ddqW338WaJBKrEjk8tRlpcci0=`

	r1, err := base64.StdEncoding.DecodeString(r)

	if err != nil {
		t.Error(err)
	}

	key := "ae3adcec76947f1701408270abd856c4"
	iv := "c558Gq0YQK2QUlMc"
	s, err := AesCbcDecrypt(key, iv, r1)

	t.Log(string(s))
	t.Log(err)

}

func TestAesCbcEncrypt(t *testing.T) {

	key := "ae3adcec76947f1701408270abd856c4"
	iv := "c558Gq0YQK2QUlMc"

	a := "12313哈哈哈哈"

	s, err := AesCbcEncrypt(key, iv, []byte(a))
	r, err := AesCbcDecrypt(key, iv, s)

	t.Log(string(r))
	t.Log(err)

}

func TestUtilCryptoMd5Lower(t *testing.T) {
	a := "page-decorate:squ_30000068__"

	r := UtilCryptoMd5Lower(a)
	t.Log(r)
}
