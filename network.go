package main

type Config struct {
	internetIPv4Gateway *Route
	internetIPv6Gateway *Route
	utunIPv4            string
	remoteAddress       string
}
