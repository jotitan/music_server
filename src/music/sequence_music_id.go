package music
import (
    "strconv"
    "sort"
)


/* A sequence give the next available music id. Based on removed music id and new available after previous max */
type Sequence struct{
    availableIds []int
    currentAvailableId int
    nextId int
}

// Create sequence by analyzing removed files and max id
func NewSequence(musicsInfo map[string]map[string]string)Sequence{
    maxId := 0
    ids := make([]int,0,len(musicsInfo))
    for _,info := range musicsInfo {
        if id,exist := info["id"] ; exist {
            if intId,err := strconv.ParseInt(id,10,32) ; err == nil {
                if maxId < int(intId) {
                    maxId = int(intId)
                }
                ids = append(ids, int(intId))
            }
        }
    }
    sequence := Sequence{nextId:maxId+1,currentAvailableId:0}
    // Sort ids and detect holes
    if len(ids) > 0 {
        sort.Ints(ids)
        previousId := ids[0]
        availableIds := make([]int,0)
        for i := 1; i < len(ids); i++ {
            if previousId +1 != ids[i] {
                // Save id from previousId+1 to current id
                for id := previousId+1 ; id < ids[i] ; id++ {
                    availableIds = append(availableIds,id)
                }
            }
        }
        if len(availableIds) > 0 {
            sequence.availableIds = availableIds
        }
    }
    return sequence
}
