package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	arg := os.Args

	//if input is empty print an error message

	if len(arg) < 2 {
		fmt.Println("Please provide an argument.")
		os.Exit(1)
	}

	//save the user input in fullinput 
	fullInput := strings.Join(arg[1:], " ")
	fmt.Println("Full input:", fullInput)

	//search for any \n to genrate new lines
	wordsOrLines := strings.Split(fullInput, "\\n")

	// open result.txt to svae the result in 
	file, err := os.Create("result.txt")
	//if any error happend then print the error message and exit
	if err != nil {
		fmt.Println("Error creating file:", err)
		os.Exit(1)
	}
	defer file.Close()

	// Create a writer to write to the file in buffered mode
	writer := bufio.NewWriter(file)

	//process each word and write the result in the file
	for _, part := range wordsOrLines {
		err := processLine(part, writer)
		if err != nil {
			fmt.Println("Error processing line:", err)
			os.Exit(1)
		}
		writer.WriteString("\n")
	}

	writer.Flush()
}

//function to process the string
func processLine(line string, writer *bufio.Writer) error {

	//store the ascii art in 2d array
	var artLines [][]string

	//count the max hight for pandding the letters that are shorter
	var maxHeight int

	// go for each char in the string
	for _, char := range line {
		var art string
		var err error

		// if you see a space then create 2d space
		if string(char) == " " {
			var builder strings.Builder
	        for i := 0; i < 7; i++ {
		    builder.WriteString(strings.Repeat(" ", 3) + "\n")}
			art= builder.String()

		} else {
		
		// if it is a char then look for it in the file that holds the ascii art using read file function 
			art, err = readfile("letters.txt", string(char))
			if err != nil {
				return err
			}
		}
		
		lines := strings.Split(strings.TrimRight(art, "\n"), "\n")

		// Track the maximum height of any character
		if len(lines) > maxHeight {
			maxHeight = len(lines)
		}

		// Append the art lines to the collection
		artLines = append(artLines, lines)
	}

	// Pad lowercase letters and characters with smaller heights to match the hight of others
	for i, lines := range artLines {
		if len(lines) < maxHeight {
			padding := maxHeight - len(lines)

			if shouldPadFromBottom(rune(line[i])) {
				artLines[i] = padFromBottom(lines, padding) 
			} else {
				artLines[i] = padFromTop(lines, padding) 
			}
		}
	}

	// Print ascii art in the file 
	for i := 0; i < maxHeight; i++ {
		for _, lines := range artLines {
			writer.WriteString(lines[i])
		}
		writer.WriteString("\n")
	}
	return nil
}

//---------------------------------------------------------------------------------------------------------------------------------------------------

//check if the letter you have is a lower case or need padding from the bottom

//if the letter is a lower case or a number then it needs padding from up to match the hight of the rest of letters
func isLowercase(char rune) bool {
	if char >= 'a' && char <= 'z' && char !='y' || char >= '0' && char <='9' {
		return true
	}
	return false
}

//see which char should be padded from the bottom 
func shouldPadFromBottom(char rune) bool {
	return char == '"'|| char == '\''  || char==':' || char=='$'||char == '[' || char == ']'|| char == '~' 
}

//---------------------------------------------------------------------------------------------------------------------------------------------------
 
//padding from the top to match hight 

func padFromTop(lines []string, padding int) []string {
	paddedLines := make([]string, padding)
	for i := 0; i < padding; i++ {
		paddedLines[i] = strings.Repeat(" ", len(lines[0])) 
	}
	return append(paddedLines, lines...)
}

func padFromBottom(lines []string, padding int) []string {
	for i := 0; i < padding; i++ {
		lines = append(lines, strings.Repeat(" ", len(lines[0]))) 
	}
	return lines
}

//---------------------------------------------------------------------------------------------------------------------------------------------------

// Function for reading the art from the file
func readfile(filename, char string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// scanner to read the file
	scanner := bufio.NewScanner(file)

	//string bulider to append the ascii art
	var artBuilder strings.Builder
	//a bool to see if the char we are looking for is in the file or no?
	var foundChar bool

	// start reading the file line by line
	for scanner.Scan() {
		line := scanner.Text()

		// if you find the char and a : after it then cont
		if strings.HasPrefix(line, char+":") {
			foundChar = true
			continue
		}

		
		//if you found the char the included it in the builder until you reach to a null or another : 
		if foundChar {
			if line == "" || strings.HasSuffix(line, ":") {
				break
			}
		//otherwise store the whole line inside the builder
			artBuilder.WriteString(line + "\n")
		}

	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	//if you didnt find anything then return an error
	if artBuilder.Len() == 0 {
		return "", fmt.Errorf("character %s not found", char)
	}

	//return the whole builder 
	return artBuilder.String(), nil
}
