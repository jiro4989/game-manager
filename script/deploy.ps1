git tag $args[0]
$env:GITHUB_TOKEN = cat ".\res\token.txt"
goreleaser --rm-dist
go install
