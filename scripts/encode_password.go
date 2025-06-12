package main

import (
	"fmt"
	"net/url"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("🔐 MongoDB Password URL Encoder")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Printf("  go run %s \"your_password_here\"\n", os.Args[0])
		fmt.Println("")
		fmt.Println("This will encode special characters in your password for MongoDB URI.")
		fmt.Println("")
		fmt.Println("Common characters that need encoding:")
		fmt.Println("  @  ->  %%40")
		fmt.Println("  :  ->  %%3A")
		fmt.Println("  /  ->  %%2F")
		fmt.Println("  ?  ->  %%3F")
		fmt.Println("  #  ->  %%23")
		fmt.Println("  %%  ->  %%25")
		fmt.Println("  +  ->  %%2B")
		fmt.Println("  =  ->  %%3D")
		fmt.Println("  &  ->  %%26")
		fmt.Println("")
		fmt.Println("Example:")
		fmt.Printf("  go run %s \"mypass@123\"\n", os.Args[0])
		fmt.Println("  Output: mypass%%40123")
		os.Exit(1)
	}

	password := os.Args[1]
	encodedPassword := url.QueryEscape(password)

	fmt.Println("🔐 MongoDB Password Encoder")
	fmt.Println("=" + fmt.Sprintf("%*s", 50, "="))
	fmt.Printf("Original password:  %s\n", password)
	fmt.Printf("Encoded password:   %s\n", encodedPassword)
	fmt.Println("=" + fmt.Sprintf("%*s", 50, "="))
	fmt.Println("")
	fmt.Println("✅ Use the encoded password in your MongoDB URI:")
	fmt.Printf("%s\n", encodedPassword)
	fmt.Println("")
	fmt.Println("💡 Copy the encoded password to your .env file")
}
