package blockchain

type InvalidPrevHashError struct{}

func (e InvalidPrevHashError) Error() string {
	return "неверный предыдущий блок"
}
