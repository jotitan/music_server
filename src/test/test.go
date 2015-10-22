package main
import (
    "fmt"
    "os"
    "bufio"
	"encoding/hex"
    "encoding/binary"
	"io/ioutil"
	"net/http"
	"sync"
	"music"
	"arguments"
)


func formatType(typeMarker int)string{
	switch typeMarker {
		case 1 : return "byte"
		case 2 : return "ascii"
		case 3 : return "short"
		case 4 : return "long"
		case 5 : return "rational"
		case 6 : return "signed byte"
		case 7 : return "undefined"
		case 8 : return "signed short"
		case 9 : return "signed long"
		case 10 : return "signed rational"
		case 11 : return "float"
		case 12 : return "double float"
		default : return "error"
	}
}

type Marker struct{
	name string
	kind int
	length int
	data interface{}

}

func (m Marker)ToString()string{
	return fmt.Sprintf("%s (%s) : %v (%d)",m.name,formatType(m.kind),m.data,m.length)
}

type IntReader interface{
	ReadInt32(tab []byte)int
	ReadInt16(tab []byte)int
	Type()string
}

type ReaderLittleIndian struct{}

func (r ReaderLittleIndian)ReadInt32(tab[]byte)int{
	return int(binary.LittleEndian.Uint32(tab))
}

func (r ReaderLittleIndian)ReadInt16(tab[]byte)int{
	return int(binary.LittleEndian.Uint16(tab))
}

func (r ReaderLittleIndian)Type()string{
	return "little"
}

type ReaderBigIndian struct{}

func (r ReaderBigIndian)ReadInt32(tab[]byte)int{
	return int(binary.BigEndian.Uint32(tab))
}

func (r ReaderBigIndian)ReadInt16(tab[]byte)int{
	return int(binary.BigEndian.Uint16(tab))
}

func (r ReaderBigIndian)Type()string{
	return "big"
}

var intReader IntReader

// Bloc of size 12
func getMarker(bloc,data []byte)Marker{
	marker := hex.EncodeToString(bloc[0:2])
	typeMarker := intReader.ReadInt16(bloc[2:4])
	length := intReader.ReadInt32(bloc[4:8])
	var formatData interface{}
	switch typeMarker {
	case 2 :
		if length > 4 {
			// Read data at offset
			offset := intReader.ReadInt32(bloc[8:])
			formatData = string(data[offset:offset+length])
		}else{
			formatData = string(bloc)
		}
	case 3 : formatData = intReader.ReadInt16(bloc[8:])
	case 4 : formatData = intReader.ReadInt32(bloc[8:])
		default:formatData = bloc[8:]
	}

	return Marker{marker,typeMarker,length,formatData}
}

func main(){
	//run()


	args := arguments.ParseArgs()


	//dico := music.LoadDictionnary(args["workingFolder"])
	music.IndexArtists(args["workingFolder"])
	//dico.Browse(args["browse"])



}

func create(){
	name := "C:\\Users\\960963\\Pictures\\RESIZER\\brut.jpg"
	data,_ := ioutil.ReadFile(name)
	for i := 0 ; i < 150 ; i++ {
		fileout,_ := os.Create(fmt.Sprintf("C:\\Users\\960963\\Pictures\\RESIZER\\brut_%d.jpg",i))
		fileout.Write(data)
		fileout.Close()
	}
}

func run(){
	nb := 50
	waiter := sync.WaitGroup{}
	waiter.Add(nb)
	for i := 0 ; i < nb ; i++ {
		go func(id int){
			url := fmt.Sprintf("http://127.0.0.1:9011/add?type=RESIZE_TASK&from=brut_%d.jpg&to=resize/resize_%d.jpg&width=200&height:150",id,id)
			http.Get(url)
			waiter.Done()
		}(i)
	}
	waiter.Wait()
}

func testReadExif() {

	path := "src/resources/img_BI.jpg"
	treat(path)
}

func treat(path string){
	f,_ := os.Open(path)

    defer f.Close()
    buf := bufio.NewReader(f)
    line,_,_ := buf.ReadLine()

    fmt.Println(len(line),line)
    fmt.Println(string(line),"\n\n")
    // Read Start marker
    pos := 0
    for {
        start := hex.EncodeToString(line[pos:pos+2])
        switch start {
            case "ffd8" : fmt.Println("Start")
            pos+=2
            case "ffe0" : fmt.Println("App marker")
                size := int(binary.BigEndian.Uint16(line[pos+2:pos+4]))
                data := line[pos+4:pos+2+size]
                fmt.Println(data,string(data))
                pos+=2+size
            case "ffe1" : fmt.Println("Exif")
                dataShifted := line[pos+10:]
				size := int(binary.BigEndian.Uint16(line[pos+2:pos+4]))
				exifHeader := string(line[pos+4:pos+10])
                switch string(line[pos+10:pos+12]) {
				case "II" : intReader = ReaderLittleIndian{}
				default : intReader = ReaderBigIndian{}
				}

				rest := string(line[pos+12:pos+18])
                fmt.Println("INFO",size,exifHeader,rest,intReader.Type())
                nbMarker := intReader.ReadInt16(line[pos+18:pos+20])
                pos+=20
                for i :=0 ; i < nbMarker ; i++{
                    marker := getMarker(line[pos:pos+12],dataShifted)
					fmt.Println(marker.ToString())
					pos+=12
                }
                return
            default:return
        }

    }
    return

    // Application marker

    // Read ifd0
    return
    // Read exif header
    // Read tiff header

    // Read nb marker

    // Read markers
    for  {
        if line[pos]!=255 {
            break
        }
		marker := hex.EncodeToString(line[pos:pos+2])
		switch marker {
			case "ffe0" : fmt.Println("INF : App marker")
			case "ffe1" : fmt.Println("INF : Tiff header")
		}
        length := int(line[pos+2])  + int(line[pos+3])
        fmt.Println("PRE",length,pos)
        data := line[pos+4:pos+2+length]
        pos+=2+length
        fmt.Println("MARKER ",pos,string(data))
    }
	line,_,_ = buf.ReadLine()
	line,_,_ = buf.ReadLine()
	line,_,_ = buf.ReadLine()
	fmt.Println("==>",string(line),"\n\n")
}

