package gonameko

type Base interface {
	GetName() string
	HealthCheck() map[string]string
}

// Base Service provide name and healthcheck method.
// Use composition to support more methods from another struct.
type Service struct {
	Name string
	*FooService
}

func (bs *Service) GetName() string {
	return bs.Name
}

func (bs *Service) HealthCheck() (map[string]string, error) {
	return map[string]string{"git_sha": "dev"}, nil
}

// Concrete service: Foo
type FooService struct{}

func (f *FooService) GetFoo() (string, error) {
	return "foo", nil
}
