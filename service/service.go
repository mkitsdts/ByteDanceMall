package service

func NewCartService() (*CartService, error) {
	s := &CartService{}
	return s, nil
}
