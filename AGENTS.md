
Please do not change the .golangci.yml file to suppress problems, fix them instead in the code

Use the following targets to check your code

    make build
    make test
    make vet
    make lint

A task is only considered complete when the following succeeds

    make build test vet lint

## Writing tests

- Build unit tests to prove that the code works 
- Use testify/assert package for checking test conditions 
- Do not write tests that only check access to struct fields work, that is a given
- When checking errors use require with an informative message about the error

    require.NoError(t, err, "descriptive string")