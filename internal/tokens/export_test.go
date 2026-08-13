package tokens

// LoadVocabForTest exposes the embedded vocabulary to the package's tests, so a
// test can assert facts about it — the longest token behind MaxTokenBytes, for
// one — without making the loader part of the package's API.
func LoadVocabForTest() (map[string]int, error) {
	return offlineLoader{}.LoadTiktokenBpe("o200k_base")
}
