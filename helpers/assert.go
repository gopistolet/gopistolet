package helpers

func Assert(condition bool, message string) {
	if !condition {
		panic("Asser failed")
	}
}