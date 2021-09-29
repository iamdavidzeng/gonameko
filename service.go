package gonameko

type Service interface {
	GetName() string
	HealthCheck() map[string]string
}

type BaseService struct {
	Name string
}

func (bs *BaseService) GetName() string {
	return bs.Name
}

func (bs *BaseService) HealthCheck() (map[string]string, error) {
	return map[string]string{"git_sha": "dev"}, nil
}
