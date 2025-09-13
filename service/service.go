package service

func NewCartService() (*CartService, error) {
	s := &CartService{}
	loopUpdateProductInfo()
	return s, nil
}
