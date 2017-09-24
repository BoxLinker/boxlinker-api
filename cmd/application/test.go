package main

import (
	"k8s.io/apimachinery/pkg/api/resource"
	"fmt"
)

func main(){
	memoryQuantity, err := resource.ParseQuantity("+1000")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%v", memoryQuantity.Value())
}
