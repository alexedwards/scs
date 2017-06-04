package cookiestore

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/databrary/scs/session"
)

var hmacKey = []byte("f71dc7e58abab014ddad2652475056f185164d262869c8931b239de52711ba87")
var blockKey = []byte("911182cec2f206986c8c82440adb7d17")

func TestNew(t *testing.T) {
	ks, err := NewKeyset(hmacKey, blockKey)
	if err != nil {
		t.Fatal(err)
	}
	c := New(ks)
	_, ok := interface{}(c).(session.Engine)
	if ok == false {
		t.Fatalf("got %v: expected %v", ok, true)
	}
}

func TestMakeToken(t *testing.T) {
	ks, err := NewKeyset(hmacKey, blockKey)
	if err != nil {
		t.Fatal(err)
	}
	c := New(ks)

	b := []byte(`{data: "lorem ipsum"}`)
	expiry := time.Now().Add(time.Minute)
	token, err := c.MakeToken(b, expiry)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Index([]byte(token), []byte("\000")) > 1 {
		t.Fatalf("got %v: expected no invalid bytes", []byte(token))
	}

	b2, found, err := c.Find(token)
	if err != nil {
		t.Fatal(err)
	}
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}
	if bytes.Equal(b, b2) != true {
		t.Fatalf("got %v: expected %v", bytes.Equal(b, b2), true)
	}

	// Check that the cookie is encrypted
	b3, err := base64.RawURLEncoding.DecodeString(strings.Split(token, "|")[0])
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(b, b3) != false {
		t.Fatalf("got %v: expected %v", bytes.Equal(b, b3), false)
	}
}

func TestMakeUnencryptedToken(t *testing.T) {
	ks, err := NewUnencryptedKeyset([]byte("f71dc7e58abab014ddad2652475056f185164d262869c8931b239de52711ba87"))
	if err != nil {
		t.Fatal(err)
	}
	c := New(ks)

	b := []byte(`{data: "lorem ipsum"}`)
	expiry := time.Now().Add(time.Minute)
	token, err := c.MakeToken(b, expiry)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Index([]byte(token), []byte("\000")) > 1 {
		t.Fatalf("got %v: expected no invalid bytes", []byte(token))
	}

	b2, found, err := c.Find(token)
	if err != nil {
		t.Fatal(err)
	}
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}
	if bytes.Equal(b, b2) != true {
		t.Fatalf("got %v: expected %v", bytes.Equal(b, b2), true)
	}

	// Check that the cookie is unencrpyted
	b3, err := base64.RawURLEncoding.DecodeString(strings.Split(token, "|")[0])
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(b, b3) != true {
		t.Fatalf("got %v: expected %v", bytes.Equal(b, b3), true)
	}
}

func TestTokenLength(t *testing.T) {
	ks, err := NewKeyset(hmacKey, blockKey)
	if err != nil {
		t.Fatal(err)
	}
	c := New(ks)

	b := []byte(`{data: "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Proin iaculis imperdiet ante, id maximus dolor blandit ut. Mauris semper, enim vel posuere vestibulum, velit diam faucibus lectus, ut feugiat ipsum erat sed nulla. Curabitur sed vestibulum dui, ac luctus metus. Aliquam in metus id dui gravida placerat ac vel ex. Cras tortor sem, laoreet et nunc iaculis, posuere pulvinar nunc. Etiam id dictum nunc, non viverra nulla. Duis facilisis nunc vel felis gravida condimentum. Etiam ultricies, sapien et euismod vehicula, arcu odio ullamcorper sapien, ac faucibus felis mauris in neque. Proin luctus nulla id suscipit tristique. Praesent vitae convallis turpis. Integer nibh metus, dictum sed urna nec, mollis feugiat sapien. Phasellus lobortis tortor ex, non maximus ante blandit ut. Nunc sed sapien luctus, gravida ante sit amet, vestibulum quam. Morbi vehicula euismod venenatis. Nulla pharetra scelerisque vehicula. Sed viverra massa eu scelerisque placerat. Sed eget risus a ligula sollicitudin scelerisque vitae vel quam. Quisque convallis sit amet ante in viverra. Quisque enim nulla, tempor vitae vestibulum vel, maximus quis libero. Vivamus commodo tempus justo, in vulputate enim sodales ut. Praesent in consectetur nibh, vitae interdum lorem. Vivamus blandit suscipit mauris eu iaculis. Donec ornare libero at lacus mattis tincidunt. Integer dictum purus id nunc malesuada, fermentum tempus ex lobortis. Cras felis nisi, commodo eget erat non, volutpat pharetra justo. Integer malesuada euismod facilisis. Maecenas volutpat risus sem, eget malesuada elit vehicula et. Aenean a arcu non nisi gravida consequat in vitae lectus. Mauris volutpat mi ac placerat iaculis. Duis interdum, elit a iaculis pretium, enim eros lacinia nunc, et pulvinar leo magna vitae purus. In volutpat egestas massa, et feugiat odio consequat et. Aliquam in varius augue, ut tempus mauris. In rutrum vehicula ullamcorper. Morbi tincidunt magna sed elit tristique suscipit. Donec sollicitudin mauris elementum lobortis malesuada. Mauris et cursus nisi. Donec congue, leo a tempus dapibus, lectus nisl varius mi, ut pulvinar felis ipsum et enim. Nunc nisi tortor, tincidunt sit amet dolor nec, sodales eleifend dolor. Fusce eget sodales urna, at pulvinar massa. Sed euismod sit amet mauris eu eleifend. Suspendisse volutpat mi quam, eget eleifend est condimentum sit amet. Praesent eget venenatis diam. Mauris hendrerit interdum bibendum. Curabitur mi est, consectetur a faucibus eu, lacinia vitae nunc. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Maecenas blandit augue ut dictum vulputate. Sed eu tristique dolor. Vivamus accumsan tincidunt fringilla. Nulla ultricies sem scelerisque metus rutrum, id tempor mauris hendrerit. Mauris at leo eget elit ultrices pulvinar vitae nec lacus. Suspendisse id posuere mi. Aenean et varius ligula, ac congue diam. Duis eget lectus enim. Proin accumsan id arcu sed sodales. Suspendisse sed arcu mattis, tristique urna in, viverra sapien. Sed quis metus venenatis leo aliquam vestibulum in a nunc. Quisque interdum vitae sem a tempor. Cras tristique egestas metus, sed accumsan neque ullamcorper et. Vestibulum imperdiet consequat lorem, sed dignissim purus hendrerit in. Fusce sollicitudin lobortis tristique. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam nisl urna, tincidunt ac purus vel, sagittis condimentum nibh. Aenean a felis ac nisl sollicitudin pharetra in sit amet metus. Nam placerat ante a nibh cursus blandit. Fusce mattis sed massa a tincidunt. Aenean sit amet justo tristique, dignissim lacus ac, ullamcorper nunc. Aliquam erat volutpat. Integer in ex elit. Sed sit amet felis dictum, posuere tellus nec, viverra ante. Ut vestibulum diam in sapien rhoncus, non condimentum metus vulputate. Vestibulum ornare nibh ipsum, sed accumsan leo cursus sed. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Vivamus et nunc at quam porta fringilla. Vivamus lacinia id ex vel aliquam. Quisque elementum nibh at enim tincidunt mattis. Morbi euismod mi eget venenatis finibus. Aliquam ut rutrum enim. Interdum et malesuada fames ac ante ipsum primis in faucibus. Vestibulum dapibus lorem lorem, vel venenatis nisi bibendum id. Aenean laoreet ipsum enim, congue pharetra nisi congue ac. Donec at odio vel leo aliquet gravida non quis turpis. Nullam faucibus mollis orci sed feugiat. Suspendisse vel felis cursus, rutrum felis a, commodo enim. Pellentesque eleifend felis dui, vel vulputate odio vehicula in. Mauris sed facilisis erat. Donec ac molestie justo, a lobortis dolor. Quisque dignissim metus sit amet neque venenatis, in dapibus risus blandit. Etiam ullamcorper enim ultricies lacus sollicitudin, a vulputate libero cursus. Interdum et malesuada fames ac ante ipsum primis in faucibus. Nulla sodales tincidunt sapien accumsan scelerisque. Duis sed pellentesque nunc, ac gravida nisl. Phasellus ullamcorper ante sed eros sagittis pretium id id est. Maecenas sed cras amet."}`)
	expiry := time.Now().Add(time.Minute)
	token, err := c.MakeToken(b, expiry)
	if err != errTokenTooLong {
		t.Fatalf("got %v: expected %q", err, errTokenTooLong)
	}
	if token != "" {
		t.Fatalf("got %s: expected %q", token, "")
	}
}

func TestTokenExpiry(t *testing.T) {
	ks, err := NewKeyset(hmacKey, blockKey)
	if err != nil {
		t.Fatal(err)
	}
	c := New(ks)

	b := []byte(`{data: "lorem ipsum"}`)
	expiry := time.Now().Add(100 * time.Millisecond)
	token, err := c.MakeToken(b, expiry)
	if err != nil {
		t.Fatal(err)
	}

	_, found, err := c.Find(token)
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if found != true {
		t.Fatalf("got %v: expected %v", true, false)
	}

	time.Sleep(100 * time.Millisecond)
	_, found, err = c.Find(token)
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", true, false)
	}
}

func TestTamperedToken(t *testing.T) {
	ks, err := NewKeyset(hmacKey, blockKey)
	if err != nil {
		t.Fatal(err)
	}
	c := New(ks)

	b := []byte(`{data: "lorem ipsum"}`)

	// Valid token set to expire in 200 years
	b2, found, err := c.Find("5BeOzBo2_NnAeGK57NTberipKRR0uVBqWS7gWr4jz-4kUd5PeQ|7778974132340921404|ZGDtOlRNCdwvuGp5kz5xFCDsFVQa_tTo-0Kbvs2G8uA")
	if err != nil {
		t.Fatal(err)
	}
	if found != true {
		t.Fatalf("got %v: expected %v", true, false)
	}
	if bytes.Equal(b, b2) == false {
		t.Fatalf("got %v: expected %v", b2, b)
	}

	// Tampered payload
	_, found, err = c.Find("5BeOzBo2_NnAeGK57NTberipKRR0uVBqWS7gWr4jz-4kUd5PeR|7778974132340921404|ZGDtOlRNCdwvuGp5kz5xFCDsFVQa_tTo-0Kbvs2G8uA")
	if err != nil {
		t.Fatal(err)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", true, false)
	}

	// Tampered timestamp
	_, found, err = c.Find("5BeOzBo2_NnAeGK57NTberipKRR0uVBqWS7gWr4jz-4kUd5PeQ|7778974132340921405|ZGDtOlRNCdwvuGp5kz5xFCDsFVQa_tTo-0Kbvs2G8uA")
	if err != nil {
		t.Fatal(err)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", true, false)
	}

	// Tampered signature
	_, found, err = c.Find("5BeOzBo2_NnAeGK57NTberipKRR0uVBqWS7gWr4jz-4kUd5PeQ|7778974132340921404|ZGDtOlRNCdwvuGp5kz5xFCDsFVQa_tTo-0Kbvs2G8uB")
	if err != nil {
		t.Fatal(err)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", true, false)
	}

	// Malformed value
	_, found, err = c.Find("5BeOzBo2_NnAeGK57NTberipKRR0uVBqWS7gWr4jz-4kUd5PeQ7778974132340921404ZGDtOlRNCdwvuGp5kz5xFCDsFVQa_tTo-0Kbvs2G8uA")
	if err != nil {
		t.Fatal(err)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", true, false)
	}

	// Empty value
	_, found, err = c.Find("")
	if err != nil {
		t.Fatal(err)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", true, false)
	}

	// Generated using a different cookie name
	_, found, err = c.Find("rUY6hQgHoX1LuF27qFCcAuER5ITPuBA-UFrL41rvVRVCLVAVKg|7778974194317574335|RiWG724uUjYzlKsdHjjBDJGEhNJy22oGA2SqMKrCdlQ")
	if err != nil {
		t.Fatal(err)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", true, false)
	}
}

func TestRotation(t *testing.T) {
	ks1, err := NewKeyset(hmacKey, blockKey)
	if err != nil {
		t.Fatal(err)
	}
	ks2, err := NewKeyset([]byte("16bd76c6372363cd9af46f5619cc406776210b6164c48fd1200119d4cfc359e6"), []byte("5f8b7a8efac2a900a0c6be609b2e0241"))
	if err != nil {
		t.Fatal(err)
	}
	ks3, err := NewKeyset([]byte("0c03fa487baa82dda09c4f12c7238370c58112a135318a6e3d4a4724a95cd2e0"), []byte("46ee77bfb95a765dfefca83bf53d5914"))
	if err != nil {
		t.Fatal(err)
	}

	// Generate a cookie token with original keyset
	c := New(ks1)
	b := []byte(`{data: "lorem ipsum"}`)
	expiry := time.Now().Add(time.Minute)
	token, err := c.MakeToken(b, expiry)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new engine with rotated keysets
	c = New(ks2, ks1)
	b2, found, err := c.Find(token)
	if err != nil {
		t.Fatal(err)
	}
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}
	if bytes.Equal(b, b2) != true {
		t.Fatalf("got %v: expected %v", bytes.Equal(b, b2), true)
	}

	// Rotate out original keyset completely
	c = New(ks3, ks2)
	_, found, err = c.Find(token)
	if err != nil {
		t.Fatal(err)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}
