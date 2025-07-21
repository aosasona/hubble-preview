package main

func main() {
	app := NewApp()

	if err := app.Run(); err != nil {
		panic(err)
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}
