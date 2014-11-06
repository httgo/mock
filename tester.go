package mock

// tester is an interface mapped to several methods in the testing package
// This is used to mock the testing package for internal tests
type tester interface {
	Error(...interface{})
	Errorf(string, ...interface{})
	Fatal(...interface{})
	Fatalf(string, ...interface{})
}
