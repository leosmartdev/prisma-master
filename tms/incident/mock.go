package incident

import "strconv"

var (
	mockIdCreatorCounter int
)

type mockIdCreator struct {
	IdCreator
}

func (mockIdCreator) Next(prefixer IdPrefixer) string {
	mockIdCreatorCounter++
	return prefixer.Prefix() + strconv.Itoa(mockIdCreatorCounter)
}

func mockIdCreatorInstance() IdCreator {
	return new(mockIdCreator)
}
