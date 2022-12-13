package db

func contains(s string, a []string) bool {
	for _, v := range a {
		if s == v {
			return true
		}
	}
	return false
}

func IsPointType(v string) bool {
	return contains(v, []string{
		"stars",
		"rank7",
		"rank14",
		"score",
		"point",
		"unlimited",
	})
}

func IsParagraphType(v string) bool {
	return contains(v, []string{
		"text",
		"twitterLink",
		"imageLink",
	})
}
