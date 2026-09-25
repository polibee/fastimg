package media

import "testing"

func TestLinkFormatsExposeAbsolutePublicURLs(t *testing.T) {
	links := LinkFormatsFromVariants("summer & sky.png", map[string]string{
		"original":  "https://img.example.com/i/42?variant=original&signature=abc",
		"thumbnail": "https://img.example.com/i/42?variant=thumbnail&signature=def",
		"medium":    "https://img.example.com/i/42?variant=medium&signature=ghi",
	})
	if links["original"] != "https://img.example.com/i/42?variant=original&signature=abc" {
		t.Fatalf("unexpected original link: %q", links["original"])
	}
	if links["url"] != links["original"] {
		t.Fatalf("url must point to the original variant: %q", links["url"])
	}
	if links["markdown"] != "![summer & sky.png]("+links["url"]+")" {
		t.Fatalf("unexpected markdown link: %q", links["markdown"])
	}
	if links["html"] != `<img src="`+links["url"]+`" alt="summer &amp; sky.png">` {
		t.Fatalf("unexpected HTML link: %q", links["html"])
	}
	if links["bbcode"] != "[img]"+links["url"]+"[/img]" {
		t.Fatalf("unexpected BBCode link: %q", links["bbcode"])
	}
}

func TestAbsolutePublicURLJoinsAppURL(t *testing.T) {
	if got := absoluteURL("https://img.example.com/", "/i/42"); got != "https://img.example.com/i/42" {
		t.Fatalf("absoluteURL() = %q", got)
	}
}
