package resource

import "testing"

func TestNameFromPathUsesConcreteAdminResourceSegment(t *testing.T) {
	tests := []struct {
		name, path, routeParam, want string
	}{
		{name: "static list", path: "/api/v1/admin/folders", routeParam: "{resource}", want: "folders"},
		{name: "nested resource action", path: "/api/v1/admin/albums/actions/archive", routeParam: "resource", want: "albums"},
		{name: "route parameter fallback", path: "/api/v1/admin/{resource}", routeParam: "advertising", want: "advertising"},
		{name: "unknown route", path: "/api/v1/health", routeParam: "", want: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := NameFromPath(test.path, test.routeParam); got != test.want {
				t.Fatalf("NameFromPath(%q, %q) = %q, want %q", test.path, test.routeParam, got, test.want)
			}
		})
	}
}
