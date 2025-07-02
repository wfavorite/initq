.PHONY: cover test default clean

# Really no *need* for a make file - except that i like the ability to
# "make cover" to see the coverage report.
#
# This is a module, so there are no 'external' tests here like might
# exist for tools.

# default is the target if none is specified.
default: test

# test simply runs go test verbosely. 
test:
	@go test -v

# cover builds a Go coverage report. This totally could be broken into
# multiple targets with the coverage.out as one of the named targets.
# But no.
cover:
	@printf "Building coverage report.\n"
	@go test -coverprofile=coverage.out
	@go tool cover -html=coverage.out

# clean (only) removes the coverage.out file (no artifacts).
clean:
	@printf "Cleaning test artifacts..."
	@rm -f coverage.out
	@printf "Done.\n"

