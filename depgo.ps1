git tag $args[0]
$env:GITHUB_TOKEN = cat "token.txt"
goreleaser --rm-dist
