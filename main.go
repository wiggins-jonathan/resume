// A simple program that converts markdown files to html.
package main

import (
    "regexp"
    "io/ioutil"
    "bytes"
    "bufio"
    "io"
    "os"
    "fmt"
    "path/filepath"
    "html/template"
    "net/http"
)

type Webpage struct {
    Title   string
    Body    template.HTML
}

// Globals for regexp. Compiling them outside of any loops is more efficient
var (
    r = regexp.MustCompile
    boldReg             = r(`\*\*(.*?)\*\*`)
    italicReg           = r(`_(.*?)_`)
    boldItalicReg       = r(`\*\*_(.*?)_\*\*`)
    strikeReg           = r(`\~\~(.*?)\~\~`)
    underscoreReg       = r(`__(.*?)__`)
    anchorReg           = r(`\[(.*?)\]\((.*?)\)`)
    escapeReg           = r(`^\>(\s)`)
    blockquoteReg       = r(`\&gt\;(.*?)$`)
    backtipReg          = r("`(.*?)`")
    horizontalRuleReg   = r(`---|___|\*\*\*`)
    ampersandReg        = r(`&`)

    h1Reg = r(`^#\s(.*?)$`)
    h2Reg = r(`^##\s(.*?)$`)
    h3Reg = r(`^###\s(.*?)$`)
    h4Reg = r(`^####\s(.*?)$`)
    h5Reg = r(`^#####\s(.*?)$`)
    h6Reg = r(`^######\s(.*?)$`)

    unorderedListReg = r(`^[-|\*|\+]\s(.*?)$`)
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Please specify a .md file to parse.")
        os.Exit(1)
    }

    input, err := ioutil.ReadFile(os.Args[1])
    checkError("Error reading .md file.", err)

    reader  := bytes.NewReader(input)
    body    := md2html(reader)

    // Slice .md extension from file & set as title
    filename    := os.Args[1]
    ext         := filepath.Ext(filename)
    title       := filename[0:len(filename) - len(ext)]

    t := template.Must(template.ParseFiles("index.html"))
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // Render templates
        data := Webpage{ Title: title, Body: template.HTML(body) }
        err := t.Execute(w, data)
        checkError("Error parsing template", err)
    })
    http.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
        // Serve style.css
        css, err := ioutil.ReadFile("style.css")
        checkError("Cannot read style.css", err)
        w.Header().Set("Content-Type", "text/css")
        w.Write(css)
    })
    fmt.Println("Server listening on port 8080")
    http.ListenAndServe(":8080", nil)
}

// Transpile markdown to html
func md2html(input io.Reader) string {

    buf := bytes.NewBuffer(nil)

    scanner := bufio.NewScanner(input)
    for scanner.Scan() {

        line := bytes.TrimSpace(scanner.Bytes())
        if len(line) == 0 { // Skip blank lines
            buf.WriteByte('\n')
            continue
        }

        // wrap bold and italic text in "<b>" and "<i>" elements
        line = boldItalicReg.ReplaceAll(line, []byte(`<b><i>$1</i></b>`))
        line = boldReg.ReplaceAll(line, []byte(`<b>$1</b>`))
        line = italicReg.ReplaceAll(line, []byte(`<i>$1</i>`))
        // wrap strikethrough text in "<s>" tags
        line = strikeReg.ReplaceAll(line, []byte(`<s>$1</s>`))
        // wrap underscored text in "<u>" tags
        line = underscoreReg.ReplaceAll(line, []byte(`<u>$1</u>`))
        // convert links to anchor tags
        line = anchorReg.ReplaceAll(line, []byte(`<a href="$2">$1</a>`))
        // escape and wrap blockquotes in "<blockquote>" tags
        line = escapeReg.ReplaceAll(line, []byte(`&gt;`))
        line = blockquoteReg.ReplaceAll(line, []byte(`<blockquote>$1</blockquote>`))
        // wrap the content of backticks inside of "<code>" tags
        line = backtipReg.ReplaceAll(line, []byte(`<code>$1</code>`))
        // Convert horizontal rule
        line = horizontalRuleReg.ReplaceAll(line, []byte(`<hr>`))
        // Convert ampersands
        line = ampersandReg.ReplaceAll(line, []byte(`&amp`))
        // convert headings
        if line[0] == '#' {
            count := bytes.Count(line, []byte(`#`))
            switch count {
            case 1: line = h1Reg.ReplaceAll(line, []byte(`<h1>$1</h1>`))
            case 2: line = h2Reg.ReplaceAll(line, []byte(`<h2>$1</h2>`))
            case 3: line = h3Reg.ReplaceAll(line, []byte(`<h3>$1</h3>`))
            case 4: line = h4Reg.ReplaceAll(line, []byte(`<h4>$1</h4>`))
            case 5: line = h5Reg.ReplaceAll(line, []byte(`<h5>$1</h5>`))
            case 6: line = h6Reg.ReplaceAll(line, []byte(`<h6>$1</h6>`))
            }
        }
        // Convert unordered lists
        if line[0] == '*' || line[0] == '-' || line[0] == '+' {
            line = unorderedListReg.ReplaceAll(line, []byte(`<ul><li>$1</li></ul>`))
        }
        buf.Write(line)
        buf.WriteByte('\n')
    }
    return buf.String()
}

func checkError(message interface{}, err error) {
    if err != nil {
        fmt.Println(message, err)
        os.Exit(1)
    }
}
