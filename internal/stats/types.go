//takes raw calendar and turns it into buckets/streak data that the render would actually want ki types

package stats


import "time"


//summary contains all processed contribution data that the renderer needs
type Summary struct {
	Login    string 
	Total    int 
	CurrentStreak   int 
	LongestStreak   int 
	FetchedAt     time.Time 
	
	Grid [][]Cell    //grid is indexed as [week][weekday]
}


type Cell struct {
	Date   time.Time 
	Count    int 
	Bucket   int 

	Empty bool  //for empty boxes 
}




