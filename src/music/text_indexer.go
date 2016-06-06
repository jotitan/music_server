package music
import (
    "sort"
    "strings"
    "math"
    "os"
    "encoding/gob"
    "path/filepath"
)

/* used to index musics and give tools to search by free text (sorted list and dicotomic search) */

const (
    tiFilename = "text_index.idx"
)

type Token struct{
    Value string
    // id of musics containing this token
    Musics []int
}

type Tokens []Token

func (t Tokens)Len() int{return len(t)}
func (t Tokens)Less(i, j int) bool {return t[i].Value < t[j].Value}
// Swap swaps the elements with indexes i and j.
func (t Tokens)Swap(i, j int){t[i],t[j] = t[j],t[i]}


type TextIndexer struct {
    // List must be sorted
    Index Tokens
    // temp map before creating list
    temp map[string][]int
}

func NewTextIndexer()TextIndexer {
    return TextIndexer{nil,make(map[string][]int)}
}

func LoadTextIndexer(folder string)TextIndexer {
    ti := TextIndexer{nil,make(map[string][]int)}
    if file,e := os.Open(filepath.Join(folder,tiFilename)) ; e == nil {
        dec := gob.NewDecoder(file)
        dec.Decode(&ti)
        file.Close()
    }
    return ti
}

func (ti TextIndexer)Filter(values ...string)[]string{
    results := make([]string,0)
    for _,value := range values {
        value = strings.ToLower(value)
        for _, e := range strings.Split(value, " ") {
            if len(e)>2 {
                //t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
                //e, _, _ := transform.String(t, "e")
                results = append(results, e)
            }
        }
    }
    return results
}

func (ti * TextIndexer)Add(value string,idMusic int){
    if list,ok := ti.temp[value] ; !ok{
        ti.temp[value] = []int{idMusic}
    }else{
        ti.temp[value] = append(list,idMusic)
    }
}

// Parse map and create list
func (ti * TextIndexer)Build(){
    ti.Index = make([]Token,0,len(ti.temp))
    for value,ids := range ti.temp {
        sort.Ints(ids)
        ti.Index = append(ti.Index,Token{value,ids})
    }
    sort.Sort(ti.Index)
}

func Intersect(a,b []int)[]int{
    if len(a) == 0 || len(b) == 0 {
        return []int{}
    }
    results := make([]int,0,int(math.Min(float64(len(a)),float64(len(b)))))
    j := 0
    for i := 0 ; i < len(a) ; i++ {
        for {
            stop := false
            switch {
                case j >= len(b) || a[i] < b[j] : stop = true
                case a[i] > b[j] : j++
                case a[i] == b[j] :
                results = append(results,a[i])
                j++
                stop = true
            }
            if stop {
                break
            }
        }
        if j >= len(b) {
            break
        }
    }
    return results
}

func (ti TextIndexer)Search(text string)[]int{
    // Convert text
    text = strings.ToLower(text)
    results := make([]int,0)
    for i,sub := range strings.Split(text," ") {
        //r := ti.subSearch(ti.Index, sub)
        r := ti.IntSearch(sub)
        if len(r) == 0 {
            return []int{}
        }
        if i == 0 {
            results = r
        }else{
            if results = Intersect(results,r) ; len(results) == 0{
                return []int{}
            }
        }
    }
    return results
}

func (ti TextIndexer)IntSearch(text string)[]int{
    if pos := ti.subSearchInt(ti.Index,text,0) ; pos != -1 {
        // search other close value
        results := ti.Index[pos].Musics
        for i := pos-1 ; i >= 0 && strings.HasPrefix(ti.Index[i].Value, text) ; i-- {
            results = append(results,ti.Index[i].Musics...)
        }
        for i := pos+1 ; i < len(ti.Index) && strings.HasPrefix(ti.Index[i].Value, text) ; i++ {
            results = append(results,ti.Index[i].Musics...)
        }
        return results
    }
    return []int{}
}

func (ti TextIndexer)subSearchInt(tokens Tokens,text string, pos int)int{
    if len(tokens) == 0 {
        return -1
    }
    center := len(tokens)/2
    t := tokens[center]
    if strings.HasPrefix(t.Value, text) {
        return pos + center
    }
    if len(tokens) == 1{
        return -1
    }
    if t.Value < text {
        return ti.subSearchInt(tokens[center:],text,center+pos)
    }
    return ti.subSearchInt(tokens[:center],text,pos)
}

func (ti TextIndexer)subSearch(tokens Tokens,text string)[]int{
    if len(tokens) == 0 {
        return []int{}
    }
    center := len(tokens)/2
    t := tokens[center]
    if strings.HasPrefix(t.Value, text) {
        return t.Musics
    }
    if len(tokens) == 1{
        return []int{}
    }
    if t.Value < text {
        return ti.subSearch(tokens[center:],text)
    }
    return ti.subSearch(tokens[:center],text)
}

// Save with naive method, gob encoder
func (ti TextIndexer)Save(folder string){
    file,_ := os.OpenFile(filepath.Join(folder,tiFilename),os.O_CREATE|os.O_TRUNC|os.O_RDWR,os.ModePerm)
    defer file.Close()
    enc := gob.NewEncoder(file)
    enc.Encode(ti)

}