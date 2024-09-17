package editor

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

// Similar to vi/vim editors, we will have 2 modes. View and edit
// mode bit 0 -> View
// mode bit 1 -> Insert/Edit
// make it a bool for efficiency? (Prolly not)

// To toggle to insert mode, press "i"
var mode int

// To adjust the size of the terminal window
var ROWS, COLS int

// To track the cursor
var currRow, currCol int

// These variables help in displaying parts of the entire text buffer
// Will help in scrolling and will help in displaying only required parts in the window
// Defaulted to 0
var relativeX, relativeY int

// This is the internal representation of the text in the editor
var textBuffer [][]rune

// targetFile is a command line argument. This file if exists populates the text buffer with its contents
var targetFile string

// To log to a file
func logger() {
	file, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Could not open log file\n")
	}
	log.SetOutput(file)
}

func readFile(target string) {
	file, err := os.Open(target)
	if err != nil { //if the file does not exist
		targetFile = target
		textBuffer = append(textBuffer, []rune{})
		return
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	lineNum := 0
	for scanner.Scan() {
		line := scanner.Text()
		textBuffer = append(textBuffer, []rune{})
		for i := 0; i < len(line); i++ {
			textBuffer[lineNum] = append(textBuffer[lineNum], rune(line[i]))
		}
		lineNum += 1
	}

	if lineNum == 0 {
		textBuffer = append(textBuffer, []rune{})
	}
}

// Adjust the offsets to determine what will be in screen and what not
func scroll() {
	// If the row goes out of the window at the bottom, extend the bottom
	if currRow >= relativeY+ROWS {
		relativeY = currRow - ROWS + 1
	}

	if currRow < relativeY {
		relativeY = currRow
	}

	if currCol >= relativeX+COLS {
		relativeX = currCol - COLS + 1
	}

	if currCol < relativeX {
		relativeX = currCol
	}
}

// This function displays the text buffer
func display() {
	var row, col int
	for row = 0; row < ROWS; row++ {
		bufRow := row + relativeY
		for col = 0; col < COLS; col++ {
			bufCol := col + relativeX

			if bufRow >= 0 && bufRow < len(textBuffer) && bufCol >= 0 && bufCol < len(textBuffer[bufRow]) {
				if textBuffer[bufRow][bufCol] != '\t' {
					termbox.SetCell(col, row, textBuffer[bufRow][bufCol], termbox.ColorLightCyan, termbox.ColorDefault)
				} else {
					//log.Println("Tab detected")
					termbox.SetCell(col, row, rune(' '), termbox.ColorGreen, termbox.ColorDefault)
				}
			} else if bufRow >= len(textBuffer) {
				termbox.SetCell(0, row, rune('*'), termbox.ColorLightCyan, termbox.ColorDefault)
			}

		}

	}
}

// To display the status bar at the bottom
// Will need: mode,File name,current row and current col
func statusBar() {
	var modeString string
	if mode == 0 {
		modeString = "--VIEW--"
	} else {
		modeString = "--EDIT--"
	}

	fileLen := strconv.Itoa(len(textBuffer))
	usedSpace := len(modeString) + len(targetFile) + len(fileLen)
	spaceLeft := COLS - usedSpace
	if spaceLeft < 0 {
		spaceLeft = 0
	}
	spacePadding := strings.Repeat(" ", spaceLeft)

	locStatus := "Row " + strconv.Itoa(currRow+1) + " Col " + strconv.Itoa(currCol+1)
	lenStatus := "  " + fileLen + " lines"
	statusMessage := modeString + " " + targetFile + lenStatus + spacePadding + locStatus
	singlePrint(0, ROWS, termbox.ColorBlack, termbox.ColorWhite, statusMessage)
}

func insertCharacter(event termbox.Event) {
	tempBuffer:=make([]rune, len(textBuffer[currRow])+1)

	// the character is to be inserted at currCol, so copy everything upto currcol in the textbuff's row into a temp buffer
	copy(tempBuffer[:currCol],textBuffer[currRow][:currCol])

	// now need to actually insert the character at the location

	if event.Key==termbox.KeySpace {
		tempBuffer[currCol]=rune(' ')
	} else if event.Key==termbox.KeyTab {
		tempBuffer[currCol]=rune('	')
	} else {
		tempBuffer[currCol]=rune(event.Ch)
	}


	// copy the rest of the data in the row
	
	copy(tempBuffer[currCol+1:],textBuffer[currRow][currCol:])

	// write back to the text buffer
	textBuffer[currRow]=tempBuffer
	currCol+=1
}

func deleteCharacter() {
	currCol-=1
	log.Println("Delete prompted")
	tempBuffer:=make([]rune,len(textBuffer[currRow])-1)

	copy(tempBuffer[:currCol],textBuffer[currRow][:currCol])

	//copy the rest skipping currCol

	copy(tempBuffer[currCol:],textBuffer[currRow][currCol+1:])

	textBuffer[currRow]=tempBuffer
}

func singlePrint(col, row int, fg, bg termbox.Attribute, message string) {
	for _, ch := range message {
		termbox.SetCell(col, row, ch, fg, bg)
		col += runewidth.RuneWidth(ch)
	}
}

// These variables help with preserving the column state on encountering a newline during traversal. Init in the RunEditor() func
var tempRowUp int
var tempRowDown int

// This handles navigation of the cursor. Still needs some work
func handleInput() {
	event := termbox.PollEvent()

	if event.Key == termbox.KeyEsc {
		mode = 0
	} else if event.Ch != 0 {
		if mode == 1 {
			log.Println("Trying to insert")
			insertCharacter(event)	
		} else {
			switch event.Ch {
			case 'q':
				termbox.Close()
				os.Exit(0)
			case 'i':
				mode = 1
			}
		}
	} else {
		switch event.Type {
		case termbox.EventKey:
			switch event.Key {
			case termbox.KeyEsc:
				mode = 0

			case termbox.KeyBackspace2:
				if mode==1 {
					deleteCharacter()
				} else {
					currCol-=1
				}
			case termbox.KeySpace:
				if mode==1 {
					insertCharacter(event)
				} else {
					currCol+=1
				}

			case termbox.KeyTab:
				if mode==1 {
					for i:=0;i<4;i++ {
						insertCharacter(event)
					}
				} else {
					log.Println("Tab does nothing in view mode")
				}
			case termbox.KeyArrowDown:
				if currRow < len(textBuffer)-1 {
					currRow += 1
					if tempRowDown != -1 {
						anotherMax := len(textBuffer[currRow])
						if anotherMax < tempRowDown {
							currCol = anotherMax
						} else {
							currCol = tempRowDown
						}
						tempRowDown = -1
					} else {
						if len(textBuffer[currRow]) == 0 {
							tempRowDown = currCol
						}
						maxCol := len(textBuffer[currRow])
						if currCol > maxCol {
							currCol = maxCol
						}
					}
				}
			case termbox.KeyArrowUp:
				if currRow > 0 {

					currRow -= 1
					if tempRowUp != -1 {
						anotherMax := len(textBuffer[currRow])
						if anotherMax < tempRowUp {
							currCol = anotherMax
						} else {
							currCol = tempRowUp
						}
						tempRowUp = -1
					} else {

						//log.Println("This is supposed to be a newline ", len(textBuffer[currRow]))
						if len(textBuffer[currRow]) == 0 {
							tempRowUp = currCol
						}
						maxCol := len(textBuffer[currRow])
						if maxCol < currCol {
							currCol = maxCol
						}
					}
				}
			case termbox.KeyArrowLeft:
				if currCol != 0 {
					currCol -= 1
				} else if currRow > 0 {
					currRow -= 1
					currCol = len(textBuffer[currRow])
				}
			case termbox.KeyArrowRight:
				if currCol < len(textBuffer[currRow]) {
					currCol += 1
				} else if currRow < len(textBuffer)-1 {
					currRow += 1
					currCol = 0
				}

			// Advanced navigations to move to start|end of a line or the entire file
			case termbox.KeyEnd: //this is fn + right arrow
				currCol = len(textBuffer[currRow])

			case termbox.KeyHome: // fn + left arrow
				currCol = 0

			case termbox.KeyPgup: //fn + arrowUp
				currRow = 0
				currCol = 0

			case termbox.KeyPgdn: //fn + arrowDown
				currRow = len(textBuffer) - 1
				currCol = 0

			default:
				log.Println("Some other key")

			}
		case termbox.EventError:
			panic(event.Err)
		}
	}
}
func RunEditor() {
	logger()

	// To build the text editor window
	err := termbox.Init()

	if err != nil {
		log.Fatal("Could not initialise termbox\n")
	}

	if len(os.Args) < 2 {
		targetFile = "default.txt"
		textBuffer = append(textBuffer, []rune{})
	} else {
		targetFile = os.Args[1]
		readFile(targetFile)
	}

	tempRowUp = -1
	tempRowDown = -1

	// The textbox runs till Escape key is pressed
	for {

		COLS, ROWS = termbox.Size()
		ROWS -= 1

		COLS = min(COLS, 100)
		//dummyPrint(0,0,termbox.ColorGreen,termbox.ColorDefault,"Sumukh")
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		scroll()
		display()
		statusBar()

		termbox.SetCursor(currCol-relativeX, currRow-relativeY)
		termbox.Flush()
		handleInput()

	}
}
