package extract

import "testing"

func TestEnsureSessionExtractorsAreLast(t *testing.T) {
	cases := []struct {
		Sequence []RequestValueExtractor
		Expected bool
	}{
		{
			Sequence: []RequestValueExtractor{
				singleSessionValue("test"),
				multiSessionValue([]string{"test"}),
			},
			Expected: true,
		},
		{
			Sequence: []RequestValueExtractor{
				singleSessionValue("test"),
				singleHeader("test"),
				multiSessionValue([]string{"test"}),
			},
			Expected: false,
		},
		{
			Sequence: []RequestValueExtractor{
				singleCookie("test"),
				Join(
					singleSessionValue("test"),
					singleHeader("test"),
				),
				singleHeader("test"),
				multiSessionValue([]string{"test"}),
			},
			Expected: false,
		},
	}

	for i, c := range cases {
		if AreSessionExtractorsLast(c.Sequence...) != c.Expected {
			t.Logf("%+v", c.Sequence)
			t.Fatalf("failed session extractor detection on case: %d", i)
		}
	}
}
