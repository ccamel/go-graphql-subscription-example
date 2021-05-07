package static

import "embed"

//go:embed graphiql/* graphql/*
var fs embed.FS // nolint:gochecknoglobals

func ReadFileStringMust(filename string) string {
	data, err := fs.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	return string(data)
}
