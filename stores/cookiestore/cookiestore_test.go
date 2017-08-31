package cookiestore

import (
	"bytes"
	"testing"
	"time"
)

var key = []byte("G_TdvPJ9T8C4p&A?Wr3YAUYW$*9vn4?t")

func TestMakeToken(t *testing.T) {
	c := New(key)

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
}

func TestTokenLength(t *testing.T) {
	c := New(key)

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
	c := New(key)

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
	c := New(key)

	b := []byte(`{data: "lorem ipsum"}`)
	c.MakeToken(b, time.Now().Add(1752000*time.Hour))

	// Valid token set to expire in 200 years
	b2, found, err := c.Find("52bFDXULZrJYHSBdA55s2d3ztM0q98DS8V1lAZaqa5uylDGhFg6Lk_JDoYt52AV4pCRnH2luvCGb2by5GFSpVcWMjRxTJAHEi4lbWBX8gx4")
	if err != nil {
		t.Fatal(err)
	}
	if found != true {
		t.Fatalf("got %v: expected %v", true, false)
	}
	if bytes.Equal(b, b2) == false {
		t.Fatalf("got %v: expected %v", b2, b)
	}

	// Tampered nonce
	_, found, err = c.Find("62bFDXULZrJYHSBdA55s2d3ztM0q98DS8V1lAZaqa5uylDGhFg6Lk_JDoYt52AV4pCRnH2luvCGb2by5GFSpVcWMjRxTJAHEi4lbWBX8gx4")
	if err != nil {
		t.Fatal(err)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", true, false)
	}

	// Tampered message
	_, found, err = c.Find("52bFDXULZrJYHSBdA55s2d3ztM0q98DS8V1lAZaqa5uylDGhFg6Lk_JDoYt52AV4pCRnH2luvCGb2by5GFSpVcWMjRxTJAHEi4lbWBX8g3")
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
}

func TestRotation(t *testing.T) {
	key1 := []byte("G_TdvPJ9T8C4p&A?Wr3YAUYW$*9vn4?t")
	key2 := []byte("cJxDdwM?yrRP6#h5^-9NSHRKm-dJbYqD")
	key3 := []byte("MHj$SQhjfnhN4J$eqvc@Mf?s29qxbMa_")

	// Generate a cookie token with original keyset
	c := New(key1)
	b := []byte(`{data: "lorem ipsum"}`)
	expiry := time.Now().Add(time.Minute)
	token, err := c.MakeToken(b, expiry)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new store with rotated keysets
	c = New(key2, key1)
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
	c = New(key3, key2, key1)
	_, found, err = c.Find(token)
	if err != nil {
		t.Fatal(err)
	}
	if found != true {
		t.Fatalf("got %v: expected %v", found, false)
	}

	// Rotate out original keyset completely
	c = New(key3, key2)
	_, found, err = c.Find(token)
	if err != nil {
		t.Fatal(err)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}
