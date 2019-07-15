SHELL=/bin/bash -o pipefail

# Formats the code
.PHONY: format
format:
		goreturns -w -local github.com/ory $$(listx .)
