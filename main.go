package main

func main() {
	server := newSever(":1935")
	server.run()
}
