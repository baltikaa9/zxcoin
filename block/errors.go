package block

type InvalidMerkleRootError struct{}

func (e InvalidMerkleRootError) Error() string {
	return "неверный хеш корня дерева Меркла"
}
