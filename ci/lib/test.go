package lib

import "dagger.io/dagger"

// Test runs go test.
func Test(client *dagger.Client, workdir string, race bool) *dagger.Container {
	src := client.Host().Directory(workdir)

	args := []string{"go", "test", "-v"}

	cgoEnabled := "0"
	if race {
		args = append(args, "-race")
		cgoEnabled = "1"
	}

	args = append(args, "./...")

	// TODO: customize container version
	return client.Container().From("golang:latest").
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithEnvVariable("CGO_ENABLED", cgoEnabled).
		WithExec(args)
}
