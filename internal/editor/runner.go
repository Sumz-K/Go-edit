package editor

import (
	"bufio"
	"log"
	"os"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

// To adjust the size of the terminal window
var ROWS,COLS int 

// These variables help in displaying parts of the entire text buffer
// Will help in scrolling and will help in displaying only required parts in the window
// Defaulted to 0
var relativeX,relativeY int

// This is the internal representation of the text in the editor
var textBuffer [][]rune 

// targetFile is a command line argument. This file if exists populates the text buffer with its contents
var targetFile string

// To log to a file
func logger() {
	file,err:=os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err!=nil {		
		log.Fatal("Could not open log file\n")
	}
	log.SetOutput(file)
}

func readFile(target string) {
	file,err:=os.Open(target)
	if err!=nil { //if the file does not exist
		targetFile=target
		textBuffer = append(textBuffer, []rune{})
		return 
	}

	defer file.Close()

	scanner:=bufio.NewScanner(file)

	lineNum:=0
	for scanner.Scan() {
		line:=scanner.Text()
		textBuffer = append(textBuffer, []rune{})
		for i:=0;i<len(line);i++ {
			textBuffer[lineNum]=append(textBuffer[lineNum],rune(line[i]))
		}
		lineNum+=1
	}

	if lineNum==0 {
		textBuffer=append(textBuffer, []rune{})
	}
}


// This function displays the text buffer 
func display() {
	var row,col int
	for row=0;row<ROWS;row++ {
		bufRow:=row+relativeY
		for col=0;col<COLS;col++ {
			bufCol:=col+relativeX

			if bufRow>=0 && bufRow<len(textBuffer) && bufCol>=0 && bufCol<len(textBuffer[bufRow]) {
				if textBuffer[bufRow][bufCol]!='\t' {
					termbox.SetCell(col,row,textBuffer[bufRow][bufCol],termbox.ColorLightCyan,termbox.ColorDefault)
				} else {
					log.Println("Tab detected")
					termbox.SetCell(col,row,rune(' '),termbox.ColorGreen,termbox.ColorDefault)
				}
			} else if bufRow>=len(textBuffer) {
				termbox.SetCell(0,row,rune('*'),termbox.ColorLightCyan,termbox.ColorDefault)
			}

		}
		
	}
}

func dummyPrint(col,row int, fg,bg termbox.Attribute,message string) {
	for _,ch:=range message {
		termbox.SetCell(col,row,ch,fg,bg)
		col+=runewidth.RuneWidth(ch)
	}
}

func RunEditor() {
	logger()

// To build the text editor window
	err:=termbox.Init()

	if err!=nil {
		log.Fatal("Could not initialise termbox\n")
	}

	if len(os.Args) < 2 {
		targetFile="default.txt"
		textBuffer = append(textBuffer, []rune{})
	} else {
		targetFile=os.Args[1]
		readFile(targetFile)
	}

// The textbox runs till Escape key is pressed
	for{

		ROWS,COLS=termbox.Size()
		//dummyPrint(0,0,termbox.ColorGreen,termbox.ColorDefault,"Sumukh")
		display()
		termbox.Flush()
		event:=termbox.PollEvent()

		if event.Type==termbox.EventKey && event.Key==termbox.KeyEsc {
			termbox.Close()
			break
		}
	}
}