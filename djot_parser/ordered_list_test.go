package djot_parser

import (
	"strings"
	"testing"
)

func TestRomanAndAlphabeticListMarkers(t *testing.T) {
	type list struct {
		kind, start string
		items       int
	}
	for _, tc := range []struct {
		markers string
		want    []list
	}{
		{"i.", []list{{"i", "", 1}}},
		{"i. j.", []list{{"a", "9", 2}}},
		{"i. ii.", []list{{"i", "", 2}}},
		{"(IV) (V)", []list{{"I", "4", 2}}},
		{"C. D. A.", []list{{"A", "3", 3}}},
		{"l. m.", []list{{"i", "50", 2}}},
		{"i) j. ii.", []list{{"i", "", 1}, {"a", "10", 1}, {"i", "2", 1}}},
		{"a. i. vi.", []list{{"a", "", 2}, {"i", "6", 1}}},
	} {
		t.Run(tc.markers, func(t *testing.T) {
			var doc strings.Builder
			for _, marker := range strings.Fields(tc.markers) {
				doc.WriteString(marker + " item\n")
			}
			got := BuildDjotAst([]byte(doc.String()))[0].Children
			if len(got) != len(tc.want) {
				t.Fatalf("got %d lists, want %d", len(got), len(tc.want))
			}
			for i, node := range got {
				actual := list{node.Attributes.Get("type"), node.Attributes.Get("start"), len(node.Children)}
				if node.Type != OrderedListNode || actual != tc.want[i] {
					t.Errorf("list %d: got %s %+v, want ordered list %+v", i, node.Type, actual, tc.want[i])
				}
			}
		})
	}
}
