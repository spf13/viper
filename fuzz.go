package viper

func FuzzGet(data []byte) int {
	_ = Get(string(data))
	return 1
}
