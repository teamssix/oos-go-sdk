package sample

import "fmt"

func GetRegionSample() {
	// New client
	client := NewClient()
	ret, err := client.GetRegions()
	if err != nil {
		HandleError(err)
	}
	fmt.Println(ret)
}
